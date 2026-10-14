// Package roomsaero is a typed client for the rooms.aero Partner API, the
// hotel-award companion to the seats.aero Partner API. It shares the same
// credentials (a Pro API key or a Login with Seats.aero OAuth token) sent in
// the Partner-Authorization header.
package roomsaero

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const BaseURL = "https://rooms.aero/partnerapi"

const maxErrorBodySnippet = 2048

type Client struct {
	apiKey string
	auth   Authorizer
	http   *http.Client
	base   string
}

// New builds a Client that authenticates with a static Partner API key via the
// Partner-Authorization header.
func New(apiKey string, httpClient *http.Client) *Client {
	return NewWithBase(apiKey, httpClient, BaseURL)
}

func NewWithBase(apiKey string, httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		apiKey: apiKey,
		http:   httpClient,
		base:   baseURL,
	}
}

// NewWithAuthorizer builds a Client that authenticates using the given
// Authorizer. The same Authorizer used for the seats.aero client works here,
// since rooms.aero accepts identical credentials.
func NewWithAuthorizer(auth Authorizer, httpClient *http.Client) *Client {
	return NewWithAuthorizerBase(auth, httpClient, BaseURL)
}

func NewWithAuthorizerBase(auth Authorizer, httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		auth: auth,
		http: httpClient,
		base: baseURL,
	}
}

// doJSON executes an HTTP request and decodes a JSON response into out.
// path may include a leading slash. query is appended as ?k=v if non-nil.
// Every rooms.aero endpoint is a GET with query parameters except Refresh
// Hotel, a POST with no body.
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, out any) error {
	endpoint := c.base + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if err := c.applyAuth(ctx, req); err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySnippet))
		body := strings.TrimSpace(string(snippet))
		apiErr := &APIError{Status: resp.StatusCode, Body: body}
		var env errorEnvelope
		if json.Unmarshal([]byte(body), &env) == nil {
			apiErr.Code = env.Error
			apiErr.Message = env.Message
		}
		return apiErr
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// applyAuth attaches authentication to req. It prefers the configured
// Authorizer; if none is set it falls back to the Partner-Authorization header
// built from apiKey (preserving the zero-value Client behavior used in tests).
func (c *Client) applyAuth(ctx context.Context, req *http.Request) error {
	if c.auth != nil {
		return c.auth.Apply(ctx, req)
	}
	req.Header.Set("Partner-Authorization", c.apiKey)
	return nil
}
