package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreatePublicChannel(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "channel-pub@example.com")

	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":        "test-public",
		"description": "A public test channel",
		"is_private":  false,
	}, cookie)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var channel struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IsPrivate bool   `json:"is_private"`
	}
	decodeJSON(t, resp, &channel)

	if channel.Name != "test-public" {
		t.Errorf("expected channel name test-public, got %s", channel.Name)
	}
	if channel.IsPrivate {
		t.Error("expected channel to be public")
	}
	if channel.ID == "" {
		t.Error("expected channel ID to be set")
	}
}

func TestCreatePrivateChannel(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "channel-priv@example.com")

	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":        "test-private",
		"description": "A private test channel",
		"is_private":  true,
	}, cookie)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	var channel struct {
		Name      string `json:"name"`
		IsPrivate bool   `json:"is_private"`
	}
	decodeJSON(t, resp, &channel)

	if !channel.IsPrivate {
		t.Error("expected channel to be private")
	}
}

func TestListUserChannels(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "channel-list@example.com")

	// Create a channel (user is auto-joined)
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":        "test-list-chan",
		"description": "Channel for list test",
	}, cookie)
	resp.Body.Close()

	// List channels
	resp = authenticatedRequest(t, server, "GET", "/api/channels", nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var channels []struct {
		Name string `json:"name"`
	}
	decodeJSON(t, resp, &channels)

	// Should have at least the channel we created (plus #general from login)
	if len(channels) < 1 {
		t.Error("expected at least one channel")
	}

	found := false
	for _, ch := range channels {
		if ch.Name == "test-list-chan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find test-list-chan in user channels")
	}
}

func TestJoinPublicChannel(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	// User 1 creates a public channel
	cookie1, _ := createTestUser(t, "channel-join1@example.com")
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":        "test-join-pub",
		"description": "Joinable channel",
		"is_private":  false,
	}, cookie1)

	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	// User 2 joins the public channel
	cookie2, _ := createTestUser(t, "channel-join2@example.com")
	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/join", nil, cookie2)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}
	decodeJSON(t, resp, &result)
	if result.Status != "joined" {
		t.Errorf("expected status joined, got %s", result.Status)
	}
}

func TestCannotJoinPrivateChannel(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	// User 1 creates a private channel
	cookie1, _ := createTestUser(t, "channel-cantjoin1@example.com")
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":        "test-cantjoin",
		"description": "Private channel",
		"is_private":  true,
	}, cookie1)

	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	// User 2 tries to join the private channel
	cookie2, _ := createTestUser(t, "channel-cantjoin2@example.com")
	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/join", nil, cookie2)

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestLeaveChannel(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "channel-leave@example.com")

	// Create a channel
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":        "test-leave-chan",
		"description": "Channel to leave",
	}, cookie)

	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	// Leave the channel
	resp = authenticatedRequest(t, server, "POST", "/api/channels/"+channel.ID+"/leave", nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}
	decodeJSON(t, resp, &result)
	if result.Status != "left" {
		t.Errorf("expected status left, got %s", result.Status)
	}
}

func TestListChannelMembers(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "channel-members@example.com")

	// Create a channel
	resp := authenticatedRequest(t, server, "POST", "/api/channels", map[string]any{
		"name":        "test-members-chan",
		"description": "Channel for members test",
	}, cookie)

	var channel struct {
		ID string `json:"id"`
	}
	decodeJSON(t, resp, &channel)

	// List members
	resp = authenticatedRequest(t, server, "GET", "/api/channels/"+channel.ID+"/members", nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var members []json.RawMessage
	decodeJSON(t, resp, &members)

	if len(members) != 1 {
		t.Errorf("expected 1 member (the creator), got %d", len(members))
	}
}
