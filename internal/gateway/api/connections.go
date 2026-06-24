package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

type ConnectionHandler struct {
	Store *store.Store
}

type createConnectionInput struct {
	Platform string `json:"platform"`
	ChatID   string `json:"chat_id"`
}

func (h *ConnectionHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	aid := r.PathValue("aid")
	connections, err := h.Store.ListConnections(aid)
	if err != nil {
		slog.Error("list connections", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	json.NewEncoder(w).Encode(connections)
}

func (h *ConnectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	aid := r.PathValue("aid")

	var in createConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	conn, err := h.Store.CreateConnection(aid, in.Platform, in.ChatID)
	if err != nil {
		slog.Error("create connection", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conn)
}

func (h *ConnectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.Store.DeleteConnection(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
