# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.1] - 2026-03-07

### Added
- **Broadcast Mode & Host Control**: Replaced the strict single-user lock with a new `Broadcaster` system. Multiple users can now connect and view the same stream simultaneously. The first user becomes the "Host" (controlling the browser), while others are "Viewers". Viewers can request control via a new UI button.
- **Process Reaper**: Implemented automatic cleanup of child processes (Xvfb, Browser) on parent crash or kill.

### Fixed
- **Status Command Output**: Added missing trailing newlines to the `status` command output for cleaner terminal display.
- **Instant Refocus Recovery**: Implemented a professional "Hard Re-sync" mechanism that clears internal WebRTC buffers upon window focus by re-attaching the media stream and requesting an immediate keyframe.
- **Latency Overhaul**: Fine-tuned the VP8 encoder (`keyframe-max-dist=10`, `error-resilient=partitions`) to ensure rapid recovery from background throttling.

## [0.7.0] - 2026-03-02

### Added
- **Hardware Acceleration**: Added support for VA-API (Intel/AMD) and NVENC (Nvidia) hardware encoders.
- **H.264 Codec Support**: New support for H.264 video codec with optimized `openh264enc` settings for ARM64 and low-power systems.
- **Dynamic Encoding**: Added `stream.encoder` and `stream.video_codec` configuration options.
- **CLI QoL Commands**: Added `vbrowser list` to show detected browsers and `vbrowser logs` for easy access to live service logs.
- **Process Reaper**: Implemented `PR_SET_PDEATHSIG` for all child processes (Xvfb, Browser) to ensure they are automatically terminated if the main process crashes or is killed.
- **Secure Authentication**: Switched from raw-text cookies to SHA-256 hashed tokens for improved session security.
- **Login UI Polish**: Added a password visibility toggle (eye icon) and improved input alignment on the login page.
- **Documentation Overhaul**: Created a dedicated `docs/` folder with detailed guides for installation, CLI, and configuration.

### Changed
- **Command Cleanup**: Removed redundant `vbrowser use` command in favor of `vbrowser config browser`.
- **Audio Isolation**: Refactored PulseAudio initialization to be non-destructive and isolated for shared environments.
- **Graceful Shutdown**: Enhanced the cleanup sequence with timeouts and more reliable process termination.

### Removed
- **Redundant Logic**: Deleted unused legacy capture package and redundant FFmpeg-based encoder logic to streamline the codebase.

## [0.6.1] - 2026-03-02

### Fixed
- **Background Tab Latency**: Implemented an immediate "jump to live" and keyframe request when returning to a backgrounded tab, eliminating unresponsiveness.

### Changed
- **Sub-Second Latency Tuning**: Optimized VP8 encoder with `lag-in-frames=0` and `rc-lookahead=0` for minimum processing delay.
- **Aggressive Buffer Management**: Reduced catch-up threshold and implemented smooth `playbackRate` speed-up (1.1x) for minor lags.

## [0.6.0] - 2026-03-02

### Added
- **Systemd Service Management**: New `vbrowser service install/uninstall` commands to run vbrowser as a robust background service with auto-restart.
- **Automatic Clipboard Sync**: Fully automatic, bi-directional clipboard synchronization between local and virtual browsers using a transparent input overlay.
- **Ultra-Low Latency Streaming**: Major GStreamer and WebRTC optimizations including PLI (Picture Loss Indication) support and refined VP8 encoding parameters.
- **Improved Input Handling**: Re-engineered input logic using a transparent textarea overlay for near-native typing and interaction reliability.

### Changed
- **Persistent Connections**: The stream now remains live in the background when the browser tab is hidden, eliminating reconnection delays.
- **Responsive Browser Flags**: Optimized Chromium and Firefox launch arguments to prevent background throttling and sleep modes.

### Fixed
- **Stop Command Reliability**: The `stop` command is now more aggressive and correctly stops the systemd service to prevent unwanted auto-restarts.

## [0.5.2] - 2026-03-01

### Added
- **Config Command**: New `config` command to manage server settings (`auth`, `port`, `browser`) from the CLI.
- **Systemd Service**: Added `service install/uninstall` commands to easily run vbrowser as a robust systemd user service.
- **Custom Login Page**: Replaced the browser's basic auth prompt with a sleek, password-only login page.
- **Cookie-Based Auth**: Implemented secure cookie-based authentication for a better user experience.
- **Enhanced Start Output**: The `start` command now displays the server URL and port when running in the background.

### Fixed
- **Daemon Detachment**: Fixed an issue where the background daemon would exit when the terminal session closed by adding proper process group detachment (`setsid`).

## [0.5.1] - 2026-03-01

### Added
- **Update Command**: New `update` command to check for the latest releases on GitHub and provide installation instructions.

## [0.5.0] - 2026-03-01

### Added
- **Daemon Mode**: The `start` command now runs vbrowser in the background by default.
- **File Logging**: Added support for logging to a file. When running as a daemon, logs are written to `~/.local/share/vbrowser/vbrowser.log` by default.
- **Foreground Flag**: Added `--foreground` (or `-f`) flag to the `start` command to run in the foreground and log to the console.
- **Log File Configuration**: Added support for `VBROWSER_LOG_FILE` environment variable and `logging.file` configuration in `config.json`.

### Changed
- **Start Behavior**: The `start` command no longer blocks the terminal unless the `--foreground` flag is used.

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

[0.7.1]: https://github.com/zulfikawr/vbrowser/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/zulfikawr/vbrowser/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/zulfikawr/vbrowser/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/zulfikawr/vbrowser/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/zulfikawr/vbrowser/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/zulfikawr/vbrowser/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/zulfikawr/vbrowser/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/zulfikawr/vbrowser/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/zulfikawr/vbrowser/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/zulfikawr/vbrowser/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/zulfikawr/vbrowser/releases/tag/v0.1.0
