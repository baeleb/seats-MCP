package mcptools

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/server"
)

func TestSearchTool_MissingOrigin(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerSearch(s, c)

	resp := callTool(t, s, "search_award_availability", map[string]any{
		"destination_airport": "NRT",
	})
	checkToolError(t, resp, "origin_airport")
}

func TestSearchTool_MissingDestination(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerSearch(s, c)

	resp := callTool(t, s, "search_award_availability", map[string]any{
		"origin_airport": "LAX",
	})
	checkToolError(t, resp, "destination_airport")
}

func TestSearchTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"ID":"a1","RouteID":"r1","Date":"2026-06-15","JAvailable":true,"JMileageCost":"75000"}],"count":1,"hasMore":false,"cursor":100}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerSearch(s, c)

	resp := callTool(t, s, "search_award_availability", map[string]any{
		"origin_airport":      "LAX",
		"destination_airport": "NRT",
		"start_date":          "2026-06-15",
		"only_direct_flights": true,
	})
	checkToolSuccess(t, resp)
}

func TestSearchTool_OptionalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("start_date") != "2026-07-01" {
			t.Errorf("expected start_date, got %s", q.Get("start_date"))
		}
		if q.Get("end_date") != "2026-07-31" {
			t.Errorf("expected end_date, got %s", q.Get("end_date"))
		}
		if q.Get("sources") != "aeroplan,united" {
			t.Errorf("expected sources, got %s", q.Get("sources"))
		}
		if q.Get("cabins") != "business,first" {
			t.Errorf("expected cabins, got %s", q.Get("cabins"))
		}
		if q.Get("carriers") != "DL,AA" {
			t.Errorf("expected carriers, got %s", q.Get("carriers"))
		}
		if q.Get("order_by") != "lowest_mileage" {
			t.Errorf("expected order_by, got %s", q.Get("order_by"))
		}
		if q.Get("only_direct_flights") != "true" {
			t.Errorf("expected only_direct_flights=true, got %s", q.Get("only_direct_flights"))
		}
		if q.Get("include_trips") != "false" {
			t.Errorf("expected include_trips=false, got %s", q.Get("include_trips"))
		}
		if q.Get("minify_trips") != "true" {
			t.Errorf("expected minify_trips=true, got %s", q.Get("minify_trips"))
		}
		if q.Get("include_filtered") != "false" {
			t.Errorf("expected include_filtered=false, got %s", q.Get("include_filtered"))
		}
		if q.Get("take") != "100" {
			t.Errorf("expected take=100, got %s", q.Get("take"))
		}
		if q.Get("skip") != "200" {
			t.Errorf("expected skip=200, got %s", q.Get("skip"))
		}
		if q.Get("cursor") != "12345" {
			t.Errorf("expected cursor=12345, got %s", q.Get("cursor"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"hasMore":false,"cursor":0}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerSearch(s, c)

	resp := callTool(t, s, "search_award_availability", map[string]any{
		"origin_airport":       "LAX",
		"destination_airport":  "NRT",
		"start_date":           "2026-07-01",
		"end_date":             "2026-07-31",
		"sources":              "aeroplan,united",
		"cabins":               "business,first",
		"carriers":             "DL,AA",
		"order_by":             "lowest_mileage",
		"only_direct_flights":  true,
		"include_trips":        false,
		"minify_trips":         true,
		"include_filtered":     false,
		"take":                 float64(100),
		"skip":                 float64(200),
		"cursor":               float64(12345),
	})
	checkToolSuccess(t, resp)
}

func TestSearchTool_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte("too many requests"))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerSearch(s, c)

	resp := callTool(t, s, "search_award_availability", map[string]any{
		"origin_airport":      "LAX",
		"destination_airport": "NRT",
	})
	checkToolError(t, resp, "429")
}

func TestSearchTool_NetworkError(t *testing.T) {
	c := seatsaero.NewWithBase("test-key", &http.Client{}, "://bad-url")
	s := server.NewMCPServer("test", "1.0")
	registerSearch(s, c)

	resp := callTool(t, s, "search_award_availability", map[string]any{
		"origin_airport":      "LAX",
		"destination_airport": "NRT",
	})
	checkToolError(t, resp, "build request")
}
