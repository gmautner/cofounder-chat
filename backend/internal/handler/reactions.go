package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"cofounder-chat/internal/database/sqlc"
)

func (h *Handler) HandleAddReaction(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	msgID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Emoji == "" {
		writeError(w, http.StatusBadRequest, "emoji is required")
		return
	}

	reaction, err := h.Queries.AddReaction(r.Context(), sqlc.AddReactionParams{
		MessageID: msgID,
		UserID:    user.ID,
		Emoji:     req.Emoji,
	})
	if err != nil {
		// Could be a conflict (already reacted) — pgx returns no rows on DO NOTHING
		slog.Error("failed to add reaction", "err", err)
		writeError(w, http.StatusConflict, "reaction already exists")
		return
	}

	// Broadcast via SSE
	h.SSEHub.Broadcast(SSEEvent{
		Type: "reaction_added",
		Data: map[string]any{
			"message_id":   msgID,
			"emoji":        req.Emoji,
			"user_id":      user.ID,
			"display_name": user.DisplayName,
		},
	})

	writeJSON(w, http.StatusCreated, reaction)
}

func (h *Handler) HandleRemoveReaction(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	msgID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	emoji := r.PathValue("emoji")
	if emoji == "" {
		writeError(w, http.StatusBadRequest, "emoji is required")
		return
	}

	err = h.Queries.RemoveReaction(r.Context(), sqlc.RemoveReactionParams{
		MessageID: msgID,
		UserID:    user.ID,
		Emoji:     emoji,
	})
	if err != nil {
		slog.Error("failed to remove reaction", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to remove reaction")
		return
	}

	// Broadcast via SSE
	h.SSEHub.Broadcast(SSEEvent{
		Type: "reaction_removed",
		Data: map[string]any{
			"message_id": msgID,
			"emoji":      emoji,
			"user_id":    user.ID,
		},
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *Handler) HandleListReactions(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	msgID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	reactions, err := h.Queries.ListMessageReactions(r.Context(), msgID)
	if err != nil {
		slog.Error("failed to list reactions", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list reactions")
		return
	}

	writeJSON(w, http.StatusOK, reactions)
}
