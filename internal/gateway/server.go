package gateway

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/rahmanramsi/aegis/internal/gateway/api"
	"github.com/rahmanramsi/aegis/internal/gateway/msg"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/gateway/ws"
)

type Server struct {
	Store      *store.Store
	Hub        *ws.Hub
	BotManager *msg.BotManager
	router     chi.Router
}

func NewServer(s *store.Store, hub *ws.Hub, bm *msg.BotManager) *Server {
	server := &Server{
		Store:      s,
		Hub:        hub,
		BotManager: bm,
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

		authH := &api.AuthHandler{Store: s.Store}
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware(s.Store))

			r.Get("/auth/me", authH.Me)

			wh := &api.WorkspaceHandler{Store: s.Store}
			r.Get("/workspaces", wh.List)
			r.Post("/workspaces", wh.Create)
			r.Get("/workspaces/{id}", wh.Get)
			r.Put("/workspaces/{id}", wh.Update)
			r.Delete("/workspaces/{id}", wh.Delete)

			dh := &api.DaemonHandler{Store: s.Store}
			r.Get("/workspaces/{wid}/daemons", dh.List)
			r.Post("/workspaces/{id}/enrollment-key", wh.GenerateEnrollmentKey)
			r.Post("/workspaces/{wid}/daemons", dh.Create)
			r.Get("/daemons/{id}", dh.Get)
			r.Delete("/daemons/{id}", dh.Delete)

			ah := &api.AgentHandler{Store: s.Store, BotManager: s.BotManager}
			r.Get("/workspaces/{wid}/agents", ah.List)
			r.Post("/workspaces/{wid}/agents", ah.Create)
			r.Get("/agents/{id}", ah.Get)
			r.Put("/agents/{id}", ah.Update)
			r.Delete("/agents/{id}", ah.Delete)

			ch := &api.ConnectionHandler{Store: s.Store}
			r.Get("/agents/{aid}/connections", ch.List)
			r.Post("/agents/{aid}/connections", ch.Create)
			r.Delete("/connections/{id}", ch.Delete)

			sh := &api.SessionHandler{Store: s.Store}
			r.Get("/connections/{cid}/sessions", sh.List)
			r.Get("/sessions/{id}/messages", sh.ListMessages)
		})
	})

	// WebSocket
	r.Get("/ws/daemon", s.Hub.ServeHTTP)

	s.router = r
}

func (s *Server) corsMiddleware() func(http.Handler) http.Handler {
	origin := "*"
	if os.Getenv("AEGIS_ENV") == "production" {
		origin = os.Getenv("AEGIS_BASE_URL")
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
