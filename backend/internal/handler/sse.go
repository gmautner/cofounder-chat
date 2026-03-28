package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// SSEEvent represents a server-sent event
type SSEEvent struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Client represents an SSE client connection
type Client struct {
	ID     string
	Events chan SSEEvent
}

// Hub manages SSE client connections and event broadcasting
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (hub *Hub) AddClient(id string) *Client {
	client := &Client{
		ID:     id,
		Events: make(chan SSEEvent, 64),
	}
	hub.mu.Lock()
	hub.clients[id] = client
	hub.mu.Unlock()
	return client
}

func (hub *Hub) RemoveClient(id string) {
	hub.mu.Lock()
	if client, ok := hub.clients[id]; ok {
		close(client.Events)
		delete(hub.clients, id)
	}
	hub.mu.Unlock()
}

func (hub *Hub) Broadcast(event SSEEvent) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for _, client := range hub.clients {
		select {
		case client.Events <- event:
		default:
			// Client buffer full, skip
			slog.Warn("SSE client buffer full, dropping event", "client_id", client.ID)
		}
	}
}

func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Use a unique client ID based on user ID + connection
	clientID := fmt.Sprintf("%v-%s", user.ID, generateToken(8))
	client := h.SSEHub.AddClient(clientID)
	defer h.SSEHub.RemoveClient(clientID)

	slog.Info("SSE client connected", "client_id", clientID, "user", user.Email)

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"client_id\":%q}\n\n", clientID)
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			slog.Info("SSE client disconnected", "client_id", clientID)
			return
		case event, ok := <-client.Events:
			if !ok {
				return
			}
			data, err := json.Marshal(event.Data)
			if err != nil {
				slog.Error("failed to marshal SSE event", "err", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}
