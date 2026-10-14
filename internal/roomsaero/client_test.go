package roomsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type headerAuthorizer struct{ key string }

func (a headerAuthorizer) Apply(_ context.Context, req *http.Request) error {
	req.Header.Set("Partner-Authorization", a.key)
	return nil
}

func TestDoJSON_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Partner-Authorization") != "test-key" {
			t.Error("missing or wrong auth header")
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Error("missing Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "yes"})
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)

	var out map[string]string
	if err := c.doJSON(context.Background(), "GET", "/test", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["ok"] != "yes" {
		t.Errorf("expected ok=yes, got %v", out)
	}
}

func TestDoJSON_QueryParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source") != "hyatt" {
			t.Errorf("expected source=hyatt, got %s", r.URL.Query().Get("source"))
		}
		if r.URL.Query().Get("nights") != "2" {
			t.Errorf("expected nights=2, got %s", r.URL.Query().Get("nights"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "yes"})
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)

	q := url.Values{}
	q.Set("source", "hyatt")
	q.Set("nights", "2")

	var out map[string]string
	if err := c.doJSON(context.Background(), "GET", "/availability", q, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_ErrorEnvelope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":"user_rate_limit_exceeded","message":"You have used all 1000 of today's API calls."}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)

	err := c.doJSON(context.Background(), "GET", "/search", nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != 429 {
		t.Errorf("expected status 429, got %d", apiErr.Status)
	}
	if apiErr.Code != "user_rate_limit_exceeded" {
		t.Errorf("expected code, got %q", apiErr.Code)
	}
	if apiErr.Message == "" {
		t.Error("expected message to be parsed")
	}
	if !strings.Contains(apiErr.Error(), "rooms.aero 429") {
		t.Errorf("unexpected error text: %s", apiErr.Error())
	}
}

func TestDoJSON_NonJSONError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("  internal error  \n"))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)

	err := c.doJSON(context.Background(), "GET", "/search", nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Body != "internal error" {
		t.Errorf("expected trimmed body, got %q", apiErr.Body)
	}
	if apiErr.Message != "" {
		t.Errorf("expected empty message for non-envelope error, got %q", apiErr.Message)
	}
	if apiErr.Error() != "rooms.aero 500: internal error" {
		t.Errorf("unexpected error text: %s", apiErr.Error())
	}
}

func TestDoJSON_TruncatesLargeErrorBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write(make([]byte, maxErrorBodySnippet+512))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)

	err := c.doJSON(context.Background(), "GET", "/search", nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if len(apiErr.Body) > maxErrorBodySnippet {
		t.Errorf("body not truncated: len %d", len(apiErr.Body))
	}
}

func TestDoJSON_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)

	var out map[string]string
	if err := c.doJSON(context.Background(), "GET", "/search", nil, &out); err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestDoJSON_NilOutput(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	if err := c.doJSON(context.Background(), "GET", "/search", nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_PathConstruction(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/availability/abc-123" {
			t.Errorf("expected /availability/abc-123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	var out map[string]string
	if err := c.doJSON(context.Background(), "GET", "/availability/abc-123", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := c.doJSON(ctx, "GET", "/search", nil, nil); err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestNew_NilHTTPClient(t *testing.T) {
	c := New("test-key", nil)
	if c.http == nil || c.http.Timeout == 0 {
		t.Fatal("expected default http client with timeout")
	}
	if c.base != BaseURL {
		t.Errorf("expected base %s, got %s", BaseURL, c.base)
	}
}

func TestNew_ProvidedHTTPClient(t *testing.T) {
	hc := &http.Client{}
	if c := New("test-key", hc); c.http != hc {
		t.Error("expected provided http client to be used")
	}
}

func TestNewWithAuthorizer_UsesAuthorizer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Partner-Authorization") != "Bearer oauth-token" {
			t.Errorf("expected bearer token, got %q", r.Header.Get("Partner-Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := NewWithAuthorizerBase(headerAuthorizer{key: "Bearer oauth-token"}, ts.Client(), ts.URL)
	var out map[string]string
	if err := c.doJSON(context.Background(), "GET", "/hotels", nil, &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
