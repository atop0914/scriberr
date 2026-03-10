package model

import (
	"testing"
)

func TestAvailableModels(t *testing.T) {
	models := AvailableModels()
	
	if len(models) == 0 {
		t.Error("Expected at least one model")
	}
	
	// Check that all expected models exist
	expectedSizes := []ModelSize{
		ModelTiny, ModelBase, ModelSmall, ModelMedium,
		ModelLarge, ModelLargeV2, ModelLargeV3,
	}
	
	if len(models) != len(expectedSizes) {
		t.Errorf("Expected %d models, got %d", len(expectedSizes), len(models))
	}
}

func TestIsValidModelSize(t *testing.T) {
	tests := []struct {
		size     string
		expected bool
	}{
		{"tiny", true},
		{"base", true},
		{"small", true},
		{"medium", true},
		{"large", true},
		{"large-v2", true},
		{"large-v3", true},
		{"invalid", false},
		{"", false},
		{"TINY", false}, // case sensitive
	}

	for _, tt := range tests {
		result := IsValidModelSize(tt.size)
		if result != tt.expected {
			t.Errorf("IsValidModelSize(%q) = %v, expected %v", tt.size, result, tt.expected)
		}
	}
}

func TestGetModelInfo(t *testing.T) {
	t.Run("valid size", func(t *testing.T) {
		info, err := GetModelInfo(ModelBase)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if info.Size != ModelBase {
			t.Errorf("Expected size %s, got %s", ModelBase, info.Size)
		}
		if info.Params == "" {
			t.Error("Expected non-empty params")
		}
	})

	t.Run("invalid size", func(t *testing.T) {
		_, err := GetModelInfo("invalid")
		if err == nil {
			t.Error("Expected error for invalid size")
		}
	})
}

func TestCacheDir(t *testing.T) {
	dir := CacheDir()
	if dir == "" {
		t.Error("Expected non-empty cache directory")
	}
}

func TestEnsureCacheDir(t *testing.T) {
	dir, err := EnsureCacheDir()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("Expected non-empty directory")
	}
}

func TestModelPath(t *testing.T) {
	tests := []struct {
		size     ModelSize
		expected string
	}{
		{ModelTiny, "ggml-tiny.bin"},
		{ModelBase, "ggml-base.bin"},
		{ModelSmall, "ggml-small.bin"},
		{ModelMedium, "ggml-medium.bin"},
		{ModelLarge, "ggml-large.bin"},
		{ModelLargeV2, "ggml-large-v2.bin"},
		{ModelLargeV3, "ggml-large-v3.bin"},
	}

	for _, tt := range tests {
		path, err := ModelPath(tt.size)
		if err != nil {
			t.Errorf("Unexpected error for %s: %v", tt.size, err)
		}
		if path == "" {
			t.Errorf("Expected non-empty path for %s", tt.size)
		}
		// Check that path ends with expected filename
		expectedFilename := tt.expected
		if len(path) < len(expectedFilename) || path[len(path)-len(expectedFilename):] != expectedFilename {
			t.Errorf("Expected path to end with %s, got %s", expectedFilename, path)
		}
	}
}

func TestIsModelCached(t *testing.T) {
	// Test with a non-existent model
	if IsModelCached("nonexistent") {
		t.Error("Expected false for non-existent model")
	}
	
	// Test with a valid but uncached model
	if IsModelCached(ModelBase) {
		t.Skip("Model already cached, skipping test")
	}
}

func TestListCachedModels(t *testing.T) {
	// This test just checks the function runs without error
	// Results depend on what's actually cached
	_ = ListCachedModels(nil) // Should not panic
	// The function may return nil or empty slice depending on cache state
}
