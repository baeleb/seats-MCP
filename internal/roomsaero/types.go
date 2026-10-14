package roomsaero

// Hotel is one property in one loyalty program. The same object is returned by
// every endpoint that mentions a hotel. Fields that the API documents as
// nullable (and omits from some responses) are pointers.
type Hotel struct {
	ID            string  `json:"id"`
	Source        string  `json:"source"`
	Program       string  `json:"program"`
	Name          string  `json:"name"`
	Brand         string  `json:"brand"`
	AwardCategory string  `json:"award_category"`
	Address       string  `json:"address"`
	City          string  `json:"city"`
	State         string  `json:"state"`
	Country       string  `json:"country"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	ImageURL      string  `json:"image_url"`
	Description   string  `json:"description"`
	InternalID    string  `json:"internal_id"`
	Open          bool    `json:"open"`
	URL           string  `json:"url"`

	// LowestAward is the cheapest award (points) seen for any upcoming date in
	// the last two weeks; nil when none is known. Populated by Get Hotels and
	// Search; omitted from availability rows.
	LowestAward *int `json:"lowest_award"`

	StandardAwardAvailabilityPercent    *int    `json:"standard_award_availability_percent"`
	StandardAwardAvailabilityWindowDays *int    `json:"standard_award_availability_window_days"`
	StandardAwardAvailabilityUpdatedAt  *string `json:"standard_award_availability_updated_at"`
	LastCheckedAt                       *string `json:"last_checked_at"`
	CreatedAt                           string  `json:"created_at"`
	UpdatedAt                           string  `json:"updated_at"`
}

// BookingLink is a deep link into the program's own booking flow for a stay.
type BookingLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Room is one room type offered on one availability, with its award and cash
// price for the stay. Only Get Availability Details returns rooms, cheapest
// award first.
type Room struct {
	ID               string `json:"id"`
	AvailabilityID   string `json:"availability_id"`
	HotelID          string `json:"hotel_id"`
	Source           string `json:"source"`
	RoomCode         string `json:"room_code"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ImageURL         string `json:"image_url"`
	Suite            bool   `json:"suite"`
	Accessible       bool   `json:"accessible"`
	AwardCost        int    `json:"award_cost"`
	CashCost         int    `json:"cash_cost"`
	CashCostCurrency string `json:"cash_cost_currency"`
	ArrivalDate      string `json:"arrival_date"`
	DepartureDate    string `json:"departure_date"`
	NumNights        int    `json:"num_nights"`
	LastCheckedAt    string `json:"last_checked_at"`
}

// Availability is one hotel on one arrival date for one stay length: the
// cheapest standard room and the cheapest suite, in points and in cash. A cost
// of 0 with its *_available flag false means the rate was not offered.
type Availability struct {
	ID            string `json:"id"`
	HotelID       string `json:"hotel_id"`
	Hotel         *Hotel `json:"hotel,omitempty"`
	Source        string `json:"source"`
	ArrivalDate   string `json:"arrival_date"`
	DepartureDate string `json:"departure_date"`
	NumNights     int    `json:"num_nights"`

	CashCurrencyCode string `json:"cash_currency_code"`

	StandardAwardAvailable bool    `json:"standard_award_available"`
	StandardAwardCost      int     `json:"standard_award_cost"`
	StandardCashAvailable  bool    `json:"standard_cash_available"`
	StandardCashCost       int     `json:"standard_cash_cost"`
	StandardCPP            float64 `json:"standard_cpp"`

	SuiteAwardAvailable bool    `json:"suite_award_available"`
	SuiteAwardCost      int     `json:"suite_award_cost"`
	SuiteCashAvailable  bool    `json:"suite_cash_available"`
	SuiteCashCost       int     `json:"suite_cash_cost"`
	SuiteCPP            float64 `json:"suite_cpp"`

	BookingLink *BookingLink `json:"booking_link,omitempty"`

	// Rooms is only populated by Get Availability Details.
	Rooms []Room `json:"rooms,omitempty"`

	LastCheckedAt string `json:"last_checked_at"`
	CreatedAt     string `json:"created_at"`
}

// SearchResult is one hotel's cheapest options across the searched dates.
type SearchResult struct {
	Hotel *Hotel `json:"hotel"`

	LowestAwardStandard int `json:"lowest_award_standard"`
	LowestAwardSuite    int `json:"lowest_award_suite"`
	LowestCashStandard  int `json:"lowest_cash_standard"`
	LowestCashSuite     int `json:"lowest_cash_suite"`

	CurrencyCode string `json:"currency_code"`

	StandardDate   string `json:"standard_date"`
	StandardNights int    `json:"standard_nights"`
	SuiteDate      string `json:"suite_date"`
	SuiteNights    int    `json:"suite_nights"`

	CPP           float64 `json:"cpp"`
	LastCheckedAt *string `json:"last_checked_at"`
}

// Alert is a notification request on the caller's rooms.aero account. Alerts
// are created and deleted on the rooms.aero site; the API only lists them.
type Alert struct {
	ID      string `json:"id"`
	HotelID string `json:"hotel_id"`
	Hotel   *Hotel `json:"hotel,omitempty"`

	DateType         string `json:"date_type"` // fixed_date | date_range | any_date
	ArrivalDateStart string `json:"arrival_date_start"`
	ArrivalDateEnd   string `json:"arrival_date_end"`
	NumNights        int    `json:"num_nights"`

	AlertMode    string `json:"alert_mode"` // one_off | continuous
	Channel      string `json:"channel"`
	EmailAddress string `json:"email_address"`

	StandardRoom bool `json:"standard_room"`
	SuiteRoom    bool `json:"suite_room"`
	MaxPoints    int  `json:"max_points"`

	Expired   bool   `json:"expired"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SearchResponse is the envelope returned by GET /search.
type SearchResponse struct {
	Data    []SearchResult `json:"data"`
	Count   int            `json:"count"`
	HasMore bool           `json:"has_more"`
	MoreURL string         `json:"more_url"`
}

// AvailabilityResponse is the envelope returned by GET /availability.
type AvailabilityResponse struct {
	Data    []Availability `json:"data"`
	Count   int            `json:"count"`
	HasMore bool           `json:"has_more"`
	MoreURL string         `json:"more_url"`
}

// HotelResponse is the envelope returned by GET /hotels.
type HotelResponse struct {
	Data    []Hotel `json:"data"`
	Count   int     `json:"count"`
	HasMore bool    `json:"has_more"`
	MoreURL string  `json:"more_url"`
}

// AlertResponse is the envelope returned by GET /alerts.
type AlertResponse struct {
	Data    []Alert `json:"data"`
	Count   int     `json:"count"`
	HasMore bool    `json:"has_more"`
	MoreURL string  `json:"more_url"`
}

// RefreshResponse is returned by POST /hotels/{id}/refresh.
type RefreshResponse struct {
	HotelID string `json:"hotel_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
