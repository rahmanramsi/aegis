package httpapi

import (
	"net/http"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

func UserFromContext(r *http.Request) *store.User {
	return store.UserFromContext(r.Context())
}
