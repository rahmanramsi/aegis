package gateway

import (
	"net/http"
	"os"

	"github.com/rahmanramsi/aegis/internal/gateway/api"
	"github.com/rahmanramsi/aegis/internal/gateway/msg"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/gateway/ws"
)

type Server struct {
	Store      *store.Store
	Hub        *ws.Hub
	BotManager *msg.BotManager
	mux        *http.ServeMux
}

func NewServer(s *store.Store, hub *ws.Hub, bm *msg.BotManager) *Server {
	server := &Server{
		Store:      s,
		Hub:        hub,
		BotManager: bm,
		mux:        http.NewServeMux(),
	}
	server.registerRoutes()
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	wh := &api.WorkspaceHandler{Store: s.Store}
	dh := &api.DaemonHandler{Store: s.Store}
	ah := &api.AgentHandler{Store: s.Store, BotManager: s.BotManager}
	ch := &api.ConnectionHandler{Store: s.Store}
	sh := &api.SessionHandler{Store: s.Store}

	mux := s.mux

	// Auth (public)
	authH := &api.AuthHandler{Store: s.Store}
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("GET /api/v1/auth/me", authH.Me)

	// Health
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Workspaces
	mux.HandleFunc("GET /api/v1/workspaces", wh.List)
	mux.HandleFunc("POST /api/v1/workspaces", wh.Create)
	mux.HandleFunc("GET /api/v1/workspaces/{id}", wh.Get)
	mux.HandleFunc("PUT /api/v1/workspaces/{id}", wh.Update)
	mux.HandleFunc("DELETE /api/v1/workspaces/{id}", wh.Delete)

	// Daemons
	mux.HandleFunc("GET /api/v1/workspaces/{wid}/daemons", dh.List)
	mux.HandleFunc("POST /api/v1/workspaces/{wid}/daemons", dh.Create)
	mux.HandleFunc("GET /api/v1/daemons/{id}", dh.Get)
	mux.HandleFunc("DELETE /api/v1/daemons/{id}", dh.Delete)

	// Agents
	mux.HandleFunc("GET /api/v1/workspaces/{wid}/agents", ah.List)
	mux.HandleFunc("POST /api/v1/workspaces/{wid}/agents", ah.Create)
	mux.HandleFunc("GET /api/v1/agents/{id}", ah.Get)
	mux.HandleFunc("PUT /api/v1/agents/{id}", ah.Update)
	mux.HandleFunc("DELETE /api/v1/agents/{id}", ah.Delete)

	// Connections
	mux.HandleFunc("GET /api/v1/agents/{aid}/connections", ch.List)
	mux.HandleFunc("POST /api/v1/agents/{aid}/connections", ch.Create)
	mux.HandleFunc("DELETE /api/v1/connections/{id}", ch.Delete)

	// Sessions
	mux.HandleFunc("GET /api/v1/connections/{cid}/sessions", sh.List)
	mux.HandleFunc("GET /api/v1/sessions/{id}/messages", sh.ListMessages)

	// WebSocket
	mux.HandleFunc("GET /ws/daemon", s.Hub.ServeHTTP)

	// Wrap with auth middleware, then CORS
	authMux := authMiddleware(mux, s.Store)
	corsMux := corsMiddleware(authMux)
	s.mux = http.NewServeMux()
	s.mux.Handle("/", corsMux)
}

func corsMiddleware(next http.Handler) http.Handler {
	origin := "*"
	if os.Getenv("AEGIS_ENV") == "production" {
		origin = os.Getenv("AEGIS_BASE_URL")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
