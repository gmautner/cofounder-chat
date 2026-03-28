package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cofounder-chat/internal/database/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

type GoogleUserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (h *Handler) baseURL(r *http.Request) string {
	if h.Config.BaseURL != "" {
		return h.Config.BaseURL
	}
	scheme := "https"
	if r.TLS == nil && !strings.HasPrefix(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "http"
	}
	return scheme + "://" + r.Host
}

func (h *Handler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateToken(16)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	params := url.Values{
		"client_id":     {h.Config.GoogleClientID},
		"redirect_uri":  {h.baseURL(r) + "/auth/google/callback"},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
	}

	http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+params.Encode(), http.StatusTemporaryRedirect)
}

func (h *Handler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		writeError(w, http.StatusBadRequest, "invalid state")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing code")
		return
	}

	// Exchange code for token
	tokenData := url.Values{
		"code":          {code},
		"client_id":     {h.Config.GoogleClientID},
		"client_secret": {h.Config.GoogleSecret},
		"redirect_uri":  {h.baseURL(r) + "/auth/google/callback"},
		"grant_type":    {"authorization_code"},
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", tokenData)
	if err != nil {
		slog.Error("token exchange failed", "err", err)
		writeError(w, http.StatusInternalServerError, "authentication failed")
		return
	}
	defer resp.Body.Close()

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		slog.Error("token decode failed", "err", err)
		writeError(w, http.StatusInternalServerError, "authentication failed")
		return
	}

	// Get user info
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userResp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("user info request failed", "err", err)
		writeError(w, http.StatusInternalServerError, "authentication failed")
		return
	}
	defer userResp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		slog.Error("user info decode failed", "err", err)
		writeError(w, http.StatusInternalServerError, "authentication failed")
		return
	}

	// Upsert user and create session
	h.createSessionForGoogleUser(w, r, userInfo)
}

func (h *Handler) createSessionForGoogleUser(w http.ResponseWriter, r *http.Request, userInfo GoogleUserInfo) {
	user, err := h.Queries.UpsertUser(r.Context(), sqlc.UpsertUserParams{
		GoogleID:    userInfo.Sub,
		Email:       userInfo.Email,
		DisplayName: userInfo.Name,
		AvatarUrl:   userInfo.Picture,
	})
	if err != nil {
		slog.Error("upsert user failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Ensure user is in #general channel
	h.ensureGeneralChannel(r.Context(), user.ID)

	token := generateToken(32)
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

	_, err = h.Queries.CreateSession(r.Context(), sqlc.CreateSessionParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		slog.Error("create session failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *Handler) ensureGeneralChannel(ctx context.Context, userID pgtype.UUID) {
	channel, err := h.Queries.GetChannelByName(ctx, "general")
	if err != nil {
		// Channel doesn't exist yet, create it
		channel, err = h.Queries.CreateChannel(ctx, sqlc.CreateChannelParams{
			Name:        "general",
			Description: "General discussion",
			IsPrivate:   false,
			CreatedBy:   userID,
		})
		if err != nil {
			slog.Error("failed to create general channel", "err", err)
			return
		}
	}
	h.Queries.AddChannelMember(ctx, sqlc.AddChannelMemberParams{
		ChannelID: channel.ID,
		UserID:    userID,
	})
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.Queries.DeleteSession(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *Handler) HandleDevLogin(w http.ResponseWriter, r *http.Request) {
	if !h.Config.DevMode {
		http.NotFound(w, r)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	defer r.Body.Close()

	var req struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	name := strings.Split(req.Email, "@")[0]

	h.createSessionForGoogleUser(w, r, GoogleUserInfo{
		Sub:     fmt.Sprintf("dev_%s", req.Email),
		Email:   req.Email,
		Name:    name,
		Picture: "",
	})
}

func (h *Handler) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func generateToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
