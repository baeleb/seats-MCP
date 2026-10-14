package roomsaero

import (
	"context"
	"net/url"
)

// HotelsParams are the inputs for GET /hotels (the hotel directory).
// All fields are optional; pass one to narrow the list.
type HotelsParams struct {
	ID         string // rooms.aero ids, comma-delimited, at most 100
	InternalID string // program property codes, comma-delimited, at most 100
	Source     string
	Country    string
	State      string
	City       string
	Brand      string
	Category   string
	Search     string // case-insensitive substring match on the name

	Skip *int
	Take *int // 1..1000
}

func (p HotelsParams) values() url.Values {
	q := url.Values{}
	setIfNotEmpty(q, "id", p.ID)
	setIfNotEmpty(q, "internal_id", p.InternalID)
	setIfNotEmpty(q, "source", p.Source)
	setIfNotEmpty(q, "country", p.Country)
	setIfNotEmpty(q, "state", p.State)
	setIfNotEmpty(q, "city", p.City)
	setIfNotEmpty(q, "brand", p.Brand)
	setIfNotEmpty(q, "category", p.Category)
	setIfNotEmpty(q, "search", p.Search)
	setIntPtr(q, "skip", p.Skip)
	setIntPtr(q, "take", p.Take)
	return q
}

// Hotels lists every property rooms.aero tracks, ordered by name.
func (c *Client) Hotels(ctx context.Context, p HotelsParams) (*HotelResponse, error) {
	out := &HotelResponse{}
	if err := c.doJSON(ctx, "GET", "/hotels", p.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// RefreshHotel asks for a hotel's rates to be retrieved again ahead of
// schedule. Each user may request one refresh every 24 hours, shared with the
// refresh button on the rooms.aero site.
func (c *Client) RefreshHotel(ctx context.Context, hotelID string) (*RefreshResponse, error) {
	out := &RefreshResponse{}
	if err := c.doJSON(ctx, "POST", "/hotels/"+url.PathEscape(hotelID)+"/refresh", nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
