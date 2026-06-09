package seatsaero

import "fmt"

type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("seats.aero %d: %s", e.Status, e.Body)
}
