package whisper

// #cgo CFLAGS: -I/root/go/pkg/mod/github.com/ggerganov/whisper.cpp@v1.8.3/include -I/root/go/pkg/mod/github.com/ggerganov/whisper.cpp@v1.8.3/ggml/include
// #cgo LDFLAGS: -L/root/go/pkg/mod/github.com/ggerganov/whisper.cpp@v1.8.3/build/src -L/root/go/pkg/mod/github.com/ggerganov/whisper.cpp@v1.8.3/build/ggml/src -lwhisper -lggml -lggml-cpu -lggml-base -lm -lpthread -ldl
// #include <whisper.h>
// #include <stdlib.h>
// #include <stdio.h>
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

type Context struct {
	ctx *C.struct_whisper_context
}

// Params wraps whisper_full_params for Go
type Params struct {
	// Model params
	NThreads int
	NAccents int
	NMaxTextLen int
	
	// Audio params
	OffsetMs int
	DurationMs int
	
	// Transcription params
	Language       string
	Translate      bool
	NoContext      bool
	SingleSegment bool
	PrintSpecial  bool
	PrintProgress bool
	PrintRealtime bool
	PrintTimestamps bool
	
	// Sampling
	MaxLen int
	MaxTokens int
	Temperature float32
}

func DefaultParams() Params {
	return Params{
		NThreads:      4,
		NAccents:     2,
		NMaxTextLen:  256,
		Language:     "auto",
		PrintProgress: false,
		PrintSpecial: false,
		PrintRealtime: false,
		PrintTimestamps: false,
		MaxLen: 0,
		MaxTokens: 0,
		Temperature: 0.4,
	}
}

// InitFromFile initializes Whisper context from a model file
func InitFromFile(modelPath string) (*Context, error) {
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))
	
	// Get default context params
	cparams := C.whisper_context_default_params()
	
	ctx := C.whisper_init_from_file_with_params(cPath, cparams)
	if ctx == nil {
		return nil, fmt.Errorf("failed to initialize whisper from file: %s", modelPath)
	}
	
	context := &Context{ctx: ctx}
	runtime.SetFinalizer(context, func(c *Context) {
		if c.ctx != nil {
			C.whisper_free(c.ctx)
		}
	})
	
	return context, nil
}

// TranscribeFromFile transcribes audio directly from file path
// Note: whisper.cpp expects raw 16kHz float audio data
// This is a simplified wrapper - for real usage you'd need to convert audio
func (c *Context) TranscribeFromFile(audioPath string, params Params) (string, error) {
	// For now, we'll use a basic approach - read audio as raw floats
	// In production, you'd use FFmpeg to convert audio to 16kHz mono float
	
	// Use the examples/talk.wav for testing
	audioFile := C.CString(audioPath)
	defer C.free(unsafe.Pointer(audioFile))
	
	// Get default full params
	wparams := C.whisper_full_default_params(C.WHISPER_SAMPLING_GREEDY)
	
	wparams.n_threads = C.int(params.NThreads)
	wparams.n_max_text_ctx = C.int(params.NMaxTextLen)
	wparams.offset_ms = C.int(params.OffsetMs)
	wparams.duration_ms = C.int(params.DurationMs)
	wparams.print_special = C.bool(params.PrintSpecial)
	wparams.print_progress = C.bool(params.PrintProgress)
	wparams.print_realtime = C.bool(params.PrintRealtime)
	wparams.print_timestamps = C.bool(params.PrintTimestamps)
	wparams.translate = C.bool(params.Translate)
	wparams.no_context = C.bool(params.NoContext)
	wparams.single_segment = C.bool(params.SingleSegment)
	wparams.max_len = C.int(params.MaxLen)
	
	if params.Language != "" && params.Language != "auto" {
		lang := C.CString(params.Language)
		defer C.free(unsafe.Pointer(lang))
		wparams.language = lang
	}
	
	// For now, we need to read actual audio data
	// This is a placeholder - in practice you'd load audio file
	result := C.whisper_full(c.ctx, wparams, nil, 0)
	
	if result != 0 {
		return "", fmt.Errorf("whisper transcription failed with code: %d", result)
	}
	
	// Get the full transcription
	nSegments := C.whisper_full_n_segments(c.ctx)
	
	var fullText string
	for i := C.int(0); i < nSegments; i++ {
		text := C.whisper_full_get_segment_text(c.ctx, i)
		fullText += C.GoString(text)
	}
	
	return fullText, nil
}

// TranscribeAudio transcribes from float32 audio samples (16kHz mono)
func (c *Context) TranscribeAudio(samples []float32, params Params) (string, error) {
	wparams := C.whisper_full_default_params(C.WHISPER_SAMPLING_GREEDY)
	
	wparams.n_threads = C.int(params.NThreads)
	wparams.n_max_text_ctx = C.int(params.NMaxTextLen)
	wparams.offset_ms = C.int(params.OffsetMs)
	wparams.duration_ms = C.int(params.DurationMs)
	wparams.print_special = C.bool(params.PrintSpecial)
	wparams.print_progress = C.bool(params.PrintProgress)
	wparams.print_realtime = C.bool(params.PrintRealtime)
	wparams.print_timestamps = C.bool(params.PrintTimestamps)
	wparams.translate = C.bool(params.Translate)
	wparams.no_context = C.bool(params.NoContext)
	wparams.single_segment = C.bool(params.SingleSegment)
	wparams.max_len = C.int(params.MaxLen)
	wparams.temperature = C.float(params.Temperature)
	
	if params.Language != "" && params.Language != "auto" {
		lang := C.CString(params.Language)
		defer C.free(unsafe.Pointer(lang))
		wparams.language = lang
	}
	
	// Convert Go slice to C array
	if len(samples) == 0 {
		return "", fmt.Errorf("no audio samples provided")
	}
	
	cSamples := (*C.float)(unsafe.Pointer(&samples[0]))
	nSamples := C.int(len(samples))
	
	result := C.whisper_full(c.ctx, wparams, cSamples, nSamples)
	
	if result != 0 {
		return "", fmt.Errorf("whisper transcription failed with code: %d", result)
	}
	
	// Get the full transcription
	nSegments := C.whisper_full_n_segments(c.ctx)
	
	var fullText string
	for i := C.int(0); i < nSegments; i++ {
		text := C.whisper_full_get_segment_text(c.ctx, i)
		fullText += C.GoString(text)
	}
	
	return fullText, nil
}

// GetSegmentCount returns the number of transcription segments
func (c *Context) GetSegmentCount() int {
	return int(C.whisper_full_n_segments(c.ctx))
}

// GetSegmentText returns the text of a specific segment
func (c *Context) GetSegmentText(index int) string {
	text := C.whisper_full_get_segment_text(c.ctx, C.int(index))
	return C.GoString(text)
}

// GetSegmentTiming returns the timing of a specific segment
func (c *Context) GetSegmentTiming(index int) (startMs, endMs int) {
	start := C.whisper_full_get_segment_t0(c.ctx, C.int(index))
	end := C.whisper_full_get_segment_t1(c.ctx, C.int(index))
	return int(start), int(end)
}

// Free releases the whisper context
func (c *Context) Free() {
	if c.ctx != nil {
		C.whisper_free(c.ctx)
		c.ctx = nil
	}
}
