package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

type WorkspaceHandler struct {
	Store *store.Store
}

type createWorkspaceInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type updateWorkspaceInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	workspaces, err := h.Store.ListWorkspaces()
	if err != nil {
		slog.Error("list workspaces", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	json.NewEncoder(w).Encode(workspaces)
}

func (h *WorkspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var in createWorkspaceInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	ws, err := h.Store.CreateWorkspace(in.Name, in.Slug)
	if err != nil {
		slog.Error("create workspace", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	ws, err := h.Store.GetWorkspace(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	var in updateWorkspaceInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	ws, err := h.Store.UpdateWorkspace(id, in.Name, in.Slug)
	if err != nil {
		slog.Error("update workspace", "err", err, "id", id)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.Store.DeleteWorkspace(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
