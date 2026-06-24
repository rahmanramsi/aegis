package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

type AuthHandler struct {
	Store *store.Store
}

type registerInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var in registerInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if in.Email == "" || in.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}
	if len(in.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 6 characters"})
		return
	}

	user, apiKey, err := h.Store.CreateUser(in.Email, in.Password)
	if err == store.ErrEmailTaken {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already registered"})
		return
	}
	if err != nil {
		slog.Error("create user", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":    user,
		"api_key": apiKey,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var in loginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	user, _, err := h.Store.LoginUser(in.Email, in.Password)
	if err == store.ErrInvalidLogin {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}
	if err != nil {
		slog.Error("login user", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	apiKey, err := h.Store.RotateAPIKey(user.ID)
	if err != nil {
		slog.Error("rotate api key", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":    user,
		"api_key": apiKey,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := UserFromContext(r)
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	json.NewEncoder(w).Encode(user)
}
