package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"cofounder-chat/internal/database/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) HandleSendChannelMessage(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	channelID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	// Check membership
	isMember, err := h.Queries.IsChannelMember(r.Context(), sqlc.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    user.ID,
	})
	if err != nil || !isMember {
		writeError(w, http.StatusForbidden, "not a member of this channel")
		return
	}

	var req struct {
		Content  string `json:"content"`
		ParentID string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	var parentID pgtype.UUID
	if req.ParentID != "" {
		parentID, err = parseUUID(req.ParentID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid parent_id")
			return
		}
	}

	msg, err := h.Queries.CreateMessage(r.Context(), sqlc.CreateMessageParams{
		UserID:    user.ID,
		ChannelID: channelID,
		ParentID:  parentID,
		Content:   req.Content,
	})
	if err != nil {
		slog.Error("failed to create message", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to send message")
		return
	}

	// Broadcast via SSE
	h.SSEHub.Broadcast(SSEEvent{
		Type: "new_message",
		Data: map[string]any{
			"message":      msg,
			"channel_id":   channelID,
			"author_name":  user.DisplayName,
			"author_avatar": user.AvatarURL,
		},
	})

	writeJSON(w, http.StatusCreated, msg)
}

func (h *Handler) HandleListChannelMessages(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	channelID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	messages, err := h.Queries.ListChannelMessages(r.Context(), channelID)
	if err != nil {
		slog.Error("failed to list channel messages", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

func (h *Handler) HandleSendConversationMessage(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	convID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	// Check membership
	isMember, err := h.Queries.IsConversationMember(r.Context(), sqlc.IsConversationMemberParams{
		ConversationID: convID,
		UserID:         user.ID,
	})
	if err != nil || !isMember {
		writeError(w, http.StatusForbidden, "not a member of this conversation")
		return
	}

	var req struct {
		Content  string `json:"content"`
		ParentID string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	var parentID pgtype.UUID
	if req.ParentID != "" {
		parentID, err = parseUUID(req.ParentID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid parent_id")
			return
		}
	}

	msg, err := h.Queries.CreateMessage(r.Context(), sqlc.CreateMessageParams{
		UserID:         user.ID,
		ConversationID: convID,
		ParentID:       parentID,
		Content:        req.Content,
	})
	if err != nil {
		slog.Error("failed to create message", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to send message")
		return
	}

	// Broadcast via SSE
	h.SSEHub.Broadcast(SSEEvent{
		Type: "new_message",
		Data: map[string]any{
			"message":         msg,
			"conversation_id": convID,
			"author_name":     user.DisplayName,
			"author_avatar":   user.AvatarURL,
		},
	})

	writeJSON(w, http.StatusCreated, msg)
}

func (h *Handler) HandleListConversationMessages(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	convID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid conversation id")
		return
	}

	// Check membership
	isMember, err := h.Queries.IsConversationMember(r.Context(), sqlc.IsConversationMemberParams{
		ConversationID: convID,
		UserID:         user.ID,
	})
	if err != nil || !isMember {
		writeError(w, http.StatusForbidden, "not a member of this conversation")
		return
	}

	messages, err := h.Queries.ListConversationMessages(r.Context(), convID)
	if err != nil {
		slog.Error("failed to list conversation messages", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

func (h *Handler) HandleUpdateMessage(w http.ResponseWriter, r *http.Request) {
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

	// Verify ownership
	existingMsg, err := h.Queries.GetMessageByID(r.Context(), msgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	if existingMsg.UserID != user.ID {
		writeError(w, http.StatusForbidden, "can only edit your own messages")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	msg, err := h.Queries.UpdateMessage(r.Context(), sqlc.UpdateMessageParams{
		ID:      msgID,
		Content: req.Content,
	})
	if err != nil {
		slog.Error("failed to update message", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update message")
		return
	}

	// Broadcast via SSE
	h.SSEHub.Broadcast(SSEEvent{
		Type: "message_updated",
		Data: msg,
	})

	writeJSON(w, http.StatusOK, msg)
}

func (h *Handler) HandleDeleteMessage(w http.ResponseWriter, r *http.Request) {
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

	// Verify ownership
	existingMsg, err := h.Queries.GetMessageByID(r.Context(), msgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}
	if existingMsg.UserID != user.ID {
		writeError(w, http.StatusForbidden, "can only delete your own messages")
		return
	}

	err = h.Queries.DeleteMessage(r.Context(), msgID)
	if err != nil {
		slog.Error("failed to delete message", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete message")
		return
	}

	// Broadcast via SSE
	h.SSEHub.Broadcast(SSEEvent{
		Type: "message_deleted",
		Data: map[string]any{
			"id":              msgID,
			"channel_id":      existingMsg.ChannelID,
			"conversation_id": existingMsg.ConversationID,
		},
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) HandleListThread(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	parentID, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid message id")
		return
	}

	// Get parent message
	parent, err := h.Queries.GetMessageByID(r.Context(), parentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "message not found")
		return
	}

	// Get thread replies
	replies, err := h.Queries.ListThreadReplies(r.Context(), parentID)
	if err != nil {
		slog.Error("failed to list thread replies", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list thread")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"parent":  parent,
		"replies": replies,
	})
}
