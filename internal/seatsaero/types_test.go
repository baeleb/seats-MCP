package seatsaero

import (
	"encoding/json"
	"testing"
)

func TestSearchParamsValues(t *testing.T) {
	tests := []struct {
		name   string
		params SearchParams
		check  func(t *testing.T, q string)
	}{
		{
			name: "required fields only",
			params: SearchParams{
				OriginAirport:      "LAX",
				DestinationAirport: "NRT",
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "origin_airport=LAX")
				assertContains(t, q, "destination_airport=NRT")
				assertNotContains(t, q, "start_date")
				assertNotContains(t, q, "end_date")
			},
		},
		{
			name: "full params with dates and sources",
			params: SearchParams{
				OriginAirport:      "SFO,LAX",
				DestinationAirport: "FRA,LHR",
				StartDate:          "2026-06-15",
				EndDate:            "2026-06-22",
				Sources:            "aeroplan,united",
				Cabins:             "business,first",
				Carriers:           "DL,AA",
				OrderBy:            "lowest_mileage",
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "origin_airport=SFO%2CLAX")
				assertContains(t, q, "destination_airport=FRA%2CLHR")
				assertContains(t, q, "start_date=2026-06-15")
				assertContains(t, q, "end_date=2026-06-22")
				assertContains(t, q, "sources=aeroplan%2Cunited")
				assertContains(t, q, "cabins=business%2Cfirst")
				assertContains(t, q, "carriers=DL%2CAA")
				assertContains(t, q, "order_by=lowest_mileage")
			},
		},
		{
			name: "bool params false",
			params: SearchParams{
				OriginAirport:      "LAX",
				DestinationAirport: "MSP",
				OnlyDirectFlights:  boolPtr(false),
				IncludeTrips:       boolPtr(false),
				MinifyTrips:        boolPtr(false),
				IncludeFiltered:    boolPtr(false),
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "only_direct_flights=false")
				assertContains(t, q, "include_trips=false")
				assertContains(t, q, "minify_trips=false")
				assertContains(t, q, "include_filtered=false")
			},
		},
		{
			name: "bool params true",
			params: SearchParams{
				OriginAirport:      "LAX",
				DestinationAirport: "MSP",
				OnlyDirectFlights:  boolPtr(true),
				IncludeTrips:       boolPtr(true),
				MinifyTrips:        boolPtr(true),
				IncludeFiltered:    boolPtr(true),
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "only_direct_flights=true")
				assertContains(t, q, "include_trips=true")
				assertContains(t, q, "minify_trips=true")
				assertContains(t, q, "include_filtered=true")
			},
		},
		{
			name: "int params",
			params: SearchParams{
				OriginAirport:      "LAX",
				DestinationAirport: "MSP",
				Take:               intPtr(100),
				Skip:               intPtr(50),
				Cursor:             intPtr(12345),
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "take=100")
				assertContains(t, q, "skip=50")
				assertContains(t, q, "cursor=12345")
			},
		},
		{
			name: "nil bool pointers omitted",
			params: SearchParams{
				OriginAirport:      "LAX",
				DestinationAirport: "MSP",
			},
			check: func(t *testing.T, q string) {
				assertNotContains(t, q, "only_direct_flights")
				assertNotContains(t, q, "include_trips")
				assertNotContains(t, q, "minify_trips")
				assertNotContains(t, q, "include_filtered")
			},
		},
		{
			name: "nil int pointers omitted",
			params: SearchParams{
				OriginAirport:      "LAX",
				DestinationAirport: "MSP",
			},
			check: func(t *testing.T, q string) {
				assertNotContains(t, q, "take=")
				assertNotContains(t, q, "skip=")
				assertNotContains(t, q, "cursor=")
			},
		},
		{
			name: "empty string params omitted",
			params: SearchParams{
				OriginAirport:      "LAX",
				DestinationAirport: "MSP",
				StartDate:          "",
				EndDate:            "",
				Sources:            "",
				Cabins:             "",
				Carriers:           "",
				OrderBy:            "",
			},
			check: func(t *testing.T, q string) {
				assertNotContains(t, q, "start_date=")
				assertNotContains(t, q, "end_date=")
				assertNotContains(t, q, "sources=")
				assertNotContains(t, q, "cabins=")
				assertNotContains(t, q, "carriers=")
				assertNotContains(t, q, "order_by=")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.params.values().Encode()
			tt.check(t, q)
		})
	}
}

func TestAvailabilityParamsValues(t *testing.T) {
	tests := []struct {
		name   string
		params AvailabilityParams
		check  func(t *testing.T, q string)
	}{
		{
			name: "source only",
			params: AvailabilityParams{
				Source: "aeroplan",
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "source=aeroplan")
				assertNotContains(t, q, "cabin")
				assertNotContains(t, q, "origin_region")
			},
		},
		{
			name: "with region and date filters",
			params: AvailabilityParams{
				Source:            "united",
				Cabin:             "business",
				StartDate:         "2026-06-01",
				EndDate:           "2026-06-30",
				OriginRegion:      "North America",
				DestinationRegion: "Asia",
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "source=united")
				assertContains(t, q, "cabin=business")
				assertContains(t, q, "start_date=2026-06-01")
				assertContains(t, q, "end_date=2026-06-30")
				assertContains(t, q, "origin_region=North+America")
				assertContains(t, q, "destination_region=Asia")
			},
		},
		{
			name: "pagination params",
			params: AvailabilityParams{
				Source: "delta",
				Take:   intPtr(200),
				Skip:   intPtr(100),
				Cursor: intPtr(9999),
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "take=200")
				assertContains(t, q, "skip=100")
				assertContains(t, q, "cursor=9999")
			},
		},
		{
			name: "bool param true",
			params: AvailabilityParams{
				Source:          "emirates",
				IncludeFiltered: boolPtr(true),
			},
			check: func(t *testing.T, q string) {
				assertContains(t, q, "include_filtered=true")
			},
		},
		{
			name: "nil pointers omitted",
			params: AvailabilityParams{
				Source: "alaska",
			},
			check: func(t *testing.T, q string) {
				assertNotContains(t, q, "take=")
				assertNotContains(t, q, "skip=")
				assertNotContains(t, q, "cursor=")
				assertNotContains(t, q, "include_filtered")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.params.values().Encode()
			tt.check(t, q)
		})
	}
}

func TestSearchResponseJSON(t *testing.T) {
	body := `{
		"data": [
			{
				"ID": "abc123",
				"RouteID": "route1",
				"Route": {
					"ID": "route1",
					"OriginAirport": "LAX",
					"OriginRegion": "North America",
					"DestinationAirport": "NRT",
					"DestinationRegion": "Asia",
					"NumDaysOut": 0,
					"Distance": 5450,
					"Source": "aeroplan"
				},
				"Date": "2026-06-15",
				"ParsedDate": "2026-06-15T00:00:00Z",
				"YAvailable": true,
				"WAvailable": false,
				"JAvailable": true,
				"FAvailable": false,
				"YMileageCost": "55000",
				"WMileageCost": "0",
				"JMileageCost": "75000",
				"FMileageCost": "0",
				"YRemainingSeats": 9,
				"WRemainingSeats": 0,
				"JRemainingSeats": 2,
				"FRemainingSeats": 0,
				"YAirlines": "UA",
				"WAirlines": "",
				"JAirlines": "AC",
				"FAirlines": "",
				"YDirect": false,
				"WDirect": false,
				"JDirect": true,
				"FDirect": false,
				"Source": "aeroplan",
				"CreatedAt": "2026-01-01T00:00:00Z",
				"UpdatedAt": "2026-06-09T00:00:00Z"
			}
		],
		"count": 1,
		"hasMore": false,
		"cursor": 12345
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Data))
	}

	a := resp.Data[0]
	if a.ID != "abc123" {
		t.Errorf("expected ID abc123, got %s", a.ID)
	}
	if a.Route.OriginAirport != "LAX" {
		t.Errorf("expected LAX, got %s", a.Route.OriginAirport)
	}
	if !a.JAvailable {
		t.Error("expected JAvailable to be true")
	}
	if a.JMileageCost != "75000" {
		t.Errorf("expected 75000, got %s", a.JMileageCost)
	}
	if a.JRemainingSeats != 2 {
		t.Errorf("expected 2 J seats, got %d", a.JRemainingSeats)
	}
	if !a.JDirect {
		t.Error("expected JDirect to be true")
	}
	if a.JAirlines != "AC" {
		t.Errorf("expected AC, got %s", a.JAirlines)
	}

	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if resp.HasMore {
		t.Error("expected HasMore false")
	}
	if resp.Cursor != 12345 {
		t.Errorf("expected cursor 12345, got %d", resp.Cursor)
	}
}

func TestAvailabilityResponseJSON(t *testing.T) {
	body := `{
		"data": [
			{
				"ID": "xyz789",
				"RouteID": "route2",
				"Route": {
					"ID": "route2",
					"OriginAirport": "SFO",
					"OriginRegion": "North America",
					"DestinationAirport": "HND",
					"DestinationRegion": "Asia",
					"NumDaysOut": 0,
					"Distance": 5150,
					"Source": "united"
				},
				"Date": "2026-06-20",
				"ParsedDate": "2026-06-20T00:00:00Z",
				"YAvailable": false,
				"WAvailable": false,
				"JAvailable": true,
				"FAvailable": false,
				"YMileageCost": "0",
				"WMileageCost": "0",
				"JMileageCost": "88000",
				"FMileageCost": "0",
				"YRemainingSeats": 0,
				"WRemainingSeats": 0,
				"JRemainingSeats": 5,
				"FRemainingSeats": 0,
				"YAirlines": "",
				"WAirlines": "",
				"JAirlines": "UA",
				"FAirlines": "",
				"YDirect": false,
				"WDirect": false,
				"JDirect": false,
				"FDirect": false,
				"Source": "united",
				"CreatedAt": "2026-01-01T00:00:00Z",
				"UpdatedAt": "2026-06-09T00:00:00Z"
			}
		],
		"count": 1,
		"hasMore": true,
		"cursor": 67890
	}`

	var resp AvailabilityResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Data))
	}

	a := resp.Data[0]
	if a.ID != "xyz789" {
		t.Errorf("expected ID xyz789, got %s", a.ID)
	}
	if a.JMileageCost != "88000" {
		t.Errorf("expected 88000, got %s", a.JMileageCost)
	}

	if resp.Count != 1 {
		t.Errorf("expected count 1, got %d", resp.Count)
	}
	if !resp.HasMore {
		t.Error("expected HasMore true")
	}
}

func TestTripsResponseJSON(t *testing.T) {
	body := `{
		"data": [
			{
				"ID": "trip1",
				"RouteID": "route1",
				"AvailabilityID": "avail1",
				"AvailabilitySegments": [
					{
						"ID": "seg1",
						"RouteID": "route1",
						"AvailabilityID": "avail1",
						"AvailabilityTripID": "trip1",
						"OriginAirport": "LAX",
						"DestinationAirport": "NRT",
						"FlightNumber": "UA123",
						"Distance": 5450,
						"FareClass": "J",
						"AircraftName": "787",
						"AircraftCode": "787",
						"DepartsAt": "2026-06-15T10:00:00Z",
						"ArrivesAt": "2026-06-16T13:00:00Z",
						"Order": 0
					}
				],
				"TotalDuration": 660,
				"Stops": 0,
				"Carriers": "UA",
				"RemainingSeats": 4,
				"MileageCost": 75000,
				"TotalTaxes": 560,
				"TaxesCurrency": "USD",
				"TaxesCurrencySymbol": "$",
				"AllianceCost": 80000,
				"FlightNumbers": "UA123",
				"Cabin": "business",
				"DepartsAt": "2026-06-15T10:00:00Z",
				"ArrivesAt": "2026-06-16T13:00:00Z",
				"CreatedAt": "2026-01-01T00:00:00Z",
				"UpdatedAt": "2026-06-09T00:00:00Z",
				"Source": "aeroplan"
			}
		],
		"origin_coordinates": {
			"Lat": 33.9425,
			"Lon": -118.4081
		},
		"destination_coordinates": {
			"Lat": 35.7647,
			"Lon": 140.3864
		},
		"booking_links": [
			{
				"label": "Book via Aeroplan",
				"link": "https://www.aircanada.com/aeroplan/redeem",
				"primary": true
			}
		]
	}`

	var resp TripsResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(resp.Data))
	}

	trip := resp.Data[0]
	if trip.ID != "trip1" {
		t.Errorf("expected ID trip1, got %s", trip.ID)
	}
	if trip.MileageCost != 75000 {
		t.Errorf("expected 75000 miles, got %d", trip.MileageCost)
	}
	if trip.TotalTaxes != 560 {
		t.Errorf("expected 560 taxes, got %d", trip.TotalTaxes)
	}
	if trip.TaxesCurrency != "USD" {
		t.Errorf("expected USD currency, got %s", trip.TaxesCurrency)
	}
	if trip.TaxesCurrencySymbol != "$" {
		t.Errorf("expected $ symbol, got %s", trip.TaxesCurrencySymbol)
	}
	if trip.AllianceCost != 80000 {
		t.Errorf("expected 80000 alliance cost, got %d", trip.AllianceCost)
	}
	if trip.RemainingSeats != 4 {
		t.Errorf("expected 4 seats, got %d", trip.RemainingSeats)
	}
	if trip.Cabin != "business" {
		t.Errorf("expected business cabin, got %s", trip.Cabin)
	}
	if trip.Stops != 0 {
		t.Errorf("expected 0 stops, got %d", trip.Stops)
	}
	if trip.TotalDuration != 660 {
		t.Errorf("expected 660 duration, got %d", trip.TotalDuration)
	}
	if trip.Carriers != "UA" {
		t.Errorf("expected UA, got %s", trip.Carriers)
	}
	if trip.Source != "aeroplan" {
		t.Errorf("expected aeroplan source, got %s", trip.Source)
	}

	if len(trip.AvailabilitySegments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(trip.AvailabilitySegments))
	}
	seg := trip.AvailabilitySegments[0]
	if seg.FlightNumber != "UA123" {
		t.Errorf("expected UA123, got %s", seg.FlightNumber)
	}
	if seg.FareClass != "J" {
		t.Errorf("expected J fare, got %s", seg.FareClass)
	}
	if seg.AircraftName != "787" {
		t.Errorf("expected 787, got %s", seg.AircraftName)
	}
	if seg.Order != 0 {
		t.Errorf("expected Order 0, got %d", seg.Order)
	}

	if resp.OriginCoordinates.Latitude != 33.9425 {
		t.Errorf("expected origin lat 33.9425, got %f", resp.OriginCoordinates.Latitude)
	}
	if resp.OriginCoordinates.Longitude != -118.4081 {
		t.Errorf("expected origin lon -118.4081, got %f", resp.OriginCoordinates.Longitude)
	}
	if resp.DestinationCoordinates.Latitude != 35.7647 {
		t.Errorf("expected dest lat 35.7647, got %f", resp.DestinationCoordinates.Latitude)
	}
	if resp.DestinationCoordinates.Longitude != 140.3864 {
		t.Errorf("expected dest lon 140.3864, got %f", resp.DestinationCoordinates.Longitude)
	}

	if len(resp.BookingLinks) != 1 {
		t.Fatalf("expected 1 booking link, got %d", len(resp.BookingLinks))
	}
	if resp.BookingLinks[0].Label != "Book via Aeroplan" {
		t.Errorf("expected label, got %s", resp.BookingLinks[0].Label)
	}
	if !resp.BookingLinks[0].Primary {
		t.Error("expected primary to be true")
	}
}

func TestLiveResponseJSON(t *testing.T) {
	body := `{
		"results": [
			{
				"ID": "live1",
				"AvailabilitySegments": [
					{
						"ID": "seg2",
						"RouteID": "route2",
						"AvailabilityID": "avail2",
						"AvailabilityTripID": "live1",
						"OriginAirport": "LAX",
						"DestinationAirport": "ICN",
						"FlightNumber": "KE12",
						"Distance": 5987,
						"FareClass": "J",
						"AircraftName": "748",
						"AircraftCode": "748",
						"DepartsAt": "2026-06-15T12:00:00Z",
						"ArrivesAt": "2026-06-16T16:00:00Z",
						"Order": 0
					}
				],
				"MileageCost": 80000,
				"TotalTaxes": 120,
				"TaxesCurrency": "KRW",
				"TaxesCurrencySymbol": "₩",
				"AllianceCost": 85000,
				"RemainingSeats": 3,
				"TotalDuration": 780,
				"Cabin": "business",
				"Stops": 0,
				"Carriers": "KE",
				"FlightNumbers": "KE12",
				"DepartsAt": "2026-06-15T12:00:00Z",
				"ArrivesAt": "2026-06-16T16:00:00Z",
				"Source": "koreanair",
				"Filtered": false
			}
		]
	}`

	var resp LiveResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	lt := resp.Results[0]
	if lt.ID != "live1" {
		t.Errorf("expected ID live1, got %s", lt.ID)
	}
	if lt.MileageCost != 80000 {
		t.Errorf("expected 80000 miles, got %d", lt.MileageCost)
	}
	if lt.TotalTaxes != 120 {
		t.Errorf("expected 120 taxes, got %d", lt.TotalTaxes)
	}
	if lt.TaxesCurrency != "KRW" {
		t.Errorf("expected KRW, got %s", lt.TaxesCurrency)
	}
	if lt.TaxesCurrencySymbol != "₩" {
		t.Errorf("expected won symbol, got %s", lt.TaxesCurrencySymbol)
	}
	if lt.AllianceCost != 85000 {
		t.Errorf("expected 85000 alliance cost, got %d", lt.AllianceCost)
	}
	if lt.RemainingSeats != 3 {
		t.Errorf("expected 3 seats, got %d", lt.RemainingSeats)
	}
	if lt.TotalDuration != 780 {
		t.Errorf("expected 780 duration, got %d", lt.TotalDuration)
	}
	if lt.Cabin != "business" {
		t.Errorf("expected business cabin, got %s", lt.Cabin)
	}
	if lt.Source != "koreanair" {
		t.Errorf("expected koreanair source, got %s", lt.Source)
	}
	if lt.Filtered {
		t.Error("expected Filtered false")
	}
}

func TestEmptySearchResponse(t *testing.T) {
	body := `{"data":[],"count":0,"hasMore":false,"cursor":0}`
	var resp SearchResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to unmarshal empty response: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 results, got %d", len(resp.Data))
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
}

func TestAvailabilityTripsRawMessage(t *testing.T) {
	body := `{
		"ID": "avail1",
		"RouteID": "r1",
		"Route": {"ID": "r1", "OriginAirport": "LAX", "OriginRegion": "North America", "DestinationAirport": "NRT", "DestinationRegion": "Asia", "NumDaysOut": 0, "Distance": 5450, "Source": "united"},
		"Date": "2026-06-15",
		"ParsedDate": "2026-06-15T00:00:00Z",
		"YAvailable": true, "WAvailable": false, "JAvailable": true, "FAvailable": false,
		"YMileageCost": "50000", "WMileageCost": "0", "JMileageCost": "75000", "FMileageCost": "0",
		"YRemainingSeats": 9, "WRemainingSeats": 0, "JRemainingSeats": 2, "FRemainingSeats": 0,
		"YAirlines": "UA", "WAirlines": "", "JAirlines": "AC", "FAirlines": "",
		"YDirect": false, "WDirect": false, "JDirect": true, "FDirect": false,
		"Source": "united", "CreatedAt": "2026-01-01T00:00:00Z", "UpdatedAt": "2026-06-09T00:00:00Z",
		"AvailabilityTrips": [{"id": 1, "name": "test"}]
	}`

	var a Availability
	if err := json.Unmarshal([]byte(body), &a); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if a.AvailabilityTrips == nil {
		t.Error("expected AvailabilityTrips to be non-nil")
	}
	if !a.YAvailable {
		t.Error("expected YAvailable true")
	}
	if !a.JDirect {
		t.Error("expected JDirect true")
	}
}

func TestRouteJSON(t *testing.T) {
	body := `{"ID": "r1", "OriginAirport": "LAX", "OriginRegion": "North America", "DestinationAirport": "NRT", "DestinationRegion": "Asia", "NumDaysOut": 5, "Distance": 5450, "Source": "aeroplan"}`
	var r Route
	if err := json.Unmarshal([]byte(body), &r); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if r.ID != "r1" {
		t.Errorf("expected ID r1, got %s", r.ID)
	}
	if r.OriginAirport != "LAX" {
		t.Errorf("expected LAX, got %s", r.OriginAirport)
	}
	if r.DestinationRegion != "Asia" {
		t.Errorf("expected Asia, got %s", r.DestinationRegion)
	}
	if r.NumDaysOut != 5 {
		t.Errorf("expected 5 days out, got %d", r.NumDaysOut)
	}
	if r.Distance != 5450 {
		t.Errorf("expected 5450 distance, got %d", r.Distance)
	}
	if r.Source != "aeroplan" {
		t.Errorf("expected aeroplan, got %s", r.Source)
	}
}

func TestSegmentJSON(t *testing.T) {
	body := `{"ID": "seg1", "RouteID": "r1", "AvailabilityID": "a1", "AvailabilityTripID": "t1", "OriginAirport": "LAX", "DestinationAirport": "NRT", "FlightNumber": "UA123", "Distance": 5450, "FareClass": "J", "AircraftName": "787", "AircraftCode": "787", "DepartsAt": "2026-06-15T10:00:00Z", "ArrivesAt": "2026-06-16T13:00:00Z", "Order": 1}`
	var s Segment
	if err := json.Unmarshal([]byte(body), &s); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if s.FlightNumber != "UA123" {
		t.Errorf("expected UA123, got %s", s.FlightNumber)
	}
	if s.FareClass != "J" {
		t.Errorf("expected J, got %s", s.FareClass)
	}
	if s.AircraftName != "787" {
		t.Errorf("expected 787, got %s", s.AircraftName)
	}
	if s.AircraftCode != "787" {
		t.Errorf("expected 787, got %s", s.AircraftCode)
	}
	if s.DepartsAt != "2026-06-15T10:00:00Z" {
		t.Errorf("unexpected departs time: %s", s.DepartsAt)
	}
	if s.ArrivesAt != "2026-06-16T13:00:00Z" {
		t.Errorf("unexpected arrives time: %s", s.ArrivesAt)
	}
	if s.Order != 1 {
		t.Errorf("expected Order 1, got %d", s.Order)
	}
}

func TestSegmentOmitEmptyFields(t *testing.T) {
	body := `{"ID": "seg1", "RouteID": "r1", "AvailabilityID": "a1", "AvailabilityTripID": "t1", "OriginAirport": "LAX", "DestinationAirport": "NRT", "FlightNumber": "UA123", "Distance": 5450, "FareClass": "J", "DepartsAt": "2026-06-15T10:00:00Z", "ArrivesAt": "2026-06-16T13:00:00Z", "Order": 0}`
	var s Segment
	if err := json.Unmarshal([]byte(body), &s); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if s.AircraftName != "" {
		t.Errorf("expected empty AircraftName, got %s", s.AircraftName)
	}
	if s.AircraftCode != "" {
		t.Errorf("expected empty AircraftCode, got %s", s.AircraftCode)
	}
}

func TestLiveParamsJSONMarshal(t *testing.T) {
	p := LiveParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
		DepartureDate:      "2026-06-15",
		Source:             "aeroplan",
		SeatCount:          intPtr(2),
		DisableFilters:     boolPtr(true),
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("failed to remarshal: %v", err)
	}

	if out["origin_airport"] != "LAX" {
		t.Errorf("expected LAX, got %v", out["origin_airport"])
	}
	if out["destination_airport"] != "NRT" {
		t.Errorf("expected NRT, got %v", out["destination_airport"])
	}
	if out["departure_date"] != "2026-06-15" {
		t.Errorf("expected 2026-06-15, got %v", out["departure_date"])
	}
	if out["source"] != "aeroplan" {
		t.Errorf("expected aeroplan, got %v", out["source"])
	}
	if out["seat_count"] != float64(2) {
		t.Errorf("expected seat_count 2, got %v", out["seat_count"])
	}
	if out["disable_filters"] != true {
		t.Errorf("expected disable_filters true, got %v", out["disable_filters"])
	}
	if _, exists := out["show_dynamic_pricing"]; exists {
		t.Error("expected show_dynamic_pricing to be omitted when nil")
	}
}

func TestLiveParamsNilOmit(t *testing.T) {
	p := LiveParams{
		OriginAirport:      "LAX",
		DestinationAirport: "NRT",
		DepartureDate:      "2026-06-15",
		Source:             "aeroplan",
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("failed to remarshal: %v", err)
	}

	for _, key := range []string{"seat_count", "disable_filters", "show_dynamic_pricing"} {
		if _, exists := out[key]; exists {
			t.Errorf("expected %s to be omitted when nil", key)
		}
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if len(s) < len(substr) {
		t.Errorf("string %q does not contain %q", s, substr)
		return
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return
		}
	}
	t.Errorf("string %q does not contain %q", s, substr)
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if len(s) < len(substr) {
		return
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			t.Errorf("string %q should not contain %q", s, substr)
			return
		}
	}
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
