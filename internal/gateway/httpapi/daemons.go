package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

type DaemonHandler struct {
	Store *store.Store
}

type createDaemonInput struct {
	Name string `json:"name"`
}

type createDaemonResponse struct {
	Daemon store.Daemon `json:"daemon"`
	Token  string       `json:"token"`
}

type daemonWithHarnesses struct {
	store.Daemon
	Harnesses     []string            `json:"harnesses"`
	HarnessModels map[string][]string `json:"harness_models"`
}

func (h *DaemonHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := store.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	daemons, err := h.Store.ListDaemonsByUser(user.ID)
	if err != nil {
		slog.Error("list daemons", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	out := make([]daemonWithHarnesses, len(daemons))
	for i, d := range daemons {
		harns, _ := h.Store.GetDaemonHarnesses(d.ID)
		var models map[string][]string
		if d.HarnessModelsJSON != "" {
			json.Unmarshal([]byte(d.HarnessModelsJSON), &models)
		}
		out[i] = daemonWithHarnesses{Daemon: d, Harnesses: harns, HarnessModels: models}
	}
	json.NewEncoder(w).Encode(out)
}

func (h *DaemonHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := store.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var in createDaemonInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	// Generate random 32-byte token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		slog.Error("generate token", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	token := hex.EncodeToString(tokenBytes)

	// Hash token for storage
	tokenHash := sha256Hex(token)

	d, err := h.Store.CreateDaemon(user.ID, in.Name, tokenHash)
	if err != nil {
		slog.Error("create daemon", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createDaemonResponse{
		Daemon: *d,
		Token:  token,
	})
}

func (h *DaemonHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	d, err := h.Store.GetDaemon(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	harnesses, err := h.Store.GetDaemonHarnesses(id)
	if err != nil {
		slog.Error("get daemon harnesses", "err", err)
		harnesses = []string{}
	}

	json.NewEncoder(w).Encode(daemonWithHarnesses{
		Daemon:    *d,
		Harnesses: harnesses,
	})
}

func (h *DaemonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteDaemon(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
