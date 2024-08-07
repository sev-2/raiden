package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	_, err := NewQuery(&mockRaidenContext).Model(articleMockModel).Delete()

	assert.NoError(t, err)
}
