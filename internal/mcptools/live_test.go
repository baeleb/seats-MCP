package mcptools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/server"
)

func TestLiveSearchTool_MissingOrigin(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerLiveSearch(s, c)

	resp := callTool(t, s, "live_search", map[string]any{
		"destination_airport": "NRT",
		"departure_date":      "2026-06-15",
		"source":              "aeroplan",
	})
	checkToolError(t, resp, "origin_airport")
}

func TestLiveSearchTool_MissingDestination(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerLiveSearch(s, c)

	resp := callTool(t, s, "live_search", map[string]any{
		"origin_airport": "LAX",
		"departure_date": "2026-06-15",
		"source":         "aeroplan",
	})
	checkToolError(t, resp, "destination_airport")
}

func TestLiveSearchTool_MissingDepartureDate(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerLiveSearch(s, c)

	resp := callTool(t, s, "live_search", map[string]any{
		"origin_airport":      "LAX",
		"destination_airport": "NRT",
		"source":              "aeroplan",
	})
	checkToolError(t, resp, "departure_date")
}

func TestLiveSearchTool_MissingSource(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerLiveSearch(s, c)

	resp := callTool(t, s, "live_search", map[string]any{
		"origin_airport":      "LAX",
		"destination_airport": "NRT",
		"departure_date":      "2026-06-15",
	})
	checkToolError(t, resp, "source")
}

func TestLiveSearchTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body seatsaero.LiveParams
		json.NewDecoder(r.Body).Decode(&body)
		if body.OriginAirport != "LAX" {
			t.Errorf("expected LAX, got %s", body.OriginAirport)
		}
		if body.DestinationAirport != "NRT" {
			t.Errorf("expected NRT, got %s", body.DestinationAirport)
		}
		if body.DepartureDate != "2026-06-15" {
			t.Errorf("expected 2026-06-15, got %s", body.DepartureDate)
		}
		if body.Source != "aeroplan" {
			t.Errorf("expected aeroplan, got %s", body.Source)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"ID":"live1","Cabin":"business","MileageCost":80000,"TotalTaxes":120,"Stops":0,"Carriers":"AC","FlightNumbers":"AC123","Filtered":false}]}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerLiveSearch(s, c)

	resp := callTool(t, s, "live_search", map[string]any{
		"origin_airport":      "LAX",
		"destination_airport": "NRT",
		"departure_date":      "2026-06-15",
		"source":              "aeroplan",
	})
	checkToolSuccess(t, resp)
}

func TestLiveSearchTool_OptionalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body seatsaero.LiveParams
		json.NewDecoder(r.Body).Decode(&body)
		if body.SeatCount == nil || *body.SeatCount != 3 {
			t.Errorf("expected seat_count=3, got %v", body.SeatCount)
		}
		if body.DisableFilters == nil || *body.DisableFilters != true {
			t.Errorf("expected disable_filters=true, got %v", body.DisableFilters)
		}
		if body.ShowDynamicPricing == nil || *body.ShowDynamicPricing != false {
			t.Errorf("expected show_dynamic_pricing=false, got %v", body.ShowDynamicPricing)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerLiveSearch(s, c)

	resp := callTool(t, s, "live_search", map[string]any{
		"origin_airport":       "SFO",
		"destination_airport":  "HND",
		"departure_date":       "2026-07-01",
		"source":               "united",
		"seat_count":           float64(3),
		"disable_filters":      true,
		"show_dynamic_pricing": false,
	})
	checkToolSuccess(t, resp)
}

func TestLiveSearchTool_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(402)
		w.Write([]byte("payment required"))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerLiveSearch(s, c)

	resp := callTool(t, s, "live_search", map[string]any{
		"origin_airport":      "LAX",
		"destination_airport": "NRT",
		"departure_date":      "2026-06-15",
		"source":              "aeroplan",
	})
	checkToolError(t, resp, "402")
}
