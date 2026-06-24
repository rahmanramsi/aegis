package gateway

import (
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/rahmanramsi/aegis/internal/gateway/api"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/gateway/ws"
)

type Server struct {
	Store    *store.Store
	Hub      *ws.Hub
	mux      *http.ServeMux
	staticFS fs.FS
}

func NewServer(s *store.Store, hub *ws.Hub, staticFS fs.FS) *Server {
	server := &Server{
		Store:    s,
		Hub:      hub,
		mux:      http.NewServeMux(),
		staticFS: staticFS,
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
	ah := &api.AgentHandler{Store: s.Store}
	ch := &api.ConnectionHandler{Store: s.Store}
	sh := &api.SessionHandler{Store: s.Store}

	mux := s.mux

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

	// Static files with SPA fallback — must be last
	fileServer := http.FileServer(http.FS(s.staticFS))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		f, err := s.staticFS.Open(path)
		if err != nil {
			r.URL.Path = "/"
		}
		if f != nil {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	})

	// Wrap with auth middleware, then CORS middleware
	authMux := authMiddleware(mux)
	corsMux := corsMiddleware(authMux)
	s.mux = http.NewServeMux()
	s.mux.Handle("/", corsMux)
}

func corsMiddleware(next http.Handler) http.Handler {
	origin := "*"
	if os.Getenv("AEGIS_ENV") == "production" {
		origin = os.Getenv("AEGIS_BASE_URL")
		if origin == "" {
			origin = ""
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
