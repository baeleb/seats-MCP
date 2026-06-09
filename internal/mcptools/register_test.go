package mcptools

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/calebhro/seats-MCP/internal/seatsaero"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestBoolPtrArg_PresentTrue(t *testing.T) {
	args := map[string]any{"my_flag": true}
	v := boolPtrArg(args, "my_flag")
	if v == nil {
		t.Fatal("expected non-nil")
	}
	if *v != true {
		t.Errorf("expected true, got %v", *v)
	}
}

func TestBoolPtrArg_PresentFalse(t *testing.T) {
	args := map[string]any{"my_flag": false}
	v := boolPtrArg(args, "my_flag")
	if v == nil {
		t.Fatal("expected non-nil")
	}
	if *v != false {
		t.Errorf("expected false, got %v", *v)
	}
}

func TestBoolPtrArg_NotPresent(t *testing.T) {
	args := map[string]any{}
	v := boolPtrArg(args, "my_flag")
	if v != nil {
		t.Errorf("expected nil, got %v", *v)
	}
}

func TestBoolPtrArg_WrongType(t *testing.T) {
	args := map[string]any{"my_flag": "true"}
	v := boolPtrArg(args, "my_flag")
	if v != nil {
		t.Errorf("expected nil for wrong type, got %v", *v)
	}
}

func TestBoolPtrArg_NilValue(t *testing.T) {
	args := map[string]any{"my_flag": nil}
	v := boolPtrArg(args, "my_flag")
	if v != nil {
		t.Errorf("expected nil, got %v", *v)
	}
}

func TestIntPtrArg_PresentFloat64(t *testing.T) {
	args := map[string]any{"take": float64(100)}
	v := intPtrArg(args, "take")
	if v == nil {
		t.Fatal("expected non-nil")
	}
	if *v != 100 {
		t.Errorf("expected 100, got %d", *v)
	}
}

func TestIntPtrArg_PresentInt(t *testing.T) {
	args := map[string]any{"take": 50}
	v := intPtrArg(args, "take")
	if v == nil {
		t.Fatal("expected non-nil")
	}
	if *v != 50 {
		t.Errorf("expected 50, got %d", *v)
	}
}

func TestIntPtrArg_PresentFloat64Fractional(t *testing.T) {
	args := map[string]any{"take": float64(100.99)}
	v := intPtrArg(args, "take")
	if v == nil {
		t.Fatal("expected non-nil")
	}
	if *v != 100 {
		t.Errorf("expected truncated 100, got %d", *v)
	}
}

func TestIntPtrArg_Zero(t *testing.T) {
	args := map[string]any{"take": float64(0)}
	v := intPtrArg(args, "take")
	if v == nil {
		t.Fatal("expected non-nil")
	}
	if *v != 0 {
		t.Errorf("expected 0, got %d", *v)
	}
}

func TestIntPtrArg_NotPresent(t *testing.T) {
	args := map[string]any{}
	v := intPtrArg(args, "take")
	if v != nil {
		t.Errorf("expected nil, got %v", *v)
	}
}

func TestIntPtrArg_WrongType(t *testing.T) {
	args := map[string]any{"take": "fifty"}
	v := intPtrArg(args, "take")
	if v != nil {
		t.Errorf("expected nil for wrong type, got %v", *v)
	}
}

func TestIntPtrArg_StringNumber(t *testing.T) {
	args := map[string]any{"take": "50"}
	v := intPtrArg(args, "take")
	if v != nil {
		t.Errorf("expected nil for string type, got %v", *v)
	}
}

func TestIntPtrArg_Bool(t *testing.T) {
	args := map[string]any{"take": true}
	v := intPtrArg(args, "take")
	if v != nil {
		t.Errorf("expected nil for bool type, got %v", *v)
	}
}

func TestJsonResult_Valid(t *testing.T) {
	type simple struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	val := simple{Name: "test", Age: 42}

	result, err := jsonResult(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	content := result.Content
	if len(content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(content))
	}

	tc, ok := content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", content[0])
	}

	var parsed simple
	if err := json.Unmarshal([]byte(tc.Text), &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if parsed.Name != "test" {
		t.Errorf("expected name test, got %s", parsed.Name)
	}
	if parsed.Age != 42 {
		t.Errorf("expected age 42, got %d", parsed.Age)
	}
}

func TestJsonResult_PrettyPrint(t *testing.T) {
	val := map[string]int{"a": 1, "b": 2}
	result, err := jsonResult(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	text := tc.Text
	if !strings.Contains(text, "\n") {
		t.Error("expected indented (pretty-printed) JSON")
	}
	if !strings.Contains(text, "  ") {
		t.Error("expected indentation spaces")
	}
}

func TestJsonResult_Array(t *testing.T) {
	val := []string{"one", "two", "three"}
	result, err := jsonResult(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	text := tc.Text
	if !strings.HasPrefix(text, "[") {
		t.Errorf("expected JSON array, got %s", text)
	}
}

func TestApiErrResult_SeatsAeroError(t *testing.T) {
	apiErr := &seatsaero.APIError{Status: 429, Body: "too many requests"}
	result := apiErrResult(apiErr)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected error result")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	text := tc.Text
	if !strings.Contains(text, "seats.aero 429") {
		t.Errorf("expected status in error text, got %s", text)
	}
	if !strings.Contains(text, "too many requests") {
		t.Errorf("expected body in error text, got %s", text)
	}
}

func TestApiErrResult_OtherError(t *testing.T) {
	generic := errors.New("connection refused")
	result := apiErrResult(generic)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected error result")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	text := tc.Text
	if !strings.Contains(text, "connection refused") {
		t.Errorf("expected error message, got %s", text)
	}
}

func TestApiErrResult_WrappedAPIError(t *testing.T) {
	apiErr := &seatsaero.APIError{Status: 404, Body: "not found"}
	wrapped := errors.New("additional context: " + apiErr.Error())

	// errors.Is uses unwrapping; errors.As is what apiErrResult uses.
	// This tests that a non-wrapped error containing the message still works.
	result := apiErrResult(wrapped)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected error result")
	}
}

func TestApiErrResult_NilError(t *testing.T) {
	result := apiErrResult(nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}
