package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult holds validation results
type ValidationResult struct {
	Valid   bool
	Errors  []ValidationError
	Warnings []ValidationError
}

// AddError adds an error to the validation result
func (r *ValidationResult) AddError(field, message string) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{Field: field, Message: message})
}

// AddWarning adds a warning to the validation result
func (r *ValidationResult) AddWarning(field, message string) {
	r.Warnings = append(r.Warnings, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are any errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// HasWarnings returns true if there are any warnings
func (r *ValidationResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// AudioFileValidator validates audio files
type AudioFileValidator struct {
	MaxFileSize   int64  // in bytes
	MinFileSize   int64  // in bytes
	AllowedFormats []string
}

// NewAudioFileValidator creates a new audio file validator
func NewAudioFileValidator() *AudioFileValidator {
	return &AudioFileValidator{
		MaxFileSize:   500 * 1024 * 1024, // 500MB default
		MinFileSize:   1024,              // 1KB minimum
		AllowedFormats: []string{".mp3", ".wav", ".m4a", ".ogg", ".flac", ".aac", ".wma", ".opus"},
	}
}

// ValidatePath validates an audio file path
func (v *AudioFileValidator) ValidatePath(path string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Check if path is empty
	if strings.TrimSpace(path) == "" {
		result.AddError("path", "file path cannot be empty")
		return result
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		result.AddError("path", fmt.Sprintf("file does not exist: %s", path))
		return result
	}

	// Check if it's a regular file
	info, err := os.Stat(path)
	if err != nil {
		result.AddError("path", fmt.Sprintf("cannot access file: %v", err))
		return result
	}

	if info.IsDir() {
		result.AddError("path", "path is a directory, not a file")
		return result
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	validFormat := false
	for _, format := range v.AllowedFormats {
		if ext == format {
			validFormat = true
			break
		}
	}

	if !validFormat {
		result.AddError("format", fmt.Sprintf("unsupported format: %s. Allowed: %v", ext, v.AllowedFormats))
	}

	// Check file size
	if info.Size() < v.MinFileSize {
		result.AddError("size", fmt.Sprintf("file too small: %d bytes (minimum: %d)", info.Size(), v.MinFileSize))
	}

	if info.Size() > v.MaxFileSize {
		result.AddError("size", fmt.Sprintf("file too large: %d bytes (maximum: %d)", info.Size(), v.MaxFileSize))
	}

	// Add warnings for unusual sizes
	if info.Size() < 10*1024 { // Less than 10KB
		result.AddWarning("size", "file is unusually small and may be corrupted or empty")
	}

	// Check file is readable
	file, err := os.Open(path)
	if err != nil {
		result.AddError("permissions", fmt.Sprintf("cannot read file: %v", err))
	} else {
		file.Close()
	}

	return result
}

// ValidatePaths validates multiple file paths
func (v *AudioFileValidator) ValidatePaths(paths []string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if len(paths) == 0 {
		result.AddError("paths", "no files provided")
		return result
	}

	for _, path := range paths {
		pathResult := v.ValidatePath(path)
		if pathResult.HasErrors() {
			for _, err := range pathResult.Errors {
				result.AddError(path, err.Message)
			}
		}
		if pathResult.HasWarnings() {
			for _, warn := range pathResult.Warnings {
				result.AddWarning(path, warn.Message)
			}
		}
	}

	return result
}

// OutputPathValidator validates output paths
type OutputPathValidator struct {
	AllowOverwrite bool
}

// NewOutputPathValidator creates a new output path validator
func NewOutputPathValidator() *OutputPathValidator {
	return &OutputPathValidator{
		AllowOverwrite: false,
	}
}

// ValidateOutputPath validates an output file path
func (v *OutputPathValidator) ValidateOutputPath(path string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if strings.TrimSpace(path) == "" {
		result.AddError("output", "output path cannot be empty")
		return result
	}

	// Check if parent directory exists
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		result.AddError("output", fmt.Sprintf("directory does not exist: %s", dir))
	}

	// Check if file already exists (if not allowing overwrite)
	if !v.AllowOverwrite {
		if _, err := os.Stat(path); err == nil {
			result.AddError("output", fmt.Sprintf("file already exists: %s (use --force to overwrite)", path))
		}
	}

	// Check if we can write to the directory
	if dir != "." {
		testFile := filepath.Join(dir, ".write_test")
		if err := os.WriteFile(testFile, []byte(""), 0644); err == nil {
			os.Remove(testFile)
		} else {
			result.AddError("permissions", fmt.Sprintf("cannot write to directory: %s", dir))
		}
	}

	return result
}

// ValidateLanguage validates a language code
func ValidateLanguage(lang string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// List of supported language codes for Whisper
	supported := map[string]string{
		"en": "English", "zh": "Chinese", "es": "Spanish", "fr": "French",
		"de": "German", "it": "Italian", "pt": "Portuguese", "ru": "Russian",
		"ja": "Japanese", "ko": "Korean", "ar": "Arabic", "hi": "Hindi",
		"nl": "Dutch", "pl": "Polish", "tr": "Turkish", "vi": "Vietnamese",
		"th": "Thai", "id": "Indonesian", "sv": "Swedish", "da": "Danish",
		"fi": "Finnish", "no": "Norwegian", "cs": "Czech", "el": "Greek",
		"he": "Hebrew", "hu": "Hungarian", "ro": "Romanian", "uk": "Ukrainian",
	}

	lang = strings.ToLower(lang)
	if lang == "auto" || lang == "" {
		return result // Auto-detection is always valid
	}

	if _, ok := supported[lang]; !ok {
		result.AddError("language", fmt.Sprintf("unsupported language: %s", lang))
	}

	return result
}
