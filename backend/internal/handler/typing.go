package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) HandleTyping(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		ChannelID      string `json:"channel_id,omitempty"`
		ConversationID string `json:"conversation_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ChannelID == "" && req.ConversationID == "" {
		writeError(w, http.StatusBadRequest, "channel_id or conversation_id is required")
		return
	}

	event := SSEEvent{
		Type: "typing",
		Data: map[string]any{
			"user_id":         user.ID,
			"display_name":    user.DisplayName,
			"channel_id":      req.ChannelID,
			"conversation_id": req.ConversationID,
		},
	}

	h.SSEHub.Broadcast(event)

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
