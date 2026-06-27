package daemonws

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/protocol"
)

func TestHubAcceptsDaemonEnrollmentToken(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	user, _, err := s.CreateUser("daemon@test.dev", "password")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	daemonToken := "daemon-enrollment-token"
	daemon, err := s.CreateDaemon(user.ID, "test-daemon", sha256Hex(daemonToken))
	if err != nil {
		t.Fatalf("create daemon: %v", err)
	}

	hub := NewHub(s)
	server := httptest.NewServer(hub)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial hub: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	err = wsjson.Write(ctx, conn, protocol.Message{
		Type:       protocol.TypeHandshake,
		DaemonID:   daemon.ID,
		Token:      daemonToken,
		DaemonName: daemon.Name,
		Harnesses:  []string{"echo"},
	})
	if err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	var response protocol.Message
	if err := wsjson.Read(ctx, conn, &response); err != nil {
		t.Fatalf("read handshake response: %v", err)
	}
	if response.Type != protocol.TypeHandshakeOK {
		t.Fatalf("handshake response type = %q, want %q", response.Type, protocol.TypeHandshakeOK)
	}

	got, err := s.GetDaemon(daemon.ID)
	if err != nil {
		t.Fatalf("get daemon: %v", err)
	}
	if got.Status != "online" {
		t.Fatalf("daemon status = %q, want online", got.Status)
	}
}

func TestHubAutoEnrollsMissingDaemonWithUserAPIKey(t *testing.T) {
	s, err := store.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	user, apiKey, err := s.CreateUser("auto@test.dev", "password")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	hub := NewHub(s)
	server := httptest.NewServer(hub)
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial hub: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	err = wsjson.Write(ctx, conn, protocol.Message{
		Type:       protocol.TypeHandshake,
		DaemonID:   "locally-generated-id",
		Token:      apiKey,
		DaemonName: "auto-daemon",
		Harnesses:  []string{"echo"},
	})
	if err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	var response protocol.Message
	if err := wsjson.Read(ctx, conn, &response); err != nil {
		t.Fatalf("read handshake response: %v", err)
	}
	if response.Type != protocol.TypeHandshakeOK {
		t.Fatalf("handshake response type = %q, want %q", response.Type, protocol.TypeHandshakeOK)
	}

	created, err := s.GetDaemonByUserAndName(user.ID, "auto-daemon")
	if err != nil {
		t.Fatalf("get auto-created daemon: %v", err)
	}
	if created.Status != "online" {
		t.Fatalf("created daemon status = %q, want online", created.Status)
	}
}
