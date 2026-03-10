package logger

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger := New()
	if logger == nil {
		t.Error("Expected non-nil logger")
	}
}

func TestLogger_Debug(t *testing.T) {
	logger := New()
	// Should not panic
	logger.Debug("test debug message", "key", "value")
}

func TestLogger_Info(t *testing.T) {
	logger := New()
	logger.Info("test info message", "key", "value")
}

func TestLogger_Warn(t *testing.T) {
	logger := New()
	logger.Warn("test warn message", "key", "value")
}

func TestLogger_Error(t *testing.T) {
	logger := New()
	logger.Error("test error message", "key", "value")
}

func TestLogger_With(t *testing.T) {
	logger := New()
	
	// Test chaining
	loggerWith := logger.With("key", "value")
	if loggerWith == nil {
		t.Error("Expected non-nil logger with context")
	}
	
	// Test that it logs with the context
	loggerWith.Info("test with context")
}

func TestLogger_SetLevel(t *testing.T) {
	logger := New()
	
	tests := []struct {
		level string
	}{
		{"debug"},
		{"info"},
		{"warn"},
		{"error"},
		{"invalid"}, // Should default to info
	}
	
	for _, tt := range tests {
		// Should not panic
		logger.SetLevel(tt.level)
	}
}

func TestFormatTime(t *testing.T) {
	// Just test the function runs without error
	_ = FormatTime(testTime)
}

var testTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
