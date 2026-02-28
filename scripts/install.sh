#!/bin/bash

# vbrowser one-line installer
# This script installs system dependencies and builds vbrowser.

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== vbrowser Installer ===${NC}"

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Please run as root or with sudo${NC}"
  exit 1
fi

# 1. Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    OS=$(uname -s)
fi

if [[ "$OS" != "ubuntu" && "$OS" != "debian" ]]; then
    echo -e "${RED}Warning: This installer is optimized for Ubuntu/Debian.${NC}"
    echo -e "Attempting to continue anyway..."
fi

# 2. Update and install system dependencies
echo -e "\n${BLUE}Step 1: Installing system dependencies...${NC}"
apt-get update
apt-get install -y xvfb xdotool pulseaudio \
    libx11-dev libxrandr-dev libxtst-dev libxfixes-dev libxcvt-dev \
    libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
    gstreamer1.0-plugins-base gstreamer1.0-plugins-good \
    gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
    gstreamer1.0-pulseaudio git make gcc

# 3. Check for Browser
echo -e "\n${BLUE}Step 2: Checking for a browser...${NC}"
if command -v chromium-browser &> /dev/null; then
    echo -e "${GREEN}✓ Chromium found${NC}"
elif command -v google-chrome &> /dev/null; then
    echo -e "${GREEN}✓ Google Chrome found${NC}"
elif command -v firefox &> /dev/null; then
    echo -e "${GREEN}✓ Firefox found${NC}"
else
    echo -e "No browser found. Installing Chromium..."
    apt-get install -y chromium-browser
fi

# 4. Check for Go
echo -e "\n${BLUE}Step 3: Checking for Go...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}✓ Go $GO_VERSION found${NC}"
else
    echo -e "Go not found. Installing Go 1.21+..."
    # Attempt to install Go via apt
    apt-get install -y golang-go || {
        echo -e "${RED}Failed to install Go via apt. Please install Go 1.21+ manually and re-run this script.${NC}"
        exit 1
    }
fi

# 5. Build vbrowser
echo -e "\n${BLUE}Step 4: Building vbrowser...${NC}"
if [ ! -f "Makefile" ]; then
    echo -e "Not in vbrowser directory. Cloning repository..."
    git clone https://github.com/zulfikawr/vbrowser.git
    cd vbrowser
fi

make build

# 6. Install binary
echo -e "\n${BLUE}Step 5: Installing vbrowser binary to /usr/local/bin...${NC}"
cp vbrowser /usr/local/bin/vbrowser
chmod +x /usr/local/bin/vbrowser

echo -e "\n${GREEN}=== Installation Complete! ===${NC}"
echo -e "You can now start vbrowser by running: ${BLUE}vbrowser start${NC}"
echo -e "For more info, run: ${BLUE}vbrowser --help${NC}"
