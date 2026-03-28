package handler

import (
	"log/slog"
	"net/http"

	"cofounder-chat/internal/database/sqlc"
)

func (h *Handler) HandleGetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	type UnreadCount struct {
		ID    any   `json:"id"`
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	var counts []UnreadCount

	// Get unread counts for all user channels
	channels, err := h.Queries.ListUserChannels(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to list user channels", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get unread counts")
		return
	}

	for _, ch := range channels {
		lastRead, err := h.Queries.GetChannelLastRead(r.Context(), sqlc.GetChannelLastReadParams{
			ChannelID: ch.ID,
			UserID:    user.ID,
		})
		if err != nil {
			continue
		}

		count, err := h.Queries.CountUnreadChannelMessages(r.Context(), sqlc.CountUnreadChannelMessagesParams{
			ChannelID: ch.ID,
			CreatedAt: lastRead,
		})
		if err != nil {
			continue
		}

		if count > 0 {
			counts = append(counts, UnreadCount{
				ID:    ch.ID,
				Type:  "channel",
				Count: count,
			})
		}
	}

	// Get unread counts for all user conversations
	conversations, err := h.Queries.ListUserConversations(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to list user conversations", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to get unread counts")
		return
	}

	for _, conv := range conversations {
		lastRead, err := h.Queries.GetConversationLastRead(r.Context(), sqlc.GetConversationLastReadParams{
			ConversationID: conv.ID,
			UserID:         user.ID,
		})
		if err != nil {
			continue
		}

		count, err := h.Queries.CountUnreadConversationMessages(r.Context(), sqlc.CountUnreadConversationMessagesParams{
			ConversationID: conv.ID,
			CreatedAt:      lastRead,
		})
		if err != nil {
			continue
		}

		if count > 0 {
			counts = append(counts, UnreadCount{
				ID:    conv.ID,
				Type:  "conversation",
				Count: count,
			})
		}
	}

	if counts == nil {
		counts = []UnreadCount{}
	}

	writeJSON(w, http.StatusOK, counts)
}

func (h *Handler) HandleMarkChannelRead(w http.ResponseWriter, r *http.Request) {
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

	err = h.Queries.UpdateChannelLastRead(r.Context(), sqlc.UpdateChannelLastReadParams{
		ChannelID: channelID,
		UserID:    user.ID,
	})
	if err != nil {
		slog.Error("failed to mark channel as read", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to mark as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

func (h *Handler) HandleMarkConversationRead(w http.ResponseWriter, r *http.Request) {
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

	err = h.Queries.UpdateConversationLastRead(r.Context(), sqlc.UpdateConversationLastReadParams{
		ConversationID: convID,
		UserID:         user.ID,
	})
	if err != nil {
		slog.Error("failed to mark conversation as read", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to mark as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "read"})
}
