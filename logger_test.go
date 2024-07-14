package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestPanic(t *testing.T) {
	assert.Panics(t, func() {
		raiden.Panic("Panic message")
	})
}
