package gateway

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/rahmanramsi/aegis/internal/gateway/daemonws"
	"github.com/rahmanramsi/aegis/internal/gateway/httpapi"
	"github.com/rahmanramsi/aegis/internal/gateway/messaging"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

type Options struct {
	APIKey  string
	Env     string
	BaseURL string
}

type Server struct {
	Store      *store.Store
	Hub        *daemonws.Hub
	BotManager *messaging.BotManager
	Options    Options
	router     chi.Router
}

func NewServer(s *store.Store, hub *daemonws.Hub, bm *messaging.BotManager, opts Options) *Server {
	server := &Server{
		Store:      s,
		Hub:        hub,
		BotManager: bm,
		Options:    opts,
	}
	// All daemons are offline on startup — they reconnect via WebSocket
	s.SetAllDaemonsOffline()
	server.registerRoutes()
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(s.corsMiddleware())

	// Public routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler)

		authH := &httpapi.AuthHandler{Store: s.Store}
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware(s.Store, s.Options.APIKey))

			r.Get("/auth/me", authH.Me)

			wh := &httpapi.WorkspaceHandler{Store: s.Store}
			r.Get("/workspaces", wh.List)
			r.Post("/workspaces", wh.Create)
			r.Get("/workspaces/{id}", wh.Get)
			r.Put("/workspaces/{id}", wh.Update)
			r.Delete("/workspaces/{id}", wh.Delete)

			dh := &httpapi.DaemonHandler{Store: s.Store}
			r.Get("/daemons", dh.List)
			r.Post("/daemons", dh.Create)
			r.Get("/daemons/{id}", dh.Get)
			r.Delete("/daemons/{id}", dh.Delete)

			ah := &httpapi.AgentHandler{Store: s.Store, BotManager: s.BotManager}
			r.Get("/workspaces/{wid}/agents", ah.List)
			r.Post("/workspaces/{wid}/agents", ah.Create)
			r.Get("/agents/{id}", ah.Get)
			r.Put("/agents/{id}", ah.Update)
			r.Delete("/agents/{id}", ah.Delete)

			sh := &httpapi.SessionHandler{Store: s.Store}
			r.Get("/agents/{aid}/chats/{cid}/sessions", sh.List)
			r.Get("/sessions/{id}/messages", sh.ListMessages)
		})
	})

	// WebSocket
	r.Get("/ws/daemon", s.Hub.ServeHTTP)

	s.router = r
}

func (s *Server) corsMiddleware() func(http.Handler) http.Handler {
	origin := "*"
	if s.Options.Env == "production" {
		origin = s.Options.BaseURL
	}
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
