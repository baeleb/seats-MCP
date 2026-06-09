# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`seats-MCP` is a Model Context Protocol (MCP) server that wraps the [seats.aero Partner API](https://developers.seats.aero/reference/getting-started-p). It exposes award flight availability search as MCP tools so AI agents can query redemption availability across major airline mileage programs.

## seats.aero API

### Authentication

All requests require a `Partner-Authorization` header with a personal API key (Pro users) or a commercial key from seats.aero.

```
Partner-Authorization: <api-key>
```

For OAuth flows (user-facing apps), the token is prefixed: `Bearer seats:ota:...`

**Rate limit:** 1,000 requests/day shared across all uses of the key.

### Base URL

```
https://seats.aero/partnerapi
```

All requests are JSON. GET params go in the query string; POST bodies use `application/json`.

### Endpoints

#### `GET /search` — Cached Search
Multi-program search across airports and date ranges (equivalent to the website's Search tab).

Key params:
- `origin_airport` / `destination_airport` — comma-delimited IATA codes (e.g. `SFO,LAX`)
- `start_date` / `end_date` — `YYYY-MM-DD`
- `sources` — comma-delimited mileage programs (e.g. `aeroplan,united`)
- `cabins` — comma-delimited cabin names (`economy`, `premium`, `business`, `first`)
- `only_direct_flights` — boolean
- `order_by` — `lowest_mileage` or default (date + available cabins)
- `include_trips` — embed trip-level detail in response (slower)
- `take` / `cursor` / `skip` — pagination (10–1000, default 500)

Returns: array of `Availability` objects + `hasMore` + `cursor`.

#### `GET /availability` — Bulk Availability
Single-program query for large result sets (equivalent to Explore tab). Requires `source`.

Key params:
- `source` — single mileage program (required)
- `cabin` — `economy | premium | business | first`
- `origin_region` / `destination_region` — `North America | South America | Africa | Asia | Europe | Oceania`
- `start_date` / `end_date`, `take` / `cursor` / `skip`

#### `GET /trips/{id}` — Get Trips
Returns detailed flight itineraries for an `Availability` object ID. The `id` comes from an Availability object's `ID` field.

Returns: `data[]` (Trip objects), `origin_coordinates`, `destination_coordinates`, `booking_links[]`.

**Note:** IDs from Live Search results are synthetic and will not work with this endpoint.

#### `POST /live` — Live Search
Real-time search for a specific route, date, and mileage program. Slower (5–15s). Use exponential backoff on retries. Failed searches don't count against the daily quota.

Required body params: `origin_airport`, `destination_airport`, `departure_date`, `source`

Optional: `seat_count` (1–9), `disable_filters`, `show_dynamic_pricing`

### Core Data Model

**Availability** — summary record for one route + date + mileage program. Fields follow a per-cabin pattern:
- `YAvailable` / `WAvailable` / `JAvailable` / `FAvailable` — boolean
- `YMileageCost` / `WMileageCost` / … — mileage price (string)
- `YRemainingSeats` / … — seat count (not all programs provide this)
- `YDirect` / … — direct-only flag
- `YAirlines` / … — operating carriers

Cabin codes: `Y` = Economy, `W` = Premium Economy, `J` = Business, `F` = First.

**Trip** — specific itinerary derived from an Availability. Contains `AvailabilitySegments[]`, `MileageCost`, `TotalTaxes`, `Cabin`, `Stops`, `FlightNumbers`, `DepartsAt`, `ArrivesAt`.

**Segment** — individual flight leg within a trip. Contains flight number, airports, times, aircraft, fare class, distance.

### Mileage Program Sources

Programs are identified by short string keys (e.g. `aeroplan`, `united`, `delta`, `emirates`). Not all programs support all cabins, seat counts, or trip-level data. Treat capability gaps as expected, not errors.

### Pagination

Use `cursor` (returned in each response) for stable pagination. `skip` is an alternative but `cursor` is preferred to avoid duplicates. Deduplicate results by `ID` field when paginating.

### Important Constraints

- Times are in **local airport timezone**, not UTC.
- Live search IDs are ephemeral — don't pass them to `/trips/{id}`.
- `include_trips=true` on `/search` is convenient but slower; prefer fetching trips separately for large result sets.
- Pro-tier keys are for non-commercial use only.
