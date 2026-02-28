# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-02-28

### Added
- **Multi-Browser Support**: Now supports both Google Chrome and Chromium.
- **Auto-Discovery**: Automatic detection of system-installed browsers (Chrome/Chromium/Chrome-Stable).
- **`vbrowser use` Command**: New CLI command to easily switch between browsers (`vbrowser use chrome` or `vbrowser use chromium`).
- **Architecture-Aware Instructions**: Installation help now correctly identifies and advises on browser availability for ARM64 vs x86_64 Linux.
- **Environment Variable**: Added `VBROWSER_BROWSER_PATH` for manual binary path override.

### Changed
- **System-First Approach**: Shifted from downloading Chromium snapshots to using system-installed versions (lighter, more stable, and easier to update).
- **Refactored Naming**: Renamed all "Chromium" specific configurations and internal logic to "Browser" for better generic support.
- **CLI Output**: Refined all CLI commands to remove extra trailing newlines for a cleaner terminal experience.
- **Improved Error Messages**: Standardized CLI error reporting with architecture-specific installation steps.

### Removed
- **Auto-Download Logic**: Removed the progress bar and snapshot downloading logic (reduced binary size and complexity).
- **Configuration Fields**: Removed `auto_download` and `download_dir` from `config.json`.
- **CLI Flags**: Removed `--no-download` flag from the `start` command.

### Fixed
- Duplicate error messages in CLI output.
- Syntax errors in browser finding logic.
- Inconsistent trailing newlines in `status` and `version` commands.

## [0.1.0] - 2026-02-24

### Added
- Initial release of vbrowser
- Single Go binary with embedded Chromium auto-download
- WebRTC-based streaming with VP8 video and Opus audio
- Native GStreamer integration for low-latency encoding
- Full mouse and keyboard interaction support
- Guacamole Keyboard library for proper key-to-keysym conversion
- Support for all keyboard layouts and international characters
- Persistent browser profile (cookies, bookmarks, passwords)
- Hot-reloadable resolution, FPS, and bitrate settings
- Simple CLI commands: start, stop, status, version
- SSH tunnel support for secure remote access
- PulseAudio integration for audio capture
- Xvfb virtual display support

### Performance
- Ultra-low latency WebRTC configuration (playoutDelayHint=0)
- Input batching for optimized mouse movement handling
- Tab visibility handling to prevent throttling lag
- 60 FPS streaming with configurable bitrate (default 8 Mbps)
- Optimized VP8 encoding settings for real-time performance

### Fixed
- XTEST BadValue errors when typing special characters
- Keyboard input lag and missing characters
- High latency when switching browser tabs
- WebRTC jitter buffer causing video delay

[0.2.0]: https://github.com/zulfikawr/vbrowser/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/zulfikawr/vbrowser/releases/tag/v0.1.0
