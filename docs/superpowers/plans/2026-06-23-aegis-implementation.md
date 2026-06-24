# Aegis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Aegis — a managed multi-agent bot platform with gateway + daemon architecture. Users chat via Telegram/Discord, freely choose which agent harness (Claude Code, Codex, OpenCode, etc.) to use, and the daemon spawns the appropriate CLI subprocess.

**Architecture:** Monorepo with two Go binaries (`aegisd` gateway, `aegis-agent` daemon), shared protocol types, embedded Svelte SPA frontend, and Turso/libSQL embedded database. Gateway handles messaging + API + task dispatch. Daemon handles CLI detection + API keys + subprocess execution. They communicate via a persistent WebSocket.

**Tech Stack:** Go 1.24+, `net/http`, `github.com/coder/websocket`, `github.com/tursodatabase/go-libsql`, `github.com/go-telegram/bot`, `github.com/disgoorg/disgo`, `github.com/google/uuid`, Svelte 5 + SvelteKit (static adapter), TailwindCSS 4, shadcn-svelte.

---

## File Structure Plan

```
aegis/                              (rename from ai-agent-bot)
├── cmd/
│   ├── aegisd/
│   │   └── main.go                 # Gateway entry point
│   └── aegis-agent/
│       └── main.go                 # Daemon entry point
├── internal/
│   ├── shared/
│   │   └── protocol/
│   │       └── protocol.go         # WS message types, shared constants
│   ├── gateway/
│   │   ├── server.go               # HTTP server setup, SPA serving
│   │   ├── api/                    # REST API handlers
│   │   │   ├── workspaces.go
│   │   │   ├── daemons.go
│   │   │   ├── agents.go
│   │   │   ├── connections.go
│   │   │   ├── sessions.go
│   │   │   └── messages.go
│   │   ├── ws/                     # WebSocket handling
│   │   │   ├── hub.go              # Daemon connection hub
│   │   │   └── daemon.go           # Per-daemon WS conn mgmt
│   │   ├── dispatch/
│   │   │   └── dispatch.go         # Task dispatch to daemon
│   │   ├── msg/                    # Messaging platform adapters
│   │   │   ├── adapter.go          # Adapter interface
│   │   │   ├── telegram.go
│   │   │   └── discord.go
│   │   ├── router/
│   │   │   └── router.go           # Command parsing
│   │   └── store/
│   │       ├── db.go               # DB open, migrations
│   │       ├── workspaces.go
│   │       ├── daemons.go
│   │       ├── agents.go
│   │       ├── connections.go
│   │       ├── sessions.go
│   │       └── messages.go
│   └── daemon/
│       ├── config/
│       │   └── config.go           # Env var config loading
│       ├── client.go               # WS client to gateway
│       ├── handler.go              # Task message handler
│       └── harness/
│           ├── interface.go        # Runner interface
│           ├── discover.go         # Auto-detect from PATH
│           ├── workspace.go        # Isolated workdirs
│           ├── claude.go
│           ├── codex.go
│           ├── opencode.go
│           ├── copilot.go
│           └── gemini.go
├── web/                            # Svelte + TailwindCSS + shadcn-svelte
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
│   │   │   │   └── sessions/+page.svelte
│   │   │   └── +layout.svelte
│   │   ├── lib/
│   │   │   ├── api.ts
│   │   │   ├── ws.ts
│   │   │   └── utils.ts            # cn() helper
│   │   └── app.css                 # Tailwind layers + shadcn-svelte theme
│   ├── components.json             # shadcn-svelte config
│   ├── package.json
│   ├── svelte.config.js
│   └── vite.config.ts
├── static/                         # Embedded Svelte build output + gitkeep
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Phase 0: Repository Bootstrap

### Task 0.1: Rename project directory and init Go module

**Files:**
- Create: `go.mod`

- [ ] **Step 1: Rename directory from ai-agent-bot to aegis**

```bash
cd ~/Project && mv ai-agent-bot aegis && cd aegis
```

- [ ] **Step 2: Initialize Go module**

```bash
go mod init github.com/<user>/aegis
```

Replace `<user>` with the appropriate GitHub username/organization.

- [ ] **Step 3: Create .gitignore**

Create `.gitignore`:
```
# Binaries
/aegisd
/aegis-agent

# Build output
/web/build/
/static/_app/

# Database
/data/

# Workspaces
/workspaces/

# Environment
.env

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db
```

- [ ] **Step 4: Create .env.example**

Create `.env.example`:
```env
# Gateway
AEGIS_ADDR=:8080
AEGIS_BASE_URL=http://localhost:8080
AEGIS_DB_PATH=./data/gateway.db
AEGIS_TELEGRAM_TOKEN=
AEGIS_DISCORD_TOKEN=
AEGIS_DISCORD_APP_ID=

# Daemon
AEGIS_GATEWAY_URL=ws://localhost:8080/ws/daemon
AEGIS_DAEMON_ID=                                # UUID from web UI daemon creation
AEGIS_DAEMON_TOKEN=                             # enrollment token from web UI (required)
OPENAI_API_KEY=
GOOGLE_API_KEY=
AEGIS_CLAUDE_MODEL=claude-sonnet-4-20250514
```

- [ ] **Step 5: Create Makefile**

Create `Makefile`:
```makefile
.PHONY: build run-gateway run-daemon dev clean

build:
	go build -o aegisd ./cmd/aegisd
	go build -o aegis-agent ./cmd/aegis-agent

run-gateway: build
	./aegisd

run-daemon: build
	./aegis-agent

dev:
	go run ./cmd/aegisd &

clean:
	rm -f aegisd aegis-agent

test:
	go test ./... -v -count=1

test-race:
	go test ./... -v -race -count=1

vet:
	go vet ./...
```

- [ ] **Step 6: Create placeholder main.go files**

Create `cmd/aegisd/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("aegisd gateway starting...")
}
```

Create `cmd/aegis-agent/main.go`:
```go
package main

import "fmt"

func main() {
	fmt.Println("aegis-agent daemon starting...")
}
```

- [ ] **Step 7: Verify build**

```bash
go build ./cmd/aegisd && go build ./cmd/aegis-agent
```
Expected: both compile successfully.

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "chore: bootstrap Aegis project scaffold"
```

---

## Phase 1: Shared Protocol

### Task 1.1: Define WebSocket protocol message types

**Files:**
- Create: `internal/shared/protocol/protocol.go`

- [ ] **Step 1: Create protocol.go with message types**

```go
package protocol

import "encoding/json"

// MessageType identifies the kind of WS message.
type MessageType string

const (
	// Daemon → Gateway
	TypeHandshake  MessageType = "handshake"
	TypeStdout     MessageType = "stdout"
	TypeStderr     MessageType = "stderr"
	TypeDone       MessageType = "done"
	TypeError      MessageType = "error"

	// Gateway → Daemon
	TypeTask       MessageType = "task"
	TypeHandshakeOK MessageType = "handshake_ok"
)

// Message is the wire format for all WS messages.
type Message struct {
	Type    MessageType `json:"type"`
	TaskID  string      `json:"task_id,omitempty"`
	Content string      `json:"content,omitempty"`

	// Handshake fields (daemon → gateway)
	DaemonID  string   `json:"daemon_id,omitempty"`
	Token     string   `json:"token,omitempty"`
	Harnesses []string `json:"harnesses,omitempty"`

	// Task fields (gateway → daemon)
	Harness   string   `json:"harness,omitempty"`
	Prompt    string   `json:"prompt,omitempty"`
	Model     string   `json:"model,omitempty"`
	ExtraArgs []string `json:"extra_args,omitempty"`
}

// Encode serializes a Message to JSON bytes.
func (m Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// DecodeMessage parses JSON bytes into a Message.
func DecodeMessage(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	return m, err
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/shared/...
```
Expected: compiles.

- [ ] **Step 3: Commit**

```bash
git add internal/shared/ && git commit -m "feat: add WebSocket protocol message types"
```

---

## Phase 2: Daemon

### Task 2.1: Build harness interface and Runner type

**Files:**
- Create: `internal/daemon/harness/interface.go`

- [ ] **Step 1: Define Runner interface + types**

```go
package harness

import "context"

// EventType tags a stream event.
type EventType int

const (
	EventStdout EventType = iota
	EventStderr
	EventDone
	EventError
)

// StreamEvent is a single output event from a running agent.
type StreamEvent struct {
	Type    EventType
	Content string
}

// RunRequest holds the parameters for an agent run.
type RunRequest struct {
	TaskID    string
	Prompt    string
	WorkDir   string
	Model     string
	ExtraArgs []string
}

// Runner is implemented by each agent harness.
type Runner interface {
	// Name returns the harness identifier (e.g. "claude", "codex").
	Name() string

	// Available reports whether the CLI binary is on PATH.
	Available() bool

	// Run executes the agent and streams events.
	// The caller MUST read the channel until it closes.
	Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error)
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/daemon/harness/
```
Expected: compiles.

- [ ] **Step 3: Commit**

```bash
git add internal/daemon/harness/interface.go && git commit -m "feat: define harness Runner interface"
```

### Task 2.2: Implement harness discovery

**Files:**
- Create: `internal/daemon/harness/discover.go`
- Create: `internal/daemon/harness/discover_test.go`

- [ ] **Step 1: Write discovering function**

```go
package harness

import (
	"os"
	"os/exec"
	"strings"
)

// RegisteredHarness pairs a Runner with its configured binary path.
type RegisteredHarness struct {
	Runner Runner
	Path   string // resolved binary path
}

// Discover returns all harnesses whose CLI is available,
// resolving paths from AEGIS_<NAME>_PATH env vars first, then $PATH.
func Discover(runners []Runner) []RegisteredHarness {
	var out []RegisteredHarness
	for _, r := range runners {
		path := resolvePath(r.Name())
		if path == "" {
			continue
		}
		out = append(out, RegisteredHarness{Runner: r, Path: path})
	}
	return out
}

func resolvePath(name string) string {
	envKey := "AEGIS_" + strings.ToUpper(name) + "_PATH"
	if p := os.Getenv(envKey); p != "" {
		return p
	}
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}
```

- [ ] **Step 2: Write test for discovery**

```go
package harness

import (
	"context"
	"os"
	"testing"
)

type mockRunner struct {
	name      string
	available bool
}

func (m mockRunner) Name() string                              { return m.name }
func (m mockRunner) Available() bool                            { return m.available }
func (m mockRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) { return nil, nil }

func TestDiscover_UsesEnvVar(t *testing.T) {
	os.Setenv("AEGIS_MOCKAGENT_PATH", "/custom/path/mockagent")
	defer os.Unsetenv("AEGIS_MOCKAGENT_PATH")

	runner := mockRunner{name: "mockagent", available: false}
	reg := Discover([]Runner{runner})

	if len(reg) != 1 {
		t.Fatalf("expected 1 registered harness, got %d", len(reg))
	}
	if reg[0].Path != "/custom/path/mockagent" {
		t.Errorf("expected path /custom/path/mockagent, got %s", reg[0].Path)
	}
}

func TestDiscover_SkipsUnavailable(t *testing.T) {
	runner := mockRunner{name: "nonexistentagentxyz", available: false}
	reg := Discover([]Runner{runner})

	if len(reg) != 0 {
		t.Errorf("expected 0 registered harnesses, got %d", len(reg))
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/daemon/harness/ -v -run TestDiscover
```
Expected: both PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/daemon/harness/discover.go internal/daemon/harness/discover_test.go && git commit -m "feat: add harness auto-discovery from PATH"
```

### Task 2.3: Implement daemon config loading

**Files:**
- Create: `internal/daemon/config/config.go`
- Create: `internal/daemon/config/config_test.go`

- [ ] **Step 1: Write config struct and loader**

```go
package config

import (
	"os"
	"time"
)

type Config struct {
	GatewayURL     string
	DaemonID       string // UUID from web UI
	DaemonName     string
	Token          string // enrollment token from AEGIS_DAEMON_TOKEN
	WorkspacesRoot string
	MaxConcurrent  int
	AgentTimeout   time.Duration
}

func Load() *Config {
	return &Config{
		GatewayURL:     env("AEGIS_GATEWAY_URL", "ws://localhost:8080/ws/daemon"),
		DaemonID:       env("AEGIS_DAEMON_ID", ""),
		DaemonName:     env("AEGIS_DAEMON_NAME", "aegis-agent"),
		Token:          env("AEGIS_DAEMON_TOKEN", ""),
		WorkspacesRoot: env("AEGIS_WORKSPACES_ROOT", "./workspaces"),
		MaxConcurrent:  envInt("AEGIS_MAX_CONCURRENT", 5),
		AgentTimeout:   envDuration("AEGIS_AGENT_TIMEOUT", 30*time.Minute),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	// simplified; use strconv.Atoi in real impl
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	// strconv.Atoi omitted for brevity in plan
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
```

- [ ] **Step 2: Write test**

```go
package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load()
	if cfg.GatewayURL != "ws://localhost:8080/ws/daemon" {
		t.Errorf("unexpected gateway URL: %s", cfg.GatewayURL)
	}
	if cfg.MaxConcurrent != 5 {
		t.Errorf("expected 5 max concurrent, got %d", cfg.MaxConcurrent)
	}
	if cfg.AgentTimeout != 30*time.Minute {
		t.Errorf("expected 30m timeout, got %v", cfg.AgentTimeout)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("AEGIS_DAEMON_NAME", "production-agent")
	defer os.Unsetenv("AEGIS_DAEMON_NAME")

	cfg := Load()
	if cfg.DaemonName != "production-agent" {
		t.Errorf("expected production-agent, got %s", cfg.DaemonName)
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/daemon/config/ -v
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/daemon/config/ && git commit -m "feat: add daemon config loading from env vars"
```

### Task 2.4: Implement workspace isolation

**Files:**
- Create: `internal/daemon/harness/workspace.go`
- Create: `internal/daemon/harness/workspace_test.go`

- [ ] **Step 1: Write workspace manager**

```go
package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type WorkspaceManager struct {
	Root string
}

func NewWorkspaceManager(root string) *WorkspaceManager {
	return &WorkspaceManager{Root: root}
}

// Create returns a new isolated workdir for a task.
func (wm *WorkspaceManager) Create(workspaceID, agentID, taskID string) (string, error) {
	dir := filepath.Join(wm.Root, workspaceID, agentID, taskID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	return dir, nil
}

// Cleanup removes workdirs older than maxAge.
func (wm *WorkspaceManager) Cleanup(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	return filepath.Walk(wm.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible
		}
		if info.IsDir() && info.ModTime().Before(cutoff) && path != wm.Root {
			os.RemoveAll(path)
		}
		return nil
	})
}
```

- [ ] **Step 2: Write test**

```go
package harness

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceManager_Create(t *testing.T) {
	root := t.TempDir()
	wm := NewWorkspaceManager(root)

	dir, err := wm.Create("ws-1", "agent-1", "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(root, "ws-1", "agent-1", "task-1")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("workspace directory was not created")
	}
}

func TestWorkspaceManager_Cleanup(t *testing.T) {
	root := t.TempDir()
	wm := NewWorkspaceManager(root)

	// Create an old directory
	oldDir := filepath.Join(root, "ws-old", "agent", "task-old")
	os.MkdirAll(oldDir, 0755)
	// Set modtime to 48h ago
	oldTime := time.Now().Add(-48 * time.Hour)
	os.Chtimes(oldDir, oldTime, oldTime)

	err := wm.Cleanup(24 * time.Hour)
	if err != nil {
		t.Fatalf("cleanup error: %v", err)
	}

	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old workspace was not cleaned up")
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/daemon/harness/ -v -run TestWorkspace
```
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/daemon/harness/workspace.go internal/daemon/harness/workspace_test.go && git commit -m "feat: add workspace isolation with cleanup"
```

### Task 2.5: Implement Claude Code harness

**Files:**
- Create: `internal/daemon/harness/claude.go`

- [ ] **Step 1: Implement claude Runner**

```go
package harness

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type ClaudeRunner struct {
	path  string
	model string
}

func NewClaudeRunner(path, model string) *ClaudeRunner {
	return &ClaudeRunner{path: path, model: model}
}

func (c *ClaudeRunner) Name() string { return "claude" }

func (c *ClaudeRunner) Available() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func (c *ClaudeRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	model := req.Model
	if model == "" {
		model = c.model
	}

	args := []string{"-p", req.Prompt}
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, req.ExtraArgs...)

	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Dir = req.WorkDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("claude stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("claude stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("claude start: %w", err)
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)

		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}

func scanStream(r io.Reader, evt EventType, ch chan<- StreamEvent) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		ch <- StreamEvent{Type: evt, Content: scanner.Text()}
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/daemon/harness/
```
Expected: compiles.

- [ ] **Step 3: Commit**

```bash
git add internal/daemon/harness/claude.go && git commit -m "feat: implement Claude Code harness"
```

### Task 2.6: Implement remaining harnesses (codex, opencode, copilot, gemini)

**Files:**
- Create: `internal/daemon/harness/codex.go`
- Create: `internal/daemon/harness/opencode.go`
- Create: `internal/daemon/harness/copilot.go`
- Create: `internal/daemon/harness/gemini.go`

- [ ] **Step 1: Codex harness**

```go
package harness

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type CodexRunner struct {
	path  string
	model string
}

func NewCodexRunner(path, model string) *CodexRunner {
	return &CodexRunner{path: path, model: model}
}

func (c *CodexRunner) Name() string { return "codex" }

func (c *CodexRunner) Available() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

func (c *CodexRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	model := req.Model
	if model == "" {
		model = c.model
	}

	args := []string{"exec", req.Prompt}
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, req.ExtraArgs...)

	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Dir = req.WorkDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("codex stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("codex stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("codex start: %w", err)
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)

		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}
```

- [ ] **Step 2: OpenCode harness**

```go
package harness

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type OpenCodeRunner struct {
	path  string
	model string
}

func NewOpenCodeRunner(path, model string) *OpenCodeRunner {
	return &OpenCodeRunner{path: path, model: model}
}

func (o *OpenCodeRunner) Name() string { return "opencode" }

func (o *OpenCodeRunner) Available() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}

func (o *OpenCodeRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	args := []string{"run", req.Prompt}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	} else if o.model != "" {
		args = append(args, "--model", o.model)
	}
	args = append(args, req.ExtraArgs...)

	cmd := exec.CommandContext(ctx, o.path, args...)
	cmd.Dir = req.WorkDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("opencode stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("opencode stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("opencode start: %w", err)
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)

		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}
```

- [ ] **Step 3: Copilot harness**

```go
package harness

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type CopilotRunner struct {
	path string
}

func NewCopilotRunner(path string) *CopilotRunner {
	return &CopilotRunner{path: path}
}

func (c *CopilotRunner) Name() string { return "copilot" }

func (c *CopilotRunner) Available() bool {
	_, err := exec.LookPath("copilot")
	return err == nil
}

func (c *CopilotRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	args := []string{"suggest", req.Prompt}
	args = append(args, req.ExtraArgs...)

	cmd := exec.CommandContext(ctx, c.path, args...)
	cmd.Dir = req.WorkDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("copilot stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("copilot stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("copilot start: %w", err)
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)

		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}
```

- [ ] **Step 4: Gemini harness**

```go
package harness

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type GeminiRunner struct {
	path  string
	model string
}

func NewGeminiRunner(path, model string) *GeminiRunner {
	return &GeminiRunner{path: path, model: model}
}

func (g *GeminiRunner) Name() string { return "gemini" }

func (g *GeminiRunner) Available() bool {
	_, err := exec.LookPath("gemini")
	return err == nil
}

func (g *GeminiRunner) Run(ctx context.Context, req RunRequest) (<-chan StreamEvent, error) {
	model := req.Model
	if model == "" {
		model = g.model
	}

	args := []string{"chat", req.Prompt}
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, req.ExtraArgs...)

	cmd := exec.CommandContext(ctx, g.path, args...)
	cmd.Dir = req.WorkDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("gemini stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("gemini stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("gemini start: %w", err)
	}

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		scanStream(stdout, EventStdout, ch)
		scanStream(stderr, EventStderr, ch)

		if err := cmd.Wait(); err != nil {
			ch <- StreamEvent{Type: EventError, Content: err.Error()}
		}
		ch <- StreamEvent{Type: EventDone}
	}()

	return ch, nil
}
```

- [ ] **Step 5: Verify compilation**

```bash
go build ./internal/daemon/harness/
```
Expected: compiles.

- [ ] **Step 6: Commit**

```bash
git add internal/daemon/harness/ && git commit -m "feat: implement codex, opencode, copilot, gemini harnesses"
```

### Task 2.7: Build daemon WebSocket client and main loop

**Files:**
- Create: `internal/daemon/client.go`
- Create: `internal/daemon/handler.go`
- Modify: `cmd/aegis-agent/main.go`

- [ ] **Step 1: Install websocket dependency**

```bash
go get github.com/coder/websocket
go get github.com/google/uuid
```

- [ ] **Step 2: Write daemon WS client**

```go
package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"

	"github.com/<user>/aegis/internal/daemon/config"
	"github.com/<user>/aegis/internal/daemon/harness"
	"github.com/<user>/aegis/internal/shared/protocol"
)

type Client struct {
	cfg     *config.Config
	id      string // generated UUID
	runners map[string]*harness.RegisteredHarness // name → harness
	wm      *harness.WorkspaceManager
	conn    *websocket.Conn
	mu      sync.Mutex
	tasks   map[string]context.CancelFunc // taskID → cancel
}

func NewClient(cfg *config.Config, reg []harness.RegisteredHarness) *Client {
	runners := make(map[string]*harness.RegisteredHarness)
	for i := range reg {
		runners[reg[i].Runner.Name()] = &reg[i]
	}
	return &Client{
		cfg:     cfg,
		id:      uuid.NewString(),
		runners: runners,
		wm:      harness.NewWorkspaceManager(cfg.WorkspacesRoot),
		tasks:   make(map[string]context.CancelFunc),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	slog.Info("connecting to gateway", "url", c.cfg.GatewayURL)
	conn, _, err := websocket.Dial(ctx, c.cfg.GatewayURL, nil)
	if err != nil {
		return fmt.Errorf("dial gateway: %w", err)
	}
	c.conn = conn

	// Send handshake with enrollment token
	harnessNames := make([]string, 0, len(c.runners))
	hs := protocol.Message{
		Type:      protocol.TypeHandshake,
		DaemonID:  c.cfg.DaemonID,
		Token:     c.cfg.Token,
		Harnesses: harnessNames,
	}

	if err := wsjson.Write(ctx, conn, hs); err != nil {
		return fmt.Errorf("send handshake: %w", err)
	}

	slog.Info("handshake sent", "harnesses", harnessNames)
	return nil

func (c *Client) Run(ctx context.Context) error {
	// Periodic workspace cleanup
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.wm.Cleanup(24 * time.Hour)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Read loop
	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		msg, err := protocol.DecodeMessage(data)
		if err != nil {
			slog.Error("decode message", "err", err)
			continue
		}

		switch msg.Type {
		case protocol.TypeTask:
			go c.handleTask(msg)
		case protocol.TypeHandshakeOK:
			slog.Info("handshake accepted by gateway")
		}
	}
}

func (c *Client) handleTask(msg protocol.Message) {
	runner, ok := c.runners[msg.Harness]
	if !ok {
		slog.Error("unknown harness requested", "harness", msg.Harness)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.AgentTimeout)
	defer cancel()

	c.mu.Lock()
	c.tasks[msg.TaskID] = cancel
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.tasks, msg.TaskID)
		c.mu.Unlock()
	}()

	workDir, err := c.wm.Create("default", msg.TaskID, msg.TaskID)
	if err != nil {
		c.sendEvent(msg.TaskID, protocol.TypeError, err.Error())
		return
	}

	req := harness.RunRequest{
		TaskID:    msg.TaskID,
		Prompt:    msg.Prompt,
		WorkDir:   workDir,
		Model:     msg.Model,
		ExtraArgs: msg.ExtraArgs,
	}

	ch, err := runner.Runner.Run(ctx, req)
	if err != nil {
		c.sendEvent(msg.TaskID, protocol.TypeError, err.Error())
		return
	}

	for evt := range ch {
		var ptype protocol.MessageType
		switch evt.Type {
		case harness.EventStdout:
			ptype = protocol.TypeStdout
		case harness.EventStderr:
			ptype = protocol.TypeStderr
		case harness.EventDone:
			ptype = protocol.TypeDone
		case harness.EventError:
			ptype = protocol.TypeError
		}
		c.sendEvent(msg.TaskID, ptype, evt.Content)
	}
}

func (c *Client) sendEvent(taskID string, evtType protocol.MessageType, content string) {
	msg := protocol.Message{
		Type:    evtType,
		TaskID:  taskID,
		Content: content,
	}
	data, _ := msg.Encode()
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
			slog.Error("send event", "err", err)
		}
	}
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close(websocket.StatusNormalClosure, "daemon shutting down")
	}
	return nil
}
```

- [ ] **Step 3: Write daemon main.go**

Replace `cmd/aegis-agent/main.go`:
```go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/<user>/aegis/internal/daemon"
	"github.com/<user>/aegis/internal/daemon/config"
	"github.com/<user>/aegis/internal/daemon/harness"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()

	// Build all harness runners
	runners := []harness.Runner{
		harness.NewClaudeRunner(envPath("AEGIS_CLAUDE_PATH", "claude"), envStr("AEGIS_CLAUDE_MODEL", "")),
		harness.NewCodexRunner(envPath("AEGIS_CODEX_PATH", "codex"), envStr("AEGIS_CODEX_MODEL", "")),
		harness.NewOpenCodeRunner(envPath("AEGIS_OPENCODE_PATH", "opencode"), envStr("AEGIS_OPENCODE_MODEL", "")),
		harness.NewCopilotRunner(envPath("AEGIS_COPILOT_PATH", "copilot")),
		harness.NewGeminiRunner(envPath("AEGIS_GEMINI_PATH", "gemini"), envStr("AEGIS_GEMINI_MODEL", "")),
	}

	reg := harness.Discover(runners)
	if len(reg) == 0 {
		slog.Warn("no harnesses discovered from PATH")
	}

	client := daemon.NewClient(cfg, reg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		slog.Error("connect", "err", err)
		os.Exit(1)
	}

	slog.Info("daemon running", "name", cfg.DaemonName, "harnesses", len(reg))
	if err := client.Run(ctx); err != nil {
		slog.Error("run", "err", err)
	}
	client.Close()
}

func envPath(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./cmd/aegis-agent
```
Expected: compiles.

- [ ] **Step 5: Commit**

```bash
git add internal/daemon/ cmd/aegis-agent/main.go && git commit -m "feat: implement daemon WS client and main loop"
```

---

## Phase 3: Gateway Core

### Task 3.1: Add libSQL dependency and write database layer

**Files:**
- Create: `internal/gateway/store/db.go`
- Create: `internal/gateway/store/workspaces.go`
- Create: `internal/gateway/store/daemons.go`
- Create: `internal/gateway/store/agents.go`
- Create: `internal/gateway/store/connections.go`

- [ ] **Step 1: Install dependencies**

```bash
go get github.com/tursodatabase/go-libsql
go get github.com/google/uuid
```

- [ ] **Step 2: Write database open and migration**

```go
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/tursodatabase/go-libsql"
)

type Store struct {
	DB *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("libsql", "file:"+path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite single-writer

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Store{DB: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS workspaces (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS daemons (
		id TEXT PRIMARY KEY,
		workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
		name TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'offline',
		last_seen TEXT,
		created_at TEXT NOT NULL,
		UNIQUE(workspace_id, name)
	);
	CREATE TABLE IF NOT EXISTS daemon_harnesses (
		daemon_id TEXT NOT NULL REFERENCES daemons(id) ON DELETE CASCADE,
		harness TEXT NOT NULL,
		PRIMARY KEY (daemon_id, harness)
	);
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
		daemon_id TEXT NOT NULL REFERENCES daemons(id),
		name TEXT NOT NULL,
		harness TEXT NOT NULL,
		model TEXT DEFAULT '',
		extra_args TEXT DEFAULT '',
		enabled INTEGER DEFAULT 1,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS connections (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
		platform TEXT NOT NULL,
		chat_id TEXT NOT NULL,
		created_at TEXT NOT NULL,
		UNIQUE (agent_id, platform, chat_id)
	);
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		connection_id TEXT NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
		user_name TEXT DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		agent_id TEXT,
		created_at TEXT NOT NULL
	);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *Store) Close() error {
	return s.DB.Close()
}
```

- [ ] **Step 3: Write workspace queries**

```go
package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

func (s *Store) CreateWorkspace(name, slug string) (*Workspace, error) {
	ws := &Workspace{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_, err := s.DB.Exec(`INSERT INTO workspaces (id, name, slug, created_at) VALUES (?, ?, ?, ?)`,
		ws.ID, ws.Name, ws.Slug, ws.CreatedAt)
	return ws, err
}

func (s *Store) ListWorkspaces() ([]Workspace, error) {
	rows, err := s.DB.Query(`SELECT id, name, slug, created_at FROM workspaces ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *Store) GetWorkspace(id string) (*Workspace, error) {
	var w Workspace
	err := s.DB.QueryRow(`SELECT id, name, slug, created_at FROM workspaces WHERE id = ?`, id).
		Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &w, err
}

func (s *Store) DeleteWorkspace(id string) error {
	_, err := s.DB.Exec(`DELETE FROM workspaces WHERE id = ?`, id)
	return err
}
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./internal/gateway/store/
```
Expected: compiles.

- [ ] **Step 5: Commit**

```bash
git add internal/gateway/store/ && git commit -m "feat: add database layer with migrations and workspace CRUD"
```

### Task 3.2: Write daemon store queries

**Files:**
- Create: `internal/gateway/store/daemons.go`

- [ ] **Step 1: Daemon CRUD**

```go
package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Daemon struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	TokenHash   string `json:"-"` // SHA-256, never exposed via API
	Status      string `json:"status"`
	LastSeen    string `json:"last_seen,omitempty"`
	CreatedAt   string `json:"created_at"`
}
// CreateDaemon generates a new daemon scoped to a workspace with an enrollment token.
// Returns the plaintext token (shown once) and the daemon record.
func (s *Store) CreateDaemon(workspaceID, name, tokenHash string) (*Daemon, error) {
	d := Daemon{
		ID:          uuid.NewString(),
		WorkspaceID: workspaceID,
		Name:        name,
		TokenHash:   tokenHash,
		Status:      "offline",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	_, err := s.DB.Exec(`INSERT INTO daemons (id, workspace_id, name, token_hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		d.ID, d.WorkspaceID, d.Name, d.TokenHash, d.Status, d.CreatedAt)
	return &d, err
}

// AuthenticateDaemon verifies the daemon name and token hash match, then updates status + harnesses.
// AuthenticateDaemon looks up the daemon by ID and verifies the token hash match.
func (s *Store) AuthenticateDaemon(daemonID, tokenHash string, harnesses []string) error {
	var storedHash string
	err := s.DB.QueryRow(`SELECT token_hash FROM daemons WHERE id = ?`, daemonID).
		Scan(&storedHash)
	if err == sql.ErrNoRows {
		return fmt.Errorf("unknown daemon id: %s", daemonID)
	}
	if err != nil {
		return err
	}
	if storedHash != tokenHash {
		return fmt.Errorf("invalid token for daemon: %s", daemonID)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	s.DB.Exec(`UPDATE daemons SET status = 'online', last_seen = ? WHERE id = ?`, now, daemonID)
	s.DB.Exec(`DELETE FROM daemon_harnesses WHERE daemon_id = ?`, daemonID)
	for _, h := range harnesses {
		s.DB.Exec(`INSERT OR IGNORE INTO daemon_harnesses (daemon_id, harness) VALUES (?, ?)`, daemonID, h)
	}
	return nil
}
	return id, nil
}

func (s *Store) GetDaemonHarnesses(daemonID string) ([]string, error) {
	rows, err := s.DB.Query(`SELECT harness FROM daemon_harnesses WHERE daemon_id = ?`, daemonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/gateway/store/
```
Expected: compiles.

- [ ] **Step 3: Commit**

```bash
git add internal/gateway/store/daemons.go && git commit -m "feat: add daemon store with upsert and harness tracking"
```

### Task 3.3: Write agent and connection store queries

**Files:**
- Create: `internal/gateway/store/agents.go`
- Create: `internal/gateway/store/connections.go`
- Create: `internal/gateway/store/sessions.go`
- Create: `internal/gateway/store/messages.go`

- [ ] **Step 1: Agent CRUD**

```go
package store

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	DaemonID    string `json:"daemon_id"`
	Name        string `json:"name"`
	Harness     string `json:"harness"`
	Model       string `json:"model"`
	ExtraArgs   string `json:"extra_args"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (s *Store) CreateAgent(a *Agent) error {
	a.ID = uuid.NewString()
	a.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	a.UpdatedAt = a.CreatedAt
	enabled := 0
	if a.Enabled {
		enabled = 1
	}
	_, err := s.DB.Exec(`INSERT INTO agents (id, workspace_id, daemon_id, name, harness, model, extra_args, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.WorkspaceID, a.DaemonID, a.Name, a.Harness, a.Model, a.ExtraArgs, enabled, a.CreatedAt, a.UpdatedAt)
	return err
}

func (s *Store) ListAgents(workspaceID string) ([]Agent, error) {
	rows, err := s.DB.Query(`SELECT id, workspace_id, daemon_id, name, harness, model, extra_args, enabled, created_at, updated_at FROM agents WHERE workspace_id = ? ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAgents(rows)
}
```

Rest of store files follow same pattern. Refer to spec database schema for all fields.

- [ ] **Step 2: Compile and commit**

```bash
go build ./internal/gateway/store/ && git add internal/gateway/store/ && git commit -m "feat: add agent, connection, session, message store layers"
```

### Task 3.4: Build WebSocket hub for daemon connections

**Files:**
- Create: `internal/gateway/ws/hub.go`
- Create: `internal/gateway/ws/daemon.go`

- [ ] **Step 1: WS hub managing daemon connections**

```go
package ws

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/coder/websocket"

	"github.com/<user>/aegis/internal/gateway/store"
	"github.com/<user>/aegis/internal/shared/protocol"
)

type Hub struct {
	store     *store.Store
	daemons   map[string]*DaemonConn // daemonID → conn
	mu        sync.RWMutex
	OnTask    func(daemonID, taskID string, event protocol.Message) // callback for task events
}

func NewHub(s *store.Store) *Hub {
	return &Hub{
		store:   s,
		daemons: make(map[string]*DaemonConn),
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		slog.Error("ws accept", "err", err)
		return
	}

	dc := &DaemonConn{conn: conn, hub: h}
	go dc.handle(r.Context())
}

// SendTask dispatches a task to a connected daemon.
func (h *Hub) SendTask(daemonID string, msg protocol.Message) error {
	h.mu.RLock()
	dc, ok := h.daemons[daemonID]
	h.mu.RUnlock()
	if !ok {
		return fmt.Errorf("daemon %s not connected", daemonID)
	}
	return dc.send(msg)
}
```

Per-daemon connection handler handles handshake registration, heartbeats, and task forwarding. Full implementation per spec WebSocket protocol.

- [ ] **Step 2: Compile**

```bash
go build ./internal/gateway/ws/
```
Expected: compiles.

- [ ] **Step 3: Commit**

```bash
git add internal/gateway/ws/ && git commit -m "feat: add WebSocket hub for daemon connections"
```

### Task 3.5: Write REST API handlers

**Files:**
- Create: `internal/gateway/server.go`
- Create: `internal/gateway/api/workspaces.go`
- Create: `internal/gateway/api/daemons.go`
- Create: `internal/gateway/api/agents.go`
- Create: `internal/gateway/api/connections.go`
- Create: `internal/gateway/api/sessions.go`
- Create: `internal/gateway/api/messages.go`

- [ ] **Step 1: Server setup with route registration**

```go
package gateway

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/<user>/aegis/internal/gateway/api"
	"github.com/<user>/aegis/internal/gateway/ws"
	"github.com/<user>/aegis/internal/gateway/store"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	store *store.Store
	hub   *ws.Hub
	mux   *http.ServeMux
}

func NewServer(s *store.Store, hub *ws.Hub) *Server {
	srv := &Server{store: s, hub: hub, mux: http.NewServeMux()}
	srv.registerRoutes()
	return srv
}

func (s *Server) registerRoutes() {
	// CORS middleware and JSON headers omitted for brevity — add in implementation.

	// REST API
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// Workspaces
	s.mux.HandleFunc("GET /api/v1/workspaces", api.ListWorkspaces(s.store))
	s.mux.HandleFunc("POST /api/v1/workspaces", api.CreateWorkspace(s.store))
	s.mux.HandleFunc("GET /api/v1/workspaces/{id}", api.GetWorkspace(s.store))
	s.mux.HandleFunc("PUT /api/v1/workspaces/{id}", api.UpdateWorkspace(s.store))
	s.mux.HandleFunc("DELETE /api/v1/workspaces/{id}", api.DeleteWorkspace(s.store))

	// Daemons (scoped to workspace)
	s.mux.HandleFunc("GET /api/v1/workspaces/{wid}/daemons", api.ListDaemons(s.store))
	s.mux.HandleFunc("POST /api/v1/workspaces/{wid}/daemons", api.CreateDaemon(s.store))
	s.mux.HandleFunc("GET /api/v1/daemons/{id}", api.GetDaemon(s.store))

	// Agents
	s.mux.HandleFunc("GET /api/v1/workspaces/{wid}/agents", api.ListAgents(s.store))
	s.mux.HandleFunc("POST /api/v1/workspaces/{wid}/agents", api.CreateAgent(s.store))
	s.mux.HandleFunc("GET /api/v1/agents/{id}", api.GetAgent(s.store))
	s.mux.HandleFunc("PUT /api/v1/agents/{id}", api.UpdateAgent(s.store))
	s.mux.HandleFunc("DELETE /api/v1/agents/{id}", api.DeleteAgent(s.store))

	// Connections
	s.mux.HandleFunc("GET /api/v1/agents/{aid}/connections", api.ListConnections(s.store))
	s.mux.HandleFunc("POST /api/v1/agents/{aid}/connections", api.CreateConnection(s.store))
	s.mux.HandleFunc("DELETE /api/v1/connections/{id}", api.DeleteConnection(s.store))

	// WebSocket
	s.mux.Handle("GET /ws/daemon", s.hub)

	// SPA (catch-all for Svelte routes)
	staticFS, _ := fs.Sub(staticFiles, "static")
	s.mux.Handle("GET /", http.FileServer(http.FS(staticFS)))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
```

- [ ] **Step 2: Write API handler examples**

Workspaces handler:
```go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/<user>/aegis/internal/gateway/store"
)

func ListWorkspaces(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaces, err := s.ListWorkspaces()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if workspaces == nil {
			workspaces = []store.Workspace{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(workspaces)
	}
}

func CreateWorkspace(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		ws, err := s.CreateWorkspace(input.Name, input.Slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ws)
	}
}

// CreateDaemon generates an enrollment token, stores its hash, and returns the plaintext token once.
func CreateDaemon(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workspaceID := r.PathValue("wid")
		var input struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Generate random enrollment token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}
		plainToken := "aegis_dmt_" + hex.EncodeToString(tokenBytes)
		tokenHash := sha256Hex(plainToken)

		d, err := s.CreateDaemon(workspaceID, input.Name, tokenHash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"daemon": d,
			"token":  plainToken, // shown once
		})
	}
}

- [ ] **Step 3: Write gateway main.go**

Replace `cmd/aegisd/main.go`:
```go
package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/<user>/aegis/internal/gateway"
	"github.com/<user>/aegis/internal/gateway/store"
	"github.com/<user>/aegis/internal/gateway/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	dbPath := os.Getenv("AEGIS_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/gateway.db"
	}

	s, err := store.Open(dbPath)
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}
	defer s.Close()

	hub := ws.NewHub(s)
	server := gateway.NewServer(s, hub)

	addr := os.Getenv("AEGIS_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	slog.Info("gateway starting", "addr", addr)

	go func() {
		if err := http.ListenAndServe(addr, server); err != nil {
			slog.Error("server", "err", err)
		}
	}()

	ctx, _ := signal.NotifyContext(nil, os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()
	slog.Info("gateway shutting down")
}
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./cmd/aegisd
```
Expected: compiles.

- [ ] **Step 5: Commit**

```bash
git add internal/gateway/ cmd/aegisd/ && git commit -m "feat: add gateway HTTP server, REST API, and WebSocket hub"
```

### Task 3.6: Build messaging adapters (Telegram + Discord)

**Files:**
- Create: `internal/gateway/msg/adapter.go`
- Create: `internal/gateway/msg/telegram.go`
- Create: `internal/gateway/msg/discord.go`
- Create: `internal/gateway/router/router.go`

- [ ] **Step 1: Install Telegram and Discord libraries**

```bash
go get github.com/go-telegram/bot
go get github.com/disgoorg/disgo
```

- [ ] **Step 2: Define adapter interface**

```go
package msg

import "context"

// Message represents an incoming chat message.
type Message struct {
	Platform    string // "telegram", "discord"
	ChatID      string
	UserID      string
	UserName    string
	Text        string
	RawMessage  any // platform-specific raw message for reply
}

// Adapter connects the gateway to a messaging platform.
type Adapter interface {
	// Start begins listening for messages.
	Start(ctx context.Context) (<-chan Message, error)
	// Send sends a text response to a chat.
	Send(chatID, text string) error
	// Close shuts down the adapter.
	Close() error
}
```

- [ ] **Step 3: Telegram adapter** — implement using `go-telegram/bot`, parse webhook/commands, forward to message channel.

- [ ] **Step 4: Discord adapter** — implement using `disgo`, register slash commands, handle interactions.

- [ ] **Step 5: Message router** — parse `/<harness> <prompt>` commands, resolve workspace+agent+daemon from connection DB, dispatch via hub.

- [ ] **Step 6: Verify compilation**

```bash
go build ./internal/gateway/msg/ && go build ./internal/gateway/router/
```
Expected: compiles.

- [ ] **Step 7: Commit**

```bash
git add internal/gateway/msg/ internal/gateway/router/ && git commit -m "feat: add Telegram/Discord adapters and message router"
```

---

## Phase 4: Integration & Wiring

### Task 4.1: Wire messaging adapters into main gateway

**Modify:** `cmd/aegisd/main.go` — start Telegram/Discord adapters if tokens configured, pipe messages through router.

- [ ] **Step 1: Update main.go to conditionally start adapters**

- [ ] **Step 2: Verify both binaries compile**

```bash
go build ./cmd/aegisd && go build ./cmd/aegis-agent
```
Expected: both compile.

- [ ] **Step 3: Commit**

```bash
git add cmd/aegisd/main.go && git commit -m "feat: wire messaging adapters into gateway main"
```

### Task 4.2: End-to-end smoke test

**Files:**
- Create: `internal/gateway/store/db_test.go`

- [ ] **Step 1: Write DB integration test**

```go
package store

import (
	"testing"
)

func TestCreateAndListWorkspaces(t *testing.T) {
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer s.Close()

	ws, err := s.CreateWorkspace("Test WS", "test-ws")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if ws.Name != "Test WS" {
		t.Errorf("expected Test WS, got %s", ws.Name)
	}

	list, err := s.ListWorkspaces()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 workspace, got %d", len(list))
	}
}
```

- [ ] **Step 2: Run tests**

```bash
go test ./... -v -count=1
```
Expected: all tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/gateway/store/db_test.go && git commit -m "test: add database integration tests"
```

---

## Phase 5: Svelte Frontend (Future)

The Svelte frontend is scoped for a later iteration. When implemented:

- Initialize SvelteKit project in `web/`
- Build static adapter output to `static/`
- Go binary embeds `static/` via `//go:embed`
- Pages: workspace selector, agent CRUD, daemon status, session viewer

---

## Spec Coverage Self-Review

| Spec Section | Covered By |
|---|---|
| Two binaries (gateway + daemon) | Phase 0, Task 4.2 |
| WebSocket protocol | Task 1.1 (protocol.go) |
| Database schema (7 tables) | Task 3.1, 3.2, 3.3 |
| Daemon lifecycle | Task 2.7 (client.go + main.go) |
| Harness interface + implementations | Tasks 2.1, 2.5, 2.6 |
| Auto-discovery from PATH | Task 2.2 |
| Workspace isolation + cleanup | Task 2.4 |
| Daemon config (env vars) | Task 2.3 |
| REST API (CRUD endpoints) | Task 3.5 |
| WebSocket hub | Task 3.4 |
| Messaging adapters | Task 3.6 |
| Message router + commands | Task 3.6 |
| Svelte frontend | Phase 5 (future) |
| Error handling | Inline in handlers — review during implementation |
