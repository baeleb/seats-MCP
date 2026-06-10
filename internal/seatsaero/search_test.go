package seatsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/search" {
			t.Errorf("expected /search, got %s", r.URL.Path)
		}
		if r.Header.Get("Partner-Authorization") != "test-key" {
			t.Error("missing auth header")
		}
		q := r.URL.Query()
		if q.Get("origin_airport") != "LAX" {
			t.Errorf("expected LAX, got %s", q.Get("origin_airport"))
		}
		if q.Get("destination_airport") != "NRT" {
			t.Errorf("expected NRT, got %s", q.Get("destination_airport"))
		}
		if q.Get("start_date") != "2026-06-15" {
			t.Errorf("expected start_date, got %s", q.Get("start_date"))
		}
		if q.Get("only_direct_flights") != "true" {
			t.Errorf("expected only_direct_flights=true, got %s", q.Get("only_direct_flights"))
		}
		if q.Get("take") != "50" {
			t.Errorf("expected take=50, got %s", q.Get("take"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SearchResponse{
			Count:   1,
			HasMore: false,
			Cursor:  100,
			Data: []Availability{
				{ID: "avail1", RouteID: "route1", Date: "2026-06-15", JAvailable: true, JMileageCost: "75000"},
			},
		})
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	resp, err := c.Search(context.Background(), SearchParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
		StartDate:          "2026-06-15",
		OnlyDirectFlights:  boolPtr(true),
		Take:               intPtr(50),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 availability, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "avail1" {
		t.Errorf("expected avail1, got %s", resp.Data[0].ID)
	}
	if !resp.Data[0].JAvailable {
		t.Error("expected JAvailable true")
	}
}

func TestSearch_MinimalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("origin_airport") != "SFO" {
			t.Errorf("expected SFO, got %s", q.Get("origin_airport"))
		}
		// optional params should not be present
		for _, key := range []string{"start_date", "sources", "cabins", "only_direct_flights", "take"} {
			if q.Has(key) {
				t.Errorf("expected %s to be absent", key)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"hasMore":false,"cursor":0}`))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	resp, err := c.Search(context.Background(), SearchParams{
		OriginAirport:      "SFO",
		DestinationAirport: "JFK",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
}

func TestSearch_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte("too many requests"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.Search(context.Background(), SearchParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Status != 429 {
		t.Errorf("expected 429, got %d", apiErr.Status)
	}
}

func TestSearch_AllOptionalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		checks := map[string]string{
			"origin_airport":       "LAX,MSP",
			"destination_airport":  "FRA,LHR",
			"start_date":           "2026-07-01",
			"end_date":             "2026-07-31",
			"sources":              "aeroplan,united",
			"cabins":               "business,first",
			"carriers":             "DL,AA",
			"order_by":             "lowest_mileage",
			"only_direct_flights":  "true",
			"include_trips":        "false",
			"minify_trips":         "true",
			"include_filtered":     "false",
			"take":                 "100",
			"skip":                 "200",
			"cursor":               "12345",
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
	_, err := c.Search(context.Background(), SearchParams{
		OriginAirport:      "LAX,MSP",
		DestinationAirport: "FRA,LHR",
		StartDate:          "2026-07-01",
		EndDate:            "2026-07-31",
		Sources:            "aeroplan,united",
		Cabins:             "business,first",
		Carriers:           "DL,AA",
		OrderBy:            "lowest_mileage",
		OnlyDirectFlights:  boolPtr(true),
		IncludeTrips:       boolPtr(false),
		MinifyTrips:        boolPtr(true),
		IncludeFiltered:    boolPtr(false),
		Take:               intPtr(100),
		Skip:               intPtr(200),
		Cursor:             intPtr(12345),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
