package postgres

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// DateTime supports multiple time formats, including ISO 8601, PostgreSQL, and MySQL timestamp formats
type DateTime struct {
	Time time.Time
}

// List of supported time formats, including ISO 8601, PostgreSQL, and MySQL timestamp formats
var timeFormats = []string{
	"2006-01-02T15:04:05.999999999Z07:00", // ISO 8601 with nanoseconds and timezone
	"2006-01-02T15:04:05.999999999Z",      // ISO 8601 with nanoseconds UTC
	"2006-01-02T15:04:05.999999999",       // ISO 8601 with nanoseconds (no timezone)
	"2006-01-02T15:04:05Z07:00",           // ISO 8601 with timezone
	"2006-01-02T15:04:05Z",                // ISO 8601 UTC
	"2006-01-02",                          // YYYY-MM-DD
	"2006-01-02 15:04:05",                 // YYYY-MM-DD HH:MM:SS
	"02-01-2006",                          // DD-MM-YYYY
	"02-01-2006 15:04:05",                 // DD-MM-YYYY HH:MM:SS
	"2006/01/02",                          // YYYY/MM/DD
	"02 Jan 2006",                         // DD Mon YYYY
	"02 Jan 2006 15:04:05",                // DD Mon YYYY HH:MM:SS
	"2006-01-02 15:04:05.999999",          // Timestamp with microseconds
	"2006-01-02 15:04:05.999999-07",       // Timestamp with microseconds and timezone
	"2006-01-02 15:04:05.999999Z07:00",    // Timestamp with timezone
	"2006-01-02T15:04:05.000000",          // ISO-like format with microseconds
}

// UnmarshalJSON parses multiple time formats, including ISO 8601, PostgreSQL, and MySQL formats
func (m *DateTime) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	for _, format := range timeFormats {
		t, err := time.Parse(format, str)
		if err == nil {
			m.Time = t
			return nil
		}
	}

	return errors.New("invalid time format: " + str)
}

// MarshalJSON ensures consistent output format (ISO 8601)
func (m DateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.Time.Format(time.RFC3339) + `"`), nil
}

// String implements fmt.Stringer
func (m DateTime) String() string {
	return m.Time.Format(time.RFC3339)
}

func (m *DateTime) Parse(value string) error {
	for _, format := range timeFormats {
		t, err := time.Parse(format, value)
		if err == nil {
			m.Time = t
			return nil
		}
	}
	return fmt.Errorf("invalid time format: %s", value)
}
