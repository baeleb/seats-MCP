package seatsaero

import "context"

// LiveParams are inputs for POST /live (live search).
// Response time is typically 5-15s. Failed searches do not count against quota.
type LiveParams struct {
	OriginAirport      string `json:"origin_airport"`
	DestinationAirport string `json:"destination_airport"`
	DepartureDate      string `json:"departure_date"` // YYYY-MM-DD
	Source             string `json:"source"`
	SeatCount          *int   `json:"seat_count,omitempty"`
	DisableFilters     *bool  `json:"disable_filters,omitempty"`
	ShowDynamicPricing *bool  `json:"show_dynamic_pricing,omitempty"`
}

func (c *Client) LiveSearch(ctx context.Context, p LiveParams) (*LiveResponse, error) {
	out := &LiveResponse{}
	if err := c.doJSON(ctx, "POST", "/live", nil, p, out); err != nil {
		return nil, err
	}
	return out, nil
}
