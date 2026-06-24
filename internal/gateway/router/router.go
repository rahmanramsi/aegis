package router

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/rahmanramsi/aegis/internal/gateway/msg"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/gateway/ws"
	"github.com/rahmanramsi/aegis/internal/shared/protocol"
)

type pendingTask struct {
	adapter   msg.Adapter
	chatID    string
	sessionID string
	agentID   string
}

type Router struct {
	Store   *store.Store
	Hub     *ws.Hub
	mu      sync.Mutex
	pending map[string]*pendingTask // taskID → callback info
}

func NewRouter(s *store.Store, hub *ws.Hub) *Router {
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
		pt.adapter.Send(pt.chatID, event.Content)
	case protocol.TypeStderr:
		pt.adapter.Send(pt.chatID, "[stderr] "+event.Content)
	case protocol.TypeDone:
		pt.adapter.Send(pt.chatID, "[done]")
	case protocol.TypeError:
		pt.adapter.Send(pt.chatID, "[error] "+event.Content)
	}
}

func (r *Router) Handle(ctx context.Context, m msg.Message, adapter msg.Adapter) {
	connection, err := r.Store.FindConnection(m.Platform, m.ChatID)
	if err != nil || connection == nil {
		adapter.Send(m.ChatID, "No agent connected to this chat. Set up a connection in the web dashboard.")
		return
	}

	agent, err := r.Store.GetAgent(connection.AgentID)
	if err != nil || agent == nil || !agent.Enabled {
		adapter.Send(m.ChatID, "Agent not found or disabled.")
		return
	}

	daemon, err := r.Store.GetDaemon(agent.DaemonID)
	if err != nil || daemon == nil || daemon.Status != "online" {
		adapter.Send(m.ChatID, "Agent daemon is offline.")
		return
	}

	session, err := r.Store.GetOrCreateSession(connection.ID, m.UserName)
	if err != nil {
		slog.Warn("router: get/create session failed", "err", err)
		adapter.Send(m.ChatID, "Internal error.")
		return
	}

	_, err = r.Store.CreateMessage(session.ID, "user", m.Text, agent.ID)
	if err != nil {
		slog.Warn("router: create message failed", "err", err)
	}

	taskID := uuid.NewString()
	taskMsg := protocol.Message{
		Type:    protocol.TypeTask,
		TaskID:  taskID,
		Harness: agent.Harness,
		Prompt:  m.Text,
		Model:   agent.Model,
	}
	if agent.ExtraArgs != "" {
		taskMsg.ExtraArgs = strings.Fields(agent.ExtraArgs)
	}

	// Register pending task for response routing
	r.mu.Lock()
	r.pending[taskID] = &pendingTask{
		adapter:   adapter,
		chatID:    m.ChatID,
		sessionID: session.ID,
		agentID:   agent.ID,
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
