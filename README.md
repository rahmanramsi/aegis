# Aegis

> **Your agents, your harness, your rules.**
> One gateway. Infinite agents. Zero vendor lock.

Aegis lets you build AI agent bots that live in Telegram, Discord, and beyond — freely choosing which agent harness (Claude Code, Codex, OpenCode, Gemini, Copilot) powers each task. No API keys ever leave your machine.

## Quick Start

```bash
# 1. Build and start the gateway
go build -o aegisd ./cmd/aegisd
cp .env.example .env
# Edit .env with your Telegram token (optional)
./aegisd

# 2. Open the web UI
open http://localhost:8080

# 3. Register a daemon and create an agent
#    - In the UI: create a workspace → create a daemon (copy the token)
#    - On your machine: run aegis-agent with that token
#    - In the UI: create an agent, pick harness + model, connect to Telegram
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Gateway (aegisd)                     │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────────┐  │
│  │ REST API  │  │   WS Hub  │  │  Embedded Svelte SPA  │  │
│  │ /api/v1/* │  │ /ws/daemon│  │  /                    │  │
│  └────┬──────┘  └────┬──────┘  └──────────────────────┘  │
│       │               │                                   │
│  ┌────┴───────────────┴──────┐                            │
│  │     libSQL (Turso) DB     │                            │
│  └───────────────────────────┘                            │
│                                                           │
│  ┌──────────────────────────────────┐                     │
│  │   Message Adapters               │                     │
│  │   Telegram Bot  │  Discord       │                     │
│  └──────────────────────────────────┘                     │
└──────────────────────┬──────────────────────────────────┘
                       │ WebSocket
┌──────────────────────┴──────────────────────────────────┐
│                 Daemon (aegis-agent)                     │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Agent Harnesses (Claude Code, Codex, …)          │   │
│  │  - Holds API keys (Anthropic, OpenAI, etc.)       │   │
│  │  - Spawns CLI, streams output to gateway          │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

**Gateway** (`aegisd`) — the control plane. Web UI, message routing, task dispatch. Zero API keys.  
**Daemon** (`aegis-agent`) — the execution plane. Runs on your machine. Holds API keys. Spawns agent CLIs.

## Environment Variables

### Gateway (`aegisd`)

| Variable | Default | Description |
|---|---|---|
| `AEGIS_ADDR` | `:8080` | HTTP listen address |
| `AEGIS_BASE_URL` | `http://localhost:8080` | Public base URL (for webhook registration and CORS) |
| `AEGIS_DB_PATH` | `./data/gateway.db` | libSQL database file path |
| `AEGIS_API_KEY` | _(empty)_ | If set, protects POST/PUT/DELETE endpoints with Bearer token auth |
| `AEGIS_ENV` | `development` | `development` or `production` (production restricts CORS origin) |
| `AEGIS_TELEGRAM_TOKEN` | _(empty)_ | Telegram Bot API token from [@BotFather](https://t.me/BotFather) |
| `AEGIS_DISCORD_TOKEN` | _(empty)_ | Discord Bot token |
| `AEGIS_DISCORD_APP_ID` | _(empty)_ | Discord Application ID |

### Daemon (`aegis-agent`)

| Variable | Default | Description |
|---|---|---|
| `AEGIS_GATEWAY_URL` | `ws://localhost:8080/ws/daemon` | Gateway WebSocket endpoint |
| `AEGIS_DAEMON_ID` | _(empty)_ | Daemon UUID from gateway registration |
| `AEGIS_DAEMON_TOKEN` | _(empty)_ | Daemon enrollment token (shown on creation) |
| `AEGIS_DAEMON_NAME` | `dev-machine` | Human-readable daemon name |
| `ANTHROPIC_API_KEY` | _(empty)_ | API key for Claude harnesses |
| `OPENAI_API_KEY` | _(empty)_ | API key for Codex/OpenAI harnesses |
| `GOOGLE_API_KEY` | _(empty)_ | API key for Gemini harnesses |
| `AEGIS_CLAUDE_MODEL` | `claude-sonnet-4-20250514` | Default Claude model |

## Setting Up Telegram

1. Create a bot with [@BotFather](https://t.me/BotFather) — copy the token.
2. Set `AEGIS_TELEGRAM_TOKEN` in your gateway `.env`.
3. Start the gateway — it auto-registers webhooks.
4. Create an agent in the web UI and connect it to a Telegram chat ID.
5. Message the bot — your agent responds via the daemon.

## Registering a Daemon

1. In the Web UI: Workspaces → open a workspace → Daemons → Create.
2. Copy the **enrollment token** shown in the response.
3. On the daemon machine, create a `.env`:
   ```
   AEGIS_GATEWAY_URL=ws://<gateway-host>:8080/ws/daemon
   AEGIS_DAEMON_ID=<daemon-uuid>
   AEGIS_DAEMON_TOKEN=<enrollment-token>
   ANTHROPIC_API_KEY=sk-ant-...
   ```
4. Run `aegis-agent` — it connects to the gateway and reports available harnesses.
5. Back in the UI: create agents using this daemon.

## API Authentication

When `AEGIS_API_KEY` is set, all mutating endpoints (POST, PUT, DELETE) require a Bearer token. Read-only endpoints (GET) and the health check are public.

```bash
# Without auth:
curl -X POST http://localhost:8080/api/v1/workspaces -d '{"name":"test"}'  # → 401

# With auth:
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H "Authorization: Bearer your-key" \
  -H "Content-Type: application/json" \
  -d '{"name":"test","slug":"test"}'  # → 201
```

In the Web UI, click the key icon in the header to enter your API key. The UI stores it in `localStorage` and includes it in all requests.

## Tech Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.24+ |
| Frontend | Svelte 5 + SvelteKit (static adapter), TailwindCSS 4, shadcn-svelte |
| Database | Turso/libSQL (embedded) |
| Messaging | Telegram Bot API, Discord Gateway |
| Communication | WebSocket (gateway ↔ daemon), REST API (web UI) |

## Status

🚧 **Pre-alpha** — spec and implementation plan complete. Building.
