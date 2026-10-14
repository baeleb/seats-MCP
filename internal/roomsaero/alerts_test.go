package roomsaero

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAlerts_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/alerts" {
			t.Errorf("expected /alerts, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("include_expired") != "true" {
			t.Errorf("expected include_expired=true, got %q", r.URL.Query().Get("include_expired"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"a1","hotel_id":"h1","alert_mode":"one_off","max_points":40000,"expired":false}],"count":1,"has_more":false}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	resp, err := c.Alerts(context.Background(), AlertsParams{IncludeExpired: boolPtr(true)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].AlertMode != "one_off" {
		t.Fatalf("unexpected alerts: %+v", resp.Data)
	}
}

func TestAlerts_OmitsDefaults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("include_expired") {
			t.Error("expected include_expired to be absent when unset")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[],"count":0,"has_more":false}`))
	}))
	defer ts.Close()

	c := NewWithBase("test-key", ts.Client(), ts.URL)
	if _, err := c.Alerts(context.Background(), AlertsParams{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
