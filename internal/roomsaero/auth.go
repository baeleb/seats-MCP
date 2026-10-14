package roomsaero

import (
	"context"
	"net/http"
)

// Authorizer attaches authentication to an outgoing rooms.aero request.
//
// rooms.aero accepts the same two credentials as the seats.aero Partner API —
// a Pro API key, or a Login with Seats.aero OAuth access token — sent in the
// Partner-Authorization header. The interface is declared here (rather than
// reused from the seatsaero package) so this package stays self-contained; Go's
// structural typing means the same Authorizer value can be shared by both
// clients.
type Authorizer interface {
	// Apply mutates req to carry valid credentials. It may perform network I/O
	// (e.g. refreshing an expired OAuth token), so ctx is threaded through.
	Apply(ctx context.Context, req *http.Request) error
}
