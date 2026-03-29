package tools

import (
	"encoding/json"
	"strings"
	"time"
)

// parseStringArray splits a comma-separated string into a string slice.
// Returns nil if the input is empty.
func parseStringArray(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parseDate parses an ISO date string. Returns nil on empty/invalid input.
func parseDate(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t
		}
	}
	return nil
}

// mustJSON serialises v to JSON, returning an error string on failure.
func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"ok":false,"error":{"code":"INTERNAL","message":"json marshal failed"}}`
	}
	return string(b)
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T { return &v }

// clamp returns n clamped between lo and hi.
func clamp(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}
