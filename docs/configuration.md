# Configuration

vbrowser uses a JSON configuration file for all its settings.

## Config File Location

Default path: `~/.config/vbrowser/config.json`

You can specify a custom config file using the `--config` flag:
```bash
./vbrowser start --config /path/to/config.json
```

## Structure

```json
{
  "server": {
    "host": "127.0.0.1",
    "port": 7070,
    "auth": {
      "enabled": false,
      "token": "yourpassword"
    }
  },
  "browser": {
    "browser_path": "/usr/bin/chromium",
    "profile_dir": "~/.local/share/vbrowser/profile",
    "window_width": 1280,
    "window_height": 720,
    "extra_args": []
  },
  "display": {
    "virtual_display": true,
    "display_num": 99,
    "depth": 24
  },
  "stream": {
    "video_codec": "vp8",
    "encoder": "software",
    "target_fps": 60,
    "max_bitrate_kbps": 4000
  },
  "logging": {
    "level": "info",
    "file": "~/.local/share/vbrowser/vbrowser.log"
  }
}
```

## Settings Reference

### Server Settings
- `server.host`: The IP address to listen on. Use `0.0.0.0` to listen on all interfaces.
- `server.port`: The HTTP port for the UI and signaling.
- `server.auth.enabled`: Whether to enable the login page.
- `server.auth.token`: The password required if authentication is enabled.

### Browser Settings
- `browser.browser_path`: Path to the browser binary (Chromium, Chrome, or Firefox).
- `browser.profile_dir`: Directory where cookies and settings are saved.
- `browser.window_width`/`height`: The resolution the browser will start with.
- `browser.extra_args`: Additional CLI flags passed directly to the browser.

### Stream Settings
- `stream.video_codec`: Either `vp8` or `h264`.
- `stream.encoder`: Encoding method:
    - `software`: Standard CPU encoding (Default).
    - `vaapi`: Hardware acceleration for Intel/AMD.
    - `nvenc`: Hardware acceleration for Nvidia.
- `stream.target_fps`: Target frames per second (1-60).
- `stream.max_bitrate_kbps`: Maximum bitrate allowed for the stream.

## Environment Variables

You can override certain settings using environment variables:

- `VBROWSER_PORT`: Override `server.port`.
- `VBROWSER_LOG_LEVEL`: Override `logging.level`.
- `VBROWSER_LOG_FILE`: Override `logging.file`.
- `VBROWSER_BROWSER_PATH`: Override `browser.browser_path`.
- `VBROWSER_ENCODER`: Override `stream.encoder`.
