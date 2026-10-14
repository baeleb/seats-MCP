package roomsaero

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBulkAvailability_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/availability" {
			t.Errorf("expected /availability, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		checks := map[string]string{
			"source":        "hyatt",
			"country":       "Japan",
			"rate_type":     "award",
			"room_type":     "standard",
			"order_by":      "lowest_award",
			"include_hotel": "false",
			"take":          "50",
		}
		for key, want := range checks {
			if got := q.Get(key); got != want {
				t.Errorf("param %s: expected %q, got %q", key, want, got)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"a1","hotel_id":"h1","source":"hyatt","standard_award_available":true,"standard_award_cost":30000}],"count":1,"has_more":true,"more_url":"https://rooms.aero/partnerapi/availability?skip=1"}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	resp, err := c.BulkAvailability(context.Background(), AvailabilityParams{
		Source:       "hyatt",
		Country:      "Japan",
		RateType:     "award",
		RoomType:     "standard",
		OrderBy:      "lowest_award",
		IncludeHotel: boolPtr(false),
		Take:         intPtr(50),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 || !resp.HasMore {
		t.Errorf("unexpected envelope: %+v", resp)
	}
	if resp.Data[0].StandardAwardCost != 30000 {
		t.Errorf("expected award cost 30000, got %d", resp.Data[0].StandardAwardCost)
	}
}

func TestAvailabilityDetail_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/availability/abc/def" {
			t.Errorf("expected escaped path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"abc/def","hotel":{"id":"h1","name":"Park Hyatt Tokyo"},"rooms":[{"id":"r1","name":"1 King Bed","award_cost":30000,"suite":false}],"standard_award_available":true}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	resp, err := c.AvailabilityDetail(context.Background(), "abc/def")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Hotel == nil || resp.Hotel.Name != "Park Hyatt Tokyo" {
		t.Fatalf("expected embedded hotel, got %+v", resp.Hotel)
	}
	if len(resp.Rooms) != 1 || resp.Rooms[0].AwardCost != 30000 {
		t.Errorf("unexpected rooms: %+v", resp.Rooms)
	}
}

func TestAvailabilityDetail_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"error":"availability_not_found","message":"No availability has that id."}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	_, err := c.AvailabilityDetail(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if apiErr, ok := err.(*APIError); !ok || apiErr.Status != 404 || apiErr.Code != "availability_not_found" {
		t.Errorf("unexpected error: %v", err)
	}
}
