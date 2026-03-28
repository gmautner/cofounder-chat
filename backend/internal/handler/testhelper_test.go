package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"cofounder-chat/internal/config"
	"cofounder-chat/internal/database"
	"cofounder-chat/internal/handler"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testPool    *pgxpool.Pool
	testHandler *handler.Handler
	testMux     *http.ServeMux
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	ctx := context.Background()
	var err error
	testPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	if err := database.RunMigrations(ctx, testPool); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Clean test data before running tests
	cleanTestData(ctx)

	cfg := &config.Config{
		Port:            "8080",
		DatabaseURL:     dbURL,
		BlobStoragePath: os.TempDir(),
		DevMode:         true,
		BaseURL:         "http://localhost:8080",
		SessionSecret:   "test-secret",
	}

	testHandler = handler.New(testPool, cfg)
	testMux = setupTestRoutes(testHandler)

	os.Exit(m.Run())
}

func cleanTestData(ctx context.Context) {
	// Clean in reverse dependency order
	testPool.Exec(ctx, "DELETE FROM reactions")
	testPool.Exec(ctx, "DELETE FROM attachments")
	testPool.Exec(ctx, "DELETE FROM messages")
	testPool.Exec(ctx, "DELETE FROM channel_members")
	testPool.Exec(ctx, "DELETE FROM channels")
	testPool.Exec(ctx, "DELETE FROM conversation_members")
	testPool.Exec(ctx, "DELETE FROM conversations")
	testPool.Exec(ctx, "DELETE FROM sessions")
	testPool.Exec(ctx, "DELETE FROM users")
}

func setupTestRoutes(h *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	// Dev login (no middleware)
	mux.HandleFunc("POST /api/dev/login", h.HandleDevLogin)

	// Protected routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /api/me", h.HandleGetMe)
	apiMux.HandleFunc("GET /api/users", h.HandleListUsers)
	apiMux.HandleFunc("GET /api/users/search", h.HandleSearchUsers)
	apiMux.HandleFunc("POST /api/channels", h.HandleCreateChannel)
	apiMux.HandleFunc("GET /api/channels", h.HandleListChannels)
	apiMux.HandleFunc("GET /api/channels/{id}", h.HandleGetChannel)
	apiMux.HandleFunc("POST /api/channels/{id}/join", h.HandleJoinChannel)
	apiMux.HandleFunc("POST /api/channels/{id}/leave", h.HandleLeaveChannel)
	apiMux.HandleFunc("GET /api/channels/{id}/members", h.HandleListChannelMembers)
	apiMux.HandleFunc("GET /api/channels/{id}/messages", h.HandleListChannelMessages)
	apiMux.HandleFunc("POST /api/channels/{id}/messages", h.HandleSendChannelMessage)
	apiMux.HandleFunc("POST /api/channels/{id}/read", h.HandleMarkChannelRead)
	apiMux.HandleFunc("POST /api/conversations", h.HandleCreateConversation)
	apiMux.HandleFunc("GET /api/conversations", h.HandleListConversations)
	apiMux.HandleFunc("GET /api/conversations/{id}/messages", h.HandleListConversationMessages)
	apiMux.HandleFunc("POST /api/conversations/{id}/messages", h.HandleSendConversationMessage)
	apiMux.HandleFunc("GET /api/conversations/{id}/members", h.HandleListConversationMembers)
	apiMux.HandleFunc("POST /api/conversations/{id}/read", h.HandleMarkConversationRead)
	apiMux.HandleFunc("PUT /api/messages/{id}", h.HandleUpdateMessage)
	apiMux.HandleFunc("DELETE /api/messages/{id}", h.HandleDeleteMessage)
	apiMux.HandleFunc("GET /api/messages/{id}/thread", h.HandleListThread)
	apiMux.HandleFunc("POST /api/messages/{id}/reactions", h.HandleAddReaction)
	apiMux.HandleFunc("DELETE /api/messages/{id}/reactions/{emoji}", h.HandleRemoveReaction)
	apiMux.HandleFunc("GET /api/messages/{id}/reactions", h.HandleListReactions)
	apiMux.HandleFunc("POST /api/upload", h.HandleFileUpload)
	apiMux.HandleFunc("GET /api/events", h.HandleSSE)
	apiMux.HandleFunc("GET /api/unread", h.HandleGetUnreadCounts)
	apiMux.HandleFunc("POST /api/typing", h.HandleTyping)

	mux.Handle("/api/", h.AuthMiddleware(apiMux))

	return mux
}

// createTestUser creates a user and returns the session cookie for authenticated requests.
func createTestUser(t *testing.T, email string) (*http.Cookie, string) {
	t.Helper()

	server := httptest.NewServer(testMux)
	defer server.Close()

	// Use a client that does NOT follow redirects so we can capture the cookie
	// from the 307 response returned by HandleDevLogin.
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	body, _ := json.Marshal(map[string]string{"email": email})
	resp, err := noRedirectClient.Post(server.URL+"/api/dev/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	defer resp.Body.Close()

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie returned")
	}

	// Get user ID
	req, _ := http.NewRequest("GET", server.URL+"/api/me", nil)
	req.AddCookie(sessionCookie)
	meResp, err := noRedirectClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	defer meResp.Body.Close()

	var user struct {
		ID string `json:"id"`
	}
	json.NewDecoder(meResp.Body).Decode(&user)

	return sessionCookie, user.ID
}

// authenticatedRequest makes an HTTP request with the given session cookie.
func authenticatedRequest(t *testing.T, server *httptest.Server, method, path string, body any, cookie *http.Cookie) *http.Response {
	t.Helper()

	var reqBody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, server.URL+path, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)

	// Don't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

// decodeJSON decodes a JSON response body into the target.
func decodeJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}
