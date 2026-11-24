package mcp

import (
	"encoding/json"
	"fmt"
	"math"
)

// Boundary: reading what a tool was called with, and describing what it takes --
// small pure helpers over decoded JSON. Nothing here knows a tool by name.

// ArgString reads a string argument, with a default.
func ArgString(args map[string]any, key, def string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return def
}

// ArgInt reads a number argument, with a default and a clamp to [min, max].
func ArgInt(args map[string]any, key string, def, min, max int) int {
	v, ok := args[key].(float64)
	if !ok {
		return def
	}
	// A JSON number is a float64, and int(v) on one that is not finite or is
	// outside int64 is undefined in Go: it wraps to a value inside the range
	// that the clamp below would then let through.
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return def
	}
	if v <= float64(min) {
		return min
	}
	if v >= float64(max) {
		return max
	}
	return int(v)
}

// ArgBool reads a boolean argument.
func ArgBool(args map[string]any, key string) bool {
	v, _ := args[key].(bool)
	return v
}

// JSON renders v as indented JSON, the form the tools answer in.
func JSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ObjectSchema builds an object schema from properties and required names.
func ObjectSchema(props map[string]any, required []string) map[string]any {
	s := map[string]any{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

// Prop is one property of a tool schema.
func Prop(typ, desc string) map[string]any {
	return map[string]any{"type": typ, "description": desc}
}

var _ = fmt.Sprintf
