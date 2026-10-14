package roomsaero

import "fmt"

// APIError is returned for any non-2xx rooms.aero response.
//
// Every rooms.aero error is JSON with a stable machine-readable code and a
// human-readable message: {"error": "...", "message": "..."}. Code and Message
// are populated when the body parses; Body always holds the (truncated) raw
// response for context.
type APIError struct {
	Status  int
	Code    string
	Message string
	Body    string
}

func (e *APIError) Error() string {
	detail := e.Message
	if detail == "" {
		detail = e.Body
	}
	return fmt.Sprintf("rooms.aero %d: %s", e.Status, detail)
}

type errorEnvelope struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
