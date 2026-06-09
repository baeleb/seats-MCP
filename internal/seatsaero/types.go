package seatsaero

import "encoding/json"

// Route is the route metadata embedded inside an Availability object.
type Route struct {
	ID                 string `json:"ID"`
	OriginAirport      string `json:"OriginAirport"`
	OriginRegion       string `json:"OriginRegion"`
	DestinationAirport string `json:"DestinationAirport"`
	DestinationRegion  string `json:"DestinationRegion"`
	NumDaysOut         int    `json:"NumDaysOut"`
	Distance           int    `json:"Distance"`
	Source             string `json:"Source"`
}

// Availability summarizes redemption availability for a route + date + program.
// Per-cabin fields use Y (economy), W (premium economy), J (business), F (first).
type Availability struct {
	ID         string `json:"ID"`
	RouteID    string `json:"RouteID"`
	Route      Route  `json:"Route"`
	Date       string `json:"Date"`
	ParsedDate string `json:"ParsedDate"`

	YAvailable bool `json:"YAvailable"`
	WAvailable bool `json:"WAvailable"`
	JAvailable bool `json:"JAvailable"`
	FAvailable bool `json:"FAvailable"`

	YMileageCost string `json:"YMileageCost"`
	WMileageCost string `json:"WMileageCost"`
	JMileageCost string `json:"JMileageCost"`
	FMileageCost string `json:"FMileageCost"`

	YRemainingSeats int `json:"YRemainingSeats"`
	WRemainingSeats int `json:"WRemainingSeats"`
	JRemainingSeats int `json:"JRemainingSeats"`
	FRemainingSeats int `json:"FRemainingSeats"`

	YAirlines string `json:"YAirlines"`
	WAirlines string `json:"WAirlines"`
	JAirlines string `json:"JAirlines"`
	FAirlines string `json:"FAirlines"`

	YDirect bool `json:"YDirect"`
	WDirect bool `json:"WDirect"`
	JDirect bool `json:"JDirect"`
	FDirect bool `json:"FDirect"`

	Source    string `json:"Source"`
	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`

	// AvailabilityTrips is only populated when include_trips=true and
	// its concrete shape varies; keep it raw so callers can decode if needed.
	AvailabilityTrips json.RawMessage `json:"AvailabilityTrips,omitempty"`
}

// SearchResponse is the envelope returned by GET /search.
type SearchResponse struct {
	Data    []Availability `json:"data"`
	Count   int            `json:"count"`
	HasMore bool           `json:"hasMore"`
	Cursor  int            `json:"cursor"`
}

// AvailabilityResponse is the envelope returned by GET /availability.
type AvailabilityResponse struct {
	Data    []Availability `json:"data"`
	Count   int            `json:"count"`
	HasMore bool           `json:"hasMore"`
	Cursor  int            `json:"cursor"`
}

// Segment is a single flight leg within a Trip.
type Segment struct {
	ID                  string `json:"ID"`
	RouteID             string `json:"RouteID"`
	AvailabilityID      string `json:"AvailabilityID"`
	AvailabilityTripID  string `json:"AvailabilityTripID"`
	OriginAirport       string `json:"OriginAirport"`
	DestinationAirport  string `json:"DestinationAirport"`
	FlightNumber        string `json:"FlightNumber"`
	Distance            int    `json:"Distance"`
	FareClass           string `json:"FareClass"`
	AircraftName        string `json:"AircraftName,omitempty"`
	AircraftCode        string `json:"AircraftCode,omitempty"`
	DepartsAt           string `json:"DepartsAt"`
	ArrivesAt           string `json:"ArrivesAt"`
	Order               int    `json:"Order"`
}

// Trip is a specific itinerary derived from an Availability.
type Trip struct {
	ID                  string    `json:"ID"`
	RouteID             string    `json:"RouteID"`
	AvailabilityID      string    `json:"AvailabilityID"`
	AvailabilitySegments []Segment `json:"AvailabilitySegments"`
	TotalDuration       int       `json:"TotalDuration"`
	Stops               int       `json:"Stops"`
	Carriers            string    `json:"Carriers"`
	RemainingSeats      int       `json:"RemainingSeats"`
	MileageCost         int       `json:"MileageCost"`
	TotalTaxes          int       `json:"TotalTaxes"`
	TaxesCurrency       string    `json:"TaxesCurrency,omitempty"`
	AllianceCost        int       `json:"AllianceCost,omitempty"`
	FlightNumbers       string    `json:"FlightNumbers"`
	Cabin               string    `json:"Cabin"`
	DepartsAt           string    `json:"DepartsAt"`
	ArrivesAt           string    `json:"ArrivesAt"`
	CreatedAt           string    `json:"CreatedAt"`
	UpdatedAt           string    `json:"UpdatedAt"`
	Source              string    `json:"Source"`
}

type Coordinates struct {
	Latitude  float64 `json:"Lat"`
	Longitude float64 `json:"Lon"`
}

type BookingLink struct {
	Label   string `json:"label"`
	Link    string `json:"link"`
	Primary bool   `json:"primary"`
}

// TripsResponse is the envelope returned by GET /trips/{id}.
type TripsResponse struct {
	Data                 []Trip       `json:"data"`
	OriginCoordinates    Coordinates  `json:"origin_coordinates"`
	DestinationCoordinates Coordinates `json:"destination_coordinates"`
	BookingLinks         []BookingLink `json:"booking_links"`
}

// LiveTrip is a trip returned by POST /live. IDs here are ephemeral and cannot
// be passed to other endpoints.
type LiveTrip struct {
	ID                   string    `json:"ID"`
	AvailabilitySegments []Segment `json:"AvailabilitySegments"`
	MileageCost          int       `json:"MileageCost"`
	TotalTaxes           int       `json:"TotalTaxes"`
	TaxesCurrency        string    `json:"TaxesCurrency,omitempty"`
	Cabin                string    `json:"Cabin"`
	Stops                int       `json:"Stops"`
	Carriers             string    `json:"Carriers"`
	FlightNumbers        string    `json:"FlightNumbers"`
	DepartsAt            string    `json:"DepartsAt"`
	ArrivesAt            string    `json:"ArrivesAt"`
	Filtered             bool      `json:"Filtered"`
}

// LiveResponse is the envelope returned by POST /live.
type LiveResponse struct {
	Results []LiveTrip `json:"results"`
}
