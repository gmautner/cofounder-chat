package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"cofounder-chat/internal/database/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) HandleCreateChannel(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	channel, err := h.Queries.CreateChannel(r.Context(), sqlc.CreateChannelParams{
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
		CreatedBy:   user.ID,
	})
	if err != nil {
		slog.Error("failed to create channel", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create channel")
		return
	}

	// Add creator as member
	err = h.Queries.AddChannelMember(r.Context(), sqlc.AddChannelMemberParams{
		ChannelID: channel.ID,
		UserID:    user.ID,
	})
	if err != nil {
		slog.Error("failed to add channel creator as member", "err", err)
	}

	writeJSON(w, http.StatusCreated, channel)
}

func (h *Handler) HandleListChannels(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	// Return channels the user is a member of
	channels, err := h.Queries.ListUserChannels(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to list channels", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list channels")
		return
	}

	writeJSON(w, http.StatusOK, channels)
}

func (h *Handler) HandleGetChannel(w http.ResponseWriter, r *http.Request) {
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

	channel, err := h.Queries.GetChannelByID(r.Context(), channelID)
	if err != nil {
		writeError(w, http.StatusNotFound, "channel not found")
		return
	}

	writeJSON(w, http.StatusOK, channel)
}

func (h *Handler) HandleJoinChannel(w http.ResponseWriter, r *http.Request) {
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

	// Check channel exists and is not private
	channel, err := h.Queries.GetChannelByID(r.Context(), channelID)
	if err != nil {
		writeError(w, http.StatusNotFound, "channel not found")
		return
	}

	if channel.IsPrivate {
		writeError(w, http.StatusForbidden, "cannot join private channel")
		return
	}

	err = h.Queries.AddChannelMember(r.Context(), sqlc.AddChannelMemberParams{
		ChannelID: channelID,
		UserID:    user.ID,
	})
	if err != nil {
		slog.Error("failed to join channel", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to join channel")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "joined"})
}

func (h *Handler) HandleLeaveChannel(w http.ResponseWriter, r *http.Request) {
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

	err = h.Queries.RemoveChannelMember(r.Context(), sqlc.RemoveChannelMemberParams{
		ChannelID: channelID,
		UserID:    user.ID,
	})
	if err != nil {
		slog.Error("failed to leave channel", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to leave channel")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

func (h *Handler) HandleListChannelMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.Queries.ListChannelMembers(r.Context(), channelID)
	if err != nil {
		slog.Error("failed to list channel members", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	writeJSON(w, http.StatusOK, members)
}

func (h *Handler) HandleListPublicChannels(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	channels, err := h.Queries.ListPublicChannels(r.Context())
	if err != nil {
		slog.Error("failed to list public channels", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list channels")
		return
	}

	writeJSON(w, http.StatusOK, channels)
}

// parseUUID converts a string UUID to pgtype.UUID
func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}
