package roomsaero

import "testing"

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name   string
		err    *APIError
		expect string
	}{
		{
			name:   "message preferred",
			err:    &APIError{Status: 400, Code: "invalid_parameter", Message: "bad nights", Body: `{"error":"invalid_parameter"}`},
			expect: "rooms.aero 400: bad nights",
		},
		{
			name:   "falls back to body",
			err:    &APIError{Status: 404, Body: "not found"},
			expect: "rooms.aero 404: not found",
		},
		{
			name:   "empty everything",
			err:    &APIError{Status: 500},
			expect: "rooms.aero 500: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, got)
			}
		})
	}
}

func TestAPIError_ImplementsErrorInterface(t *testing.T) {
	var e error = &APIError{Status: 400, Message: "test"}
	if e == nil {
		t.Fatal("APIError does not satisfy error interface")
	}
}
