package gateway

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

func authMiddleware(next http.Handler, s *store.Store) http.Handler {
	adminKey := os.Getenv("AEGIS_API_KEY")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r) {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			writeAuthError(w, "missing API key")
			return
		}

		if adminKey != "" && token == adminKey {
			next.ServeHTTP(w, r)
			return
		}

		user, err := s.VerifyAPIKey(token)
		if err != nil {
			writeAuthError(w, "invalid API key")
			return
		}

		ctx := context.WithValue(r.Context(), store.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}

func isPublicPath(r *http.Request) bool {
	path := r.URL.Path

	if path == "/api/v1/health" {
		return true
	}
	if strings.HasPrefix(path, "/api/v1/auth/") {
		return true
	}
	if strings.HasPrefix(path, "/ws/") {
		return true
	}
	if !strings.HasPrefix(path, "/api/") {
		return true
	}
	return false
}
