# Aegis — Design Spec (v2: Gateway + Daemon)

**Date**: 2026-06-23
**Status**: Draft

## Overview

`Aegis` is a managed multi-agent bot platform with two components:

1. **Gateway** (`aegisd`) — server/cloud binary. Web UI, message routing, task dispatch, workspace/agent management. Holds NO API keys.
2. **Daemon** (`aegis-agent`) — runs on user machines. Holds API keys, auto-detects agent CLIs from `$PATH`, executes tasks, streams output back to gateway.

This mirrors multica's architecture: the gateway is the control plane; the daemon is the execution plane. Users interact via Telegram/Discord, freely choose which agent to use per task, and the daemon spawns the appropriate CLI.

```
┌──────────────────────────────────────────────────┐
│  Gateway (server)                                │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ Telegram │  │ Discord  │  │  Web UI       │  │
│  │ Adapter  │  │ Adapter  │  │  (Svelte SPA) │  │
│  └────┬─────┘  └────┬─────┘  └───────┬───────┘  │
│       └──────┬──────┘                │          │
│              ▼                       ▼          │
│  ┌───────────────────┐  ┌────────────────────┐  │
│  │  Message Router   │  │  REST API          │  │
│  └────────┬──────────┘  └────────────────────┘  │
│           ▼                                      │
│  ┌────────────────────────────────────────┐     │
│  │  Task Dispatcher (WebSocket to daemon) │     │
│  └────────────────┬───────────────────────┘     │
│                   │                              │
│  ┌────────────────┴───────────────────────┐     │
│  │  Turso/libSQL (embedded)               │     │
│  │  workspaces, agents, daemons, sessions │     │
│  └────────────────────────────────────────┘     │
└──────────────────┬───────────────────────────────┘
                   │ WebSocket (persistent)
        ┌──────────┴──────────┐
        ▼                     ▼
┌───────────────┐    ┌───────────────┐
│  Daemon "dev" │    │  Daemon "ops" │
│  API keys:    │    │  API keys:    │
│  ANTHROPIC    │    │  ANTHROPIC    │
│  OPENAI       │    │  GOOGLE       │
│               │    │               │
│  Harnesses:   │    │  Harnesses:   │
│  claude ✓     │    │  claude ✓     │
│  codex ✓      │    │  gemini ✓     │
│  opencode ✓   │    │               │
└───────────────┘    └───────────────┘
```

---

## Architecture

### Two Binaries

| Binary | Role | Location |
|---|---|---|
| `aegisd` | Gateway server — HTTP, messaging, WebSocket hub, Web UI | Cloud / VPS |
| `aegis-agent` | Daemon — CLI detection, task execution, API keys | User machines |

### Communication Flow

```
1. Daemon starts → connects to gateway via WebSocket
2. Daemon sends handshake: { daemon_id, name, harnesses: ["claude", "codex"] }
3. Gateway registers daemon + available harnesses in DB
4. User sends message in Telegram: "/codex fix this bug"
5. Gateway: resolve workspace → connection → agent → daemon
6. Gateway → Daemon (WS): { task_id, harness: "codex", prompt: "...", model: "..." }
7. Daemon: spawn `codex exec "fix this bug"`, stream stdout/stderr back via WS
8. Gateway: relay stream to user Telegram chat
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Gateway backend | Go 1.24+ |
| Daemon | Go 1.24+ |
| Frontend | Svelte 5 + SvelteKit (static adapter), TailwindCSS 4, shadcn-svelte |
| Database (gateway) | libSQL (Turso embedded) via `github.com/tursodatabase/go-libsql` |
| Telegram | `github.com/go-telegram/bot` |
| Discord | `github.com/disgoorg/disgo` |
| HTTP router | `net/http` (Go 1.22+ enhanced ServeMux) |
| WebSocket | `github.com/coder/websocket` |
| Config | Environment variables (`.env` file) |
| Logging | `log/slog` (stdlib structured logging) |

---

## Database Schema (Gateway)

```sql
-- Workspace: organizational boundary
CREATE TABLE workspaces (
    id          TEXT PRIMARY KEY,          -- UUID
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    created_at  TEXT NOT NULL
);

-- Registered daemon instances (scoped to a workspace)
CREATE TABLE daemons (
    id            TEXT PRIMARY KEY,          -- UUID
    workspace_id  TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,             -- user-defined name e.g. "dev-machine"
    token_hash    TEXT NOT NULL,             -- SHA-256 of daemon enrollment token
    status        TEXT NOT NULL DEFAULT 'offline',
    last_seen     TEXT,
    created_at    TEXT NOT NULL,
    UNIQUE(workspace_id, name)
);

-- Harnesses reported by a daemon
CREATE TABLE daemon_harnesses (
    daemon_id   TEXT NOT NULL REFERENCES daemons(id) ON DELETE CASCADE,
    harness     TEXT NOT NULL,             -- "claude", "codex", "opencode", ...
    PRIMARY KEY (daemon_id, harness)
);

-- Agent: a bot persona backed by a daemon harness
CREATE TABLE agents (
    id            TEXT PRIMARY KEY,        -- UUID
    workspace_id  TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    daemon_id     TEXT NOT NULL REFERENCES daemons(id),
    name          TEXT NOT NULL,            -- display name e.g. "Code Reviewer"
    harness       TEXT NOT NULL,            -- "claude", "codex", "gemini", ...
    model         TEXT DEFAULT '',           -- model override
    extra_args    TEXT DEFAULT '',           -- extra CLI args
    enabled       INTEGER DEFAULT 1,
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);

-- Connection: maps an agent to a messaging platform chat
CREATE TABLE connections (
    id          TEXT PRIMARY KEY,          -- UUID
    agent_id    TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    platform    TEXT NOT NULL,             -- "telegram", "discord"
    chat_id     TEXT NOT NULL,             -- Telegram chat ID / Discord channel ID
    created_at  TEXT NOT NULL,
    UNIQUE (agent_id, platform, chat_id)
);

-- Sessions: conversation state per connection
CREATE TABLE sessions (
    id            TEXT PRIMARY KEY,        -- UUID
    connection_id TEXT NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    user_name     TEXT DEFAULT '',          -- display name of the chat user
    created_at    TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);

-- Message history
CREATE TABLE messages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role        TEXT NOT NULL,             -- "user", "assistant"
    content     TEXT NOT NULL,
    agent_id    TEXT,
    workspaces_id TEXT,
    created_at  TEXT NOT NULL
);
```

---

## Daemon

### Lifecycle

```
1. Start → load config (API keys + enrollment token from env vars)
2. Auto-detect harnesses from $PATH
3. Connect to gateway via WebSocket
4. Send handshake: { token, name, harnesses: [...] }
5. Gateway verifies token hash for pre-registered daemon → marks online, updates harnesses
6. Idle loop: wait for tasks from gateway
7. On task: spawn harness CLI, stream output back via WS
8. On disconnect: gateway marks daemon offline
```

```json
{
  "type": "handshake",
  "daemon_id": "uuid-from-web-ui",
2. Admin copies daemon ID + token to daemon machine: `export AEGIS_DAEMON_ID=<uuid>; export AEGIS_DAEMON_TOKEN=aegis_dmt_...`
3. Daemon connects via WebSocket, sends `{ type: "handshake", daemon_id, token }`
4. Gateway looks up daemon by id → verifies `sha256(token)` matches stored hash
5. Match → gateway updates status to `online`, refreshes harnesses
6. Unknown id or wrong token → gateway rejects with `handshake_error`, closes connection

Unauthenticated WS connections are closed immediately. There is no self-registration path.

### Handshake Message (Daemon → Gateway)

```json
{
  "type": "handshake",
  "daemon_id": "uuid-from-web-ui",
  "token": "aegis_dmt_xxxx",
  "harnesses": ["claude", "codex", "opencode"]
}

### Task Message (Gateway → Daemon)

```json
{
  "type": "task",
  "task_id": "uuid",
  "harness": "codex",
  "prompt": "fix the login bug",
  "model": "gpt-5.1-codex",
  "extra_args": ["--sandbox", "read-only"]
}
```

### Stream Events (Daemon → Gateway)

```json
{"type": "stdout", "task_id": "uuid", "content": "Analyzing..."}
{"type": "stderr", "task_id": "uuid", "content": "Warning: ..."}
{"type": "done",  "task_id": "uuid"}
{"type": "error", "task_id": "uuid", "content": "timeout"}
```

### Config (Daemon)

```env
# Gateway connection
AEGIS_GATEWAY_URL=ws://localhost:8080/ws/daemon
AEGIS_DAEMON_TOKEN=aegis_dmt_...   # enrollment token (required)
AEGIS_DAEMON_NAME=dev-machine

# API keys (never leave this machine)
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=...

# Agent path overrides (optional)
AEGIS_CLAUDE_PATH=/custom/path/claude
AEGIS_CODEX_PATH=
AEGIS_OPENCODE_PATH=
AEGIS_COPILOT_PATH=
AEGIS_GEMINI_PATH=

# Agent model defaults (optional)
AEGIS_CLAUDE_MODEL=claude-sonnet-4-20250514
AEGIS_CODEX_MODEL=gpt-5.1-codex
AEGIS_GEMINI_MODEL=gemini-2.5-pro

# Workspace isolation
AEGIS_WORKSPACES_ROOT=./workspaces

# Limits
AEGIS_MAX_CONCURRENT=5
AEGIS_AGENT_TIMEOUT=30m
```

---

## Harness Interface (Daemon)

```go
// Runner is implemented by each agent harness.
type Runner interface {
    // Name returns the harness identifier (e.g. "claude", "codex").
    Name() string

    // Available reports whether the CLI binary is on PATH.
    Available() bool

    // Run executes the agent and streams events.
    Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error)
}

type RunRequest struct {
    TaskID    string
    Prompt    string
    WorkDir   string
    Model     string
    ExtraArgs []string
}

type StreamEvent struct {
    Type    EventType // Stdout, Stderr, Done, Error
    Content string
}
```

### Agent CLI Invocation

| Harness | Command |
|---|---|
| `claude` | `claude -p "<prompt>" --model <model>` |
| `codex` | `codex exec "<prompt>" --model <model>` |
| `opencode` | `opencode run "<prompt>"` |
| `copilot` | `copilot suggest "<prompt>"` |
| `gemini` | `gemini chat "<prompt>"` |
| `hermes` | `hermes chat --model <model> "<prompt>"` |
| `openclaw` | `openclaw run "<prompt>"` |

### Binary discovery order (each harness)
1. `AEGIS_<NAME>_PATH` env var
2. `exec.LookPath("<name>")`
3. Report unavailable if not found

---

## REST API (Gateway)

Base path: `/api/v1`

### Workspaces

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/workspaces` | List workspaces |
| `POST` | `/api/v1/workspaces` | Create workspace |
| `GET` | `/api/v1/workspaces/:id` | Get workspace |
| `PUT` | `/api/v1/workspaces/:id` | Update workspace |
| `DELETE` | `/api/v1/workspaces/:id` | Delete workspace |

### Daemons

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/workspaces/:wid/daemons` | List daemons in workspace |
| `POST` | `/api/v1/workspaces/:wid/daemons` | Create daemon (generates enrollment token) |
| `GET` | `/api/v1/daemons/:id` | Get daemon detail |
| `DELETE` | `/api/v1/daemons/:id` | Remove daemon |
| `GET` | `/api/v1/workspaces/:wid/agents` | List agents in workspace |
| `POST` | `/api/v1/workspaces/:wid/agents` | Create agent |
| `GET` | `/api/v1/agents/:id` | Get agent detail |
| `PUT` | `/api/v1/agents/:id` | Update agent |
| `DELETE` | `/api/v1/agents/:id` | Delete agent |

### Connections

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/agents/:aid/connections` | List connections for agent |
| `POST` | `/api/v1/agents/:aid/connections` | Add connection |
| `DELETE` | `/api/v1/connections/:id` | Remove connection |

### Sessions

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/connections/:cid/sessions` | Sessions for connection |
| `GET` | `/api/v1/sessions/:id` | Get session + messages |
| `DELETE` | `/api/v1/sessions/:id` | Clear session + messages |

### Messages

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/sessions/:id/messages` | Messages (paginated) |

### System

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/health` | Health check |

### WebSocket

| Path | Description |
|---|---|
| `GET /ws/daemon` | Daemon connection — handshake, tasks, stream events |
| `GET /ws/sessions/:id` | Real-time agent output for a session (web UI) |

---

## Message Commands

| Command | Description |
|---|---|
| `/claude <prompt>` | Run with Claude Code |
| `/codex <prompt>` | Run with Codex |
| `/open <prompt>` | Run with OpenCode |
| `/copilot <prompt>` | Run with GitHub Copilot |
| `/gemini <prompt>` | Run with Gemini |
| `/hermes <prompt>` | Run with Hermes |
| `/openclaw <prompt>` | Run with OpenClaw |
| `/clear` | Reset session history |
| `/help` | Show available commands |

---

## Web Dashboard (Svelte)

### Pages

| Route | Purpose |
|---|---|
| `/` | Dashboard — daemon status, agent status |
| `/<ws-slug>` | Workspace overview |
| `/<ws-slug>/agents` | Agent list |
| `/<ws-slug>/agents/new` | Create agent (pick daemon → harness → model) |
| `/<ws-slug>/agents/:id` | Edit agent + manage connections |
| `/<ws-slug>/daemons` | Daemon list + status |
| `/<ws-slug>/sessions` | Session history |

### Design Philosophy

- Dark theme, minimal, terminal-inspired
- Real-time WebSocket updates
- Mobile-responsive

---

## Gateway Config

```env
# Server
AEGIS_ADDR=:8080
AEGIS_BASE_URL=http://localhost:8080

# Database
AEGIS_DB_PATH=./data/gateway.db

# Telegram
AEGIS_TELEGRAM_TOKEN=

# Discord
AEGIS_DISCORD_TOKEN=
AEGIS_DISCORD_APP_ID=
```

---

## Workspace Isolation (Daemon)

Each task from the daemon gets an isolated directory:

```
./workspaces/
  <workspace_id>/
    <agent_id>/
      2026-06-23-001/
      2026-06-23-002/
```

Auto-cleanup: directories older than 24h removed on startup and hourly.

---

## Error Handling

| Scenario | Behavior |
|---|---|
| Daemon offline | Reply "Agent unavailable — daemon offline" |
| Harness not on daemon | Reply "Agent 'codex' not found on daemon 'dev'" |
| Agent timeout | Reply "Agent timed out after 30m" + partial output |
| Invalid command | Reply with help + available agents for this chat |
| Concurrency limit | Reply "Daemon busy — 5 tasks running" |
| Platform API error | Log, retry 3x, then notify user |

---

## Non-Goals (v1)

These are explicitly excluded from v1 scope:

- AI API key management in gateway (keys live on daemon machines only)
- WhatsApp, Slack, or other messaging platforms (Telegram + Discord only)
- Multi-turn conversation memory (single prompt per message)
- SaaS billing, RBAC, workspace member roles
- Agent-to-agent delegation or skill plugin system
- OAuth / SSO for web dashboard authentication
- Docker deployment config

---

## Directory Structure

```
Aegis/
├── cmd/
│   ├── aegisd/                   # Gateway binary
│   │   └── main.go
│   └── aegis-agent/              # Daemon binary
│       └── main.go
├── internal/
│   ├── gateway/                 # Gateway-specific
│   │   ├── api/                 # REST API
│   │   │   ├── server.go
│   │   │   ├── workspaces.go
│   │   │   ├── daemons.go
│   │   │   ├── agents.go
│   │   │   ├── connections.go
│   │   │   ├── sessions.go
│   │   │   └── messages.go
│   │   ├── msg/                 # messaging adapters
│   │   │   ├── telegram.go
│   │   │   └── discord.go
│   │   ├── router/              # command parser → dispatch
│   │   │   └── router.go
│   │   ├── dispatcher/          # task dispatch to daemon via WS
│   │   │   └── dispatcher.go
│   │   └── store/               # database layer
│   │       ├── db.go
│   │       ├── workspaces.go
│   │       ├── daemons.go
│   │       ├── agents.go
│   │       ├── connections.go
│   │       ├── sessions.go
│   │       └── messages.go
│   ├── daemon/                   # Daemon-specific
│   │   ├── daemon.go             # main loop, WS client, handshake
│   │   ├── harness/
│   │   │   ├── interface.go
│   │   │   ├── discover.go       # auto-detect from PATH
│   │   │   ├── workspace.go      # isolated workdirs
│   │   │   ├── claude.go
│   │   │   ├── codex.go
│   │   │   ├── opencode.go
│   │   │   ├── copilot.go
│   │   │   ├── gemini.go
│   │   │   ├── hermes.go
│   │   │   └── openclaw.go
│   │   └── config/
│   │       └── config.go
│   └── shared/                   # Shared between gateway and daemon
│       └── protocol/             # WebSocket message types
│           └── messages.go
├── web/                          # Svelte frontend
│   ├── src/
│   │   ├── routes/
│   │   │   ├── +page.svelte
│   │   │   ├── [workspace]/
│   │   │   │   ├── +page.svelte
│   │   │   │   ├── agents/
│   │   │   │   │   ├── +page.svelte
│   │   │   │   │   ├── new/+page.svelte
│   │   │   │   │   └── [id]/+page.svelte
│   │   │   │   ├── daemons/+page.svelte
│   │   │   │   └── sessions/
│   │   │   │       ├── +page.svelte
│   │   │   │       └── [id]/+page.svelte
│   │   │   └── +layout.svelte
│   │   ├── lib/
│   │   │   ├── api.ts
│   │   │   └── ws.ts
│   │   └── app.css
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.ts
├── static/
│   └── .gitkeep
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```
