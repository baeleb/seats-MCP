package seatsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIKeyAuthorizer(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://x", nil)
	if err := (APIKeyAuthorizer{Key: "k"}).Apply(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Partner-Authorization") != "k" {
		t.Errorf("header = %q", req.Header.Get("Partner-Authorization"))
	}
}

func TestOAuthAuthorizer(t *testing.T) {
	now := time.Now()
	tok := &Token{AccessToken: "tok123", Expiry: now.Add(time.Hour)}
	ts := NewTokenSource(OAuthConfig{}, tok, nil, nil)

	req, _ := http.NewRequest("GET", "http://x", nil)
	if err := (OAuthAuthorizer{Source: ts}).Apply(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Partner-Authorization"); got != "Bearer tok123" {
		t.Errorf("header = %q, want Bearer tok123", got)
	}
}

func TestOAuthAuthorizerPropagatesError(t *testing.T) {
	ts := NewTokenSource(OAuthConfig{}, nil, nil, nil) // no token, no refresh
	req, _ := http.NewRequest("GET", "http://x", nil)
	if err := (OAuthAuthorizer{Source: ts}).Apply(context.Background(), req); err != ErrNoToken {
		t.Errorf("err = %v, want ErrNoToken", err)
	}
}

func TestResolveAuthorizer(t *testing.T) {
	now := time.Now()
	validTok := &Token{AccessToken: "a", Expiry: now.Add(time.Hour)}
	refreshOnly := &Token{AccessToken: "a", RefreshToken: "r", Expiry: now.Add(-time.Hour)}
	expiredNoRefresh := &Token{AccessToken: "a", Expiry: now.Add(-time.Hour)}

	tsWith := func(tok *Token) *TokenSource { return NewTokenSource(OAuthConfig{}, tok, nil, nil) }

	t.Run("prefers oauth when valid", func(t *testing.T) {
		auth, mode, err := ResolveAuthorizer("key", tsWith(validTok))
		if err != nil || mode != "oauth" {
			t.Fatalf("mode = %q, err = %v", mode, err)
		}
		if _, ok := auth.(OAuthAuthorizer); !ok {
			t.Errorf("auth = %T, want OAuthAuthorizer", auth)
		}
	})

	t.Run("prefers oauth when only refreshable", func(t *testing.T) {
		_, mode, err := ResolveAuthorizer("key", tsWith(refreshOnly))
		if err != nil || mode != "oauth" {
			t.Fatalf("mode = %q, err = %v", mode, err)
		}
	})

	t.Run("falls back to api key when token unusable", func(t *testing.T) {
		_, mode, err := ResolveAuthorizer("key", tsWith(expiredNoRefresh))
		if err != nil || mode != "api_key" {
			t.Fatalf("mode = %q, err = %v", mode, err)
		}
	})

	t.Run("api key when no token source", func(t *testing.T) {
		auth, mode, err := ResolveAuthorizer("key", nil)
		if err != nil || mode != "api_key" {
			t.Fatalf("mode = %q, err = %v", mode, err)
		}
		if _, ok := auth.(APIKeyAuthorizer); !ok {
			t.Errorf("auth = %T, want APIKeyAuthorizer", auth)
		}
	})

	t.Run("error when nothing available", func(t *testing.T) {
		_, _, err := ResolveAuthorizer("", nil)
		if err != ErrNoCredentials {
			t.Errorf("err = %v, want ErrNoCredentials", err)
		}
	})
}

// TestClientUsesOAuthAuthorizer verifies the bearer token reaches the wire on a
// real Partner API call routed through doJSON, in the documented
// Partner-Authorization header.
func TestClientUsesOAuthAuthorizer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Partner-Authorization") != "Bearer live-token" {
			t.Errorf("partner-authorization = %q", r.Header.Get("Partner-Authorization"))
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "yes"})
	}))
	defer srv.Close()

	now := time.Now()
	ts := NewTokenSource(OAuthConfig{}, &Token{AccessToken: "live-token", Expiry: now.Add(time.Hour)}, nil, nil)
	c := NewWithAuthorizerBase(OAuthAuthorizer{Source: ts}, srv.Client(), srv.URL)

	var out map[string]string
	if err := c.doJSON(context.Background(), "GET", "/search", nil, nil, &out); err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	if out["ok"] != "yes" {
		t.Errorf("out = %v", out)
	}
}
