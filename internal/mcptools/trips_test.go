package mcptools

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/server"
)

func TestGetTripsTool_MissingID(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	c := seatsaero.New("test-key", nil)
	registerGetTrips(s, c)

	resp := callTool(t, s, "get_trips", map[string]any{})
	checkToolError(t, resp, "availability_id")
}

func TestGetTripsTool_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"ID":"trip1","Cabin":"business","MileageCost":75000}],"origin_coordinates":{"Lat":0,"Lon":0},"destination_coordinates":{"Lat":0,"Lon":0},"booking_links":[]}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerGetTrips(s, c)

	resp := callTool(t, s, "get_trips", map[string]any{
		"availability_id": "avail-123",
	})
	checkToolSuccess(t, resp)
}

func TestGetTripsTool_WithFiltered(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("include_filtered") != "true" {
			t.Error("expected include_filtered=true")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"origin_coordinates":{"Lat":0,"Lon":0},"destination_coordinates":{"Lat":0,"Lon":0},"booking_links":[]}`))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerGetTrips(s, c)

	resp := callTool(t, s, "get_trips", map[string]any{
		"availability_id":  "avail-456",
		"include_filtered": true,
	})
	checkToolSuccess(t, resp)
}

func TestGetTripsTool_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer ts.Close()

	c := seatsaero.NewWithBase("test-key", ts.Client(), ts.URL)
	s := server.NewMCPServer("test", "1.0")
	registerGetTrips(s, c)

	resp := callTool(t, s, "get_trips", map[string]any{
		"availability_id": "nonexistent",
	})
	checkToolError(t, resp, "404")
}
