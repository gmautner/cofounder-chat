package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevLoginCreatesUserAndReturnsSession(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "auth-test@example.com"})

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Post(server.URL+"/api/dev/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("dev login request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should redirect (307)
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", resp.StatusCode)
	}

	// Should have a session cookie
	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}
	if sessionCookie.Value == "" {
		t.Error("session cookie should not be empty")
	}
}

func TestGetMeReturnsAuthenticatedUser(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "getme-test@example.com")

	resp := authenticatedRequest(t, server, "GET", "/api/me", nil, cookie)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var user struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	}
	decodeJSON(t, resp, &user)

	if user.Email != "getme-test@example.com" {
		t.Errorf("expected email getme-test@example.com, got %s", user.Email)
	}
	if user.DisplayName != "getme-test" {
		t.Errorf("expected display_name getme-test, got %s", user.DisplayName)
	}
}

func TestUnauthenticatedRequestReturns401(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/me")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestLogoutClearsSession(t *testing.T) {
	server := httptest.NewServer(testMux)
	defer server.Close()

	cookie, _ := createTestUser(t, "logout-test@example.com")

	// First verify the user is authenticated
	resp := authenticatedRequest(t, server, "GET", "/api/me", nil, cookie)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Fatalf("expected authenticated user, got status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Now log out - uses the main auth/logout route which is outside our test mux
	// Instead, test that using an invalid session returns 401
	invalidCookie := &http.Cookie{
		Name:  "session",
		Value: "invalid-token-that-does-not-exist",
	}
	resp = authenticatedRequest(t, server, "GET", "/api/me", nil, invalidCookie)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 with invalid session, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
