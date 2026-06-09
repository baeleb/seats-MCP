package seatsaero

import (
	"context"
	"net/url"
	"strconv"
)

// Trips fetches the detailed flight itineraries for an Availability ID.
// IDs returned by /live are ephemeral and will not work here.
func (c *Client) Trips(ctx context.Context, availabilityID string, includeFiltered bool) (*TripsResponse, error) {
	q := url.Values{}
	if includeFiltered {
		q.Set("include_filtered", strconv.FormatBool(true))
	}
	out := &TripsResponse{}
	if err := c.doJSON(ctx, "GET", "/trips/"+url.PathEscape(availabilityID), q, nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
