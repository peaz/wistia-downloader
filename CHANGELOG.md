# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.2] - 2025-09-15

### Added
- **Homebrew Support**: Added official Homebrew formula for easy installation on macOS
  - Created dedicated Homebrew tap at `peaz/homebrew-tap`
  - Users can now install with `brew install peaz/tap/wistia-downloader`
  - Updated installation documentation with Homebrew instructions

### Changed
- Updated README.md with Homebrew installation section
- Reorganized installation instructions to prioritize Homebrew for macOS users

## [1.1.1] - 2025-09-15

### Enhanced
- **URL Parsing**: Enhanced `-url` option to support detecting media IDs from `/medias/` URL patterns
  - Now supports URLs like `https://example.wistia.com/medias/abc123def456`
  - Automatically extracts `abc123def456` as the media ID from the URL path
  - Maintains backward compatibility with existing `wmediaid` parameter detection
- **Filename Handling**: Improved file naming with proper URL decoding and filesystem safety
  - HTML entities (like `&amp;`, `&lt;`, `&gt;`) are now properly decoded
  - Filesystem-safe decoded characters (like `&`) are preserved in filenames
  - Unsafe characters are replaced with underscores for better compatibility
  - Added handling for control characters and other problematic filename characters

### Changed
- Enhanced `extractVideoIDFromURL()` function to detect multiple URL patterns
- Improved `createSafeFilename()` and `generateVideoFilename()` functions with HTML entity decoding
- Added `html` package import for proper HTML entity handling
- **Code Quality**: Refactored user choice handling from if-else chain to tagged switch statement
- **Code Cleanup**: Removed unused `outputFlag` parameter from `handleChannelDownload()` function

### Technical
- Better regex patterns for URL parsing and character replacement
- Enhanced filename sanitization with expanded unsafe character detection
- Improved code readability and maintainability

## [1.1.0] - 2025-08-30

### Added
- **Channel Downloads**: Added support for downloading entire Wistia channels with all videos
- Channel detection for channel URLs (e.g., `https://fast.wistia.com/embed/channel/...`)
- User confirmation prompts for bulk channel downloads
- Video count and section breakdown display for channels
- Organized directory structure creation for channel downloads
- Descriptive filename generation based on video titles and sections
- Download progress tracking and summary for channel downloads
- Automatic detection of channel pages vs individual video pages

### Changed
- Improved error handling and user feedback
- Enhanced command-line interface for better user experience
- Updated build system to support new channel functionality
- Expanded documentation with channel download examples and features
- **Build System**: Improved build script to create consistently named binaries
  - All binaries are now named `wistia-downloader` (or `wistia-downloader.exe` on Windows) regardless of platform
  - Archives maintain platform-specific names but contain a consistently named binary
  - Build directory contains only distribution archives for cleaner releases

### Technical
- Significant code refactoring in `src/main.go` (544+ lines added, 68 removed)
- Updated all platform binaries (macOS, Linux, Windows for AMD64 and ARM64)
- Updated distribution archives from 1.0.1 to 1.1.0 versions
- Enhanced build script and Makefile

## [1.0.1] - 2025-08-28

### Added
- Initial release with core Wistia video downloading functionality
- Support for direct video ID downloads
- Support for extracting video ID from Wistia page URLs  
- Support for extracting video ID from HTML snippets (Copy link functionality)
- Cross-platform support (macOS Intel/Apple Silicon, Linux AMD64/ARM64, Windows AMD64/ARM64)
- Command-line interface with `-id`, `-url`, `-clipboard`, and `-o` options
- Pre-built binaries and distribution archives
- Comprehensive build system with `build.sh` script and Makefile
- Basic documentation and usage examples

### Technical
- Built with Go 1.25.0+
- Cross-compilation support for multiple architectures
- Optimized binary builds with distribution packaging

---

## Release Notes

### Version 1.1.0 Highlights

The major feature in version 1.1.0 is **Channel Downloads** - you can now download entire Wistia channels with all their videos in one command. The tool automatically:

- Detects when you're providing a channel URL vs individual video
- Shows you how many videos will be downloaded and their organization
- Asks for confirmation before starting bulk downloads
- Creates organized folder structures
- Uses descriptive filenames based on video titles

Example channel download:
```bash
./wistia-downloader -url "https://fast.wistia.com/embed/channel/m9k8d7f2jq?wchannelid=m9k8d7f2jq"
```

### Version 1.0.1 Foundation

Version 1.0.1 established the core functionality for individual video downloads with support for:
- Direct video IDs
- Page URLs 
- HTML snippets from Wistia's "Copy link" feature
- Multiple platform support with optimized binaries
