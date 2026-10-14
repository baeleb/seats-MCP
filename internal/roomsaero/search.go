package roomsaero

import (
	"context"
	"net/url"
)

// SearchParams are the inputs for GET /search.
//
// StartDate and EndDate are required. Describe the area with a full bounding
// box (SWLat, SWLng, NELat, NELng) or a Location to geocode. Source, MaxPoints
// and MinCPP are optional filters; Skip/Take paginate.
type SearchParams struct {
	StartDate string // required, YYYY-MM-DD
	EndDate   string // required, YYYY-MM-DD

	SWLat *float64
	SWLng *float64
	NELat *float64
	NELng *float64

	Location  string // place to geocode, ignored when a full box is given
	Nights    *int   // 1..5
	Source    string // comma-delimited program keys
	MaxPoints *int
	MinCPP    *float64

	Skip *int
	Take *int // 1..500
}

func (p SearchParams) values() url.Values {
	q := url.Values{}
	q.Set("start_date", p.StartDate)
	q.Set("end_date", p.EndDate)
	setFloatPtr(q, "sw_lat", p.SWLat)
	setFloatPtr(q, "sw_lng", p.SWLng)
	setFloatPtr(q, "ne_lat", p.NELat)
	setFloatPtr(q, "ne_lng", p.NELng)
	setIfNotEmpty(q, "location", p.Location)
	setIntPtr(q, "nights", p.Nights)
	setIfNotEmpty(q, "source", p.Source)
	setIntPtr(q, "max_points", p.MaxPoints)
	setFloatPtr(q, "min_cpp", p.MinCPP)
	setIntPtr(q, "skip", p.Skip)
	setIntPtr(q, "take", p.Take)
	return q
}

// Search finds hotels with award availability in an area and date range,
// cheapest award first.
func (c *Client) Search(ctx context.Context, p SearchParams) (*SearchResponse, error) {
	out := &SearchResponse{}
	if err := c.doJSON(ctx, "GET", "/search", p.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}
