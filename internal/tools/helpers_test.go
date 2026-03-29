package tools

import (
	"strings"
	"testing"
)

func TestParseStringArray(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", nil},
		{"  ,  , ", nil},
	}
	for _, tc := range cases {
		got := parseStringArray(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("parseStringArray(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("parseStringArray(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestParseDate(t *testing.T) {
	if parseDate("") != nil {
		t.Error("expected nil for empty string")
	}
	d := parseDate("2025-12-31T00:00:00Z")
	if d == nil {
		t.Error("expected non-nil for valid date")
	}
	if d.Year() != 2025 {
		t.Errorf("unexpected year: %d", d.Year())
	}
}

func TestMustJSON(t *testing.T) {
	out := mustJSON(map[string]string{"key": "value"})
	if !strings.Contains(out, `"key"`) || !strings.Contains(out, `"value"`) {
		t.Errorf("unexpected JSON output: %s", out)
	}
}

func TestClamp(t *testing.T) {
	if clamp(5, 1, 10) != 5 {
		t.Error("expected 5")
	}
	if clamp(-1, 1, 10) != 1 {
		t.Error("expected min")
	}
	if clamp(100, 1, 10) != 10 {
		t.Error("expected max")
	}
}
