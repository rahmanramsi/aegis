package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

type SessionHandler struct {
	Store *store.Store
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cid := r.PathValue("cid")
	sessions, err := h.Store.ListSessions(cid)
	if err != nil {
		slog.Error("list sessions", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

func (h *SessionHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")

	limit := 50
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	messages, err := h.Store.ListMessages(id, limit, offset)
	if err != nil {
		slog.Error("list messages", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	json.NewEncoder(w).Encode(messages)
}
