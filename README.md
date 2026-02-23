# vbrowser

> Self-hosted virtual browser that streams Chromium via WebRTC

A single Go binary that launches a Chromium instance on a remote server and streams it to your local browser via WebRTC, accessible over SSH tunnel.

## Features

- 🚀 Single static binary - high-performance Go + GStreamer
- ⚡ Buttery smooth 60 FPS - ultra-low latency streaming
- 🔄 Auto-downloads correct Chromium build for your OS
- 🎥 Native GStreamer VP8 encoding (Mimicking Neko architecture)
- 🔊 Full audio support via PulseAudio + Opus codec
- 🔒 Secure over SSH tunnel (no cloud dependency)
- 🖱️ Full interaction support (Mouse, Keyboard, Scroll, Shortcuts)
- 💾 Persistent browser profile (cookies, bookmarks, passwords)
- ⚙️ Hot-reloadable resolution, FPS, and bitrate
- 🔧 Simple CLI (start, stop, status)

## Quick Start

### Installation

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install -y xvfb xdotool pulseaudio libgstreamer1.0-dev \
    gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
    gstreamer1.0-pulseaudio

# Clone the repository
git clone https://github.com/zulfikawr/vbrowser.git
cd vbrowser

# Build
make build
```

### Usage

```bash
# On your remote server
./vbrowser start

# In another terminal (or from local machine)
ssh -L 7070:localhost:7070 user@remote-server

# Open in your local browser
# http://localhost:7070
```

## Requirements

### Server (Linux)
- Go 1.21+ (for building)
- GStreamer 1.0+: `libgstreamer1.0-dev` and plugins (base, good, bad, ugly)
- PulseAudio: `pulseaudio` and `gstreamer1.0-pulseaudio`
- Xvfb & xdotool: `sudo apt-get install xvfb xdotool`
- ~500MB disk space for Chromium
- ~500MB RAM

### Client
- Any modern browser with WebRTC support
- SSH access to server

## Configuration

Default config location: `~/.config/vbrowser/config.json`

```json
{
  "server": {
    "host": "127.0.0.1",
    "port": 7070
  },
  "browser": {
    "auto_download": true,
    "window_width": 1920,
    "window_height": 1080
  },
  "stream": {
    "video_codec": "vp8",
    "target_fps": 60,
    "max_bitrate_kbps": 8000
  }
}
```

See `configs/default.json` for all options.

## CLI Commands

### start
Start the vbrowser daemon:
```bash
vbrowser start [flags]

Flags:
  -f, --foreground     Run in foreground (don't daemonize)
      --port int       Override server port
      --no-download    Skip Chromium auto-download
```

### stop
Stop the running daemon:
```bash
vbrowser stop
```

### status
Show daemon status:
```bash
vbrowser status
```

Output:
```
● vbrowser is running
  PID:        12345
  URL:        http://127.0.0.1:7070
  Chromium:   /home/user/.local/share/vbrowser/chromium/chrome
  Display:    :99 (1920x1080)
```

### version
Print version info:
```bash
vbrowser version
```

## Development

### Build
```bash
make build
```

### Test
```bash
make test
```

### Lint
```bash
make lint
```

### Cross-compile
```bash
GOOS=linux GOARCH=amd64 go build -o dist/vbrowser-linux-amd64 ./cmd/vbrowser
```

## Project Structure

```
vbrowser/
├── cmd/vbrowser/          # Main entry point
├── internal/
│   ├── browser/           # Chromium management & download
│   ├── capture/           # Legacy capture (keeping for reference)
│   ├── cmd/               # CLI commands
│   ├── config/            # Configuration
│   ├── platform/          # Platform-specific code
│   ├── process/           # PID file management
│   └── stream/            # WebRTC streaming
├── pkg/
│   ├── gst/               # Native GStreamer CGo bindings
│   └── server/            # HTTP server & signaling
└── configs/               # Example configs
```

## Troubleshooting

### Missing Dependencies
```bash
sudo apt-get install xvfb xdotool pulseaudio libgstreamer1.0-dev gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly gstreamer1.0-pulseaudio
```

### Port already in use
Change port in config or use `--port` flag:
```bash
./vbrowser start --port 8080
```

### Chromium won't start
Check logs:
```bash
./vbrowser start --foreground --log-level debug
```

### No video stream
1. Check WebSocket connection in browser console
2. Verify Chromium is running: `ps aux | grep chrome`
3. Check Xvfb is running: `ps aux | grep Xvfb`
4. Check GStreamer pipeline is active in logs.

## Contributing

Contributions welcome! Please read `CONTRIBUTING.md` first.

## License

MIT License - see `LICENSE` file for details.

## Acknowledgments

- [m1k1o/neko](https://github.com/m1k1o/neko) - Architecture inspiration for low-latency GStreamer streaming
- [pion/webrtc](https://github.com/pion/webrtc) - Pure Go WebRTC implementation
- [chromedp](https://github.com/chromedp/chromedp) - Chrome DevTools Protocol
- [GStreamer](https://gstreamer.freedesktop.org/) - Multimedia framework

## Author

Built by [@zulfikawr](https://github.com/zulfikawr)
