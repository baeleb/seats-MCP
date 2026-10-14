package roomsaero

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHotels_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hotels" {
			t.Errorf("expected /hotels, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("source") != "hyatt" {
			t.Errorf("expected source=hyatt, got %q", q.Get("source"))
		}
		if q.Get("search") != "park" {
			t.Errorf("expected search=park, got %q", q.Get("search"))
		}
		if q.Get("take") != "1000" {
			t.Errorf("expected take=1000, got %q", q.Get("take"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"h1","source":"hyatt","name":"Park Hyatt Tokyo","lowest_award":30000}],"count":1,"has_more":false}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	resp, err := c.Hotels(context.Background(), HotelsParams{
		Source: "hyatt",
		Search: "park",
		Take:   intPtr(1000),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].Name != "Park Hyatt Tokyo" {
		t.Fatalf("unexpected hotels: %+v", resp.Data)
	}
	if resp.Data[0].LowestAward == nil || *resp.Data[0].LowestAward != 30000 {
		t.Errorf("expected lowest_award 30000, got %v", resp.Data[0].LowestAward)
	}
}

func TestRefreshHotel_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/hotels/h1/refresh" {
			t.Errorf("expected /hotels/h1/refresh, got %s", r.URL.Path)
		}
		w.WriteHeader(202)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"hotel_id":"h1","status":"queued","message":"Refreshed shortly."}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	resp, err := c.RefreshHotel(context.Background(), "h1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "queued" || resp.HotelID != "h1" {
		t.Errorf("unexpected refresh response: %+v", resp)
	}
}

func TestRefreshHotel_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":"refresh_limit_exceeded","message":"One refresh every 24 hours."}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	_, err := c.RefreshHotel(context.Background(), "h1")
	if apiErr, ok := err.(*APIError); !ok || apiErr.Status != 429 {
		t.Errorf("unexpected error: %v", err)
	}
}
