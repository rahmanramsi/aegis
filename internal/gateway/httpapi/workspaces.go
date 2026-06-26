package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

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

	user := UserFromContext(r)
	var workspaces []store.Workspace
	var err error

	if user != nil {
		workspaces, err = h.Store.ListUserWorkspaces(user.ID)
	} else {
		workspaces, err = h.Store.ListWorkspaces()
	}
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

	// Auto-add creator as admin member
	if user := UserFromContext(r); user != nil {
		if err := h.Store.AddMember(ws.ID, user.ID, "admin"); err != nil {
			slog.Warn("add workspace member", "err", err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if !h.checkAccess(r, id, w) {
		return
	}

	ws, err := h.Store.GetWorkspace(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	json.NewEncoder(w).Encode(ws)
}

func (h *WorkspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if !h.checkAccess(r, id, w) {
		return
	}

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
	id := chi.URLParam(r, "id")

	if !h.checkAccess(r, id, w) {
		return
	}

	if err := h.Store.DeleteWorkspace(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// checkAccess verifies the authenticated user is a member of the workspace.
// Admin key users (no user context) always pass.
func (h *WorkspaceHandler) checkAccess(r *http.Request, workspaceID string, w http.ResponseWriter) bool {
	user := UserFromContext(r)
	if user == nil {
		return true // admin key access
	}
	isMember, err := h.Store.IsMember(workspaceID, user.ID)
	if err != nil || !isMember {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
