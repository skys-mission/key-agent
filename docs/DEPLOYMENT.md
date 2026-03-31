# Key Agent Deployment Guide

This guide covers deploying key-agent as a system service on different platforms.

## Table of Contents

- [macOS](#macos)
- [Linux](#linux)
- [Windows](#windows)
- [Docker](#docker)

---

## macOS

macOS uses `launchd` to manage background services.

### Installation

```bash
# 1. Copy binary to system path
sudo cp key-agent /usr/local/bin/
sudo chmod +x /usr/local/bin/key-agent

# 2. Copy LaunchAgent plist
cp configs/com.skys-mission.key-agent.plist ~/Library/LaunchAgents/

# 3. Edit the plist to set your username
sed -i '' "s/YOUR_USERNAME/$(whoami)/g" ~/Library/LaunchAgents/com.skys-mission.key-agent.plist

# 4. Load the service
launchctl load ~/Library/LaunchAgents/com.skys-mission.key-agent.plist
```

### Service Management

```bash
# Check status
launchctl list | grep key-agent

# Start service
launchctl start com.skys-mission.key-agent

# Stop service
launchctl stop com.skys-mission.key-agent

# Restart service
launchctl stop com.skys-mission.key-agent && launchctl start com.skys-mission.key-agent

# Unload (disable autostart)
launchctl unload ~/Library/LaunchAgents/com.skys-mission.key-agent.plist

# View logs
tail -f /tmp/key-agent.log
```

### LaunchAgent Configuration

The plist file is located at `configs/com.skys-mission.key-agent.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.skys-mission.key-agent</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/key-agent</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <key>StandardOutPath</key>
    <string>/tmp/key-agent.log</string>

    <key>StandardErrorPath</key>
    <string>/tmp/key-agent.log</string>
</dict>
</plist>
```

### Custom Configuration

To use a custom config file, modify the plist:

```xml
<key>ProgramArguments</key>
<array>
    <string>/usr/local/bin/key-agent</string>
    <string>--config</string>
    <string>/Users/yourname/.key-agent/config.yaml</string>
</array>
```

---

## Linux

Linux uses `systemd` to manage services.

### Installation

```bash
# 1. Copy binary
sudo cp key-agent /usr/local/bin/
sudo chmod +x /usr/local/bin/key-agent

# 2. Create systemd service file
sudo tee /etc/systemd/system/key-agent.service > /dev/null << 'EOF'
[Unit]
Description=Key Agent - Local Key-Value and Secrets Management
After=network.target

[Service]
Type=simple
User=%USER%
Group=%USER%
ExecStart=/usr/local/bin/key-agent
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 3. Replace %USER% with your username
sudo sed -i "s/%USER%/$(whoami)/g" /etc/systemd/system/key-agent.service

# 4. Enable and start
sudo systemctl daemon-reload
sudo systemctl enable key-agent
sudo systemctl start key-agent
```

### Service Management

```bash
# Check status
sudo systemctl status key-agent

# Start
sudo systemctl start key-agent

# Stop
sudo systemctl stop key-agent

# Restart
sudo systemctl restart key-agent

# View logs
journalctl -u key-agent -f
```

### User-level Service

For user-level installation (no sudo required for service management):

```bash
# Create user service directory
mkdir -p ~/.config/systemd/user/

# Create service file
cat > ~/.config/systemd/user/key-agent.service << 'EOF'
[Unit]
Description=Key Agent - Local Key-Value and Secrets Management

[Service]
Type=simple
ExecStart=/usr/local/bin/key-agent
Restart=on-failure

[Install]
WantedBy=default.target
EOF

# Enable and start
systemctl --user daemon-reload
systemctl --user enable key-agent
systemctl --user start key-agent
```

---

## Windows

Windows uses Windows Service or Task Scheduler.

### Option 1: Task Scheduler (Recommended)

1. Open Task Scheduler
2. Create a new task:
   - **General**: Name "Key Agent", Run only when user is logged on
   - **Triggers**: At log on
   - **Actions**: Start a program → `C:\Program Files\key-agent\key-agent.exe`
   - **Conditions**: Uncheck "Start the task only if the computer is on AC power"

### Option 2: Using NSSM (Non-Sucking Service Manager)

```powershell
# Download nssm from https://nssm.cc/download

# Install service
nssm install KeyAgent "C:\Program Files\key-agent\key-agent.exe"

# Start service
nssm start KeyAgent

# Stop service
nssm stop KeyAgent

# Remove service
nssm remove KeyAgent confirm
```

---

## Docker

### Environment Variables

Key Agent supports the following environment variables for Docker deployments:

| Variable | Description |
|----------|-------------|
| `KEY_AGENT_MASTER_KEY_BACKEND` | Set to `file` for containers (required) |
| `KEY_AGENT_MASTER_KEY` | Base64-encoded 32-byte master key (recommended) |
| `KEY_AGENT_PASSPHRASE` | Passphrase for encrypted master key file |

### Master Key Options

**Option 1: Direct Master Key Injection (Recommended)**

```bash
# Generate a master key
openssl rand -base64 32

# Run with the master key
docker run -d \
  --name key-agent \
  -p 127.0.0.1:8080:8080 \
  -v key-agent-data:/data \
  -e KEY_AGENT_MASTER_KEY_BACKEND=file \
  -e KEY_AGENT_MASTER_KEY=<your-base64-encoded-key> \
  key-agent:latest
```

**Option 2: Passphrase-Based Encryption**

```bash
docker run -d \
  --name key-agent \
  -p 127.0.0.1:8080:8080 \
  -v key-agent-data:/data \
  -e KEY_AGENT_MASTER_KEY_BACKEND=file \
  -e KEY_AGENT_PASSPHRASE=your-secure-passphrase \
  key-agent:latest
```

### Build and Run

```bash
# Build
docker build -t key-agent:latest .

# Run (with master key)
docker run -d \
  --name key-agent \
  -p 127.0.0.1:8080:8080 \
  -v key-agent-data:/data \
  -e KEY_AGENT_MASTER_KEY_BACKEND=file \
  -e KEY_AGENT_MASTER_KEY=$(openssl rand -base64 32) \
  key-agent:latest

# View logs (contains root token on first run)
docker logs -f key-agent
```

### Docker Compose

```yaml
version: '3.8'
services:
  key-agent:
    image: key-agent:latest
    container_name: key-agent
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - key-agent-data:/data
    environment:
      - KEY_AGENT_MASTER_KEY_BACKEND=file
      - KEY_AGENT_MASTER_KEY=${KEY_AGENT_MASTER_KEY}
      # Or use passphrase:
      # - KEY_AGENT_PASSPHRASE=your-secure-passphrase
    restart: unless-stopped

volumes:
  key-agent-data:
```

### Data Persistence

| Scenario | Data Retained |
|----------|---------------|
| `docker compose restart` | ✅ Yes |
| `docker compose down && up` | ✅ Yes (named volume) |
| `docker compose down -v` | ❌ No (volume deleted) |

**Important**: If using `KEY_AGENT_MASTER_KEY`, you must provide the same key when restarting. If using `KEY_AGENT_PASSPHRASE`, the master key file is stored in the volume.

---

## Verification

After deployment, verify the service is running:

```bash
# Check health endpoint
curl http://127.0.0.1:8080/health

# Expected response
{"status":"healthy","version":"1.0.0"}
```

## Troubleshooting

### Service won't start

```bash
# Check if port is in use
lsof -i :8080    # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Check logs
# macOS: /tmp/key-agent.log
# Linux: journalctl -u key-agent
# Windows: Event Viewer
```

### Permission denied

```bash
# Ensure binary is executable
chmod +x /usr/local/bin/key-agent

# Ensure data directory is writable
chmod 700 ~/.key-agent
```

### Master key issues

```bash
# Set passphrase for file backend
export KEY_AGENT_PASSPHRASE="your-passphrase"

# Or specify backend
key-agent --master-key-backend file
```
