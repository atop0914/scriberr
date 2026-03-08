package model

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"scriberr/internal/logger"
)

// Downloader handles model downloads
type Downloader struct {
	log      *logger.Logger
	client   *http.Client
	progress func(downloaded, total int64)
}

// NewDownloader creates a new model downloader
func NewDownloader(log *logger.Logger) *Downloader {
	return &Downloader{
		log: log,
		client: &http.Client{
			Timeout: 0, // No timeout for large downloads
		},
	}
}

// SetProgressCallback sets the progress callback
func (d *Downloader) SetProgressCallback(callback func(downloaded, total int64)) {
	d.progress = callback
}

// Download downloads a model and returns the path to the downloaded file
func (d *Downloader) Download(size ModelSize) (string, error) {
	// Get model info
	info, err := GetModelInfo(size)
	if err != nil {
		return "", err
	}

	// Check if already cached
	if IsModelCached(size) {
		d.log.Info("Model already cached", "size", size)
		path, _ := ModelPath(size)
		return path, nil
	}

	// Ensure cache directory
	cacheDir, err := EnsureCacheDir()
	if err != nil {
		return "", err
	}

	// Create temp file path
	tempPath := filepath.Join(cacheDir, fmt.Sprintf("ggml-%s.bin.tmp", size))
	finalPath := filepath.Join(cacheDir, fmt.Sprintf("ggml-%s.bin", size))

	d.log.Info("Downloading model", "size", size, "url", info.DownloadURL)

	// Download the file
	req, err := http.NewRequest("GET", info.DownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Get total size
	totalSize := resp.ContentLength

	// Create temp file
	outFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer outFile.Close()

	// Download with progress
	var downloaded int64
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
				return "", fmt.Errorf("failed to write: %w", writeErr)
			}
			downloaded += int64(n)

			// Call progress callback
			if d.progress != nil && totalSize > 0 {
				d.progress(downloaded, totalSize)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("download error: %w", err)
		}
	}

	// Sync and close
	if err := outFile.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync: %w", err)
	}
	outFile.Close()

	// Rename to final path
	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", fmt.Errorf("failed to rename: %w", err)
	}

	d.log.Info("Model downloaded successfully", "size", size, "path", finalPath)
	return finalPath, nil
}

// DownloadWithRetry downloads with retry logic
func (d *Downloader) DownloadWithRetry(size ModelSize, maxRetries int) (string, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		path, err := d.Download(size)
		if err == nil {
			return path, nil
		}
		lastErr = err
		d.log.Warn("Download failed, retrying...", "attempt", i+1, "error", err)
	}
	return "", fmt.Errorf("download failed after %d retries: %w", maxRetries, lastErr)
}

// DeleteModel deletes a cached model
func (d *Downloader) DeleteModel(size ModelSize) error {
	path, err := ModelPath(size)
	if err != nil {
		return err
	}

	if !IsModelCached(size) {
		return fmt.Errorf("model not cached: %s", size)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	d.log.Info("Model deleted", "size", size)
	return nil
}

// ClearCache deletes all cached models
func (d *Downloader) ClearCache() error {
	cacheDir, err := EnsureCacheDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "ggml-") && strings.HasSuffix(name, ".bin") {
			path := filepath.Join(cacheDir, name)
			if err := os.Remove(path); err != nil {
				d.log.Warn("Failed to delete model file", "path", path, "error", err)
			}
		}
	}

	d.log.Info("Cache cleared")
	return nil
}
