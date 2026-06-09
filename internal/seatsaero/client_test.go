package seatsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

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

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	var out map[string]string
	err := c.doJSON(context.Background(), "GET", "/test", nil, nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["ok"] != "yes" {
		t.Errorf("expected ok=yes, got %v", out)
	}
}

func TestDoJSON_QueryParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source") != "aeroplan" {
			t.Errorf("expected source=aeroplan, got %s", r.URL.Query().Get("source"))
		}
		if r.URL.Query().Get("take") != "50" {
			t.Errorf("expected take=50, got %s", r.URL.Query().Get("take"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "yes"})
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	q := url.Values{}
	q.Set("source", "aeroplan")
	q.Set("take", "50")

	var out map[string]string
	err := c.doJSON(context.Background(), "GET", "/search", q, nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_POSTBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["origin_airport"] != "LAX" {
			t.Errorf("expected LAX in body, got %v", body["origin_airport"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	body := LiveParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
		DepartureDate:      "2026-06-15",
		Source:             "aeroplan",
	}

	var out map[string]string
	err := c.doJSON(context.Background(), "POST", "/live", nil, &body, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", out)
	}
}

func TestDoJSON_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte("too many requests"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	var out map[string]string
	err := c.doJSON(context.Background(), "GET", "/search", nil, nil, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 429 {
		t.Errorf("expected status 429, got %d", apiErr.Status)
	}
	if apiErr.Body != "too many requests" {
		t.Errorf("expected body 'too many requests', got %q", apiErr.Body)
	}
}

func TestDoJSON_404Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	err := c.doJSON(context.Background(), "GET", "/trips/unknown", nil, nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("expected status 404, got %d", apiErr.Status)
	}
}

func TestDoJSON_500Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	var out map[string]string
	err := c.doJSON(context.Background(), "GET", "/search", nil, nil, &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDoJSON_MalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	var out map[string]string
	err := c.doJSON(context.Background(), "GET", "/search", nil, nil, &out)
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestDoJSON_NilOutput(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	// out is nil — should discard body without error
	err := c.doJSON(context.Background(), "GET", "/search", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoJSON_LargeErrorBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		large := make([]byte, maxErrorBodySnippet+1024)
		for i := range large {
			large[i] = 'x'
		}
		w.Write(large)
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	err := c.doJSON(context.Background(), "GET", "/search", nil, nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if len(apiErr.Body) > maxErrorBodySnippet {
		t.Errorf("body truncated to %d, got len %d", maxErrorBodySnippet, len(apiErr.Body))
	}
}

func TestDoJSON_TrimmedErrorBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("  error message with spaces  \n"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	err := c.doJSON(context.Background(), "GET", "/search", nil, nil, nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Body != "error message with spaces" {
		t.Errorf("expected trimmed body, got %q", apiErr.Body)
	}
}

func TestDoJSON_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out map[string]string
	err := c.doJSON(ctx, "GET", "/search", nil, nil, &out)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestDoJSON_PathConstruction(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trips/abc-123" {
			t.Errorf("expected path /trips/abc-123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}

	var out map[string]string
	err := c.doJSON(context.Background(), "GET", "/trips/abc-123", nil, nil, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_NilHTTPClient(t *testing.T) {
	c := New("test-key", nil)
	if c.http == nil {
		t.Fatal("expected http client to be set")
	}
	if c.http.Timeout == 0 {
		t.Error("expected timeout to be set")
	}
}

func TestNew_ProvidedHTTPClient(t *testing.T) {
	hc := &http.Client{}
	c := New("test-key", hc)
	if c.http != hc {
		t.Error("expected provided http client to be used")
	}
}

func TestNew_BaseURL(t *testing.T) {
	c := New("test-key", nil)
	if c.base != BaseURL {
		t.Errorf("expected BaseURL %s, got %s", BaseURL, c.base)
	}
}

func TestDoJSON_MarshalBodyError(t *testing.T) {
	c := &Client{apiKey: "test-key", http: &http.Client{}, base: "http://localhost"}

	// channel cannot be JSON-marshaled and will cause an error
	body := make(chan int)
	err := c.doJSON(context.Background(), "POST", "/live", nil, body, nil)
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestDoJSON_BadURL(t *testing.T) {
	c := &Client{apiKey: "test-key", http: &http.Client{}, base: "://bad-url"}

	err := c.doJSON(context.Background(), "GET", "/search", nil, nil, nil)
	if err == nil {
		t.Fatal("expected url error, got nil")
	}
}
