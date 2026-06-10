package seatsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrips_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/trips/avail-123" {
			t.Errorf("expected /trips/avail-123, got %s", r.URL.Path)
		}
		if r.Header.Get("Partner-Authorization") != "test-key" {
			t.Error("missing auth header")
		}
		if r.URL.Query().Has("include_filtered") {
			t.Error("expected include_filtered to be absent")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TripsResponse{
			Data: []Trip{
				{ID: "trip1", RouteID: "route1", AvailabilityID: "avail-123", Cabin: "business", MileageCost: 75000},
			},
		})
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	resp, err := c.Trips(context.Background(), "avail-123", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "trip1" {
		t.Errorf("expected trip1, got %s", resp.Data[0].ID)
	}
	if resp.Data[0].MileageCost != 75000 {
		t.Errorf("expected 75000, got %d", resp.Data[0].MileageCost)
	}
}

func TestTrips_WithFiltered(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trips/avail-456" {
			t.Errorf("expected /trips/avail-456, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("include_filtered") != "true" {
			t.Errorf("expected include_filtered=true, got %s", r.URL.Query().Get("include_filtered"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"origin_coordinates":{"Lat":0,"Lon":0},"destination_coordinates":{"Lat":0,"Lon":0},"booking_links":[]}`))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.Trips(context.Background(), "avail-456", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrips_URLEncoding(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/trips/avail%2Fwith%2Fslashes" && r.URL.Path != "/trips/avail/with/slashes" {
			t.Errorf("expected encoded path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"origin_coordinates":{"Lat":0,"Lon":0},"destination_coordinates":{"Lat":0,"Lon":0},"booking_links":[]}`))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.Trips(context.Background(), "avail/with/slashes", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrips_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.Trips(context.Background(), "nonexistent", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Status != 404 {
		t.Errorf("expected 404, got %d", apiErr.Status)
	}
}
