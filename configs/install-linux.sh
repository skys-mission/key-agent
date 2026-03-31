#!/bin/bash
# Key Agent Linux Installation Script
# Usage: ./install-linux.sh [--user]

set -e

USER=$(whoami)
BINDIR=/usr/local/bin
SERVICE_FILE="/etc/systemd/system/key-agent.service"
USER_SERVICE_FILE="$HOME/.config/systemd/user/key-agent.service"
USER_MODE=false

# Parse arguments
if [ "$1" == "--user" ]; then
    USER_MODE=true
fi

echo "=== Key Agent Linux Installation ==="
echo ""

# Check if running from correct directory
if [ ! -f "key-agent" ]; then
    echo "Error: key-agent binary not found in current directory"
    echo "Please run this script from the project root after building."
    exit 1
fi

if [ "$USER_MODE" = true ]; then
    echo "Installing as user service..."
    echo ""

    # Install binary
    echo "[1/3] Installing key-agent binary to $BINDIR..."
    sudo cp key-agent "$BINDIR/key-agent"
    sudo chmod +x "$BINDIR/key-agent"
    echo "    Done."

    # Create user service directory
    mkdir -p "$HOME/.config/systemd/user/"

    # Create service file
    echo "[2/3] Creating systemd user service..."
    cat > "$USER_SERVICE_FILE" << EOF
[Unit]
Description=Key Agent - Local Key-Value and Secrets Management

[Service]
Type=simple
ExecStart=$BINDIR/key-agent
Restart=on-failure

[Install]
WantedBy=default.target
EOF
    echo "    Done: $USER_SERVICE_FILE"

    # Enable and start
    echo "[3/3] Enabling and starting service..."
    systemctl --user daemon-reload
    systemctl --user enable key-agent
    systemctl --user start key-agent
    echo "    Done."

    echo ""
    echo "=== Installation Complete ==="
    echo ""
    echo "Service status:"
    systemctl --user status key-agent --no-pager || true
    echo ""
    echo "Logs: journalctl --user -u key-agent -f"
    echo ""

else
    echo "Installing as system service..."
    echo ""

    # Install binary
    echo "[1/3] Installing key-agent binary to $BINDIR..."
    sudo cp key-agent "$BINDIR/key-agent"
    sudo chmod +x "$BINDIR/key-agent"
    echo "    Done."

    # Create service file
    echo "[2/3] Creating systemd service..."
    sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=Key Agent - Local Key-Value and Secrets Management
After=network.target

[Service]
Type=simple
User=$USER
Group=$USER
ExecStart=$BINDIR/key-agent
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    echo "    Done: $SERVICE_FILE"

    # Enable and start
    echo "[3/3] Enabling and starting service..."
    sudo systemctl daemon-reload
    sudo systemctl enable key-agent
    sudo systemctl start key-agent
    echo "    Done."

    echo ""
    echo "=== Installation Complete ==="
    echo ""
    echo "Service status:"
    sudo systemctl status key-agent --no-pager || true
    echo ""
    echo "Logs: sudo journalctl -u key-agent -f"
    echo ""
fi
