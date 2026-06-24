package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

type AgentHandler struct {
	Store      *store.Store
	BotManager interface {
		AddBot(ctx context.Context, token string) error
		RemoveBotByHash(hash string)
	}
}

type createAgentInput struct {
	DaemonID      string `json:"daemon_id"`
	Name          string `json:"name"`
	Harness       string `json:"harness"`
	Model         string `json:"model"`
	ExtraArgs     string `json:"extra_args"`
	Enabled       *bool  `json:"enabled"`
	Personality   string `json:"personality"`
	TelegramToken string `json:"telegram_token"`
}

type updateAgentInput struct {
	Name        string `json:"name"`
	Harness     string `json:"harness"`
	Model       string `json:"model"`
	ExtraArgs   string `json:"extra_args"`
	Enabled     *bool  `json:"enabled"`
	Personality string `json:"personality"`
}

func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	wid := chi.URLParam(r, "wid")
	agents, err := h.Store.ListAgents(wid)
	if err != nil {
		slog.Error("list agents", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	json.NewEncoder(w).Encode(agents)
}

func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	wid := chi.URLParam(r, "wid")

	user := store.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return
	}

	var in createAgentInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	daemon, err := h.Store.GetDaemon(in.DaemonID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "daemon not found"})
		return
	}
	if daemon.UserID != user.ID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "daemon does not belong to your account"})
		return
	}

	// Validate: harness must exist in daemon_harnesses
	harnesses, err := h.Store.GetDaemonHarnesses(in.DaemonID)
	if err != nil {
		slog.Error("get daemon harnesses", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	found := false
	for _, h := range harnesses {
		if h == in.Harness {
			found = true
			break
		}
	}
	if !found {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "harness not available on daemon"})
		return
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}

	tokenHash := ""
	if in.TelegramToken != "" {
		tokenHash = sha256Hex(in.TelegramToken)
	}

	agent, err := h.Store.CreateAgent(wid, in.DaemonID, in.Name, in.Harness, in.Model, in.ExtraArgs, in.Personality, tokenHash, enabled)
	if err != nil {
		slog.Error("create agent", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	// Start Telegram bot for this agent
	if in.TelegramToken != "" && h.BotManager != nil {
		if err := h.BotManager.AddBot(context.Background(), in.TelegramToken); err != nil {
			slog.Warn("start bot", "err", err)
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent":          agent,
		"telegram_token": in.TelegramToken,
	})

}

func (h *AgentHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")
	agent, err := h.Store.GetAgent(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	json.NewEncoder(w).Encode(agent)
}

func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	var in updateAgentInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	// Get existing agent to preserve fields if not provided
	existing, err := h.Store.GetAgent(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	name := existing.Name
	harness := existing.Harness
	model := existing.Model
	extraArgs := existing.ExtraArgs
	personality := existing.Personality
	enabled := existing.Enabled

	if in.Name != "" {
		name = in.Name
	}
	if in.Harness != "" {
		harness = in.Harness
	}
	if in.Model != "" {
		model = in.Model
	}
	if in.ExtraArgs != "" {
		extraArgs = in.ExtraArgs
	}
	if in.Personality != "" {
		personality = in.Personality
	}
	if in.Enabled != nil {
		enabled = *in.Enabled
	}

	agent, err := h.Store.UpdateAgent(id, name, harness, model, extraArgs, personality, enabled)
	if err != nil {
		slog.Error("update agent", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	json.NewEncoder(w).Encode(agent)
}

func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	agent, err := h.Store.GetAgent(id)
	if err == nil && agent.TelegramTokenHash != "" && h.BotManager != nil {
		h.BotManager.RemoveBotByHash(agent.TelegramTokenHash)
	}

	if err := h.Store.DeleteAgent(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
