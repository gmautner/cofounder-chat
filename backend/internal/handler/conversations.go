package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"cofounder-chat/internal/database/sqlc"
)

func (h *Handler) HandleCreateConversation(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		UserID string `json:"user_id"` // the other user to start a conversation with
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	otherUserID, err := parseUUID(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	// Check if a conversation already exists between these two users
	existingConvID, err := h.Queries.FindExistingConversation(r.Context(), sqlc.FindExistingConversationParams{
		UserID:   user.ID,
		UserID_2: otherUserID,
	})
	if err == nil {
		// Conversation already exists, return it
		writeJSON(w, http.StatusOK, map[string]any{"id": existingConvID, "existing": true})
		return
	}

	// Create new conversation
	conv, err := h.Queries.CreateConversation(r.Context())
	if err != nil {
		slog.Error("failed to create conversation", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	// Add both members
	err = h.Queries.AddConversationMember(r.Context(), sqlc.AddConversationMemberParams{
		ConversationID: conv.ID,
		UserID:         user.ID,
	})
	if err != nil {
		slog.Error("failed to add conversation member", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	err = h.Queries.AddConversationMember(r.Context(), sqlc.AddConversationMemberParams{
		ConversationID: conv.ID,
		UserID:         otherUserID,
	})
	if err != nil {
		slog.Error("failed to add conversation member", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	writeJSON(w, http.StatusCreated, conv)
}

func (h *Handler) HandleListConversations(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	conversations, err := h.Queries.ListUserConversations(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to list conversations", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list conversations")
		return
	}

	// Enrich each conversation with member info
	type ConversationWithMembers struct {
		ID        any    `json:"id"`
		CreatedAt any    `json:"created_at"`
		Members   []any  `json:"members"`
	}

	var result []ConversationWithMembers
	for _, conv := range conversations {
		members, err := h.Queries.ListConversationMembers(r.Context(), conv.ID)
		if err != nil {
			slog.Error("failed to list conversation members", "err", err)
			continue
		}

		memberList := make([]any, 0, len(members))
		for _, m := range members {
			memberList = append(memberList, map[string]any{
				"id":           m.ID,
				"display_name": m.DisplayName,
				"avatar_url":   m.AvatarUrl,
				"email":        m.Email,
			})
		}

		result = append(result, ConversationWithMembers{
			ID:        conv.ID,
			CreatedAt: conv.CreatedAt,
			Members:   memberList,
		})
	}

	if result == nil {
		result = []ConversationWithMembers{}
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) HandleListConversationMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.Queries.ListConversationMembers(r.Context(), convID)
	if err != nil {
		slog.Error("failed to list conversation members", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	writeJSON(w, http.StatusOK, members)
}
