package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddReaction(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "react-add@example.com")

	// Create a channel and send a message
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-react-add",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "React to this",
	}, cookie)
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &msg)

	// Add a reaction
	resp = authenticatedRequest(t, server, "POST", "/api/messages/"+msg.ID+"/reactions", map[string]any{
		"emoji": "thumbsup",
	}, cookie)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var reaction struct {
		Emoji string `json:"emoji"`
	}
	decodeJSON(t, resp, &reaction)

	if reaction.Emoji != "thumbsup" {
		t.Errorf("expected emoji 'thumbsup', got '%s'", reaction.Emoji)
	}
}

func TestRemoveReaction(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "react-remove@example.com")

	// Create a channel, send a message, add a reaction
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-react-rm",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "React then remove",
	}, cookie)
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &msg)

	resp = authenticatedRequest(t, server, "POST", "/api/messages/"+msg.ID+"/reactions", map[string]any{
		"emoji": "heart",
	}, cookie)
	resp.Body.Close()

	// Remove the reaction
	resp = authenticatedRequest(t, server, "DELETE", "/api/messages/"+msg.ID+"/reactions/heart", nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}
	decodeJSON(t, resp, &result)
	if result.Status != "removed" {
		t.Errorf("expected status removed, got %s", result.Status)
	}
}

func TestListReactions(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "react-list@example.com")

	// Create a channel, send a message, add reactions
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name": "test-react-list",
	}, cookie)
	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/messages", map[string]any{
		"content": "List reactions here",
	}, cookie)
	var msg struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &msg)

	// Add two different reactions
	resp = authenticatedRequest(t, server, "POST", "/api/messages/"+msg.ID+"/reactions", map[string]any{
		"emoji": "thumbsup",
	}, cookie)
	resp.Body.Close()

	resp = authenticatedRequest(t, server, "POST", "/api/messages/"+msg.ID+"/reactions", map[string]any{
		"emoji": "heart",
	}, cookie)
	resp.Body.Close()

	// List reactions
	resp = authenticatedRequest(t, server, "GET", "/api/messages/"+msg.ID+"/reactions", nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var reactions []struct {
		Emoji       string `json:"emoji"`
		DisplayName string `json:"display_name"`
	}
	decodeJSON(t, resp, &reactions)

	if len(reactions) != 2 {
		t.Fatalf("expected 2 reactions, got %d", len(reactions))
	}

	emojis := map[string]bool{}
	for _, r := range reactions {
		emojis[r.Emoji] = true
	}
	if !emojis["thumbsup"] {
		t.Error("expected thumbsup reaction")
	}
	if !emojis["heart"] {
		t.Error("expected heart reaction")
	}
}
