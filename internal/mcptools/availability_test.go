package mcptools

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/server"
)

func TestBulkAvailabilityTool_MissingSource(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerBulkAvailability(s, c)

	resp := callTool(t, s, "bulk_availability", map[string]any{
		"cabin": "business",
	})
	checkToolError(t, resp, "source")
}

func TestBulkAvailabilityTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source") != "aeroplan" {
			t.Errorf("expected source=aeroplan, got %s", r.URL.Query().Get("source"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"ID":"a1","RouteID":"r1","Date":"2026-06-15","JAvailable":true}],"count":1,"hasMore":false,"cursor":0}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerBulkAvailability(s, c)

	resp := callTool(t, s, "bulk_availability", map[string]any{
		"source": "aeroplan",
	})
	checkToolSuccess(t, resp)
}

func TestBulkAvailabilityTool_AllParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		checks := map[string]string{
			"source":             "united",
			"cabin":              "first",
			"start_date":         "2026-06-01",
			"end_date":           "2026-06-30",
			"origin_region":      "Europe",
			"destination_region": "Asia",
			"include_filtered":   "true",
			"take":               "500",
			"skip":               "100",
			"cursor":             "999",
		}
		for key, expected := range checks {
			if q.Get(key) != expected {
				t.Errorf("param %s: expected %q, got %q", key, expected, q.Get(key))
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"hasMore":false,"cursor":0}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerBulkAvailability(s, c)

	resp := callTool(t, s, "bulk_availability", map[string]any{
		"source":             "united",
		"cabin":              "first",
		"start_date":         "2026-06-01",
		"end_date":           "2026-06-30",
		"origin_region":      "Europe",
		"destination_region": "Asia",
		"include_filtered":   true,
		"take":               float64(500),
		"skip":               float64(100),
		"cursor":             float64(999),
	})
	checkToolSuccess(t, resp)
}

func TestBulkAvailabilityTool_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("bad request"))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerBulkAvailability(s, c)

	resp := callTool(t, s, "bulk_availability", map[string]any{
		"source": "invalid",
	})
	checkToolError(t, resp, "400")
}
