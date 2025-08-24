# Wistia Downloader

A command-line tool to download videos from Wistia hosting platform. Supports multiple input methods and cross-platform compilation.

## Features

- Download videos using Wistia video ID
- Extract video ID from Wistia page URLs
- Extract video ID from HTML snippets (from "Copy link" functionality)

## Installation

### Pre-built Binaries

Download the latest pre-built binaries from the releases section for your platform:

- **macOS**: 
  - Intel: [`wistia-downloader-macos-amd64`](build/wistia-downloader-macos-amd64) | [`tar.gz`](build/wistia-downloader-macos-amd64-1.0.0.tar.gz)
  - Apple Silicon: [`wistia-downloader-macos-arm64`](build/wistia-downloader-macos-arm64) | [`tar.gz`](build/wistia-downloader-macos-arm64-1.0.0.tar.gz)
- **Linux**: 
  - Intel/AMD: [`wistia-downloader-linux-amd64`](build/wistia-downloader-linux-amd64) | [`tar.gz`](build/wistia-downloader-linux-amd64-1.0.0.tar.gz)
  - ARM: [`wistia-downloader-linux-arm64`](build/wistia-downloader-linux-arm64) | [`tar.gz`](build/wistia-downloader-linux-arm64-1.0.0.tar.gz)
- **Windows**: 
  - Intel/AMD: [`wistia-downloader-windows-amd64.exe`](build/wistia-downloader-windows-amd64.exe) | [`zip`](build/wistia-downloader-windows-amd64.exe-1.0.0.zip)
  - ARM: [`wistia-downloader-windows-arm64.exe`](build/wistia-downloader-windows-arm64.exe) | [`zip`](build/wistia-downloader-windows-arm64.exe-1.0.0.zip)

### Build from Source

#### Prerequisites

- Go 1.25.0 or later
- Git

#### Building

```bash
# Clone the repository
git clone <repository-url>
cd wistia-downloader

# Build for all platforms
./build.sh

# Or build for current platform only
make build-local

# Or use Go directly
go build -o wistia-downloader ./src
```

## Usage

The tool supports three different input methods:

### 1. Direct Video ID

```bash
./wistia-downloader -id tra6gsm6rl -o my-video.mp4
```

### 2. Wistia Page URL

```bash
./wistia-downloader -url "https://workato.wistia.com/a/9cqncbbbtw05gh8" -o my-video.mp4
```

### 3. HTML Snippet (from "Copy link")

```bash
./wistia-downloader -clipboard '<div class="wistia_responsive_padding">...' -o my-video.mp4
```

### Command Line Options

- `-id <videoID>`: Wistia video ID (e.g., tra6gsm6rl)
- `-url <pageURL>`: Main Wistia page URL
- `-clipboard <htmlSnippet>`: HTML snippet containing wvideo parameter
- `-o <filename>`: Output filename (default: video.mp4)

## Building

### Build Script

The project includes a comprehensive build script (`build.sh`) that creates binaries for multiple platforms:

```bash
# Build all platforms
./build.sh

# Build with custom version
VERSION=1.0.0 ./build.sh
```

**Supported Platforms:**
- macOS (Intel and Apple Silicon)
- Linux (Intel/AMD64 and ARM64)
- Windows (Intel/AMD64 and ARM64)

The build script will:
1. Create optimized binaries for all platforms
2. Generate distribution archives (`.tar.gz` for Unix, `.zip` for Windows)
3. Display build summary with file sizes

### Makefile

For convenience, a Makefile is also provided:

```bash
# Show available targets
make help

# Build for all platforms
make build

# Build for current platform only
make build-local

# Clean build artifacts
make clean

# Run tests
make test

# Run locally with arguments
make run ARGS="-id tra6gsm6rl"

# Install dependencies
make install-deps

# Development build (with debug info)
make build-dev
```
