package seatsaero

import "testing"

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name   string
		err    *APIError
		expect string
	}{
		{
			name:   "standard error",
			err:    &APIError{Status: 400, Body: "bad request"},
			expect: "seats.aero 400: bad request",
		},
		{
			name:   "no body",
			err:    &APIError{Status: 500, Body: ""},
			expect: "seats.aero 500: ",
		},
		{
			name:   "404 not found",
			err:    &APIError{Status: 404, Body: "not found"},
			expect: "seats.aero 404: not found",
		},
		{
			name:   "429 rate limit",
			err:    &APIError{Status: 429, Body: "too many requests"},
			expect: "seats.aero 429: too many requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, got)
			}
		})
	}
}

func TestAPIError_ImplementsErrorInterface(t *testing.T) {
	var e error = &APIError{Status: 400, Body: "test"}
	if e == nil {
		t.Fatal("APIError does not satisfy error interface")
	}
}
