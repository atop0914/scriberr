package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	
	if cfg == nil {
		t.Error("Expected non-nil config")
	}
	
	if cfg.App.Name != "scriberr" {
		t.Errorf("Expected app name 'scriberr', got %s", cfg.App.Name)
	}
	
	if cfg.Model.Size != "base" {
		t.Errorf("Expected default model size 'base', got %s", cfg.Model.Size)
	}
	
	if cfg.Model.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", cfg.Model.MaxRetries)
	}
}

func TestLoad_NonExistent(t *testing.T) {
	// Should return default config when file doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if cfg == nil {
		t.Error("Expected non-nil config")
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	// Create a temp directory with invalid config
	tmpDir := t.TempDir()
	invalidConfig := filepath.Join(tmpDir, "invalid.yaml")
	
	// Write invalid YAML
	os.WriteFile(invalidConfig, []byte("invalid: yaml: content:"), 0644)
	
	// Set env to point to this file
	os.Setenv("SCIBERR_CONFIG", invalidConfig)
	defer os.Unsetenv("SCIBERR_CONFIG")
	
	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestConfig_Save(t *testing.T) {
	cfg := Default()
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	
	err := cfg.Save(tmpFile)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	// Verify the file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
	
	// Verify we can load it back
	loaded, err := Load()
	if err != nil {
		t.Errorf("Unexpected error loading saved config: %v", err)
	}
	
	if loaded.App.Name != cfg.App.Name {
		t.Errorf("Expected app name %s, got %s", cfg.App.Name, loaded.App.Name)
	}
}

func TestLogConfig(t *testing.T) {
	cfg := Default()
	
	// Test default log config
	if cfg.Log.Level == "" {
		t.Error("Expected non-empty log level")
	}
	
	if cfg.Log.Format == "" {
		t.Error("Expected non-empty log format")
	}
}

func TestModelConfig(t *testing.T) {
	cfg := Default()
	
	// Test model sizes
	validSizes := []string{"tiny", "base", "small", "medium", "large", "large-v2", "large-v3"}
	
	for _, size := range validSizes {
		cfg.Model.Size = size
		if cfg.Model.Size != size {
			t.Errorf("Expected size %s, got %s", size, cfg.Model.Size)
		}
	}
}
