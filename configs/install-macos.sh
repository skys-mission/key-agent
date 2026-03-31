#!/bin/bash
# Key Agent macOS Installation Script
# Usage: ./install-macos.sh

set -e

USER=$(whoami)
BINDIR=/usr/local/bin
PLIST="$HOME/Library/LaunchAgents/com.skys-mission.key-agent.plist"
SOURCE_PLIST="$(dirname "$0")/../configs/com.skys-mission.key-agent.plist"

echo "=== Key Agent macOS Installation ==="
echo ""

# Check if running from correct directory
if [ ! -f "key-agent" ]; then
    echo "Error: key-agent binary not found in current directory"
    echo "Please run this script from the project root after building."
    exit 1
fi

# Install binary
echo "[1/4] Installing key-agent binary to $BINDIR..."
sudo cp key-agent "$BINDIR/key-agent"
sudo chmod +x "$BINDIR/key-agent"
echo "    Done."

# Install plist
echo "[2/4] Installing LaunchAgent plist..."
sed "s/YOUR_USERNAME/$USER/g" "$SOURCE_PLIST" > "$PLIST"
echo "    Done: $PLIST"

# Load service
echo "[3/4] Loading service..."
launchctl load "$PLIST"
echo "    Done."

# Start service
echo "[4/4] Starting service..."
launchctl start com.skys-mission.key-agent
echo "    Done."

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Service status:"
launchctl list | grep key-agent
echo ""
echo "Logs: tail -f /tmp/key-agent.log"
echo ""
