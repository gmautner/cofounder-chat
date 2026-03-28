package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const userContextKey contextKey = "user"

type AuthUser struct {
	ID          pgtype.UUID `json:"id"`
	Email       string      `json:"email"`
	DisplayName string      `json:"display_name"`
	AvatarURL   string      `json:"avatar_url"`
}

func GetUserFromContext(ctx context.Context) *AuthUser {
	user, ok := ctx.Value(userContextKey).(*AuthUser)
	if !ok {
		return nil
	}
	return user
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		session, err := h.Queries.GetSessionByToken(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid session")
			return
		}

		user := &AuthUser{
			ID:          session.UserID,
			Email:       session.Email,
			DisplayName: session.DisplayName,
			AvatarURL:   session.AvatarUrl,
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
