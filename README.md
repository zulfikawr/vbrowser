# vbrowser

> Self-hosted virtual browser that streams Chrome, Chromium, or Firefox via WebRTC

vbrowser is a high-performance, single-binary tool that launches a browser on a remote server and streams it directly to your local browser using WebRTC and GStreamer. It's designed for low-latency interactions, secure browsing, and lightweight remote access.

## Features

- 🚀 **Single Binary**: High-performance Go + GStreamer backend.
- ⚡ **Buttery Smooth**: Ultra-low latency 60 FPS streaming.
- 🌐 **Native Browser**: Uses system-installed Chrome, Chromium, or Firefox.
- 🎥 **Flexible Encoding**: Native GStreamer VP8/H.264 with Hardware Acceleration support.
- 🔊 **Audio Support**: Full audio via PulseAudio + Opus codec.
- 🔒 **Secure**: Runs over SSH tunnel or Cloudflare, with optional password protection.
- 🖱️ **Full Interaction**: Mouse, Keyboard, Scroll, and Shortcut support.
- 💾 **Persistence**: Preserves browser profiles (cookies, bookmarks, passwords).
- 📋 **Auto Clipboard**: Seamless bi-directional clipboard synchronization.

## Quick Start

### 1. Install
```bash
curl -sSL https://raw.githubusercontent.com/zulfikawr/vbrowser/main/scripts/install.sh | sudo bash
```

### 2. Start
```bash
./vbrowser start
```

### 3. Access
Open `http://localhost:7070` in your web browser.

## Documentation

For detailed information, please refer to the following guides:

- [**Installation**](docs/installation.md): System requirements and manual setup.
- [**Usage**](docs/usage.md): Starting, stopping, and accessing vbrowser.
- [**CLI Commands**](docs/cli.md): Full command reference (`list`, `logs`, `config`, etc.).
- [**Configuration**](docs/configuration.md): Detailed explanation of all settings and environment variables.
- [**Troubleshooting**](docs/troubleshooting.md): Solutions to common issues.
- [**Development**](docs/development.md): Contributing and project structure.

## Acknowledgments

- [m1k1o/neko](https://github.com/m1k1o/neko) - Architecture inspiration for low-latency streaming.
- [pion/webrtc](https://github.com/pion/webrtc) - Pure Go WebRTC implementation.
- [GStreamer](https://gstreamer.freedesktop.org/) - Multimedia framework.

## License

MIT License - see [LICENSE](LICENSE) for details.

---
Built by [@zulfikawr](https://github.com/zulfikawr)
