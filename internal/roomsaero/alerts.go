package roomsaero

import (
	"context"
	"net/url"
)

// AlertsParams are the inputs for GET /alerts.
type AlertsParams struct {
	IncludeExpired *bool
	Skip           *int
	Take           *int // 1..500
}

func (p AlertsParams) values() url.Values {
	q := url.Values{}
	setBoolPtr(q, "include_expired", p.IncludeExpired)
	setIntPtr(q, "skip", p.Skip)
	setIntPtr(q, "take", p.Take)
	return q
}

// Alerts lists the alerts on your rooms.aero account, newest first.
func (c *Client) Alerts(ctx context.Context, p AlertsParams) (*AlertResponse, error) {
	out := &AlertResponse{}
	if err := c.doJSON(ctx, "GET", "/alerts", p.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}
