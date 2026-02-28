# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-02-28

### Added
- **Firefox Support**: Full support for running Firefox as the virtual browser engine.
- **Silent & Robust Start**: Automatic profile initialization to bypass "Welcome" screens and automatic clearing of stale lock files.
- **Firefox CDP Integration**: Full compatibility with Remote Debugging for Firefox.
- **Snap Compatibility**: Support for Ubuntu's Firefox Snap via a dedicated non-hidden profile path (`~/vbrowser_firefox_profile`).
- **Improved Browser Discovery**: Added system detection for Firefox across Linux, macOS, and Windows.

### Changed
- **Driver Abstraction**: Refactored browser management to handle engine-specific CLI flags (Gecko vs Blink) and ensure proper window sizing across all browsers.
- **Profile Isolation**: Separated Firefox and Chromium profiles into dedicated subdirectories to ensure stability and prevent engine conflicts.

## [0.3.0] - 2026-02-28

### Added
- **New Settings UI**: Replaced the top-hover menu with a sleek, non-intrusive side-panel drawer.
- **Ghost-State Trigger**: Added a semi-transparent gear icon in the bottom-right corner that pulses on load to guide the user.
- **Connection Heartbeat**: Implemented a 5-second WebSocket ping/pong mechanism to prevent random disconnects.
- **Auto-Reconnect on Focus**: Stream now automatically disconnects when the tab is backgrounded and reconnects on focus, preventing lag buildup from browser throttling.

### Changed
- **Increased Timeouts**: Updated server-side Read/Write timeouts to 1 hour to support long-running sessions.
- **Minimal Interface**: Cleaned up the overlay UI to provide a more immersive, distraction-free browser stream.

### Fixed
- **Top Tab Interference**: Fixed the issue where the settings menu would block access to the browser's top tabs.
- **Tab Switching Lag**: Eliminated the high latency and "catch-up" lag that occurred when returning to a backgrounded vbrowser tab.
- **Random Disconnects**: Resolved periodic connection drops caused by intermediate proxy or router inactivity timeouts.

## [0.2.0] - 2026-02-28

### Added
- **Multi-Browser Support**: Now supports both Google Chrome and Chromium.
- **Auto-Discovery**: Automatic detection of system-installed browsers (Chrome/Chromium/Chrome-Stable).
- **`vbrowser use` Command**: New CLI command to easily switch between browsers (`vbrowser use chrome`, `chromium`, or `firefox`).
- **Architecture-Aware Instructions**: Installation help now correctly identifies and advises on browser availability for ARM64 vs x86_64 Linux.
- **Environment Variable**: Added `VBROWSER_BROWSER_PATH` for manual binary path override.

### Changed
- **System-First Approach**: Shifted from downloading Chromium snapshots to using system-installed versions.
- **Refactored Naming**: Renamed all "Chromium" specific configurations and internal logic to "Browser".
- **CLI Output**: Refined all CLI commands to remove extra trailing newlines.
- **Improved Error Messages**: Standardized CLI error reporting with architecture-specific installation steps.

### Removed
- **Auto-Download Logic**: Removed the snapshot downloading logic (reduced binary size and complexity).
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

[0.4.0]: https://github.com/zulfikawr/vbrowser/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/zulfikawr/vbrowser/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/zulfikawr/vbrowser/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/zulfikawr/vbrowser/releases/tag/v0.1.0
