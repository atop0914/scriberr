package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// FFmpeg provides FFmpeg command wrapper functionality
type FFmpeg struct {
	ffmpegPath string
	ffprobePath string
}

// New creates a new FFmpeg instance
func New() (*FFmpeg, error) {
	ffmpegPath, err := findExecutable("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}

	ffprobePath, err := findExecutable("ffprobe")
	if err != nil {
		return nil, fmt.Errorf("ffprobe not found: %w", err)
	}

	return &FFmpeg{
		ffmpegPath: ffmpegPath,
		ffprobePath: ffprobePath,
	}, nil
}

func findExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err == nil {
		return path, nil
	}
	return "", fmt.Errorf("%s not found in PATH", name)
}

// StreamInfo contains information about an audio/video stream
type StreamInfo struct {
	Index     int
	CodecType string
	CodecName string
	SampleRate int
	Channels  int
	Duration float64
	BitRate   int
}

// FileInfo contains overall file metadata
type FileInfo struct {
	Format   string
	Duration float64
	BitRate  int
	Size     int64
	Streams  []StreamInfo
}

// ProbeFile extracts metadata from a media file
func (f *FFmpeg) ProbeFile(path string) (*FileInfo, error) {
	cmd := exec.Command(f.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	return parseProbeOutput(string(output))
}

var durationRegex = regexp.MustCompile(`"duration":\s*"?(\d+\.?\d*)"?`)
var bitRateRegex = regexp.MustCompile(`"bit_rate":\s*"?(\d+\.?\d*)"?`)
var sizeRegex = regexp.MustCompile(`"size":\s*"?(\d+\.?\d*)"?`)
var formatRegex = regexp.MustCompile(`"format_name":\s*"([^"]+)"`)
var codecTypeRegex = regexp.MustCompile(`"codec_type":\s*"([^"]+)"`)
var codecNameRegex = regexp.MustCompile(`"codec_name":\s*"([^"]+)"`)
var sampleRateRegex = regexp.MustCompile(`"sample_rate":\s*"(\d+)"`)
var channelsRegex = regexp.MustCompile(`"channels":\s*(\d+)`)
var streamIndexRegex = regexp.MustCompile(`"index":\s*(\d+)`)

func parseProbeOutput(output string) (*FileInfo, error) {
	info := &FileInfo{
		Streams: []StreamInfo{},
	}

	// Extract format
	if match := formatRegex.FindStringSubmatch(output); match != nil {
		info.Format = match[1]
	}

	// Extract duration
	if match := durationRegex.FindStringSubmatch(output); match != nil {
		if d, err := strconv.ParseFloat(match[1], 64); err == nil {
			info.Duration = d
		}
	}

	// Extract bit rate
	if match := bitRateRegex.FindStringSubmatch(output); match != nil {
		if br, err := strconv.Atoi(match[1]); err == nil {
			info.BitRate = br
		}
	}

	// Extract size
	if match := sizeRegex.FindStringSubmatch(output); match != nil {
		if s, err := strconv.ParseFloat(match[1], 64); err == nil {
			info.Size = int64(s)
		}
	}

	// Extract stream information
	// This is a simplified parser - in production you'd use JSON parsing
	streamBlocks := strings.Split(output, "}")
	for _, block := range streamBlocks {
		if !strings.Contains(block, "codec_type") {
			continue
		}

		stream := StreamInfo{}

		// Find stream index
		if match := streamIndexRegex.FindStringSubmatch(block); match != nil {
			if idx, err := strconv.Atoi(match[1]); err == nil {
				stream.Index = idx
			}
		}

		// Find codec type
		if match := codecTypeRegex.FindStringSubmatch(block); match != nil {
			stream.CodecType = match[1]
		}

		// Find codec name
		if match := codecNameRegex.FindStringSubmatch(block); match != nil {
			stream.CodecName = match[1]
		}

		// Find sample rate (for audio)
		if match := sampleRateRegex.FindStringSubmatch(block); match != nil {
			if sr, err := strconv.Atoi(match[1]); err == nil {
				stream.SampleRate = sr
			}
		}

		// Find channels (for audio)
		if match := channelsRegex.FindStringSubmatch(block); match != nil {
			if ch, err := strconv.Atoi(match[1]); err == nil {
				stream.Channels = ch
			}
		}

		// Only add audio streams
		if stream.CodecType == "audio" {
			info.Streams = append(info.Streams, stream)
		}
	}

	return info, nil
}

// ConvertOptions contains options for audio conversion
type ConvertOptions struct {
	SampleRate   int    // Target sample rate (e.g., 16000)
	Channels    int    // Target number of channels (1 = mono)
	Codec       string // Audio codec (e.g., "pcm_s16le")
	BitRate     int    // Target bit rate in kbps
	Overwrite   bool   // Overwrite output file
}

// DefaultConvertOptions returns the default conversion options for Whisper
func DefaultConvertOptions() *ConvertOptions {
	return &ConvertOptions{
		SampleRate: 16000,
		Channels:   1,
		Codec:      "pcm_s16le",
		Overwrite:  true,
	}
}

// Convert converts an audio file to the specified format
func (f *FFmpeg) Convert(ctx context.Context, inputPath, outputPath string, opts *ConvertOptions) error {
	if opts == nil {
		opts = DefaultConvertOptions()
	}

	args := []string{}

	// Add input
	args = append(args, "-i", inputPath)

	// Add audio filters
	filters := []string{}
	
	if opts.SampleRate > 0 {
		filters = append(filters, fmt.Sprintf("aresample=%d", opts.SampleRate))
	}
	
	if opts.Channels > 0 {
		filters = append(filters, fmt.Sprintf("ac=%d", opts.Channels))
	}

	if len(filters) > 0 {
		args = append(args, "-af", strings.Join(filters, ","))
	}

	// Add codec
	if opts.Codec != "" {
		args = append(args, "-c:a", opts.Codec)
	}

	// Add bit rate
	if opts.BitRate > 0 {
		args = append(args, "-b:a", fmt.Sprintf("%dk", opts.BitRate))
	}

	// Overwrite
	if opts.Overwrite {
		args = append(args, "-y")
	}

	// Output path
	args = append(args, outputPath)

	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg conversion failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// ConvertToWAV converts audio to WAV format optimized for Whisper
func (f *FFmpeg) ConvertToWAV(ctx context.Context, inputPath, outputPath string) error {
	return f.Convert(ctx, inputPath, outputPath, &ConvertOptions{
		SampleRate: 16000,
		Channels:   1,
		Codec:      "pcm_s16le",
		Overwrite:  true,
	})
}

// Check checks if FFmpeg and FFprobe are available
func Check() (ffmpeg, ffprobe bool, err error) {
	_, err = findExecutable("ffmpeg")
	ffmpeg = err == nil

	_, err = findExecutable("ffprobe")
	ffprobe = err == nil

	return ffmpeg, ffprobe, nil
}
