# Development Guide

This guide is for contributors and developers who want to modify vbrowser.

## Build from Source

Requirements: Go 1.21+ and GStreamer headers.

```bash
# Build the binary
make build

# Run tests
make test

# Run linter
make lint
```

## Project Structure

- `cmd/vbrowser/`: Main entry point and CLI command definitions.
- `internal/browser/`: Browser lifecycle management (Chrome/Firefox).
- `internal/cmd/`: Implementation of CLI commands.
- `internal/config/`: Configuration parsing and persistence.
- `internal/platform/`: OS-specific code (X11, etc.).
- `internal/process/`: Process management and PID tracking.
- `internal/stream/`: WebRTC session and GStreamer integration.
- `pkg/gst/`: CGo bindings for GStreamer.
- `pkg/server/`: HTTP server and WebSocket signaling.
- `pkg/server/ui/`: Embedded frontend assets.

## GStreamer Bindings

vbrowser uses CGo to interface with GStreamer. The code is located in `pkg/gst/`. If you modify the GStreamer pipelines or properties, you may need to update the C code in `gst.c` and `gst.h`.

## WebSocket Signaling

The signaling protocol uses JSON over WebSockets. The main message loop is in `pkg/server/signaling.go`.

**Message Types:**
- `offer`: SDP offer from client.
- `answer`: SDP answer from server.
- `candidate`: ICE candidate exchange.
- `input`: Mouse and keyboard events.
- `config`: Dynamic configuration updates.
- `pli`: Request for a new keyframe.
- `clipboard`: Clipboard synchronization.

## UI Development

The UI is a single-page application built with vanilla JavaScript.
- `index.html`: Main structure and CSS.
- `client.js`: WebRTC client and input handling.
- `guacamole-keyboard.js`: External library for keyboard event mapping.
