package routing

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/rahmanramsi/aegis/internal/gateway/daemonws"
	"github.com/rahmanramsi/aegis/internal/gateway/messaging"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/protocol"
)

type pendingTask struct {
	adapter   messaging.Adapter
	chatID    string
	sessionID string
	agentID   string
	stream    messaging.StreamSender
}

type Router struct {
	Store   *store.Store
	Hub     *daemonws.Hub
	mu      sync.Mutex
	pending map[string]*pendingTask // taskID → callback info
}

func NewRouter(s *store.Store, hub *daemonws.Hub) *Router {
	r := &Router{
		Store:   s,
		Hub:     hub,
		pending: make(map[string]*pendingTask),
	}
	hub.OnTaskEvent = r.onTaskEvent
	return r
}

func (r *Router) onTaskEvent(taskID string, event protocol.Message) {
	r.mu.Lock()
	pt, ok := r.pending[taskID]
	if event.Type == protocol.TypeDone || event.Type == protocol.TypeError {
		delete(r.pending, taskID)
	}
	r.mu.Unlock()

	if !ok {
		return
	}
	switch event.Type {
	case protocol.TypeStdout:
		pt.stream.Append(event.Content)
	case protocol.TypeStderr:
		// ignore — most CLIs mirror stdout to stderr
	case protocol.TypeDone:
		pt.stream.Done()
		// Save assistant response to session
		r.Store.CreateMessage(pt.sessionID, "assistant", event.Content, pt.agentID)
	case protocol.TypeError:
		pt.stream.Error(event.Content)
	}
}

func (r *Router) Handle(ctx context.Context, m messaging.Message, adapter messaging.Adapter) {
	adapter.Send(m.ChatID, "Direct routing not configured. Use an agent via web dashboard.")
}

func (r *Router) HandleWithAgent(ctx context.Context, m messaging.Message, adapter messaging.Adapter, agent *store.Agent) {
	// Show typing indicator while processing
	if ta, ok := adapter.(interface{ SendTyping(string) error }); ok {
		ta.SendTyping(m.ChatID)
	}

	if agent == nil || !agent.Enabled {
		adapter.Send(m.ChatID, "Agent not found or disabled.")
		return
	}

	daemon, err := r.Store.GetDaemon(agent.DaemonID)
	if err != nil || daemon == nil || daemon.Status != "online" {
		adapter.Send(m.ChatID, "Agent daemon is offline.")
		return
	}

	session, err := r.Store.GetOrCreateSessionByAgent(agent.ID, m.ChatID, m.UserName)
	if err != nil {
		slog.Warn("router: get/create session failed", "err", err)
		adapter.Send(m.ChatID, "Internal error.")
		return
	}

	_, err = r.Store.CreateMessage(session.ID, "user", m.Text, agent.ID)
	if err != nil {
		slog.Warn("router: create message failed", "err", err)
	}

	// Build prompt with personality if set
	prompt := m.Text
	if agent.Personality != "" {
		prompt = agent.Personality + "\n\nUser: " + m.Text
	}

	taskID := uuid.NewString()
	taskMsg := protocol.Message{
		Type:      protocol.TypeTask,
		TaskID:    taskID,
		Harness:   agent.Harness,
		Prompt:    prompt,
		Model:     agent.Model,
		SessionID: session.ID,
	}
	if agent.ExtraArgs != "" {
		taskMsg.ExtraArgs = strings.Fields(agent.ExtraArgs)
	}

	r.mu.Lock()
	r.pending[taskID] = &pendingTask{
		adapter:   adapter,
		chatID:    m.ChatID,
		sessionID: session.ID,
		agentID:   agent.ID,
		stream:    adapter.SendStream(m.ChatID),
	}
	r.mu.Unlock()

	if err := r.Hub.SendTask(agent.DaemonID, taskMsg); err != nil {
		r.mu.Lock()
		delete(r.pending, taskID)
		r.mu.Unlock()
		slog.Warn("router: dispatch failed", "daemon", agent.DaemonID, "err", err)
		adapter.Send(m.ChatID, "Failed to dispatch task: "+err.Error())
		return
	}

	slog.Info("router: task dispatched", "task_id", taskID, "agent", agent.ID, "daemon", agent.DaemonID, "chat", m.ChatID)
}
