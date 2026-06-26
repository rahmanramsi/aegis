package daemonws

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/coder/websocket"

	"github.com/rahmanramsi/aegis/internal/protocol"
)

type DaemonConn struct {
	ID   string
	Conn *websocket.Conn
	hub  *Hub
}

func (dc *DaemonConn) readLoop(ctx context.Context) {
	defer dc.hub.removeDaemon(dc.ID)
	for {
		_, data, err := dc.Conn.Read(ctx)
		if err != nil {
			return
		}
		msg, err := protocol.DecodeMessage(bytes.TrimSpace(data))
		if err != nil {
			continue
		}
		slog.Debug("daemon event", "id", dc.ID, "type", msg.Type, "task", msg.TaskID)
		if dc.hub.OnTaskEvent != nil && msg.TaskID != "" {
			dc.hub.OnTaskEvent(msg.TaskID, msg)
		}
	}
}
