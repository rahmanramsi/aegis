# Aegis systemd Deployment

## Prerequisites

- Linux with systemd (Ubuntu 20.04+, Debian 11+, Rocky 9+, etc.)
- Go toolchain installed (or pre-built binaries)
- A user account to run the service (e.g. `aegis`)

## Setup

### 1. Create a dedicated user (recommended)

```bash
sudo useradd -r -m -d /home/aegis -s /usr/sbin/nologin aegis
```

### 2. Create environment file

```bash
sudo mkdir -p /home/aegis/.aegis
sudo tee /home/aegis/.aegis/env <<EOF
AEGIS_GATEWAY_URL=ws://your-gateway:8080/ws/daemon
AEGIS_DAEMON_ID=agent-01
AEGIS_DAEMON_NAME=production-agent
AEGIS_DAEMON_TOKEN=your-secret-token
AEGIS_WORKSPACES_ROOT=/home/aegis/aegis/workspaces
AEGIS_MAX_CONCURRENT=5
AEGIS_AGENT_TIMEOUT=30m
AEGIS_LOG_LEVEL=info
EOF
sudo chmod 600 /home/aegis/.aegis/env
sudo chown -R aegis:aegis /home/aegis/.aegis
```

### 3. Place binaries

```bash
sudo mkdir -p /home/aegis/aegis
sudo cp aegis-agent aegisd /home/aegis/aegis/
sudo chown aegis:aegis /home/aegis/aegis/aegis-agent /home/aegis/aegis/aegisd
sudo chmod +x /home/aegis/aegis/aegis-agent /home/aegis/aegis/aegisd
```

### 4. Install systemd units

```bash
sudo cp deploy/aegis-agent.service /etc/systemd/system/aegis-agent@.service
sudo cp deploy/aegisd.service /etc/systemd/system/aegisd@.service
sudo systemctl daemon-reload
```

### 5. Start services

```bash
# Start the gateway
sudo systemctl enable --now aegisd@aegis.service

# Start the agent
sudo systemctl enable --now aegis-agent@aegis.service
```

### 6. Check status

```bash
sudo systemctl status aegisd@aegis.service
sudo systemctl status aegis-agent@aegis.service
sudo journalctl -u aegis-agent@aegis.service -f
```

## Environment Variables

| Variable              | Default                          | Description                  |
|-----------------------|----------------------------------|------------------------------|
| `AEGIS_GATEWAY_URL`   | `ws://localhost:8080/ws/daemon`  | Gateway WebSocket endpoint   |
| `AEGIS_DAEMON_ID`     | _(required)_                     | Unique daemon identifier     |
| `AEGIS_DAEMON_NAME`   | `aegis-agent`                    | Display name for this agent  |
| `AEGIS_DAEMON_TOKEN`  | _(required)_                     | Authentication token         |
| `AEGIS_WORKSPACES_ROOT`| `./workspaces`                  | Working directory root       |
| `AEGIS_MAX_CONCURRENT` | `5`                             | Max concurrent agent tasks   |
| `AEGIS_AGENT_TIMEOUT`  | `30m`                           | Per-task timeout             |
| `AEGIS_LOG_LEVEL`      | `info`                          | Log level: `info` or `debug` |

## Daemon Behavior

- **Auto-reconnect**: On disconnect, the daemon reconnects with exponential backoff (1s → 2s → 4s → 8s → 16s → 32s max). On each reconnect it sends a fresh handshake.
- **Keepalive**: Sends WebSocket pings every 30 seconds to detect dead connections.
- **Graceful shutdown**: Responds to SIGINT/SIGTERM by cancelling in-flight tasks and closing the WebSocket connection with a normal closure code.
- **systemd auto-restart**: `Restart=always` with a 5-second delay ensures the daemon survives crashes and upgrade cycles.
