package audio

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestSupportedFormats(t *testing.T) {
	expectedFormats := []string{
		".mp3", ".wav", ".m4a", ".ogg", ".flac", ".aac", ".wma", ".opus",
	}
	
	if len(SupportedFormats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(SupportedFormats))
	}
	
	for i, format := range expectedFormats {
		if SupportedFormats[i] != format {
			t.Errorf("Expected format %s at index %d, got %s", format, i, SupportedFormats[i])
		}
	}
}

func TestGetSupportedFormats(t *testing.T) {
	formats := GetSupportedFormats()
	if len(formats) == 0 {
		t.Error("Expected non-empty formats list")
	}
}

func TestProcessorFindFFmpeg(t *testing.T) {
	// This test checks if ffmpeg can be found
	// It may skip if ffmpeg is not installed
	proc, err := NewProcessor()
	if err != nil {
		t.Skipf("FFmpeg not available: %v", err)
	}
	
	if proc == nil {
		t.Error("Expected non-nil processor")
	}
	
	if proc.ffmpegPath == "" {
		t.Error("Expected non-empty ffmpeg path")
	}
}

func TestIsFFmpegAvailable(t *testing.T) {
	// Just test the function runs without panic
	_ = IsFFmpegAvailable()
}

func TestAudioFile(t *testing.T) {
	// Test the AudioFile struct
	audioFile := AudioFile{
		Path:       "/test/path.mp3",
		Format:     ".mp3",
		Duration:   time.Second * 30,
		SampleRate: 44100,
		Channels:   2,
		BitRate:    320,
		Valid:      true,
	}
	
	if audioFile.Path != "/test/path.mp3" {
		t.Errorf("Expected path /test/path.mp3, got %s", audioFile.Path)
	}
	
	if audioFile.Duration != time.Second*30 {
		t.Errorf("Expected duration 30s, got %s", audioFile.Duration)
	}
}

func TestConvertToTemp(t *testing.T) {
	proc, err := NewProcessor()
	if err != nil {
		t.Skipf("FFmpeg not available: %v", err)
	}
	
	// Create a temp file that doesn't exist to test error handling
	tmpFile := filepath.Join(t.TempDir(), "nonexistent.wav")
	ctx := context.Background()
	
	_, err = proc.ConvertToTemp(ctx, tmpFile)
	// Should fail because the input file doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent input file")
	}
}

func TestValidateFile(t *testing.T) {
	proc, err := NewProcessor()
	if err != nil {
		t.Skipf("FFmpeg not available: %v", err)
	}
	
	tmpDir := t.TempDir()
	
	// Test with non-existent file
	_, err = proc.ValidateFile(filepath.Join(tmpDir, "nonexistent.mp3"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	
	// Test with valid audio file (we can't easily create one, so skip)
}

func TestNewProcessorWithProgress(t *testing.T) {
	// Test with a mock progress tracker
	progress := &mockProgressTracker{}
	
	proc, err := NewProcessorWithProgress(progress)
	if err != nil {
		t.Skipf("FFmpeg not available: %v", err)
	}
	
	if proc == nil {
		t.Error("Expected non-nil processor")
	}
	
	if proc.progress == nil {
		t.Error("Expected progress tracker to be set")
	}
}

// mockProgressTracker implements ProgressTracker for testing
type mockProgressTracker struct {
	started    bool
	completed  bool
	hasError   bool
	currentVal int
}

func (m *mockProgressTracker) Start(total int, message string) {
	m.started = true
}

func (m *mockProgressTracker) Update(current int) {
	m.currentVal = current
}

func (m *mockProgressTracker) Complete() {
	m.completed = true
}

func (m *mockProgressTracker) Error(err error) {
	m.hasError = true
}
