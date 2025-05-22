package postgres

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type Date struct {
	time.Time
}

// Ensure Date implements sql.Scanner and driver.Valuer interfaces
var _ driver.Valuer = (*Date)(nil)
var _ fmt.Stringer = (*Date)(nil)

const dateFormat = "2006-01-02"

func (d *Date) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		d.Time = v.UTC().Truncate(24 * time.Hour)
		return nil
	case string:
		t, err := time.Parse(dateFormat, v)
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	case []byte:
		t, err := time.Parse(dateFormat, string(v))
		if err != nil {
			return err
		}
		d.Time = t
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Date", value)
	}
}

func (d Date) Value() (driver.Value, error) {
	return d.Format(dateFormat), nil
}

func (d Date) String() string {
	return d.Format(dateFormat)
}

// JSON Marshal/Unmarshal support
func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Format(dateFormat))), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}
