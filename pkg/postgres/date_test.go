package postgres_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDate_ScanFromString(t *testing.T) {
	var d postgres.Date
	err := d.Scan("2025-05-22")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Format("2006-01-02") != "2025-05-22" {
		t.Errorf("expected 2025-05-22, got %s", d.String())
	}
}

func TestDate_ScanFromString_Err(t *testing.T) {
	var d postgres.Date
	err := d.Scan("2025-05-22-01")
	assert.Error(t, err)
}

func TestDate_ScanFromBytes(t *testing.T) {
	var d postgres.Date
	err := d.Scan([]byte("2025-05-22"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Format("2006-01-02") != "2025-05-22" {
		t.Errorf("expected 2025-05-22, got %s", d.String())
	}
}

func TestDate_ScanFromTime(t *testing.T) {
	expected := time.Date(2025, 5, 22, 12, 0, 0, 0, time.UTC)
	var d postgres.Date
	err := d.Scan(expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.Time.Equal(expected.Truncate(24 * time.Hour)) {
		t.Errorf("expected %v, got %v", expected, d.Time)
	}
}

func TestDateValue(t *testing.T) {
	d := postgres.Date{Time: time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC)}
	val, err := d.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if str, ok := val.(string); !ok || str != "2025-05-22" {
		t.Errorf("expected string 2025-05-22, got %v", val)
	}
}

func TestDate_JSONMarshaling(t *testing.T) {
	d := postgres.Date{Time: time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC)}
	jsonBytes, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `"2025-05-22"`
	if string(jsonBytes) != expected {
		t.Errorf("expected %s, got %s", expected, string(jsonBytes))
	}
}

func TestDate_JSONUnmarshaling(t *testing.T) {
	var d postgres.Date
	err := json.Unmarshal([]byte(`"2025-05-22"`), &d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.String() != "2025-05-22" {
		t.Errorf("expected 2025-05-22, got %s", d.String())
	}
}

func TestDate_ScanInvalidType(t *testing.T) {
	var d postgres.Date
	err := d.Scan(123)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
