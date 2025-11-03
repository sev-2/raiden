package utils

import (
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

type customZero struct{ v string }

func (c customZero) IsZero() bool { return c.v == "" }

func TestEmptyOrDefault(t *testing.T) {
	require.Equal(t, "fallback", EmptyOrDefault("", "fallback"))
	require.Equal(t, "value", EmptyOrDefault("value", "fallback"))

	var ptr *string
	def := "ptr"
	require.Equal(t, &def, EmptyOrDefault(ptr, &def))

	strict := "strict"
	checker := func(v *string) bool { return v != nil && *v == "strict" }
	require.Equal(t, &def, EmptyOrDefault(&strict, &def, checker))
}

func TestEmptyOrError(t *testing.T) {
	require.NoError(t, EmptyOrError("value", "must not be empty"))
	err := EmptyOrError("", "must not be empty")
	require.Error(t, err)

	val := "trigger"
	checker := func(v *string) bool { return v != nil && *v == "trigger" }
	require.Error(t, EmptyOrError(&val, "blocked", checker))
}

func TestIsEmptyGenericScenarios(t *testing.T) {
	require.True(t, isEmptyGeneric(""))
	require.False(t, isEmptyGeneric("text"))

	require.True(t, isEmptyGeneric([]int{}))
	require.False(t, isEmptyGeneric([]int{1}))

	var ptr *int
	require.True(t, isEmptyGeneric(ptr))

	now := time.Now()
	require.True(t, isEmptyGeneric(time.Time{}))
	require.False(t, isEmptyGeneric(now))

	require.True(t, isEmptyGeneric(customZero{}))
	require.False(t, isEmptyGeneric(customZero{v: "non-empty"}))
}

func TestUnwrapAndTryIsZero(t *testing.T) {
	original := customZero{}

	iz, ok := tryIsZero(reflect.ValueOf(original))
	require.True(t, ok)
	require.True(t, iz)

	nonZero := customZero{v: "x"}
	nz, ok := tryIsZero(reflect.ValueOf(nonZero))
	require.True(t, ok)
	require.False(t, nz)

	nested := &original
	uv, wasNil := unwrap(reflect.ValueOf(nested))
	require.False(t, wasNil)
	require.True(t, uv.IsValid())
}

func TestUnwrapNil(t *testing.T) {
	var ptr *string
	uv, wasNil := unwrap(reflect.ValueOf(ptr))
	require.True(t, wasNil)
	require.False(t, uv.IsValid())
}

func TestHttpConnErrorMapping(t *testing.T) {
	tests := []struct {
		err      error
		expected string
		known    bool
	}{
		{fasthttp.ErrTimeout, "timeout", true},
		{fasthttp.ErrNoFreeConns, "conn_limit", true},
		{fasthttp.ErrConnectionClosed, "conn_close", true},
		{&net.OpError{Op: "dial", Err: errors.New("boom"), Net: "tcp"}, "timeout", true},
		{errors.New("other"), "", false},
	}

	for _, tt := range tests {
		name, known := httpConnError(tt.err)
		require.Equal(t, tt.known, known)
		require.Equal(t, tt.expected, name)
	}
}
