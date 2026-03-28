package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"cofounder-chat/internal/config"
	"cofounder-chat/internal/database/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB      *pgxpool.Pool
	Queries *sqlc.Queries
	Config  *config.Config
	SSEHub  *Hub
}

func New(db *pgxpool.Pool, cfg *config.Config) *Handler {
	return &Handler{
		DB:      db,
		Queries: sqlc.New(db),
		Config:  cfg,
		SSEHub:  NewHub(),
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
