package seatsaero

import (
	"context"
	"net/url"
	"strconv"
)

// SearchParams are the inputs for GET /search (cached search).
// Only OriginAirport and DestinationAirport are required; the rest are zero-valued
// optional. Strings already comma-delimited are passed through.
type SearchParams struct {
	OriginAirport      string // required
	DestinationAirport string // required
	StartDate          string // YYYY-MM-DD
	EndDate            string // YYYY-MM-DD
	Sources            string // comma-delimited program keys
	Cabins             string // comma-delimited: economy,premium,business,first
	Carriers           string // comma-delimited carrier codes
	OrderBy            string // "lowest_mileage" or ""
	OnlyDirectFlights  *bool
	IncludeTrips       *bool
	MinifyTrips        *bool
	IncludeFiltered    *bool
	Take               *int
	Skip               *int
	Cursor             *int
}

func (p SearchParams) values() url.Values {
	q := url.Values{}
	q.Set("origin_airport", p.OriginAirport)
	q.Set("destination_airport", p.DestinationAirport)
	setIfNotEmpty(q, "start_date", p.StartDate)
	setIfNotEmpty(q, "end_date", p.EndDate)
	setIfNotEmpty(q, "sources", p.Sources)
	setIfNotEmpty(q, "cabins", p.Cabins)
	setIfNotEmpty(q, "carriers", p.Carriers)
	setIfNotEmpty(q, "order_by", p.OrderBy)
	setBoolPtr(q, "only_direct_flights", p.OnlyDirectFlights)
	setBoolPtr(q, "include_trips", p.IncludeTrips)
	setBoolPtr(q, "minify_trips", p.MinifyTrips)
	setBoolPtr(q, "include_filtered", p.IncludeFiltered)
	setIntPtr(q, "take", p.Take)
	setIntPtr(q, "skip", p.Skip)
	setIntPtr(q, "cursor", p.Cursor)
	return q
}

func (c *Client) Search(ctx context.Context, p SearchParams) (*SearchResponse, error) {
	out := &SearchResponse{}
	if err := c.doJSON(ctx, "GET", "/search", p.values(), nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func setIfNotEmpty(q url.Values, key, val string) {
	if val != "" {
		q.Set(key, val)
	}
}

func setBoolPtr(q url.Values, key string, val *bool) {
	if val != nil {
		q.Set(key, strconv.FormatBool(*val))
	}
}

func setIntPtr(q url.Values, key string, val *int) {
	if val != nil {
		q.Set(key, strconv.Itoa(*val))
	}
}
