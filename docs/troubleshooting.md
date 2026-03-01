# Troubleshooting

If you encounter issues with vbrowser, follow these steps.

## Common Issues

### "Disconnected" message in UI
1. **Check if the binary is running**: Run `vbrowser status`.
2. **Check the logs**: Run `vbrowser logs` to see if there are GStreamer or WebSocket errors.
3. **Verify Port**: Ensure the port (default 7070) is open and not blocked by a firewall.
4. **SSH Tunnel**: If accessing remotely, ensure your SSH tunnel is active: `ssh -L 7070:localhost:7070 user@remote`.

### No Video or Frozen Screen
1. **Check Xvfb**: Ensure Xvfb is running (`ps aux | grep Xvfb`).
2. **Check Browser**: Ensure the browser process is running.
3. **Restart Service**: Try `vbrowser stop && vbrowser start`.

### Can't Type or Click
1. **Focus**: Click once anywhere on the video area to ensure the input overlay has focus.
2. **Browser Flags**: If the browser feels sluggish, ensure `vbrowser logs` doesn't show "GStreamer pipeline errors".

### Port Already in Use
If you see `bind: address already in use`, either stop the existing process or change the port:
```bash
./vbrowser config port 8080
./vbrowser start
```

## Debugging

To get the most detailed information, run vbrowser in the foreground with debug logging:

```bash
./vbrowser start --foreground --log-level debug
```

This will output every WebRTC event, GStreamer sample, and input event to your terminal.

## Reporting Bugs

If you find a persistent bug, please open an issue on the [GitHub repository](https://github.com/zulfikawr/vbrowser/issues) with:
1. Your OS and architecture (e.g., Ubuntu 24.04 ARM64).
2. The browser you are using (e.g., Chromium).
3. The output of `./vbrowser logs`.
