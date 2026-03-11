# Scriberr

A fast, local audio transcription tool powered by OpenAI's Whisper model. Scriberr runs entirely on your machine - no cloud APIs, no data leaves your computer.

## Features

- **Local Transcription** - All processing happens on your device using Whisper.cpp
- **Multiple Formats** - Support for MP3, WAV, M4A, OGG, FLAC, AAC, WMA, OPUS
- **Multiple Output Formats** - Export as plain text, JSON, SRT, or VTT
- **Batch Processing** - Transcribe multiple files at once with glob patterns
- **Model Options** - Choose from tiny to large-v3 models based on your needs
- **Language Support** - Auto-detect or specify source language
- **Timestamps** - Optional timestamped output for subtitle generation

## Installation

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [FFmpeg](https://ffmpeg.org/download.html) (for audio format conversion)
- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp) (optional, for native builds)

### Build from Source

```bash
git clone https://github.com/yourusername/scriberr.git
cd scriberr
go build -o scriberr ./cmd
```

### Pre-built Binary

Download the latest release for your platform from the releases page.

## Quick Start

```bash
# Transcribe a single file
./scriberr transcribe audio.mp3

# Specify output format
./scriberr transcribe audio.mp3 --format json

# Use a specific model
./scriberr transcribe audio.mp3 --model small

# Batch process multiple files
./scriberr transcribe "*.mp3" --batch

# Output to a specific directory
./scriberr transcribe audio.mp3 --output ./transcriptions/
```

## Usage

### Basic Transcription

```bash
# Transcribe to stdout (text format)
./scriberr transcribe audio.mp3

# Transcribe to file
./scriberr transcribe audio.mp3 -o output.txt

# Specify format
./scriberr transcribe audio.mp3 --format srt -o subtitles.srt
```

### Model Selection

Available models (size/VRAM/speed tradeoffs):

| Model   | Size  | VRAM  | Use Case |
|---------|-------|-------|----------|
| tiny    | 39 MB | ~1 GB | Fastest, lowest accuracy |
| base    | 74 MB | ~1 GB | Good balance |
| small   | 244 MB| ~2 GB | Better accuracy |
| medium  | 769 MB| ~5 GB | High accuracy |
| large   | 1550 MB| ~10 GB| Highest accuracy |
| large-v2| 1550 MB| ~10 GB| Improved v2 weights |
| large-v3| 1550 MB| ~10 GB| Latest version |

```bash
./scriberr transcribe audio.mp3 --model base
```

### Language Options

```bash
# Auto-detect language (default)
./scriberr transcribe audio.mp3 --language auto

# Specify language (e.g., English, Chinese, Spanish)
./scriberr transcribe audio.mp3 --language en
./scriberr transcribe audio.mp3 --language zh
./scriberr transcribe audio.mp3 --language es
```

### Batch Processing

```bash
# Process all files matching a pattern
./scriberr transcribe "*.mp3" --batch

# Recursive directory scanning
./scriberr transcribe ./audiofolder --recursive

# Output to directory
./scriberr transcribe "*.mp3" --output-dir ./transcripts/
```

### Output Formats

- **text** - Plain text transcript
- **json** - Structured JSON with segments and timestamps
- **srt** - SubRip subtitle format
- **vtt** - WebVTT subtitle format

```bash
./scriberr transcribe audio.mp3 --format json
./scriberr transcribe audio.mp3 --format srt
```

### Advanced Options

```bash
# Use multiple threads for faster processing
./scriberr transcribe audio.mp3 --threads 8

# Include timestamps in output
./scriberr transcribe audio.mp3 --timestamps

# Translate to English
./scriberr transcribe audio.mp3 --translate

# Verbose output
./scriberr transcribe audio.mp3 --verbose
```

## Configuration

Create a `config.yaml` file to set defaults:

```yaml
app:
  name: scriberr
  version: 0.1.0
  environment: development

model:
  size: base
  cacheDir: ~/.cache/whisper
  maxRetries: 3

log:
  level: info
  format: json
```

### Environment Variables

- `SCIBERR_CONFIG` - Path to config file
- `SCIBERR_LOG_LEVEL` - Log level (debug, info, warn, error)

## Architecture

```
scriberr/
├── cmd/              # CLI entry point
├── internal/
│   ├── audio/        # Audio processing
│   ├── commands/    # CLI commands
│   ├── config/      # Configuration management
│   ├── logger/      # Logging
│   ├── model/       # Model management
│   ├── output/      # Output formatters
│   ├── progress/    # Progress display
│   ├── validator/   # Input validation
│   └── whisper/     # Whisper.cpp integration
├── pkg/ffmpeg/      # FFmpeg wrapper
└── whisper/         # Whisper.cpp bindings
```

## Development

### Running Tests

```bash
go test ./...
```

### Code Structure

- `internal/commands/` - CLI command implementations
- `internal/output/` - Output formatters (text, JSON, SRT, VTT)
- `internal/whisper/` - Whisper model interface
- `internal/model/` - Model download and management

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [Whisper](https://github.com/openai/whisper) - OpenAI's Whisper model
- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp) - C++ port of Whisper
- [urfave/cli](https://github.com/urfave/cli) - CLI framework
