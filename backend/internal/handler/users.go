package handler

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	users, err := h.Queries.ListUsers(r.Context())
	if err != nil {
		slog.Error("failed to list users", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) HandleSearchUsers(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q parameter is required")
		return
	}

	users, err := h.Queries.SearchUsers(r.Context(), pgtype.Text{String: q, Valid: true})
	if err != nil {
		slog.Error("failed to search users", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to search users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}
