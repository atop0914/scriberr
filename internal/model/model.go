package model

import (
	"fmt"
	"os"
	"path/filepath"

	"scriberr/internal/logger"
)

// ModelSize represents the Whisper model size
type ModelSize string

const (
	ModelTiny    ModelSize = "tiny"
	ModelBase    ModelSize = "base"
	ModelSmall   ModelSize = "small"
	ModelMedium  ModelSize = "medium"
	ModelLarge   ModelSize = "large"
	ModelLargeV2 ModelSize = "large-v2"
	ModelLargeV3 ModelSize = "large-v3"
)

// ModelInfo contains information about a model
type ModelInfo struct {
	Name        string    `json:"name"`
	Size        ModelSize `json:"size"`
	Params      string    `json:"params"`
	FileSize    string    `json:"file_size"`
	DownloadURL string    `json:"download_url"`
	Description string    `json:"description"`
}

// AvailableModels returns the list of available Whisper models
func AvailableModels() []ModelInfo {
	return []ModelInfo{
		{
			Name:        "tiny",
			Size:        ModelTiny,
			Params:      "39M",
			FileSize:    "~75 MB",
			DownloadURL: "https://huggingface.co/datasets/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
			Description: "Tiny model - fastest, lowest accuracy",
		},
		{
			Name:        "base",
			Size:        ModelBase,
			Params:      "74M",
			FileSize:    "~142 MB",
			DownloadURL: "https://huggingface.co/datasets/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
			Description: "Base model - balanced speed and accuracy",
		},
		{
			Name:        "small",
			Size:        ModelSmall,
			Params:      "244M",
			FileSize:    "~466 MB",
			DownloadURL: "https://huggingface.co/datasets/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
			Description: "Small model - good accuracy",
		},
		{
			Name:        "medium",
			Size:        ModelMedium,
			Params:      "769M",
			FileSize:    "~1.5 GB",
			DownloadURL: "https://huggingface.co/datasets/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin",
			Description: "Medium model - high accuracy",
		},
		{
			Name:        "large",
			Size:        ModelLarge,
			Params:      "1550M",
			FileSize:    "~3.0 GB",
			DownloadURL: "https://huggingface.co/datasets/ggerganov/whisper.cpp/resolve/main/ggml-large.bin",
			Description: "Large model - highest accuracy (v1)",
		},
		{
			Name:        "large-v2",
			Size:        ModelLargeV2,
			Params:      "1550M",
			FileSize:    "~3.0 GB",
			DownloadURL: "https://huggingface.co/datasets/ggerganov/whisper.cpp/resolve/main/ggml-large-v2.bin",
			Description: "Large v2 model - highest accuracy",
		},
		{
			Name:        "large-v3",
			Size:        ModelLargeV3,
			Params:      "1550M",
			FileSize:    "~3.0 GB",
			DownloadURL: "https://huggingface.co/datasets/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin",
			Description: "Large v3 model - latest highest accuracy",
		},
	}
}

// GetModelInfo returns info for a specific model size
func GetModelInfo(size ModelSize) (*ModelInfo, error) {
	for _, m := range AvailableModels() {
		if m.Size == size {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("unknown model size: %s", size)
}

// IsValidModelSize checks if the given size is valid
func IsValidModelSize(size string) bool {
	for _, m := range AvailableModels() {
		if string(m.Size) == size {
			return true
		}
	}
	return false
}

// CacheDir returns the model cache directory path
func CacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to temp directory
		return filepath.Join(os.TempDir(), "scriberr", "models")
	}
	return filepath.Join(homeDir, ".scriberr", "models")
}

// EnsureCacheDir creates the cache directory if it doesn't exist
func EnsureCacheDir() (string, error) {
	cacheDir := CacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	return cacheDir, nil
}

// ModelPath returns the path to a cached model file
func ModelPath(size ModelSize) (string, error) {
	cacheDir, err := EnsureCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, fmt.Sprintf("ggml-%s.bin", size)), nil
}

// IsModelCached checks if a model is already downloaded
func IsModelCached(size ModelSize) bool {
	path, err := ModelPath(size)
	if err != nil {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Size() > 0
}

// ListCachedModels returns a list of cached models
func ListCachedModels(log *logger.Logger) []ModelInfo {
	cacheDir, err := EnsureCacheDir()
	if err != nil {
		log.Warn("Failed to access cache directory", "error", err)
		return []ModelInfo{}
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return []ModelInfo{}
	}

	var cached []ModelInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Extract model size from filename like "ggml-tiny.bin"
		if len(name) > 6 && name[:5] == "ggml-" && name[len(name)-4:] == ".bin" {
			sizeStr := name[5 : len(name)-4]
			if info, err := GetModelInfo(ModelSize(sizeStr)); err == nil {
				// Get actual file size
				path := filepath.Join(cacheDir, name)
				if stat, err := os.Stat(path); err == nil {
					info.FileSize = formatFileSize(stat.Size())
				}
				cached = append(cached, *info)
			}
		}
	}
	return cached
}

// CacheStats returns cache statistics
func CacheStats(log *logger.Logger) (totalModels int, totalSize int64, err error) {
	cached := ListCachedModels(log)
	totalModels = len(cached)

	for _, m := range cached {
		path, err := ModelPath(m.Size)
		if err != nil {
			continue
		}
		if stat, err := os.Stat(path); err == nil {
			totalSize += stat.Size()
		}
	}
	return totalModels, totalSize, nil
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
