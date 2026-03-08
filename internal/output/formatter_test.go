package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestTextFormatter(t *testing.T) {
	transcription := &Transcription{
		Text:     "Hello world this is a test",
		Language: "en",
		Model:    "base",
		Timestamp: time.Now(),
		Segments: []Segment{
			{Index: 1, Start: 0.0, End: 1.5, Text: "Hello world", Timestamp: "00:00:00,000"},
			{Index: 2, Start: 1.5, End: 3.0, Text: "this is a test", Timestamp: "00:00:01,500"},
		},
	}

	formatter := &TextFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(transcription, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	expected := "Hello world this is a test"
	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

func TestJSONFormatter(t *testing.T) {
	transcription := &Transcription{
		Text:     "Hello world",
		Language: "en",
		Model:    "base",
		Timestamp: time.Now(),
		Segments: []Segment{
			{Index: 1, Start: 0.0, End: 1.5, Text: "Hello world", Timestamp: "00:00:00,000"},
		},
	}

	formatter := &JSONFormatter{Pretty: true}
	var buf bytes.Buffer
	err := formatter.Format(transcription, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	
	// Check that JSON contains the expected fields (accounting for formatting)
	if !strings.Contains(output, `"text"`) || !strings.Contains(output, `"Hello world"`) {
		t.Errorf("Expected JSON to contain text, got: %s", output)
	}
	if !strings.Contains(output, `"language"`) || !strings.Contains(output, `"en"`) {
		t.Errorf("Expected JSON to contain language, got: %s", output)
	}
}

func TestSRTFormatter(t *testing.T) {
	transcription := &Transcription{
		Text:     "Hello world",
		Language: "en",
		Timestamp: time.Now(),
		Segments: []Segment{
			{Index: 1, Start: 0.0, End: 1.5, Text: "Hello world", Timestamp: "00:00:00,000"},
			{Index: 2, Start: 1.5, End: 3.0, Text: "This is a test", Timestamp: "00:00:01,500"},
		},
	}

	formatter := &SRTFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(transcription, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	
	// Check SRT format
	if !strings.Contains(output, "1\n") {
		t.Errorf("Expected SRT to contain segment index 1")
	}
	if !strings.Contains(output, "00:00:00,000 --> 00:00:01,500") {
		t.Errorf("Expected SRT to contain correct timestamp, got: %s", output)
	}
	if !strings.Contains(output, "Hello world") {
		t.Errorf("Expected SRT to contain text, got: %s", output)
	}
}

func TestVTTFormatter(t *testing.T) {
	transcription := &Transcription{
		Text:     "Hello world",
		Language: "en",
		Timestamp: time.Now(),
		Segments: []Segment{
			{Index: 1, Start: 0.0, End: 1.5, Text: "Hello world", Timestamp: "00:00:00.000"},
		},
	}

	formatter := &VTTFormatter{}
	var buf bytes.Buffer
	err := formatter.Format(transcription, &buf)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	output := buf.String()
	
	// Check VTT format
	if !strings.Contains(output, "WEBVTT") {
		t.Errorf("Expected VTT to contain WEBVTT header")
	}
	if !strings.Contains(output, "00:00:00.000 --> 00:00:01.500") {
		t.Errorf("Expected VTT to contain correct timestamp, got: %s", output)
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"text", FormatText},
		{"txt", FormatText},
		{"TEXT", FormatText},
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"srt", FormatSRT},
		{"SRT", FormatSRT},
		{"vtt", FormatVTT},
		{"VTT", FormatVTT},
		{"unknown", FormatText}, // default
	}

	for _, tc := range tests {
		result := ParseFormat(tc.input)
		if result != tc.expected {
			t.Errorf("ParseFormat(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestGetFormatter(t *testing.T) {
	if GetFormatter(FormatText) == nil {
		t.Error("GetFormatter(FormatText) returned nil")
	}
	if GetFormatter(FormatJSON) == nil {
		t.Error("GetFormatter(FormatJSON) returned nil")
	}
	if GetFormatter(FormatSRT) == nil {
		t.Error("GetFormatter(FormatSRT) returned nil")
	}
	if GetFormatter(FormatVTT) == nil {
		t.Error("GetFormatter(FormatVTT) returned nil")
	}
}

func TestTextFormatterExtension(t *testing.T) {
	formatter := &TextFormatter{}
	if formatter.Extension() != ".txt" {
		t.Errorf("Expected .txt, got %s", formatter.Extension())
	}
}

func TestJSONFormatterExtension(t *testing.T) {
	formatter := &JSONFormatter{}
	if formatter.Extension() != ".json" {
		t.Errorf("Expected .json, got %s", formatter.Extension())
	}
}

func TestSRTFormatterExtension(t *testing.T) {
	formatter := &SRTFormatter{}
	if formatter.Extension() != ".srt" {
		t.Errorf("Expected .srt, got %s", formatter.Extension())
	}
}

func TestVTTFormatterExtension(t *testing.T) {
	formatter := &VTTFormatter{}
	if formatter.Extension() != ".vtt" {
		t.Errorf("Expected .vtt, got %s", formatter.Extension())
	}
}
