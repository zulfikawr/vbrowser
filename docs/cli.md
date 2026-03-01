# CLI Commands

vbrowser provides a rich set of CLI commands for management.

## `start`
Start the vbrowser daemon.
```bash
vbrowser start [flags]
```
**Flags:**
- `-f, --foreground`: Run in the current terminal session (don't daemonize).
- `--port`: Override the configured server port.

## `stop`
Stop the running vbrowser daemon and all associated processes (Xvfb, Browser, etc.).
```bash
vbrowser stop
```

## `status`
Display whether vbrowser is running and show its current PID, URL, and browser resolution.
```bash
vbrowser status
```

## `config`
Manage settings directly from the terminal.
```bash
# Toggle authentication
vbrowser config auth on <password>
vbrowser config auth off

# Set server port
vbrowser config port 8080

# Set preferred browser (chrome, chromium, or firefox)
vbrowser config browser chromium
```

## `service`
Manage the systemd user service.
```bash
# Install and start the service
vbrowser service install

# Stop and remove the service
vbrowser service uninstall
```

## `list`
List all detected browser binaries on your system.
```bash
vbrowser list
```

## `logs`
Stream live service logs from the systemd journal.
```bash
vbrowser logs
```

## `update`
Check GitHub for the latest vbrowser release and get update instructions.
```bash
vbrowser update
```

## `version`
Print version and build information.
```bash
vbrowser version
```
