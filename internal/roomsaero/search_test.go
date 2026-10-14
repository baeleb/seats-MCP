package roomsaero

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
		checks := map[string]string{
			"start_date": "2026-11-20",
			"end_date":   "2026-11-27",
			"location":   "Tokyo",
			"nights":     "2",
			"source":     "hyatt,hilton",
			"max_points": "30000",
			"min_cpp":    "2.5",
			"take":       "100",
		}
		for key, want := range checks {
			if got := q.Get(key); got != want {
				t.Errorf("param %s: expected %q, got %q", key, want, got)
			}
		}
		// Bounding box coords should be omitted when nil.
		for _, key := range []string{"sw_lat", "sw_lng", "ne_lat", "ne_lng", "skip"} {
			if q.Has(key) {
				t.Errorf("expected %s to be absent", key)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SearchResponse{
			Count: 1,
			Data:  []SearchResult{{Hotel: &Hotel{ID: "h1", Name: "Park Hyatt Tokyo"}, CPP: 3}},
		})
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	resp, err := c.Search(context.Background(), SearchParams{
		StartDate: "2026-11-20",
		EndDate:   "2026-11-27",
		Location:  "Tokyo",
		Nights:    intPtr(2),
		Source:    "hyatt,hilton",
		MaxPoints: intPtr(30000),
		MinCPP:    floatPtr(2.5),
		Take:      intPtr(100),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected 1 result, got count=%d len=%d", resp.Count, len(resp.Data))
	}
	if resp.Data[0].Hotel.Name != "Park Hyatt Tokyo" {
		t.Errorf("unexpected hotel name %q", resp.Data[0].Hotel.Name)
	}
}

func TestSearch_BoundingBox(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("sw_lat") != "-1.5" {
			t.Errorf("expected sw_lat=-1.5, got %q", q.Get("sw_lat"))
		}
		if q.Get("ne_lng") != "139.7" {
			t.Errorf("expected ne_lng=139.7, got %q", q.Get("ne_lng"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"has_more":false}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	if _, err := c.Search(context.Background(), SearchParams{
		StartDate: "2026-11-20",
		EndDate:   "2026-11-27",
		SWLat:     floatPtr(-1.5),
		SWLng:     floatPtr(100),
		NELat:     floatPtr(35),
		NELng:     floatPtr(139.7),
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearch_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"missing_parameter","message":"Provide a bounding box or a location."}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	_, err := c.Search(context.Background(), SearchParams{StartDate: "2026-11-20", EndDate: "2026-11-27"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != 400 || apiErr.Code != "missing_parameter" {
		t.Errorf("unexpected error: %+v", apiErr)
	}
}
