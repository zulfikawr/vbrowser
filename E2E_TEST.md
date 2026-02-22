# End-to-End Streaming Test

## Prerequisites
- Linux system with Xvfb installed: `sudo apt-get install xvfb`
- Go 1.21 or later

## Test Procedure

### 1. Build the binary
```bash
make build
```

### 2. Start vbrowser (with auto-download)
```bash
./vbrowser start --foreground --log-level debug
```

Expected output:
- "fetching latest Chromium revision" (first run only)
- "downloading Chromium" (first run only)
- "Xvfb started"
- "Chromium started"
- "HTTP server starting"
- "vbrowser started"

### 3. Open the viewer
In your local browser, navigate to:
```
http://localhost:7070
```

Expected behavior:
- Status shows "● Connecting..."
- WebSocket connects
- Status changes to "● Connected"
- Video stream appears showing Chromium window

### 4. Verify streaming
Check the following:
- [ ] Video stream is visible
- [ ] Video shows Chromium browser window
- [ ] Frame rate is smooth (30 FPS)
- [ ] No significant lag (<100ms)
- [ ] Status indicator shows "● Connected" (green)

### 5. Check logs
Server logs should show:
```
INFO WebRTC session created
INFO starting capture loop
INFO connection state changed state=connected
```

Browser console should show:
```
WebSocket connected
Received track: video
Connection state: connected
```

### 6. Test reconnection
1. Stop the server (Ctrl+C)
2. Observe status changes to "● Disconnected"
3. Restart the server
4. Observe automatic reconnection
5. Video stream resumes

### 7. Stop the server
```bash
./vbrowser stop
```

Expected:
- Graceful shutdown
- All processes terminated
- PID file removed

## Troubleshooting

### No video stream
- Check Xvfb is running: `ps aux | grep Xvfb`
- Check Chromium is running: `ps aux | grep chrome`
- Check WebSocket connection in browser console
- Verify display number matches config (default: :99)

### Connection fails
- Check firewall allows port 7070
- Verify server is listening: `netstat -tlnp | grep 7070`
- Check browser console for errors

### Black screen
- Chromium may still be loading
- Wait 5-10 seconds for initial render
- Check Chromium logs in profile directory

### High CPU usage
- Normal during streaming (VP8 encoding)
- Reduce FPS in config if needed
- Reduce resolution if needed

## Performance Metrics

Expected performance on modern hardware:
- CPU usage: 20-40% (single core)
- Memory: ~500MB (Chromium + vbrowser)
- Network: ~2-4 Mbps (depends on bitrate config)
- Latency: 50-100ms (localhost)

## Success Criteria

✅ All checks passed:
- [ ] Server starts without errors
- [ ] Chromium launches successfully
- [ ] WebSocket connection establishes
- [ ] WebRTC peer connection succeeds
- [ ] Video stream displays in browser
- [ ] Frame rate is acceptable
- [ ] Reconnection works
- [ ] Graceful shutdown works
