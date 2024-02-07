package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestMarshallColumnTag(t *testing.T) {
	tag := "name:name;type:varchar(10);nullable"

	column := raiden.MarshalColumnTag(tag)

	assert.Equal(t, "name", column.Name)
	assert.Equal(t, "varchar(10)", column.Type)
	assert.Equal(t, true, column.Nullable)
}
