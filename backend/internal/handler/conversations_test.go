package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateConversation(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie1, _ := createTestUser(t, "conv-create1@example.com")
	_, user2ID := createTestUser(t, "conv-create2@example.com")

	resp := authenticatedRequest(t, server, "POST", "/api/conversations", map[string]any{
		"user_id": user2ID,
	}, cookie1)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &conv)

	if conv.ID == "" {
		t.Error("expected conversation ID to be set")
	}
}

func TestDuplicateConversationReturnsExisting(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie1, _ := createTestUser(t, "conv-dup1@example.com")
	_, user2ID := createTestUser(t, "conv-dup2@example.com")

	// Create conversation first time
	resp := authenticatedRequest(t, server, "POST", "/api/conversations", map[string]any{
		"user_id": user2ID,
	}, cookie1)

	var firstConv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &firstConv)

	// Create conversation again with same users
	resp = authenticatedRequest(t, server, "POST", "/api/conversations", map[string]any{
		"user_id": user2ID,
	}, cookie1)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for existing conversation, got %d", resp.StatusCode)
	}

	var secondConv struct {
		ID       string `json:"id"`
		Existing bool   `json:"existing"`
	}
	decodeJSON(t, resp, &secondConv)

	if !secondConv.Existing {
		t.Error("expected existing flag to be true")
	}
	if secondConv.ID != firstConv.ID {
		t.Errorf("expected same conversation ID %s, got %s", firstConv.ID, secondConv.ID)
	}
}

func TestListConversationsIncludesMembers(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie1, _ := createTestUser(t, "conv-list1@example.com")
	_, user2ID := createTestUser(t, "conv-list2@example.com")

	// Create a conversation
	resp := authenticatedRequest(t, server, "POST", "/api/conversations", map[string]any{
		"user_id": user2ID,
	}, cookie1)
	resp.Body.Close()

	// List conversations
	resp = authenticatedRequest(t, server, "GET", "/api/conversations", nil, cookie1)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var conversations []struct {
		ID      string            `json:"id"`
		Members []json.RawMessage `json:"members"`
	}
	decodeJSON(t, resp, &conversations)

	if len(conversations) < 1 {
		t.Fatal("expected at least one conversation")
	}

	// Find our conversation and check it has 2 members
	found := false
	for _, conv := range conversations {
		if len(conv.Members) == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find a conversation with 2 members")
	}
}

func TestSendAndListConversationMessages(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie1, _ := createTestUser(t, "conv-msg1@example.com")
	cookie2, user2ID := createTestUser(t, "conv-msg2@example.com")

	// Create a conversation
	resp := authenticatedRequest(t, server, "POST", "/api/conversations", map[string]any{
		"user_id": user2ID,
	}, cookie1)

	var conv struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &conv)

	// User 1 sends a message
	resp = authenticatedRequest(t, server, "POST", "/api/conversations/"+conv.ID+"/messages", map[string]any{
		"content": "Hello from user 1",
	}, cookie1)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// User 2 sends a message
	resp = authenticatedRequest(t, server, "POST", "/api/conversations/"+conv.ID+"/messages", map[string]any{
		"content": "Hello from user 2",
	}, cookie2)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// List messages
	resp = authenticatedRequest(t, server, "GET", "/api/conversations/"+conv.ID+"/messages", nil, cookie1)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var messages []struct {
		Content    string `json:"content"`
		AuthorName string `json:"author_name"`
	}
	decodeJSON(t, resp, &messages)

	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Content != "Hello from user 1" {
		t.Errorf("expected first message 'Hello from user 1', got '%s'", messages[0].Content)
	}
	if messages[1].Content != "Hello from user 2" {
		t.Errorf("expected second message 'Hello from user 2', got '%s'", messages[1].Content)
	}
}
