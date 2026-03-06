package audio

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SupportedFormats defines the audio formats supported by scriberr
var SupportedFormats = []string{
	".mp3",
	".wav",
	".m4a",
	".ogg",
	".flac",
	".aac",
	".wma",
	".opus",
}

// ProgressTracker interface for tracking processing progress
type ProgressTracker interface {
	Start(total int, message string)
	Update(current int)
	Complete()
	Error(err error)
}

// Processor handles audio file processing with FFmpeg
type Processor struct {
	ffmpegPath string
	progress   ProgressTracker
}

// AudioFile represents an audio file with metadata
type AudioFile struct {
	Path            string
	Format          string
	Duration        time.Duration
	SampleRate      int
	Channels        int
	BitRate         int
	Valid           bool
	ValidationError error
}

// NewProcessor creates a new audio processor
func NewProcessor() (*Processor, error) {
	ffmpegPath, err := findFFmpeg()
	if err != nil {
		return nil, fmt.Errorf("FFmpeg not found: %w", err)
	}

	return &Processor{
		ffmpegPath: ffmpegPath,
	}, nil
}

// NewProcessorWithProgress creates a new audio processor with progress tracking
func NewProcessorWithProgress(progress ProgressTracker) (*Processor, error) {
	proc, err := NewProcessor()
	if err != nil {
		return nil, err
	}
	proc.progress = progress
	return proc, nil
}

// findFFmpeg searches for FFmpeg in common locations
func findFFmpeg() (string, error) {
	// Check if ffmpeg is in PATH
	path, err := exec.LookPath("ffmpeg")
	if err == nil {
		return path, nil
	}

	// Check common installation locations
	locations := []string{
		"/usr/bin/ffmpeg",
		"/usr/local/bin/ffmpeg",
		"/opt/homebrew/bin/ffmpeg",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("ffmpeg not found in PATH or common locations")
}

// ValidateFile checks if a file is a valid audio file
func (p *Processor) ValidateFile(path string) (*AudioFile, error) {
	// Check file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &AudioFile{
			Path:            path,
			Valid:           false,
			ValidationError: fmt.Errorf("file does not exist"),
		}, fmt.Errorf("file does not exist: %s", path)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	valid := false
	for _, format := range SupportedFormats {
		if ext == format {
			valid = true
			break
		}
	}

	if !valid {
		return &AudioFile{
			Path:            path,
			Format:          ext,
			Valid:           false,
			ValidationError: fmt.Errorf("unsupported audio format: %s", ext),
		}, fmt.Errorf("unsupported audio format: %s", ext)
	}

	// Get file metadata using FFprobe
	audioFile, err := p.getAudioMetadata(path)
	if err != nil {
		return &AudioFile{
			Path:            path,
			Format:          ext,
			Valid:           false,
			ValidationError: err,
		}, err
	}

	audioFile.Valid = true
	return audioFile, nil
}

// getAudioMetadata extracts metadata from an audio file
func (p *Processor) getAudioMetadata(path string) (*AudioFile, error) {
	audioFile := &AudioFile{
		Path:   path,
		Format: strings.ToLower(filepath.Ext(path)),
	}

	// Use ffprobe to get duration
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		// If ffprobe fails, try to at least get file size
		audioFile.Duration = 0
		return audioFile, nil
	}

	// Parse the output to extract duration
	durationStr := extractJSONValue(string(output), "duration")
	if durationStr != "" {
		var duration float64
		if _, err := fmt.Sscanf(durationStr, "%f", &duration); err == nil {
			audioFile.Duration = time.Duration(duration * float64(time.Second))
		}
	}

	// Extract sample rate
	sampleRateStr := extractJSONValue(string(output), "sample_rate")
	if sampleRateStr != "" {
		var sampleRate int
		if _, err := fmt.Sscanf(sampleRateStr, "%d", &sampleRate); err == nil {
			audioFile.SampleRate = sampleRate
		}
	}

	// Extract channels
	channelsStr := extractJSONValue(string(output), "channels")
	if channelsStr != "" {
		var channels int
		if _, err := fmt.Sscanf(channelsStr, "%d", &channels); err == nil {
			audioFile.Channels = channels
		}
	}

	// Extract bit rate
	bitRateStr := extractJSONValue(string(output), "bit_rate")
	if bitRateStr != "" {
		var bitRate int
		if _, err := fmt.Sscanf(bitRateStr, "%d", &bitRate); err == nil {
			audioFile.BitRate = bitRate / 1000 // Convert to kbps
		}
	}

	return audioFile, nil
}

// extractJSONValue is a simple helper to extract values from ffprobe JSON output
func extractJSONValue(json, key string) string {
	searchStr := fmt.Sprintf("%q:", key)
	idx := strings.Index(json, searchStr)
	if idx == -1 {
		return ""
	}

	// Find the value after the key
	rest := json[idx+len(searchStr):]

	// Skip whitespace and handle different value types
	rest = strings.TrimSpace(rest)

	// Handle string values
	if strings.HasPrefix(rest, `"`) {
		start := strings.Index(rest, `"`) + 1
		end := strings.Index(rest[start:], `"`)
		if end > 0 {
			return rest[start : start+end]
		}
	}

	// Handle number values
	var value string
	for _, c := range rest {
		if (c >= '0' && c <= '9') || c == '.' || c == '-' {
			value += string(c)
		} else if len(value) > 0 {
			break
		}
	}

	return value
}

// Convert converts an audio file to WAV format (required by Whisper)
func (p *Processor) Convert(ctx context.Context, inputPath, outputPath string) error {
	if p.progress != nil {
		p.progress.Start(100, "Converting audio...")
		p.progress.Update(10)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		if p.progress != nil {
			p.progress.Error(err)
		}
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Run FFmpeg conversion
	cmd := exec.CommandContext(ctx, p.ffmpegPath,
		"-i", inputPath,
		"-ar", "16000",        // Whisper requires 16kHz
		"-ac", "1",            // Mono
		"-c:a", "pcm_s16le",  // 16-bit PCM WAV
		"-y",                 // Overwrite output
		outputPath,
	)

	if p.progress != nil {
		p.progress.Update(30)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if p.progress != nil {
			p.progress.Error(err)
		}
		return fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	if p.progress != nil {
		p.progress.Update(100)
		p.progress.Complete()
	}

	return nil
}

// ConvertToTemp converts audio to a temporary WAV file
func (p *Processor) ConvertToTemp(ctx context.Context, inputPath string) (string, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "scriberr-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFile.Close()
	tmpPath := tmpFile.Name()

	// Convert
	if err := p.Convert(ctx, inputPath, tmpPath); err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	return tmpPath, nil
}

// IsFFmpegAvailable checks if FFmpeg is installed
func IsFFmpegAvailable() bool {
	_, err := findFFmpeg()
	return err == nil
}

// GetSupportedFormats returns the list of supported audio formats
func GetSupportedFormats() []string {
	return SupportedFormats
}
