# vbrowser

> Self-hosted virtual browser that streams Chromium via WebRTC

A single Go binary that launches a Chromium instance on a remote server and streams it to your local browser via WebRTC, accessible over SSH tunnel.

## Features

- 🚀 Single static binary - no runtime dependencies
- 🔄 Auto-downloads correct Chromium build for your OS
- 🎥 Real-time streaming via WebRTC (VP8)
- 🔒 Secure over SSH tunnel (no cloud dependency)
- 💾 Persistent browser profile (cookies, bookmarks, passwords)
- ⚙️ Configurable resolution, FPS, and bitrate
- 🔧 Simple CLI (start, stop, status)

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/zulfikawr/vbrowser.git
cd vbrowser

# Build
make build

# Or install directly
make install
```

### Usage

```bash
# On your remote server
vbrowser start

# In another terminal (or from local machine)
ssh -L 7070:localhost:7070 user@remote-server

# Open in your local browser
open http://localhost:7070
```

## Requirements

### Server (Linux)
- Go 1.21+ (for building)
- Xvfb: `sudo apt-get install xvfb`
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
    "target_fps": 30,
    "max_bitrate_kbps": 4000
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

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                        SERVER                           │
│                                                         │
│   ┌──────────┐           ┌───────────┐                 │
│   │ vbrowser │◄─────────►│ Chromium  │                 │
│   │  daemon  │    CDP    │ (Xvfb)    │                 │
│   │          │           └───────────┘                 │
│   │  ┌─────┐ │                                         │
│   │  │WebRTC│◄──────────┐                             │
│   │  └─────┘ │           │                             │
│   └──────────┘           │                             │
│        │                 │                             │
│   :7070 HTTP + WS        │                             │
└────────┼─────────────────┼─────────────────────────────┘
         │                 │
         │  SSH Tunnel     │ WebRTC Stream
         │                 │
┌────────┼─────────────────┼─────────────────────────────┐
│        │                 │      LOCAL MACHINE          │
│   localhost:7070         │                             │
│   ┌─────────────┐        │                             │
│   │   Browser   │◄───────┘                             │
│   └─────────────┘                                      │
└─────────────────────────────────────────────────────────┘
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
│   ├── capture/           # Screen capture (Xvfb)
│   ├── cmd/               # CLI commands
│   ├── config/            # Configuration
│   ├── platform/          # Platform-specific code
│   ├── process/           # PID file management
│   └── stream/            # WebRTC streaming
├── pkg/server/            # HTTP server & signaling
└── configs/               # Example configs
```

## Troubleshooting

### Xvfb not found
```bash
sudo apt-get install xvfb
```

### Port already in use
Change port in config or use `--port` flag:
```bash
vbrowser start --port 8080
```

### Chromium won't start
Check logs:
```bash
vbrowser start --foreground --log-level debug
```

### No video stream
1. Check WebSocket connection in browser console
2. Verify Chromium is running: `ps aux | grep chrome`
3. Check Xvfb is running: `ps aux | grep Xvfb`

## Performance Tuning

### Reduce CPU usage
Lower FPS or resolution in config:
```json
{
  "stream": {
    "target_fps": 15,
    "max_bitrate_kbps": 2000
  },
  "browser": {
    "window_width": 1280,
    "window_height": 720
  }
}
```

### Reduce latency
Increase bitrate and FPS:
```json
{
  "stream": {
    "target_fps": 60,
    "max_bitrate_kbps": 8000
  }
}
```

## Roadmap

- [x] Phase 1: Foundation (CLI, config, process management)
- [x] Phase 2: Streaming MVP (capture, WebRTC, UI)
- [ ] Phase 3: Input forwarding (mouse, keyboard)
- [ ] Phase 4: Audio streaming
- [ ] Phase 5: Multi-session support

See `tasks.md` for detailed roadmap.

## Contributing

Contributions welcome! Please read `CONTRIBUTING.md` first.

## License

MIT License - see `LICENSE` file for details.

## Acknowledgments

- [pion/webrtc](https://github.com/pion/webrtc) - Pure Go WebRTC implementation
- [chromedp](https://github.com/chromedp/chromedp) - Chrome DevTools Protocol
- [kbinani/screenshot](https://github.com/kbinani/screenshot) - Screen capture

## Author

Built by [@zulfikawr](https://github.com/zulfikawr)
