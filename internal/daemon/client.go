package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/rahmanramsi/aegis/internal/config"
	"github.com/rahmanramsi/aegis/internal/daemon/harness"
	"github.com/rahmanramsi/aegis/internal/protocol"
)

type Client struct {
	cfg     *config.Config
	runners map[string]*harness.RegisteredHarness
	wm      *harness.WorkspaceManager
	conn    *websocket.Conn
	mu      sync.Mutex
	tasks   map[string]context.CancelFunc
}

func NewClient(cfg *config.Config, reg []harness.RegisteredHarness) *Client {
	runners := make(map[string]*harness.RegisteredHarness)
	for i := range reg {
		runners[reg[i].Runner.Name()] = &reg[i]
	}
	return &Client{
		cfg:     cfg,
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

	harnessNames := make([]string, 0, len(c.runners))
	harnessModels := make(map[string][]string, len(c.runners))
	for name, rh := range c.runners {
		harnessNames = append(harnessNames, name)
		if models, err := rh.Runner.Models(ctx); err == nil && len(models) > 0 {
			harnessModels[name] = models
		}
	}

	hs := protocol.Message{
		Type:         protocol.TypeHandshake,
		DaemonID:     c.cfg.DaemonID,
		Token:        c.cfg.DaemonToken,
		DaemonName:   c.cfg.DaemonName,
		Harnesses:    harnessNames,
		HarnessModels: harnessModels,
	}

	if err := wsjson.Write(ctx, conn, hs); err != nil {
		return fmt.Errorf("send handshake: %w", err)
	}

	slog.Info("handshake sent", "harnesses", harnessNames)
	return nil
}


// Run connects to the gateway and runs the read loop with auto-reconnect
// and exponential backoff. It blocks until ctx is cancelled.
func (c *Client) Run(ctx context.Context) error {
	// Cleanup goroutine — runs once for the lifetime of Run.
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

	// Keepalive goroutine — sends pings every 30s.
	go c.keepalive(ctx)

	backoff := 1 * time.Second
	maxBackoff := 32 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := c.Connect(ctx); err != nil {
			slog.Error("connect failed", "err", err, "retry_in", backoff)
			select {
			case <-time.After(backoff):
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		backoff = 1 * time.Second // reset on successful connect

		slog.Info("daemon running", "name", c.cfg.DaemonName)
		err := c.readLoop(ctx)
		slog.Error("disconnected", "err", err)
		c.Close()
	}
}

// readLoop reads messages from the connection until it fails.
func (c *Client) readLoop(ctx context.Context) error {
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

// keepalive sends WebSocket pings every 30 seconds.
func (c *Client) keepalive(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			if c.conn != nil {
				c.conn.Ping(ctx)
			}
			c.mu.Unlock()
		case <-ctx.Done():
			return
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
		SessionID: msg.SessionID,
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
