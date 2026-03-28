package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cofounder-chat/internal/config"
	"cofounder-chat/internal/database"
	"cofounder-chat/internal/handler"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Connect to database with retry
	var pool *pgxpool.Pool
	for i := range 6 {
		pool, err = pgxpool.New(ctx, cfg.DatabaseURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				break
			}
			pool.Close()
		}
		delay := time.Second * (1 << i)
		slog.Warn("database not ready, retrying", "attempt", i+1, "delay", delay, "err", err)
		time.Sleep(delay)
	}
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations
	if err := database.RunMigrations(ctx, pool); err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(cfg.BlobStoragePath, 0755); err != nil {
		slog.Error("failed to create uploads directory", "err", err)
		os.Exit(1)
	}

	h := handler.New(pool, cfg)
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /up", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Auth routes (no middleware needed)
	mux.HandleFunc("GET /auth/google/login", h.HandleGoogleLogin)
	mux.HandleFunc("GET /auth/google/callback", h.HandleGoogleCallback)
	mux.HandleFunc("POST /auth/logout", h.HandleLogout)

	// Dev login (only registered when DEV_MODE=1)
	if cfg.DevMode {
		slog.Warn("DEV_MODE is enabled — dev login endpoint is active")
		mux.HandleFunc("POST /api/dev/login", h.HandleDevLogin)
	}

	// API routes (protected)
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /api/me", h.HandleGetMe)
	apiMux.HandleFunc("GET /api/users", h.HandleListUsers)
	apiMux.HandleFunc("GET /api/users/search", h.HandleSearchUsers)

	// Channel routes
	apiMux.HandleFunc("POST /api/channels", h.HandleCreateChannel)
	apiMux.HandleFunc("GET /api/channels", h.HandleListChannels)
	apiMux.HandleFunc("GET /api/channels/{id}", h.HandleGetChannel)
	apiMux.HandleFunc("POST /api/channels/{id}/join", h.HandleJoinChannel)
	apiMux.HandleFunc("POST /api/channels/{id}/leave", h.HandleLeaveChannel)
	apiMux.HandleFunc("GET /api/channels/{id}/members", h.HandleListChannelMembers)
	apiMux.HandleFunc("GET /api/channels/{id}/messages", h.HandleListChannelMessages)
	apiMux.HandleFunc("POST /api/channels/{id}/messages", h.HandleSendChannelMessage)
	apiMux.HandleFunc("POST /api/channels/{id}/read", h.HandleMarkChannelRead)

	// Conversation (DM) routes
	apiMux.HandleFunc("POST /api/conversations", h.HandleCreateConversation)
	apiMux.HandleFunc("GET /api/conversations", h.HandleListConversations)
	apiMux.HandleFunc("GET /api/conversations/{id}/messages", h.HandleListConversationMessages)
	apiMux.HandleFunc("POST /api/conversations/{id}/messages", h.HandleSendConversationMessage)
	apiMux.HandleFunc("GET /api/conversations/{id}/members", h.HandleListConversationMembers)
	apiMux.HandleFunc("POST /api/conversations/{id}/read", h.HandleMarkConversationRead)

	// Message routes
	apiMux.HandleFunc("PUT /api/messages/{id}", h.HandleUpdateMessage)
	apiMux.HandleFunc("DELETE /api/messages/{id}", h.HandleDeleteMessage)
	apiMux.HandleFunc("GET /api/messages/{id}/thread", h.HandleListThread)

	// Reaction routes
	apiMux.HandleFunc("POST /api/messages/{id}/reactions", h.HandleAddReaction)
	apiMux.HandleFunc("DELETE /api/messages/{id}/reactions/{emoji}", h.HandleRemoveReaction)
	apiMux.HandleFunc("GET /api/messages/{id}/reactions", h.HandleListReactions)

	// File upload
	apiMux.HandleFunc("POST /api/upload", h.HandleFileUpload)

	// SSE
	apiMux.HandleFunc("GET /api/events", h.HandleSSE)

	// Unread counts
	apiMux.HandleFunc("GET /api/unread", h.HandleGetUnreadCounts)

	// Typing indicator
	apiMux.HandleFunc("POST /api/typing", h.HandleTyping)

	// Wrap API routes with auth middleware
	mux.Handle("/api/", h.AuthMiddleware(apiMux))

	// Serve uploaded files
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.BlobStoragePath))))

	// Frontend static files
	frontendDist := "frontend/dist"
	if _, err := os.Stat(frontendDist); err != nil {
		frontendDist = "../frontend/dist"
	}

	fs := http.FileServer(http.Dir(frontendDist))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path != "/" {
			if _, err := os.Stat(filepath.Join(frontendDist, filepath.Clean(r.URL.Path))); err == nil {
				fs.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(frontendDist, "index.html"))
	})

	slog.Info("server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
