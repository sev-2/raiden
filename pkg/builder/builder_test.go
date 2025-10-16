package builder_test

import (
	"testing"

	builder "github.com/sev-2/raiden/pkg/builder"
	"github.com/stretchr/testify/require"
)

type sampleModel struct {
	Field string `json:"field"`
}

func TestColOfReturnsColumnName(t *testing.T) {
	m := sampleModel{}
	col := builder.ColOf(&m, &m.Field)
	require.Equal(t, "field", col)
}

func TestColOfPanicsWhenModelNotStructPointer(t *testing.T) {
	field := "value"
	expected := "builder: model must resolve to struct (got *string)"
	require.PanicsWithValue(t, expected, func() {
		builder.ColOf(&field, &field)
	})
}

func TestColOfPanicsWhenFieldPointerNil(t *testing.T) {
	m := sampleModel{}
	require.PanicsWithValue(t, "builder: fieldPtr must be a pointer to a struct field (e.g., &m.Title)", func() {
		builder.ColOf(&m, (*string)(nil))
	})
}

func TestColOfWithStructValue(t *testing.T) {
	m := sampleModel{}
	col := builder.ColOf(m, &m.Field)
	require.Equal(t, "field", col)
}
