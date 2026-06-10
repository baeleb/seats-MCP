package seatsaero

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const BaseURL = "https://seats.aero/partnerapi"

const maxErrorBodySnippet = 2048

type Client struct {
	apiKey string
	http   *http.Client
	base   string
}

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

// doJSON executes an HTTP request and decodes a JSON response into out.
// path may include a leading slash. query is appended as ?k=v if non-nil.
// body, if non-nil, is JSON-encoded and sent as the request body.
func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body, out any) error {
	endpoint := c.base + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Partner-Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySnippet))
		return &APIError{Status: resp.StatusCode, Body: strings.TrimSpace(string(snippet))}
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
