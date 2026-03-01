# Usage

This guide covers how to start and access your virtual browser.

## Starting the Server

The simplest way to start vbrowser is by using the `start` command:

```bash
./vbrowser start
```

By default, vbrowser will daemonize (run in the background). You can see the logs at `~/.local/share/vbrowser/vbrowser.log` or by running `vbrowser logs`.

### Background vs Foreground

- **Daemon Mode (Default)**: Returns immediately to the terminal.
  ```bash
  ./vbrowser start
  ```
- **Foreground Mode**: Useful for debugging.
  ```bash
  ./vbrowser start --foreground
  ```

## Accessing the Browser

By default, vbrowser listens on `127.0.0.1:7070`.

### Local Access (on the same machine)
Simply open your web browser and navigate to:
`http://localhost:7070`

### Remote Access (via SSH Tunnel)
If vbrowser is running on a remote server, use an SSH tunnel to access it securely:

```bash
ssh -L 7070:localhost:7070 user@remote-server
```

Then open `http://localhost:7070` on your local machine.

### Remote Access (via Cloudflare Tunnel)
If you have `cloudflared` installed, you can expose vbrowser to a custom domain:

1. Create a tunnel: `cloudflared tunnel create vbrowser`
2. Route a domain: `cloudflared tunnel route dns vbrowser browser.yourdomain.com`
3. Run the tunnel pointing to port 7070.

## Running as a Service (Recommended)

To ensure vbrowser starts automatically on boot and restarts if it crashes, install it as a systemd user service:

```bash
./vbrowser service install
```

Manage the service using standard systemctl commands:
```bash
systemctl --user status vbrowser.service
systemctl --user stop vbrowser.service
systemctl --user start vbrowser.service
```

## Stopping the Server

To stop the background daemon and clean up all associated processes (Xvfb, Browser, PulseAudio):

```bash
./vbrowser stop
```
