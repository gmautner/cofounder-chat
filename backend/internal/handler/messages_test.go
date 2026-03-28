package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendChannelMessage(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "msg-send@example.com")

	// Create a channel
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-msg-send",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	// Send a message
	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "Hello, world!",
	}, cookie)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var msg struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
	decodeJSON(t, resp, &msg)

	if msg.Content != "Hello, world!" {
		t.Errorf("expected content 'Hello, world!', got '%s'", msg.Content)
	}
	if msg.ID == "" {
		t.Error("expected message ID to be set")
	}
}

func TestListChannelMessages(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "msg-list@example.com")

	// Create a channel
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-msg-list",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	// Send two messages
	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "First message",
	}, cookie)
	resp.Body.Close()

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "Second message",
	}, cookie)
	resp.Body.Close()

	// List messages
	resp = authenticatedRequest(t, server, "GET", "/api/channels/"+channel.ID+"/messages", nil, cookie)

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

	if messages[0].Content != "First message" {
		t.Errorf("expected first message content 'First message', got '%s'", messages[0].Content)
	}
	if messages[1].Content != "Second message" {
		t.Errorf("expected second message content 'Second message', got '%s'", messages[1].Content)
	}
}

func TestEditMessage(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "msg-edit@example.com")

	// Create a channel and send a message
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-msg-edit",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "Original content",
	}, cookie)
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &msg)

	// Edit the message
	resp = authenticatedRequest(t, server, "PUT", "/api/messages/"+msg.ID, map[string]any{
		"content": "Edited content",
	}, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var edited struct {
		Content  string `json:"content"`
		IsEdited bool   `json:"is_edited"`
	}
	decodeJSON(t, resp, &edited)

	if edited.Content != "Edited content" {
		t.Errorf("expected content 'Edited content', got '%s'", edited.Content)
	}
	if !edited.IsEdited {
		t.Error("expected is_edited to be true")
	}
}

func TestEditMessageOnlyAuthor(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie1, _ := createTestUser(t, "msg-edit-author1@example.com")
	cookie2, _ := createTestUser(t, "msg-edit-author2@example.com")

	// User 1 creates channel and sends a message
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":       "test-edit-author",
		"is_private": false,
	}, cookie1)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "User 1 message",
	}, cookie1)
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &msg)

	// User 2 tries to edit User 1's message
	resp = authenticatedRequest(t, server, "PUT", "/api/messages/"+msg.ID, map[string]any{
		"content": "Hacked content",
	}, cookie2)

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestDeleteMessage(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "msg-delete@example.com")

	// Create a channel and send a message
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-msg-delete",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "To be deleted",
	}, cookie)
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &msg)

	// Delete the message
	resp = authenticatedRequest(t, server, "DELETE", "/api/messages/"+msg.ID, nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}
	decodeJSON(t, resp, &result)
	if result.Status != "deleted" {
		t.Errorf("expected status deleted, got %s", result.Status)
	}
}

func TestDeleteMessageOnlyAuthor(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie1, _ := createTestUser(t, "msg-del-author1@example.com")
	cookie2, _ := createTestUser(t, "msg-del-author2@example.com")

	// User 1 creates channel and sends a message
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":       "test-del-author",
		"is_private": false,
	}, cookie1)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "User 1 message",
	}, cookie1)
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &msg)

	// User 2 tries to delete User 1's message
	resp = authenticatedRequest(t, server, "DELETE", "/api/messages/"+msg.ID, nil, cookie2)

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestThreadReplies(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "msg-thread@example.com")

	// Create a channel and send a parent message
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-msg-thread",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "Parent message",
	}, cookie)
	var parent struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &parent)

	// Send thread replies
	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content":   "Reply 1",
		"parent_id": parent.ID,
	}, cookie)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for reply 1, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content":   "Reply 2",
		"parent_id": parent.ID,
	}, cookie)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 for reply 2, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// List thread
	resp = authenticatedRequest(t, server, "GET", "/api/messages/"+parent.ID+"/thread", nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var thread struct {
		Parent struct {
			Content string `json:"content"`
		} `json:"parent"`
		Replies []json.RawMessage `json:"replies"`
	}
	decodeJSON(t, resp, &thread)

	if thread.Parent.Content != "Parent message" {
		t.Errorf("expected parent content 'Parent message', got '%s'", thread.Parent.Content)
	}
	if len(thread.Replies) != 2 {
		t.Errorf("expected 2 replies, got %d", len(thread.Replies))
	}
}
