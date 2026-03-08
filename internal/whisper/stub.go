//go:build no_cgo || !cgo

package whisper

import (
	"fmt"
)

// Context is a stub implementation when CGO is disabled
type Context struct {
	// stub implementation
}

// Params wraps whisper_full_params for Go (stub)
type Params struct {
	// Model params
	NThreads      int
	NAccents      int
	NMaxTextLen   int
	OffsetMs      int
	DurationMs    int
	Language      string
	Translate     bool
	NoContext     bool
	SingleSegment bool
	PrintSpecial  bool
	PrintProgress bool
	PrintRealtime bool
	PrintTimestamps bool
	MaxLen        int
	MaxTokens     int
	Temperature   float32
}

func DefaultParams() Params {
	return Params{
		NThreads:       4,
		NAccents:      2,
		NMaxTextLen:   256,
		Language:      "auto",
		PrintProgress: false,
		PrintSpecial:  false,
		PrintRealtime: false,
		PrintTimestamps: false,
		MaxLen:        0,
		MaxTokens:     0,
		Temperature:   0.4,
	}
}

// InitFromFile is a stub that returns an error
func InitFromFile(modelPath string) (*Context, error) {
	return nil, fmt.Errorf("whisper C library not available (build with CGO_ENABLED=1)")
}

// TranscribeFromFile is a stub that returns an error
func (c *Context) TranscribeFromFile(audioPath string, params Params) (string, error) {
	return "", fmt.Errorf("whisper C library not available")
}

// TranscribeAudio is a stub that returns an error
func (c *Context) TranscribeAudio(samples []float32, params Params) (string, error) {
	return "", fmt.Errorf("whisper C library not available")
}

// GetSegmentCount returns 0
func (c *Context) GetSegmentCount() int {
	return 0
}

// GetSegmentText returns empty string
func (c *Context) GetSegmentText(index int) string {
	return ""
}

// GetSegmentTiming returns 0, 0
func (c *Context) GetSegmentTiming(index int) (startMs, endMs int) {
	return 0, 0
}

// Free is a no-op in stub
func (c *Context) Free() {
	// no-op
}
