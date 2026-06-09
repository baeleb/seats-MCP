package seatsaero

import (
	"context"
	"net/url"
)

// AvailabilityParams are inputs for GET /availability (bulk availability).
// Source is required.
type AvailabilityParams struct {
	Source            string // required, e.g. "aeroplan"
	Cabin             string // economy | premium | business | first
	StartDate         string
	EndDate           string
	OriginRegion      string
	DestinationRegion string
	IncludeFiltered   *bool
	Take              *int
	Skip              *int
	Cursor            *int
}

func (p AvailabilityParams) values() url.Values {
	q := url.Values{}
	q.Set("source", p.Source)
	setIfNotEmpty(q, "cabin", p.Cabin)
	setIfNotEmpty(q, "start_date", p.StartDate)
	setIfNotEmpty(q, "end_date", p.EndDate)
	setIfNotEmpty(q, "origin_region", p.OriginRegion)
	setIfNotEmpty(q, "destination_region", p.DestinationRegion)
	setBoolPtr(q, "include_filtered", p.IncludeFiltered)
	setIntPtr(q, "take", p.Take)
	setIntPtr(q, "skip", p.Skip)
	setIntPtr(q, "cursor", p.Cursor)
	return q
}

func (c *Client) BulkAvailability(ctx context.Context, p AvailabilityParams) (*AvailabilityResponse, error) {
	out := &AvailabilityResponse{}
	if err := c.doJSON(ctx, "GET", "/availability", p.values(), nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
