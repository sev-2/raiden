package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Date struct {
	Time  time.Time
	Valid bool // NEW: true if not NULL
}

var _ driver.Valuer = (*Date)(nil)
var _ fmt.Stringer = (*Date)(nil)

const dateFormat = "2006-01-02"

// Scan implements the sql.Scanner interface
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.Time, d.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		d.Time = v.UTC().Truncate(24 * time.Hour)
	case string:
		t, err := time.Parse(dateFormat, v)
		if err != nil {
			return err
		}
		d.Time = t
	case []byte:
		t, err := time.Parse(dateFormat, string(v))
		if err != nil {
			return err
		}
		d.Time = t
	default:
		return fmt.Errorf("cannot scan type %T into Date", value)
	}

	d.Valid = true
	return nil
}

// Value implements the driver.Valuer interface
func (d Date) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Time.Format(dateFormat), nil
}

// String implements fmt.Stringer
func (d Date) String() string {
	if !d.Valid {
		return "null"
	}
	return d.Time.Format(dateFormat)
}

// MarshalJSON handles null and valid dates
func (d Date) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(d.Time.Format(dateFormat))
}

// UnmarshalJSON supports null and valid date strings
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		d.Valid = false
		return nil
	}
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return err
	}
	d.Time = t
	d.Valid = true
	return nil
}
