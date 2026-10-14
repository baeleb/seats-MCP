package mcptools

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baeleb/seats-MCP/internal/roomsaero"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func roomsServer(t *testing.T, c *roomsaero.Client) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "1.0")
	RegisterRooms(s, c)
	return s
}

func TestFloatPtrArg(t *testing.T) {
	if v := floatPtrArg(map[string]any{"x": 2.5}, "x"); v == nil || *v != 2.5 {
		t.Errorf("expected 2.5, got %v", v)
	}
	if v := floatPtrArg(map[string]any{"x": 3}, "x"); v == nil || *v != 3 {
		t.Errorf("expected 3, got %v", v)
	}
	if v := floatPtrArg(map[string]any{}, "x"); v != nil {
		t.Errorf("expected nil when absent, got %v", *v)
	}
	if v := floatPtrArg(map[string]any{"x": "nope"}, "x"); v != nil {
		t.Errorf("expected nil for wrong type, got %v", *v)
	}
}

func TestRegisterRooms_AllToolsPresent(t *testing.T) {
	s := roomsServer(t, roomsaero.NewWithBase("key", nil, "http://localhost"))
	// A read-only tool with no required args should succeed against a stub.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"has_more":false}`))
	}))
	defer ts.Close()
	s = roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))

	for _, name := range []string{
		"search_hotel_availability",
		"bulk_hotel_availability",
		"get_hotel_availability_details",
		"list_hotels",
		"refresh_hotel",
		"list_hotel_alerts",
	} {
		found := false
		for _, tl := range s.ListTools() {
			if tl.Tool.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestSearchHotelAvailability_MissingDates(t *testing.T) {
	s := roomsServer(t, roomsaero.NewWithBase("key", nil, "http://localhost"))
	resp := callTool(t, s, "search_hotel_availability", map[string]any{"location": "Tokyo"})
	checkToolError(t, resp, "start_date")
}

func TestSearchHotelAvailability_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("location") != "Tokyo" {
			t.Errorf("expected location=Tokyo, got %q", r.URL.Query().Get("location"))
		}
		if r.URL.Query().Get("nights") != "2" {
			t.Errorf("expected nights=2, got %q", r.URL.Query().Get("nights"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"hotel":{"id":"h1","name":"Park Hyatt Tokyo"},"cpp":3}],"count":1,"has_more":false}`))
	}))
	defer ts.Close()

	s := roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))
	resp := callTool(t, s, "search_hotel_availability", map[string]any{
		"start_date": "2026-11-20",
		"end_date":   "2026-11-27",
		"location":   "Tokyo",
		"nights":     2,
	})
	checkToolSuccess(t, resp)
}

func TestBulkHotelAvailability_RequiresSourceOrHotel(t *testing.T) {
	s := roomsServer(t, roomsaero.NewWithBase("key", nil, "http://localhost"))
	resp := callTool(t, s, "bulk_hotel_availability", map[string]any{"start_date": "2026-11-20"})
	checkToolError(t, resp, "source or hotel_id")
}

func TestBulkHotelAvailability_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("source") != "hyatt" {
			t.Errorf("expected source=hyatt, got %q", r.URL.Query().Get("source"))
		}
		if r.URL.Query().Get("include_hotel") != "false" {
			t.Errorf("expected include_hotel=false, got %q", r.URL.Query().Get("include_hotel"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"a1","hotel_id":"h1"}],"count":1,"has_more":false}`))
	}))
	defer ts.Close()

	s := roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))
	resp := callTool(t, s, "bulk_hotel_availability", map[string]any{
		"source":        "hyatt",
		"include_hotel": false,
	})
	checkToolSuccess(t, resp)
}

func TestGetHotelAvailabilityDetails_MissingID(t *testing.T) {
	s := roomsServer(t, roomsaero.NewWithBase("key", nil, "http://localhost"))
	resp := callTool(t, s, "get_hotel_availability_details", map[string]any{})
	checkToolError(t, resp, "availability_id")
}

func TestGetHotelAvailabilityDetails_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/availability/a1" {
			t.Errorf("expected /availability/a1, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"a1","rooms":[{"id":"r1","name":"1 King Bed"}]}`))
	}))
	defer ts.Close()

	s := roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))
	resp := callTool(t, s, "get_hotel_availability_details", map[string]any{"availability_id": "a1"})
	checkToolSuccess(t, resp)
}

func TestListHotels_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("search") != "park" {
			t.Errorf("expected search=park, got %q", r.URL.Query().Get("search"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"h1","name":"Park Hyatt Tokyo"}],"count":1,"has_more":false}`))
	}))
	defer ts.Close()

	s := roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))
	resp := callTool(t, s, "list_hotels", map[string]any{"search": "park"})
	checkToolSuccess(t, resp)
}

func TestRefreshHotel_MissingID(t *testing.T) {
	s := roomsServer(t, roomsaero.NewWithBase("key", nil, "http://localhost"))
	resp := callTool(t, s, "refresh_hotel", map[string]any{})
	checkToolError(t, resp, "hotel_id")
}

func TestRefreshHotel_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/hotels/h1/refresh" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(202)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"hotel_id":"h1","status":"queued"}`))
	}))
	defer ts.Close()

	s := roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))
	resp := callTool(t, s, "refresh_hotel", map[string]any{"hotel_id": "h1"})
	checkToolSuccess(t, resp)
}

func TestListHotelAlerts_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("include_expired") != "true" {
			t.Errorf("expected include_expired=true, got %q", r.URL.Query().Get("include_expired"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"al1","alert_mode":"one_off"}],"count":1,"has_more":false}`))
	}))
	defer ts.Close()

	s := roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))
	resp := callTool(t, s, "list_hotel_alerts", map[string]any{"include_expired": true})
	checkToolSuccess(t, resp)
}

func TestRoomsTool_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"hotel_not_found","message":"No hotel has that id."}`))
	}))
	defer ts.Close()

	s := roomsServer(t, roomsaero.NewWithBase("key", ts.Client(), ts.URL))
	resp := callTool(t, s, "refresh_hotel", map[string]any{"hotel_id": "missing"})
	checkToolError(t, resp, "rooms.aero 404")
}

func TestApiErrResult_RoomsAeroError(t *testing.T) {
	result := apiErrResult(&roomsaero.APIError{Status: 429, Code: "user_rate_limit_exceeded", Message: "quota used"})
	if result == nil || !result.IsError {
		t.Fatal("expected error result")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if tc.Text != "rooms.aero 429: quota used" {
		t.Errorf("unexpected text: %q", tc.Text)
	}
}
