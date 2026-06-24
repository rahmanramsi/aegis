package ws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/protocol"
)

type TaskCallback func(taskID string, event protocol.Message)

type Hub struct {
	Store        *store.Store
	mu           sync.RWMutex
	daemons      map[string]*DaemonConn
	OnTaskEvent  TaskCallback // called when daemon sends stdout/done/error for a task
}

func NewHub(s *store.Store) *Hub {
	return &Hub{
		Store:   s,
		daemons: make(map[string]*DaemonConn),
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("ws handler panic", "panic", rec)
		}
	}()
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.Error("ws accept", "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var msg protocol.Message
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		slog.Error("read handshake", "err", err)
		conn.Close(websocket.StatusPolicyViolation, "expected handshake")
		return
	}

	if msg.Type != protocol.TypeHandshake {
		slog.Warn("expected handshake", "got", msg.Type)
		wsjson.Write(ctx, conn, protocol.Message{Type: protocol.TypeError, Content: "expected handshake"})
		conn.Close(websocket.StatusPolicyViolation, "expected handshake")
		return
	}

	tokenHash := sha256Hex(msg.Token)
	user, err := h.Store.GetUserByAPIKey(tokenHash)
	if err != nil {
		slog.Warn("ws: invalid API key", "err", err)
		wsjson.Write(ctx, conn, protocol.Message{Type: protocol.TypeError, Content: "invalid API key"})
		conn.Close(websocket.StatusPolicyViolation, "invalid API key")
		return
	}

	daemonID := msg.DaemonID
	if daemonID == "" {
		daemonID = uuid.NewString()
	}

	d, err := h.Store.GetDaemon(daemonID)
	if err != nil {
		name := msg.DaemonName
		if name == "" {
			name = "auto"
		}
		// Try to find existing daemon by user+name (daemon regenerates UUID each start)
		d, err = h.Store.GetDaemonByUserAndName(user.ID, name)
		if err != nil {
			d, err = h.Store.CreateDaemon(user.ID, name, tokenHash)
			if err != nil {
				slog.Error("ws: create daemon", "err", err)
				wsjson.Write(ctx, conn, protocol.Message{Type: protocol.TypeError, Content: "failed to create daemon"})
				conn.Close(websocket.StatusInternalError, "create daemon failed")
				return
			}
		}
		daemonID = d.ID
	}
	msg.DaemonID = daemonID

	modelsJSON, _ := json.Marshal(msg.HarnessModels)
	if err := h.Store.AuthenticateDaemon(daemonID, tokenHash, msg.Harnesses, string(modelsJSON)); err != nil {
		slog.Error("auth daemon", "err", err, "daemon_id", daemonID)
		wsjson.Write(ctx, conn, protocol.Message{Type: protocol.TypeError, Content: err.Error()})
		conn.Close(websocket.StatusPolicyViolation, "auth failed")
		return
	}

	wsjson.Write(ctx, conn, protocol.Message{Type: protocol.TypeHandshakeOK})

	dc := &DaemonConn{
		ID:   msg.DaemonID,
		Conn: conn,
		hub:  h,
	}

	h.mu.Lock()
	h.daemons[msg.DaemonID] = dc
	h.mu.Unlock()

	slog.Info("daemon connected", "id", msg.DaemonID, "harnesses", msg.Harnesses)

	go dc.readLoop(context.Background())
}

func (h *Hub) SendTask(daemonID string, msg protocol.Message) error {
	h.mu.RLock()
	dc, ok := h.daemons[daemonID]
	h.mu.RUnlock()
	if !ok {
		return fmt.Errorf("daemon %s not connected", daemonID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return wsjson.Write(ctx, dc.Conn, msg)
}

func (h *Hub) removeDaemon(id string) {
	h.mu.Lock()
	delete(h.daemons, id)
	h.mu.Unlock()
	h.Store.SetDaemonOffline(id)
	slog.Info("daemon disconnected", "id", id)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
