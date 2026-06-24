package gateway

import (
	"net/http"
	"os"
	"strings"
)

func authMiddleware(next http.Handler) http.Handler {
	apiKey := os.Getenv("AEGIS_API_KEY")
	if apiKey == "" {
		// No auth configured — allow all
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public paths: health, GET read endpoints, WebSocket
		if isPublicPath(r) {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Bearer ") || strings.TrimPrefix(token, "Bearer ") != apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isPublicPath(r *http.Request) bool {
	path := r.URL.Path

	// Health endpoint is always public
	if path == "/api/v1/health" {
		return true
	}

	// WebSocket handshake (auth handled by daemon token)
	if strings.HasPrefix(path, "/ws/") {
		return true
	}

	// Static files (non-API paths)
	if !strings.HasPrefix(path, "/api/") {
		return true
	}

	// All /api/* paths require auth when AEGIS_API_KEY is set
	return false
}
