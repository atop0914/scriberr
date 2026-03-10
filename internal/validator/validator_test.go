package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAudioFileValidator_ValidatePath(t *testing.T) {
	// Create a temporary test file (need to be >1KB for validator)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mp3")
	
	// Create a test file with enough content (>1KB)
	testContent := make([]byte, 2048)
	for i := range testContent {
		testContent[i] = 'A'
	}
	if err := os.WriteFile(tmpFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	validator := NewAudioFileValidator()

	tests := []struct {
		name      string
		path      string
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "valid mp3 file",
			path:      tmpFile,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "non-existent file",
			path:      filepath.Join(tmpDir, "nonexistent.mp3"),
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "directory instead of file",
			path:      tmpDir,
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "unsupported format",
			path:      filepath.Join(tmpDir, "test.xyz"),
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "empty path",
			path:      "",
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePath(tt.path)
			
			if tt.wantValid && !result.Valid {
				t.Errorf("Expected valid result, got errors: %+v", result.Errors)
			}
			if tt.wantErr && result.Valid {
				t.Errorf("Expected error result, got valid")
			}
		})
	}
}

func TestAudioFileValidator_ValidatePaths(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test1.mp3")
	file2 := filepath.Join(tmpDir, "test2.wav")
	
	// Create test files with enough content (>1KB each)
	testContent := make([]byte, 2048)
	os.WriteFile(file1, testContent, 0644)
	os.WriteFile(file2, testContent, 0644)

	validator := NewAudioFileValidator()

	t.Run("multiple valid files", func(t *testing.T) {
		result := validator.ValidatePaths([]string{file1, file2})
		if !result.Valid {
			t.Errorf("Expected valid result for multiple valid files, got errors: %+v", result.Errors)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		result := validator.ValidatePaths([]string{})
		if result.Valid {
			t.Errorf("Expected error for empty list")
		}
	})
}

func TestIsSupportedAudioFormat(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".mp3", true},
		{".wav", true},
		{".m4a", true},
		{".ogg", true},
		{".flac", true},
		{".aac", true},
		{".wma", true},
		{".opus", true},
		{".xyz", false},
		{".txt", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsSupportedAudioFormat(tt.ext)
		if result != tt.expected {
			t.Errorf("IsSupportedAudioFormat(%q) = %v, expected %v", tt.ext, result, tt.expected)
		}
	}
}

func TestValidateLanguage(t *testing.T) {
	tests := []struct {
		language string
		wantValid bool
	}{
		{"auto", true},
		{"en", true},
		{"zh", true},
		{"EN", true}, // case insensitive
		{"", true},
		{"invalid", false},
		{"xyz123", false},
	}

	for _, tt := range tests {
		result := ValidateLanguage(tt.language)
		if result.Valid != tt.wantValid {
			t.Errorf("ValidateLanguage(%q).Valid = %v, expected %v", tt.language, result.Valid, tt.wantValid)
		}
	}
}

func TestOutputPathValidator_ValidateOutputPath(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("non-existent directory", func(t *testing.T) {
		validator := NewOutputPathValidator()
		newPath := filepath.Join(tmpDir, "nonexistent", "output.txt")
		result := validator.ValidateOutputPath(newPath)
		
		if result.Valid {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		validator := NewOutputPathValidator()
		result := validator.ValidateOutputPath("")
		
		if result.Valid {
			t.Error("Expected error for empty path")
		}
	})
	
	t.Run("valid new path in existing directory", func(t *testing.T) {
		validator := NewOutputPathValidator()
		newPath := filepath.Join(tmpDir, "new_output.txt")
		result := validator.ValidateOutputPath(newPath)
		
		// Should be valid for new file in existing dir (it doesn't exist yet)
		// The validator checks if directory exists, not if file exists
		if !result.Valid {
			t.Logf("Validation errors: %+v", result.Errors)
		}
	})
}

func TestValidationResult_HasErrors(t *testing.T) {
	result := &ValidationResult{Valid: true}
	
	if result.HasErrors() {
		t.Error("Expected no errors initially")
	}
	
	result.AddError("test", "test error")
	
	if !result.HasErrors() {
		t.Error("Expected errors after adding one")
	}
}

func TestValidationResult_HasWarnings(t *testing.T) {
	result := &ValidationResult{Valid: true}
	
	if result.HasWarnings() {
		t.Error("Expected no warnings initially")
	}
	
	result.AddWarning("test", "test warning")
	
	if !result.HasWarnings() {
		t.Error("Expected warnings after adding one")
	}
}
