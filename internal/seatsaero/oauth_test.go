package seatsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// fixedClock returns a now func anchored at t.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAuthCodeURL(t *testing.T) {
	cfg := OAuthConfig{
		ClientID:    "cid",
		RedirectURI: "https://app.example/callback",
		BaseURL:     "https://seats.aero/oauth2",
	}
	ts := NewTokenSource(cfg, nil, nil, nil)
	got := ts.AuthCodeURL("xyz")

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	if u.Path != "/oauth2/consent" {
		t.Errorf("path = %q, want /oauth2/consent", u.Path)
	}
	q := u.Query()
	checks := map[string]string{
		"response_type": "code",
		"client_id":     "cid",
		"redirect_uri":  "https://app.example/callback",
		"state":         "xyz",
		"scope":         "openid",
	}
	for k, want := range checks {
		if got := q.Get(k); got != want {
			t.Errorf("query %s = %q, want %q", k, got, want)
		}
	}
}

func TestExchange(t *testing.T) {
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("path = %q, want /token", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("missing JSON content-type")
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken:  "seats:ota:abc",
			TokenType:    "Bearer",
			ExpiresIn:    3599,
			RefreshToken: "seats:otr:def",
			Scope:        "openid",
		})
	}))
	defer srv.Close()

	now := time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
	ts := NewTokenSource(OAuthConfig{
		ClientID:     "cid",
		ClientSecret: "secret",
		RedirectURI:  "https://app.example/cb",
		BaseURL:      srv.URL,
	}, nil, nil, nil)
	ts.now = fixedClock(now)

	tok, err := ts.Exchange(context.Background(), "the-code", "the-state")
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if tok.AccessToken != "seats:ota:abc" {
		t.Errorf("access token = %q", tok.AccessToken)
	}
	if tok.RefreshToken != "seats:otr:def" {
		t.Errorf("refresh token = %q", tok.RefreshToken)
	}
	wantExpiry := now.Add(3599 * time.Second)
	if !tok.Expiry.Equal(wantExpiry) {
		t.Errorf("expiry = %v, want %v", tok.Expiry, wantExpiry)
	}
	// Request body should carry the auth-code grant fields.
	for k, want := range map[string]string{
		"grant_type":    "authorization_code",
		"code":          "the-code",
		"state":         "the-state",
		"scope":         "openid",
		"client_id":     "cid",
		"client_secret": "secret",
		"redirect_uri":  "https://app.example/cb",
	} {
		if gotBody[k] != want {
			t.Errorf("body[%s] = %q, want %q", k, gotBody[k], want)
		}
	}
}

func TestTokenCachesValidToken(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "new", ExpiresIn: 3600})
	}))
	defer srv.Close()

	now := time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
	tok := &Token{AccessToken: "cached", RefreshToken: "r", Expiry: now.Add(time.Hour)}
	ts := NewTokenSource(OAuthConfig{BaseURL: srv.URL}, tok, nil, nil)
	ts.now = fixedClock(now)

	got, err := ts.Token(context.Background())
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if got != "cached" {
		t.Errorf("token = %q, want cached", got)
	}
	if calls != 0 {
		t.Errorf("expected no refresh call, got %d", calls)
	}
}

func TestTokenRefreshesExpired(t *testing.T) {
	var gotBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "fresh", ExpiresIn: 3600})
	}))
	defer srv.Close()

	now := time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
	store := &MemoryTokenStore{}
	expired := &Token{AccessToken: "stale", RefreshToken: "the-refresh", Expiry: now.Add(-time.Hour)}
	ts := NewTokenSource(OAuthConfig{ClientID: "cid", ClientSecret: "s", BaseURL: srv.URL}, expired, store, nil)
	ts.now = fixedClock(now)

	got, err := ts.Token(context.Background())
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if got != "fresh" {
		t.Errorf("token = %q, want fresh", got)
	}
	if gotBody["grant_type"] != "refresh_token" || gotBody["refresh_token"] != "the-refresh" {
		t.Errorf("unexpected refresh body: %v", gotBody)
	}
	// Refreshed token should be persisted.
	saved, _ := store.Load()
	if saved == nil || saved.AccessToken != "fresh" {
		t.Errorf("store not updated: %+v", saved)
	}
}

func TestTokenRefreshKeepsOldRefreshTokenWhenOmitted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No refresh_token in the response.
		json.NewEncoder(w).Encode(tokenResponse{AccessToken: "fresh", ExpiresIn: 3600})
	}))
	defer srv.Close()

	now := time.Now()
	expired := &Token{AccessToken: "stale", RefreshToken: "keep-me", Expiry: now.Add(-time.Hour)}
	ts := NewTokenSource(OAuthConfig{ClientID: "c", ClientSecret: "s", BaseURL: srv.URL}, expired, nil, nil)

	if _, err := ts.Token(context.Background()); err != nil {
		t.Fatalf("token: %v", err)
	}
	if got := ts.Peek(); got.RefreshToken != "keep-me" {
		t.Errorf("refresh token = %q, want keep-me", got.RefreshToken)
	}
}

func TestTokenNoRefreshToken(t *testing.T) {
	ts := NewTokenSource(OAuthConfig{}, &Token{AccessToken: "x"}, nil, nil)
	_, err := ts.Token(context.Background())
	if err != ErrNoToken {
		t.Errorf("err = %v, want ErrNoToken", err)
	}
}

func TestTokenNilToken(t *testing.T) {
	ts := NewTokenSource(OAuthConfig{}, nil, nil, nil)
	_, err := ts.Token(context.Background())
	if err != ErrNoToken {
		t.Errorf("err = %v, want ErrNoToken", err)
	}
}

func TestTokenRefreshError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("invalid_grant"))
	}))
	defer srv.Close()

	now := time.Now()
	expired := &Token{AccessToken: "stale", RefreshToken: "r", Expiry: now.Add(-time.Hour)}
	ts := NewTokenSource(OAuthConfig{ClientID: "c", ClientSecret: "s", BaseURL: srv.URL}, expired, nil, nil)

	_, err := ts.Token(context.Background())
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %T, want *APIError", err)
	}
	if apiErr.Status != 400 {
		t.Errorf("status = %d, want 400", apiErr.Status)
	}
}

func TestPostTokenEmptyAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(tokenResponse{ExpiresIn: 3600})
	}))
	defer srv.Close()

	ts := NewTokenSource(OAuthConfig{ClientID: "c", ClientSecret: "s", BaseURL: srv.URL}, nil, nil, nil)
	_, err := ts.Exchange(context.Background(), "code", "state")
	if err == nil {
		t.Fatal("expected error for empty access_token")
	}
}

func TestUserInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/userinfo" {
			if r.Header.Get("Authorization") != "Bearer the-access" {
				t.Errorf("authorization = %q", r.Header.Get("Authorization"))
			}
			json.NewEncoder(w).Encode(UserInfo{
				Name:                "Ada Lovelace",
				Email:               "ada@example.com",
				SubscriptionExpires: 1800000000,
			})
			return
		}
		t.Errorf("unexpected path %q", r.URL.Path)
	}))
	defer srv.Close()

	now := time.Now()
	tok := &Token{AccessToken: "the-access", Expiry: now.Add(time.Hour)}
	ts := NewTokenSource(OAuthConfig{BaseURL: srv.URL}, tok, nil, nil)

	info, err := ts.UserInfo(context.Background())
	if err != nil {
		t.Fatalf("userinfo: %v", err)
	}
	if info.Name != "Ada Lovelace" || info.Email != "ada@example.com" {
		t.Errorf("info = %+v", info)
	}
}

func TestOAuthConfigComplete(t *testing.T) {
	cases := []struct {
		cfg  OAuthConfig
		want bool
	}{
		{OAuthConfig{ClientID: "a", ClientSecret: "b", RedirectURI: "c"}, true},
		{OAuthConfig{ClientID: "a", ClientSecret: "b"}, false},
		{OAuthConfig{}, false},
	}
	for i, c := range cases {
		if got := c.cfg.Complete(); got != c.want {
			t.Errorf("case %d: Complete() = %v, want %v", i, got, c.want)
		}
	}
}

func TestTokenValid(t *testing.T) {
	now := time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		tok  *Token
		want bool
	}{
		{"nil", nil, false},
		{"empty access", &Token{Expiry: now.Add(time.Hour)}, false},
		{"zero expiry", &Token{AccessToken: "x"}, false},
		{"valid", &Token{AccessToken: "x", Expiry: now.Add(time.Hour)}, true},
		{"expired", &Token{AccessToken: "x", Expiry: now.Add(-time.Hour)}, false},
		{"within skew", &Token{AccessToken: "x", Expiry: now.Add(30 * time.Second)}, false},
	}
	for _, c := range cases {
		if got := c.tok.Valid(now); got != c.want {
			t.Errorf("%s: Valid() = %v, want %v", c.name, got, c.want)
		}
	}
}
