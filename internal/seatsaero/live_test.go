package seatsaero

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLiveSearch_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/live" {
			t.Errorf("expected /live, got %s", r.URL.Path)
		}
		if r.Header.Get("Partner-Authorization") != "test-key" {
			t.Error("missing auth header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing Content-Type header")
		}
		var body LiveParams
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
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
		json.NewEncoder(w).Encode(LiveResponse{
			Results: []LiveTrip{
				{
					ID:            "live1",
					Cabin:         "business",
					MileageCost:   80000,
					TotalTaxes:    120,
					Stops:         0,
					Carriers:      "AC",
					FlightNumbers: "AC123",
				},
			},
		})
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	resp, err := c.LiveSearch(context.Background(), LiveParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
		DepartureDate:      "2026-06-15",
		Source:             "aeroplan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].ID != "live1" {
		t.Errorf("expected live1, got %s", resp.Results[0].ID)
	}
	if resp.Results[0].MileageCost != 80000 {
		t.Errorf("expected 80000, got %d", resp.Results[0].MileageCost)
	}
}

func TestLiveSearch_OptionalParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body LiveParams
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

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.LiveSearch(context.Background(), LiveParams{
		OriginAirport:      "SFO",
		DestinationAirport: "HND",
		DepartureDate:      "2026-07-01",
		Source:             "united",
		SeatCount:          intPtr(3),
		DisableFilters:     boolPtr(true),
		ShowDynamicPricing: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLiveSearch_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(402)
		w.Write([]byte("payment required"))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	_, err := c.LiveSearch(context.Background(), LiveParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
		DepartureDate:      "2026-06-15",
		Source:             "aeroplan",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Status != 402 {
		t.Errorf("expected 402, got %d", apiErr.Status)
	}
}

func TestLiveSearch_EmptyResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[]}`))
	}))
	defer ts.Close()

	c := &Client{apiKey: "test-key", http: ts.Client(), base: ts.URL}
	resp, err := c.LiveSearch(context.Background(), LiveParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
		DepartureDate:      "2026-06-15",
		Source:             "aeroplan",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Results))
	}
}
