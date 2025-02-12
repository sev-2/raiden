package postgres_test

import (
	"encoding/json"
	"testing"

	"github.com/sev-2/raiden/pkg/postgres"
)

// TestUnmarshalPoint verifies if the POINT string from Supabase is correctly parsed into the Point struct.
func TestUnmarshalPoint(t *testing.T) {
	tests := []struct {
		input    string
		expected postgres.Point
		hasError bool
	}{
		{`"(144.9631,-37.8136)"`, postgres.Point{144.9631, -37.8136}, false},
		{`"(0,0)"`, postgres.Point{0, 0}, false},
		{`"(-122.4194,37.7749)"`, postgres.Point{-122.4194, 37.7749}, false},
		{`"(invalid, data)"`, postgres.Point{}, true}, // Invalid input
		{`"()"`, postgres.Point{}, true},              // Empty POINT
	}

	for _, test := range tests {
		var p postgres.Point
		err := json.Unmarshal([]byte(test.input), &p)

		if (err != nil) != test.hasError {
			t.Errorf("Unexpected error for input %s: %v", test.input, err)
		}

		if !test.hasError && (p.X != test.expected.X || p.Y != test.expected.Y) {
			t.Errorf("Expected %+v, got %+v", test.expected, p)
		}
	}
}

// TestMarshalPoint verifies if the Point struct is correctly converted into a POINT string.
func TestMarshalPoint(t *testing.T) {
	tests := []struct {
		input    postgres.Point
		expected string
	}{
		{postgres.Point{144.9631, -37.8136}, `"(144.963100,-37.813600)"`},
		{postgres.Point{0, 0}, `"(0.000000,0.000000)"`},
		{postgres.Point{-122.4194, 37.7749}, `"(-122.419400,37.774900)"`},
	}

	for _, test := range tests {
		data, err := json.Marshal(test.input)
		if err != nil {
			t.Errorf("Unexpected error for input %+v: %v", test.input, err)
		}

		if string(data) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(data))
		}
	}
}
