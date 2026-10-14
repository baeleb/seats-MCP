package roomsaero

import (
	"context"
	"net/url"
)

// AvailabilityParams are the inputs for GET /availability (bulk availability).
//
// Source is required unless HotelID is given (and vice versa). By default rows
// must offer an award in some room class; RateType and RoomType change that.
type AvailabilityParams struct {
	Source  string
	HotelID string // comma-delimited, at most 50

	StartDate string // YYYY-MM-DD
	EndDate   string // YYYY-MM-DD
	Nights    *int   // 1..5

	Brand    string
	Category string
	Country  string

	RateType string // award | cash | any (default award)
	RoomType string // standard | suite | any (default any)

	MaxPoints    *int
	MinCPP       *float64
	OrderBy      string // arrival_date | lowest_award | lowest_cash | last_checked
	IncludeHotel *bool

	Skip *int
	Take *int // 1..500
}

func (p AvailabilityParams) values() url.Values {
	q := url.Values{}
	setIfNotEmpty(q, "source", p.Source)
	setIfNotEmpty(q, "hotel_id", p.HotelID)
	setIfNotEmpty(q, "start_date", p.StartDate)
	setIfNotEmpty(q, "end_date", p.EndDate)
	setIntPtr(q, "nights", p.Nights)
	setIfNotEmpty(q, "brand", p.Brand)
	setIfNotEmpty(q, "category", p.Category)
	setIfNotEmpty(q, "country", p.Country)
	setIfNotEmpty(q, "rate_type", p.RateType)
	setIfNotEmpty(q, "room_type", p.RoomType)
	setIntPtr(q, "max_points", p.MaxPoints)
	setFloatPtr(q, "min_cpp", p.MinCPP)
	setIfNotEmpty(q, "order_by", p.OrderBy)
	setBoolPtr(q, "include_hotel", p.IncludeHotel)
	setIntPtr(q, "skip", p.Skip)
	setIntPtr(q, "take", p.Take)
	return q
}

// BulkAvailability retrieves every upcoming date with a rate for a program or
// a set of hotels, one object per hotel, check-in date and stay length.
func (c *Client) BulkAvailability(ctx context.Context, p AvailabilityParams) (*AvailabilityResponse, error) {
	out := &AvailabilityResponse{}
	if err := c.doJSON(ctx, "GET", "/availability", p.values(), out); err != nil {
		return nil, err
	}
	return out, nil
}

// AvailabilityDetail retrieves one availability with every room type, prices
// and a booking link for that exact stay.
func (c *Client) AvailabilityDetail(ctx context.Context, availabilityID string) (*Availability, error) {
	out := &Availability{}
	if err := c.doJSON(ctx, "GET", "/availability/"+url.PathEscape(availabilityID), nil, out); err != nil {
		return nil, err
	}
	return out, nil
}
