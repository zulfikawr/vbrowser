#!/bin/bash
set -e

echo "=== PulseAudio Audio Test ==="
echo

# Check if PulseAudio is running
if ! pgrep -x pulseaudio > /dev/null; then
    echo "❌ PulseAudio is not running"
    exit 1
fi
echo "✓ PulseAudio is running"

# List sinks
echo
echo "Available sinks:"
pactl list short sinks

# Check for vbrowser sink
SINK_NAME="vbrowser-99"
if pactl list short sinks | grep -q "$SINK_NAME"; then
    echo "✓ Found $SINK_NAME"
else
    echo "❌ $SINK_NAME not found"
    exit 1
fi

# Check monitor source
echo
echo "Available sources:"
pactl list short sources | grep monitor

if pactl list short sources | grep -q "${SINK_NAME}.monitor"; then
    echo "✓ Found ${SINK_NAME}.monitor"
else
    echo "❌ ${SINK_NAME}.monitor not found"
    exit 1
fi

# Test audio capture from monitor
echo
echo "Testing audio capture (5 seconds)..."
timeout 5 gst-launch-1.0 pulsesrc device="${SINK_NAME}.monitor" ! fakesink || true

echo
echo "=== Test complete ==="
