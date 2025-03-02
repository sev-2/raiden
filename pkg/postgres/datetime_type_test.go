package postgres_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/postgres"
)

func TestDateTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"\"2025-03-02T15:04:05Z\"", "2025-03-02T15:04:05Z", true},
		{"\"2025-03-02 15:04:05\"", "2025-03-02T15:04:05Z", true},
		{"\"02-03-2025\"", "2025-03-02T00:00:00Z", true},
		{"\"02-03-2025 15:04:05\"", "2025-03-02T15:04:05Z", true},
		{"\"2025/03/02\"", "2025-03-02T00:00:00Z", true},
		{"\"02 Mar 2025\"", "2025-03-02T00:00:00Z", true},
		{"\"02 Mar 2025 15:04:05\"", "2025-03-02T15:04:05Z", true},
		{"\"invalid-date\"", "", false},
	}

	for _, test := range tests {
		var dt postgres.DateTime
		err := json.Unmarshal([]byte(test.input), &dt)
		if test.valid {
			if err != nil {
				t.Errorf("failed to parse valid time %s: %v", test.input, err)
			} else if dt.Time.Format(time.RFC3339) != test.expected {
				t.Errorf("expected %s, got %s", test.expected, dt.Time.Format(time.RFC3339))
			}
		} else {
			if err == nil {
				t.Errorf("expected error for invalid input %s, got none", test.input)
			}
		}
	}
}

func TestDateTime_MarshalJSON(t *testing.T) {
	dt := postgres.DateTime{Time: time.Date(2025, 3, 2, 15, 4, 5, 0, time.UTC)}
	expected := "\"2025-03-02T15:04:05Z\""
	b, err := json.Marshal(dt)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	if string(b) != expected {
		t.Errorf("expected %s, got %s", expected, string(b))
	}
}

func TestDateTime_String(t *testing.T) {
	dt := postgres.DateTime{Time: time.Date(2025, 3, 2, 15, 4, 5, 0, time.UTC)}
	expected := "2025-03-02T15:04:05Z"
	if dt.String() != expected {
		t.Errorf("expected %s, got %s", expected, dt.String())
	}
}
