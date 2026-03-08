package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"scriberr/internal/config"
	"scriberr/internal/logger"
	"scriberr/internal/model"
	"scriberr/internal/output"
	"scriberr/internal/validator"
	"scriberr/internal/whisper"

	"github.com/urfave/cli/v2"
)

// TranscribeCommand creates the transcribe command with batch processing
func TranscribeCommand(log *logger.Logger, cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:        "transcribe",
		Aliases:     []string{"t"},
		Usage:       "Transcribe audio files to text",
		Description: "Transcribe one or more audio files to text with support for multiple output formats (text, JSON, SRT, VTT)",
		Category:    "Transcription",
		ArgsUsage:   "<audio-file> [audio-files...]",
		Flags: []cli.Flag{
			// Output format flags
			&cli.StringFlag{
				Name:  "format",
				Aliases: []string{"f"},
				Usage: "Output format: text, json, srt, vtt",
				Value: "text",
			},
			&cli.StringFlag{
				Name:  "output",
				Aliases: []string{"o"},
				Usage: "Output file or directory (default: stdout)",
			},
			&cli.BoolFlag{
				Name:  "output-dir",
				Usage: "Treat output as directory (for batch processing)",
			},
			
			// Model flags
			&cli.StringFlag{
				Name:  "model",
				Aliases: []string{"m"},
				Usage: "Model size: tiny, base, small, medium, large, large-v2, large-v3",
				Value: cfg.Model.Size,
			},
			&cli.StringFlag{
				Name:  "language",
				Aliases: []string{"l"},
				Usage: "Language code (auto for auto-detect)",
				Value: "auto",
			},
			
			// Processing flags
			&cli.IntFlag{
				Name:  "threads",
				Usage: "Number of CPU threads to use",
				Value: 4,
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Print verbose output",
			},
			&cli.BoolFlag{
				Name:  "translate",
				Usage: "Translate to English",
			},
			&cli.BoolFlag{
				Name:  "timestamps",
				Aliases: []string{"ts"},
				Usage: "Include timestamps in output",
			},
			
			// Batch processing flags
			&cli.BoolFlag{
				Name:  "batch",
				Aliases: []string{"b"},
				Usage: "Process multiple files in batch mode",
			},
			&cli.StringFlag{
				Name:  "pattern",
				Usage: "Glob pattern for batch processing (e.g., '*.mp3')",
			},
			&cli.BoolFlag{
				Name:  "recursive",
				Aliases: []string{"r"},
				Usage: "Recursively find files in directories",
			},
			
			// Output control
			&cli.BoolFlag{
				Name:  "quiet",
				Aliases: []string{"q"},
				Usage: "Suppress non-error output",
			},
		},
		Action: func(cCtx *cli.Context) error {
			return runTranscribe(log, cfg, cCtx)
		},
	}
}

// TranscribeCommands returns all transcription-related commands
func TranscribeCommands(log *logger.Logger, cfg *config.Config) []*cli.Command {
	return []*cli.Command{
		TranscribeCommand(log, cfg),
	}
}

// runTranscribe handles the actual transcription logic
func runTranscribe(log *logger.Logger, cfg *config.Config, cCtx *cli.Context) error {
	// Parse output format
	formatStr := cCtx.String("format")
	outFormat := output.ParseFormat(formatStr)
	
	// Get model
	modelSize := cCtx.String("model")
	if !model.IsValidModelSize(modelSize) {
		return fmt.Errorf("invalid model size: %s", modelSize)
	}
	
	// Check if model is cached
	if !model.IsModelCached(model.ModelSize(modelSize)) {
		fmt.Printf("Model '%s' not cached. Downloading...\n", modelSize)
		downloader := model.NewDownloader(log)
		path, err := downloader.Download(model.ModelSize(modelSize))
		if err != nil {
			return fmt.Errorf("failed to download model: %w", err)
		}
		fmt.Printf("Model downloaded to: %s\n", path)
	}
	
	// Get language
	language := cCtx.String("language")
	
	// Get output path
	outPath := cCtx.String("output")
	outDir := cCtx.Bool("output-dir")
	
	// Handle batch processing
	pattern := cCtx.String("pattern")
	recursive := cCtx.Bool("recursive")
	
	var inputFiles []string
	
	// If pattern is provided, use glob
	if pattern != "" {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %w", err)
		}
		inputFiles = matches
	} else {
		// Use command line arguments
		inputFiles = cCtx.Args().Slice()
	}
	
	// If recursive, find all audio files in directories
	if recursive {
		var audioDirs []string
		for _, f := range inputFiles {
			info, err := os.Stat(f)
			if err == nil && info.IsDir() {
				audioDirs = append(audioDirs, f)
			}
		}
		
		for _, dir := range audioDirs {
			audioFiles, err := findAudioFilesRecursive(dir)
			if err != nil {
				log.Warn("Failed to scan directory", "dir", dir, "error", err)
				continue
			}
			inputFiles = append(inputFiles, audioFiles...)
		}
	}
	
	// Filter to only audio files
	inputFiles = filterAudioFiles(inputFiles)
	
	if len(inputFiles) == 0 {
		return fmt.Errorf("no audio files provided")
	}
	
	// Process files
	quiet := cCtx.Bool("quiet")
	verbose := cCtx.Bool("verbose")
	
	// Handle single vs batch output
	if len(inputFiles) == 1 && !outDir && outPath != "" {
		// Single file with output path - process directly
		err := transcribeSingleFile(log, inputFiles[0], outPath, outFormat, modelSize, language, cCtx, quiet, verbose)
		if err != nil {
			return fmt.Errorf("transcription failed: %w", err)
		}
		if !quiet {
			fmt.Printf("✓ Transcribed: %s -> %s\n", inputFiles[0], outPath)
		}
	} else if len(inputFiles) > 1 || outDir {
		// Batch processing
		batchOutputDir := outPath
		if batchOutputDir == "" {
			batchOutputDir = "."
		}
		
		// Ensure output directory exists
		if err := os.MkdirAll(batchOutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		
		successCount := 0
		failCount := 0
		
		for i, inputFile := range inputFiles {
			if !quiet {
				fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(inputFiles), inputFile)
			}
			
			// Generate output filename
			basename := filepath.Base(inputFile)
			ext := output.GetFormatter(outFormat).Extension()
			outputFile := filepath.Join(batchOutputDir, basename+ext)
			
			err := transcribeSingleFile(log, inputFile, outputFile, outFormat, modelSize, language, cCtx, quiet, verbose)
			if err != nil {
				log.Error("Transcription failed", "file", inputFile, "error", err)
				failCount++
				continue
			}
			
			successCount++
			if !quiet {
				fmt.Printf("  ✓ Saved: %s\n", outputFile)
			}
		}
		
		if !quiet {
			fmt.Printf("\nBatch complete: %d succeeded, %d failed\n", successCount, failCount)
		}
	} else {
		// Single file to stdout
		transcription, err := transcribeToTranscription(log, inputFiles[0], outFormat, modelSize, language, cCtx, verbose)
		if err != nil {
			return fmt.Errorf("transcription failed: %w", err)
		}
		
		formatter := output.GetFormatter(outFormat)
		if err := formatter.Format(transcription, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
	
	return nil
}

// transcribeSingleFile transcribes a single audio file to the specified output path
func transcribeSingleFile(
	log *logger.Logger,
	inputPath string,
	outputPath string,
	format output.Format,
	modelSize string,
	language string,
	cCtx *cli.Context,
	quiet bool,
	verbose bool,
) error {
	// Validate input file
	if err := validator.ValidateAudioFile(inputPath); err != nil {
		return fmt.Errorf("invalid audio file: %w", err)
	}
	
	// Load model (simplified - in production would use whisper.cpp)
	ctx, err := loadWhisperModel(modelSize, log)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}
	defer ctx.Free()
	
	// Process audio (simplified - would use ffmpeg to convert to raw audio)
	audioData, err := loadAudioData(inputPath, log)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}
	
	// Transcribe
	params := loadTranscribeParams(cCtx, language)
	result, err := ctx.TranscribeAudio(audioData, params)
	if err != nil {
		return fmt.Errorf("transcription error: %w", err)
	}
	
	// Get timing information
	segmentCount := ctx.GetSegmentCount()
	timings := make([]struct{ Start, End int }, segmentCount)
	for i := 0; i < segmentCount; i++ {
		start, end := ctx.GetSegmentTiming(i)
		timings[i] = struct{ Start, End int }{start, end}
	}
	
	// Build transcription output
	transcription := output.BuildSegments(result, timings, modelSize, language)
	
	// Write to file
	if err := output.WriteToFile(transcription, format, outputPath); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	
	return nil
}

// transcribeToTranscription transcribes a file and returns the transcription object
func transcribeToTranscription(
	log *logger.Logger,
	inputPath string,
	format output.Format,
	modelSize string,
	language string,
	cCtx *cli.Context,
	verbose bool,
) (*output.Transcription, error) {
	// Validate input file
	if err := validator.ValidateAudioFile(inputPath); err != nil {
		return nil, fmt.Errorf("invalid audio file: %w", err)
	}
	
	// Load model
	ctx, err := loadWhisperModel(modelSize, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}
	defer ctx.Free()
	
	// Process audio
	audioData, err := loadAudioData(inputPath, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio: %w", err)
	}
	
	// Transcribe
	params := loadTranscribeParams(cCtx, language)
	result, err := ctx.TranscribeAudio(audioData, params)
	if err != nil {
		return nil, fmt.Errorf("transcription error: %w", err)
	}
	
	// Get timing information
	segmentCount := ctx.GetSegmentCount()
	timings := make([]struct{ Start, End int }, segmentCount)
	for i := 0; i < segmentCount; i++ {
		start, end := ctx.GetSegmentTiming(i)
		timings[i] = struct{ Start, End int }{start, end}
	}
	
	// Build transcription output
	return output.BuildSegments(result, timings, modelSize, language), nil
}

// loadWhisperModel loads the Whisper model (simplified wrapper)
func loadWhisperModel(size string, log *logger.Logger) (*whisper.Context, error) {
	modelPath, err := model.ModelPath(model.ModelSize(size))
	if err != nil {
		return nil, err
	}
	return whisper.InitFromFile(modelPath)
}

// loadAudioData loads audio data from file (simplified - would use ffmpeg)
func loadAudioData(path string, log *logger.Logger) ([]float32, error) {
	// This is a placeholder - in production would use ffmpeg to convert to 16kHz mono float
	// For now, return empty to trigger the error handling in whisper
	return nil, fmt.Errorf("audio conversion not implemented - use ffmpeg to convert to 16kHz mono WAV")
}

// loadTranscribeParams loads transcription parameters from CLI context
func loadTranscribeParams(cCtx *cli.Context, language string) whisper.Params {
	params := whisper.DefaultParams()
	params.Language = language
	params.NThreads = cCtx.Int("threads")
	params.PrintProgress = cCtx.Bool("verbose")
	params.PrintTimestamps = cCtx.Bool("timestamps")
	params.Translate = cCtx.Bool("translate")
	return params
}

// findAudioFilesRecursive finds all audio files in a directory recursively
func findAudioFilesRecursive(dir string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		if validator.IsSupportedAudioFormat(ext) {
			files = append(files, path)
		}
		
		return nil
	})
	
	return files, err
}

// filterAudioFiles filters a list to only include supported audio files
func filterAudioFiles(files []string) []string {
	var result []string
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if validator.IsSupportedAudioFormat(ext) {
			result = append(result, f)
		}
	}
	return result
}
