package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Format defines the output format type
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
	FormatSRT  Format = "srt"
	FormatVTT  Format = "vtt"
)

// Segment represents a transcription segment with timing
type Segment struct {
	Index     int     `json:"index"`
	Start     float64 `json:"start"`     // Start time in seconds
	End       float64 `json:"end"`       // End time in seconds
	Text      string  `json:"text"`       // Transcribed text
	Timestamp string  `json:"timestamp"` // Formatted timestamp (for SRT/VTT)
}

// Transcription holds the complete transcription result
type Transcription struct {
	Text      string    `json:"text"`
	Segments  []Segment `json:"segments,omitempty"`
	Language  string    `json:"language,omitempty"`
	Duration  float64   `json:"duration,omitempty"`
	Model     string    `json:"model,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Formatter is the interface for output formatters
type Formatter interface {
	Format(transcription *Transcription, w io.Writer) error
	Extension() string
}

// TextFormatter outputs plain text
type TextFormatter struct{}

func (f *TextFormatter) Format(transcription *Transcription, w io.Writer) error {
	_, err := w.Write([]byte(transcription.Text))
	return err
}

func (f *TextFormatter) Extension() string {
	return ".txt"
}

// JSONFormatter outputs JSON
type JSONFormatter struct {
	Pretty bool
}

func (f *JSONFormatter) Format(transcription *Transcription, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if f.Pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(transcription)
}

func (f *JSONFormatter) Extension() string {
	return ".json"
}

// SRTFormatter outputs SubRip subtitle format
type SRTFormatter struct{}

func (f *SRTFormatter) Format(transcription *Transcription, w io.Writer) error {
	for _, seg := range transcription.Segments {
		fmt.Fprintf(w, "%d\n", seg.Index)
		fmt.Fprintf(w, "%s --> %s\n", formatSRTTime(seg.Start), formatSRTTime(seg.End))
		fmt.Fprintf(w, "%s\n\n", strings.TrimSpace(seg.Text))
	}
	return nil
}

func (f *SRTFormatter) Extension() string {
	return ".srt"
}

// VTTFormatter outputs WebVTT subtitle format
type VTTFormatter struct{}

func (f *VTTFormatter) Format(transcription *Transcription, w io.Writer) error {
	fmt.Fprintln(w, "WEBVTT")
	fmt.Fprintln(w)

	for _, seg := range transcription.Segments {
		fmt.Fprintf(w, "%d\n", seg.Index)
		fmt.Fprintf(w, "%s --> %s\n", formatVTTTime(seg.Start), formatVTTTime(seg.End))
		fmt.Fprintf(w, "%s\n\n", strings.TrimSpace(seg.Text))
	}
	return nil
}

func (f *VTTFormatter) Extension() string {
	return ".vtt"
}

// GetFormatter returns the appropriate formatter for the given format
func GetFormatter(format Format) Formatter {
	switch format {
	case FormatText:
		return &TextFormatter{}
	case FormatJSON:
		return &JSONFormatter{Pretty: true}
	case FormatSRT:
		return &SRTFormatter{}
	case FormatVTT:
		return &VTTFormatter{}
	default:
		return &TextFormatter{}
	}
}

// ParseFormat parses a format string and returns the Format type
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "text", "txt":
		return FormatText
	case "json":
		return FormatJSON
	case "srt":
		return FormatSRT
	case "vtt":
		return FormatVTT
	default:
		return FormatText
	}
}

// WriteToFile writes the transcription to a file
func WriteToFile(transcription *Transcription, format Format, outputPath string) error {
	formatter := GetFormatter(format)

	// If no output path specified, use input filename with appropriate extension
	if outputPath == "" {
		return nil // Caller will handle output
	}

	// If output path is a directory, generate filename
	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		// Generate output filename based on transcription metadata
		outputPath = filepath.Join(outputPath, "transcription"+formatter.Extension())
	}

	w, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer w.Close()

	return formatter.Format(transcription, w)
}

// formatSRTTime formats time for SRT (HH:MM:SS,mmm)
func formatSRTTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

// formatVTTTime formats time for VTT (HH:MM:SS.mmm)
func formatVTTTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

// BuildTextOutput builds plain text from segments
func BuildTextOutput(segments []Segment) string {
	var sb strings.Builder
	for _, seg := range segments {
		sb.WriteString(seg.Text)
		sb.WriteString(" ")
	}
	return strings.TrimSpace(sb.String())
}

// BuildSegments builds Segment array from raw text and timing info
func BuildSegments(text string, timings []struct{ Start, End int }, model string, language string) *Transcription {
	segments := make([]Segment, 0, len(timings))
	
	for i, t := range timings {
		// Extract text for this segment (simplified - in reality would need proper segment boundaries)
		segmentText := text
		if len(timings) > 1 {
			// For multiple segments, we need to split text appropriately
			// This is a simplified approach
			segmentText = fmt.Sprintf("[%d]", i+1)
		}
		
		segments = append(segments, Segment{
			Index:     i + 1,
			Start:     float64(t.Start) / 1000.0, // Convert ms to seconds
			End:       float64(t.End) / 1000.0,
			Text:      segmentText,
			Timestamp: formatSRTTime(float64(t.Start) / 1000.0),
		})
	}
	
	// If no timings provided, create a single segment
	if len(timings) == 0 {
		segments = append(segments, Segment{
			Index:     1,
			Start:     0,
			End:       0,
			Text:      text,
			Timestamp: "00:00:00,000",
		})
	}
	
	return &Transcription{
		Text:      text,
		Language:  language,
		Segments:  segments,
		Model:     model,
		Timestamp: time.Now(),
	}
}
