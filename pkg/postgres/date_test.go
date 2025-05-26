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
	assert.NoError(t, err)
	assert.True(t, d.Valid)
	assert.Equal(t, "2025-05-22", d.String())
}

func TestDate_ScanFromString_Err(t *testing.T) {
	var d postgres.Date
	err := d.Scan("2025-05-22-01")
	assert.Error(t, err)
	assert.False(t, d.Valid)
}

func TestDate_ScanFromBytes(t *testing.T) {
	var d postgres.Date
	err := d.Scan([]byte("2025-05-22"))
	assert.NoError(t, err)
	assert.True(t, d.Valid)
	assert.Equal(t, "2025-05-22", d.String())
}

func TestDate_ScanFromTime(t *testing.T) {
	expected := time.Date(2025, 5, 22, 12, 0, 0, 0, time.UTC)
	var d postgres.Date
	err := d.Scan(expected)
	assert.NoError(t, err)
	assert.True(t, d.Valid)
	assert.True(t, d.Time.Equal(expected.Truncate(24*time.Hour)))
}

func TestDate_ScanNil(t *testing.T) {
	var d postgres.Date
	err := d.Scan(nil)
	assert.NoError(t, err)
	assert.False(t, d.Valid)
	assert.Equal(t, "null", d.String())
}

func TestDate_ScanInvalidType(t *testing.T) {
	var d postgres.Date
	err := d.Scan(123)
	assert.Error(t, err)
	assert.False(t, d.Valid)
}

func TestDateValue(t *testing.T) {
	d := postgres.Date{
		Time:  time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
	val, err := d.Value()
	assert.NoError(t, err)
	assert.Equal(t, "2025-05-22", val)
}

func TestDateValue_Null(t *testing.T) {
	d := postgres.Date{Valid: false}
	val, err := d.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestDate_JSONMarshaling(t *testing.T) {
	d := postgres.Date{
		Time:  time.Date(2025, 5, 22, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
	jsonBytes, err := json.Marshal(d)
	assert.NoError(t, err)
	assert.Equal(t, `"2025-05-22"`, string(jsonBytes))
}

func TestDate_JSONMarshaling_Null(t *testing.T) {
	d := postgres.Date{Valid: false}
	jsonBytes, err := json.Marshal(d)
	assert.NoError(t, err)
	assert.Equal(t, "null", string(jsonBytes))
}

func TestDate_JSONUnmarshaling(t *testing.T) {
	var d postgres.Date
	err := json.Unmarshal([]byte(`"2025-05-22"`), &d)
	assert.NoError(t, err)
	assert.True(t, d.Valid)
	assert.Equal(t, "2025-05-22", d.String())
}

func TestDate_JSONUnmarshaling_Null(t *testing.T) {
	var d postgres.Date
	err := json.Unmarshal([]byte(`null`), &d)
	assert.NoError(t, err)
	assert.False(t, d.Valid)
	assert.Equal(t, "null", d.String())
}
