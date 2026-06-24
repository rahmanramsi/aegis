package gateway

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

func authMiddleware(s *store.Store) func(http.Handler) http.Handler {
	adminKey := os.Getenv("AEGIS_API_KEY")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}
