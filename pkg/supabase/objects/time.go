package objects

import (
	"fmt"
	"strings"
	"time"
)

func NewSupabaseTime(newTime time.Time) *SupabaseTime {
	return &SupabaseTime{
		Time: newTime,
	}
}

type SupabaseTime struct {
	time.Time
}

func (mt *SupabaseTime) UnmarshalJSON(b []byte) error {
	layout := "2006-01-02 00:00:00+00"
	dateString := string(b)
	dateString = strings.ReplaceAll(dateString, "\"", "")
	t, err := time.Parse(layout, dateString)
	if err != nil {
		return err
	}
	mt.Time = t
	return nil
}

func (mt SupabaseTime) MarshalJSON() ([]byte, error) {
	layout := "2006-01-02 00:00:00+00"
	return []byte(fmt.Sprintf(`"%s"`, mt.Time.Format(layout))), nil
}
