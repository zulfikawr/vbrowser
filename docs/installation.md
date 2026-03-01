# Installation

There are two ways to install vbrowser: using the automated one-line installer or performing a manual installation.

## Requirements

### Server (Linux)
- **Go 1.21+** (for building from source)
- **GStreamer 1.0+**: Core library and plugins (base, good, bad, ugly)
- **X11 Libraries**: libx11, libxrandr, libxtst, libxfixes, libxcvt
- **PulseAudio**: pulseaudio and gstreamer1.0-pulseaudio
- **Xvfb & xdotool**: Virtual framebuffer and input automation
- **Disk Space**: ~100MB for the binary and browser profiles
- **RAM**: ~500MB (depending on the browser and workload)

### Client
- Any modern web browser with **WebRTC** support.
- SSH access to the server (if using a tunnel).

## Automated Installation (Recommended)

The easiest way to install vbrowser and all its system dependencies is via the one-line installer:

```bash
curl -sSL https://raw.githubusercontent.com/zulfikawr/vbrowser/main/scripts/install.sh | sudo bash
```

This script will:
1. Install necessary system packages (APT).
2. Install Chromium browser if no browser is found.
3. Download and build the vbrowser binary.
4. Set up the default configuration.

## Manual Installation

If you prefer to control each step, follow these instructions:

### 1. Install System Dependencies (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y xvfb xdotool pulseaudio 
    libx11-dev libxrandr-dev libxtst-dev libxfixes-dev libxcvt-dev 
    libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev 
    gstreamer1.0-plugins-base gstreamer1.0-plugins-good 
    gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly 
    gstreamer1.0-pulseaudio gstreamer1.0-tools
```

### 2. Install a Browser

Choose at least one of the following:

- **Chromium (Recommended for ARM64)**:
  ```bash
  sudo apt-get install -y chromium-browser
  ```
- **Google Chrome (x86_64 only)**:
  ```bash
  wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
  sudo apt install ./google-chrome-stable_current_amd64.deb
  ```
- **Firefox**:
  ```bash
  sudo apt-get install -y firefox
  ```

### 3. Clone and Build

```bash
git clone https://github.com/zulfikawr/vbrowser.git
cd vbrowser
make build
```

The resulting `vbrowser` binary will be in the root of the repository.
