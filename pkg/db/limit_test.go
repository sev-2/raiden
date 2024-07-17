package db

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Model(model).Limit(15)

	assert.Equal(t, 15, q.LimitValue)
}

func TestOffset(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Model(model).Offset(30)

	assert.Equal(t, 30, q.OffsetValue)
}

func TestLimitOffset(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).Model(model).Limit(15).Offset(30)

	assert.Equal(t, 15, q.LimitValue)
	assert.Equal(t, 30, q.OffsetValue)
}
