package seatsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBulkAvailability_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/availability" {
			t.Errorf("expected /availability, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("source") != "aeroplan" {
			t.Errorf("expected aeroplan, got %s", q.Get("source"))
		}
		if q.Get("cabin") != "business" {
			t.Errorf("expected business, got %s", q.Get("cabin"))
		}
		if q.Get("origin_region") != "North America" {
			t.Errorf("expected North America, got %s", q.Get("origin_region"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AvailabilityResponse{
			Count:   2,
			HasMore: true,
			Cursor:  50,
			Data: []Availability{
				{ID: "a1", RouteID: "r1", Date: "2026-06-15", JAvailable: true},
				{ID: "a2", RouteID: "r2", Date: "2026-06-16", FAvailable: true},
			},
		})
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	resp, err := c.BulkAvailability(context.Background(), AvailabilityParams{
		Source:       "aeroplan",
		Cabin:        "business",
		OriginRegion: "North America",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("expected count 2, got %d", resp.Count)
	}
	if !resp.HasMore {
		t.Error("expected HasMore true")
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 availabilities, got %d", len(resp.Data))
	}
}

func TestBulkAvailability_MinimalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("source") != "united" {
			t.Errorf("expected united, got %s", q.Get("source"))
		}
		for _, key := range []string{"cabin", "origin_region", "take"} {
			if q.Has(key) {
				t.Errorf("expected %s to be absent", key)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"hasMore":false,"cursor":0}`))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	resp, err := c.BulkAvailability(context.Background(), AvailabilityParams{
		Source: "united",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
}

func TestBulkAvailability_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("missing required field"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.BulkAvailability(context.Background(), AvailabilityParams{
		Source: "invalid",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Status != 400 {
		t.Errorf("expected 400, got %d", apiErr.Status)
	}
}

func TestBulkAvailability_AllParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		checks := map[string]string{
			"source":             "delta",
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

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.BulkAvailability(context.Background(), AvailabilityParams{
		Source:            "delta",
		Cabin:             "first",
		StartDate:         "2026-06-01",
		EndDate:           "2026-06-30",
		OriginRegion:      "Europe",
		DestinationRegion: "Asia",
		IncludeFiltered:   boolPtr(true),
		Take:              intPtr(500),
		Skip:              intPtr(100),
		Cursor:            intPtr(999),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
