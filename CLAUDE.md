# CLAUDE.md

Guidance for coding agents working with this repository.

## Build

```sh
go build -o bin/seats-mcp .
```

Requires Go 1.26+.

## Test

There is no test suite yet. When adding tests:

```sh
go test ./...
```

Use the standard `testing` package. Mock HTTP calls with `httptest.NewServer`. Tests live alongside source as `*_test.go`.

## Architecture

```
main.go                        Entry point — reads env, wires up HTTP client → seats.aero client → MCP server
internal/
  seatsaero/                   Typed API client for seats.aero Partner API
    client.go                  HTTP client core: doJSON() handles auth, errors, JSON encoding
    types.go                   All data types: Availability, Trip, Segment, SearchResponse, etc.
    search.go                  GET /search — cached multi-program search
    availability.go            GET /availability — single-program bulk fetch
    trips.go                   GET /trips/{id} — flight-level detail from an Availability ID
    live.go                    POST /live — real-time search
    errors.go                  APIError type for non-2xx responses
  mcptools/                    MCP tool layer — wraps seats.aero client methods as MCP tools
    register.go                Register() wires all tools; jsonResult, apiErrResult, arg helpers
    search.go                  "search_award_availability" tool
    availability.go            "bulk_availability" tool
    trips.go                   "get_trips" tool
    live.go                    "live_search" tool
```

**Flow:** `main.go` → `mcptools.Register(s, client)` → each tool handler calls `seatsaero.Client` methods → `doJSON()` → seats.aero Partner API.

## Conventions

- **Optional parameters** use `*bool` or `*int` in the seatsaero param structs. The MCP tool layer reads them from `map[string]any` args via `boolPtrArg()` / `intPtrArg()` helpers.
- **Error handling:** The seatsaero layer returns Go errors (including `*APIError` for non-2xx). The mcptools layer converts the Go error to an MCP tool error result via `apiErrResult()` — the second return (Go error) is reserved for transport failures and is always `nil` after an API error.
- **Required MCP params** use `req.RequireString()`. Optional string params use `req.GetString(key, default)`. Optional bool/int use `boolPtrArg()` / `intPtrArg()`.
- **JSON serialization** in MCP results goes through `jsonResult()` which marshals with indentation.
- **Context** is threaded through `doJSON` via `http.NewRequestWithContext`. Don't drop it.
- **The server logs only to stderr.** stdout is the MCP transport (stdio). Use `fmt.Fprintln(os.Stderr, ...)` for logging.

## API key constraints

- Live search (`/live`) is **not available to Pro users** — requires a commercial agreement.
- 1,000 requests/day shared across all uses of a key.
- Pro keys are non-commercial use only.

## seats.aero API quick reference

- **Auth:** `Partner-Authorization` header with the API key.
- **Base URL:** `https://seats.aero/partnerapi`
- **Cabin codes:** `Y` = economy, `W` = premium economy, `J` = business, `F` = first.
- **Pagination:** prefer `cursor` over `skip` to avoid duplicates.
- **Live search IDs** are ephemeral — don't pass to `/trips/{id}`.
- **Times** are in local airport timezone, not UTC.

Full API docs: https://developers.seats.aero/reference/getting-started-p
