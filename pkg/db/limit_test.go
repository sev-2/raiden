package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Limit(15)

	assert.Equal(t, 15, q.LimitValue)
}

func TestOffset(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Offset(30)

	assert.Equal(t, 30, q.OffsetValue)
}

func TestLimitOffset(t *testing.T) {
	q := NewQuery(&mockRaidenContext).Model(articleMockModel).Limit(15).Offset(30)

	assert.Equal(t, 15, q.LimitValue)
	assert.Equal(t, 30, q.OffsetValue)
}
