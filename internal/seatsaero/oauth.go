package seatsaero

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuthBaseURL is the root for seats.aero's OAuth2 endpoints (/consent, /token,
// /userinfo). Override via OAuthConfig.BaseURL in tests.
const OAuthBaseURL = "https://seats.aero/oauth2"

// OAuthScope is the only scope seats.aero accepts.
const OAuthScope = "openid"

// expirySkew is subtracted from a token's expiry so we refresh slightly early
// rather than racing the server clock.
const expirySkew = 60 * time.Second

// ErrNoToken is returned when a TokenSource has no usable token and no refresh
// token to obtain one — the user must run `seats-mcp auth login`.
var ErrNoToken = errors.New("no OAuth token available; run `seats-mcp auth login`")

// OAuthConfig holds the registered application credentials needed to run the
// OAuth2 authorization-code flow against seats.aero.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string // must exactly match a redirect URI registered for the app
	BaseURL      string // defaults to OAuthBaseURL when empty
}

func (c OAuthConfig) baseURL() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return OAuthBaseURL
}

// Complete reports whether the config has the credentials required to exchange
// and refresh tokens.
func (c OAuthConfig) Complete() bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.RedirectURI != ""
}

// Token is a persisted OAuth2 token set. Expiry is an absolute instant derived
// from the token endpoint's expires_in.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	Expiry       time.Time `json:"expiry"`
}

// Valid reports whether the access token is present and not within expirySkew
// of expiring, relative to now.
func (t *Token) Valid(now time.Time) bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	if t.Expiry.IsZero() {
		return false
	}
	return now.Before(t.Expiry.Add(-expirySkew))
}

// tokenResponse mirrors the JSON body returned by POST /oauth2/token.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// UserInfo is the response from GET /oauth2/userinfo.
type UserInfo struct {
	Name                string `json:"name"`
	Email               string `json:"email"`
	SubscriptionExpires int64  `json:"subscription_expires"`
}

// TokenSource yields valid OAuth2 access tokens, refreshing and persisting them
// as needed. It is safe for concurrent use.
type TokenSource struct {
	cfg   OAuthConfig
	http  *http.Client
	store TokenStore
	now   func() time.Time

	mu  sync.Mutex
	tok *Token
}

// NewTokenSource builds a TokenSource. tok is the initial token (may be nil if
// the caller will Exchange a code first). store, if non-nil, receives refreshed
// tokens. httpClient defaults to a 30s-timeout client.
func NewTokenSource(cfg OAuthConfig, tok *Token, store TokenStore, httpClient *http.Client) *TokenSource {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &TokenSource{
		cfg:   cfg,
		http:  httpClient,
		store: store,
		now:   time.Now,
		tok:   tok,
	}
}

// AuthCodeURL builds the consent URL the user must visit to authorize the app.
func (ts *TokenSource) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", ts.cfg.ClientID)
	q.Set("redirect_uri", ts.cfg.RedirectURI)
	q.Set("state", state)
	q.Set("scope", OAuthScope)
	return ts.cfg.baseURL() + "/consent?" + q.Encode()
}

// Token returns a valid access token, refreshing if the cached one is expired.
func (ts *TokenSource) Token(ctx context.Context) (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.tok.Valid(ts.now()) {
		return ts.tok.AccessToken, nil
	}
	if ts.tok == nil || ts.tok.RefreshToken == "" {
		return "", ErrNoToken
	}

	refreshed, err := ts.refresh(ctx, ts.tok.RefreshToken)
	if err != nil {
		return "", err
	}
	ts.setLocked(refreshed)
	return refreshed.AccessToken, nil
}

// Exchange trades an authorization code (from the consent redirect) for a token
// set and caches/persists it. state must match the value used in AuthCodeURL.
func (ts *TokenSource) Exchange(ctx context.Context, code, state string) (*Token, error) {
	body := map[string]string{
		"client_id":     ts.cfg.ClientID,
		"client_secret": ts.cfg.ClientSecret,
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  ts.cfg.RedirectURI,
		"state":         state,
		"scope":         OAuthScope,
	}
	tok, err := ts.postToken(ctx, body)
	if err != nil {
		return nil, err
	}
	ts.mu.Lock()
	ts.setLocked(tok)
	ts.mu.Unlock()
	return tok, nil
}

// refresh exchanges a refresh token for a fresh token set. The caller holds mu.
func (ts *TokenSource) refresh(ctx context.Context, refreshToken string) (*Token, error) {
	body := map[string]string{
		"client_id":     ts.cfg.ClientID,
		"client_secret": ts.cfg.ClientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	tok, err := ts.postToken(ctx, body)
	if err != nil {
		return nil, err
	}
	// seats.aero may or may not rotate the refresh token; keep the prior one if
	// the response omits it.
	if tok.RefreshToken == "" {
		tok.RefreshToken = refreshToken
	}
	return tok, nil
}

// postToken POSTs a JSON body to the token endpoint and decodes the response
// into a Token with an absolute Expiry.
func (ts *TokenSource) postToken(ctx context.Context, body map[string]string) (*Token, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal token request: %w", err)
	}
	endpoint := ts.cfg.baseURL() + "/token"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := ts.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySnippet))
		return nil, &APIError{Status: resp.StatusCode, Body: strings.TrimSpace(string(snippet))}
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if tr.AccessToken == "" {
		return nil, errors.New("token endpoint returned an empty access_token")
	}

	return &Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    tr.TokenType,
		Scope:        tr.Scope,
		Expiry:       ts.now().Add(time.Duration(tr.ExpiresIn) * time.Second),
	}, nil
}

// UserInfo fetches the authenticated user's profile using a valid access token.
func (ts *TokenSource) UserInfo(ctx context.Context) (*UserInfo, error) {
	tok, err := ts.Token(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := ts.cfg.baseURL() + "/userinfo"
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/json")

	resp, err := ts.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySnippet))
		return nil, &APIError{Status: resp.StatusCode, Body: strings.TrimSpace(string(snippet))}
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode userinfo response: %w", err)
	}
	return &info, nil
}

// Peek returns the currently cached token without refreshing. Useful for status
// reporting. The returned pointer is a copy and safe to read.
func (ts *TokenSource) Peek() *Token {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.tok == nil {
		return nil
	}
	cp := *ts.tok
	return &cp
}

// setLocked caches tok and best-effort persists it. The caller holds mu.
func (ts *TokenSource) setLocked(tok *Token) {
	ts.tok = tok
	if ts.store != nil {
		// Persistence failures must not break live requests; surface via stderr
		// is the caller's job. We intentionally ignore the error here.
		_ = ts.store.Save(tok)
	}
}
