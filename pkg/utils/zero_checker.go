package utils

import (
	"errors"
	"reflect"
	"time"
)

// Optional: callers can override what "empty" means for T.
type ZeroChecker[T any] func(T) bool

// EmptyOrDefault returns def when s is considered "empty".
// Works whether T is a value or a pointer type.
func EmptyOrDefault[T any](s T, def T, check ...ZeroChecker[T]) T {
	// Custom checker on T-as-provided (pointer or value).
	if len(check) > 0 && check[0](s) {
		return def
	}
	if isEmptyGeneric(s) {
		return def
	}
	return s
}

// EmptyOrError returns an error when s is considered "empty".
// Works whether T is a value or a pointer type.
func EmptyOrError[T any](s T, message string, check ...ZeroChecker[T]) error {
	if len(check) > 0 && check[0](s) {
		return errors.New(message)
	}
	if isEmptyGeneric(s) {
		return errors.New(message)
	}
	return nil
}

// ---- helpers ----

type isZeroer interface{ IsZero() bool }

// isEmptyGeneric unwraps pointers/interfaces (recursively) and
// then applies sensible emptiness rules:
// - nil pointers / nil interfaces -> empty
// - string: ""
// - slices/maps/chans: len==0
// - arrays: len==0
// - funcs: nil
// - time.Time: IsZero()
// - any type with IsZero() bool (value or pointer receiver)
// - fallback: reflect.Value.IsZero()
func isEmptyGeneric[T any](v T) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}

	// Try IsZero on the original value (covers value receiver cases on T).
	if iz, ok := tryIsZero(rv); ok {
		return iz
	}

	// Unwrap pointers/interfaces; nil at any level => empty.
	uv, wasNil := unwrap(rv)
	if wasNil {
		return true
	}

	// Try IsZero again on unwrapped value (and its address).
	if iz, ok := tryIsZero(uv); ok {
		return iz
	}

	// Special-case time.Time
	if uv.Type() == reflect.TypeOf(time.Time{}) {
		return uv.Interface().(time.Time).IsZero()
	}

	switch uv.Kind() {
	case reflect.String:
		return uv.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Chan, reflect.Array:
		return uv.Len() == 0
	case reflect.Func:
		return uv.IsNil()
	case reflect.Interface:
		return uv.IsNil()
	}

	// Numbers, bools, structs, etc.
	return uv.IsZero()
}

// unwrap recursively dereferences pointers/interfaces.
// Returns (zero, true) if a nil is encountered.
func unwrap(rv reflect.Value) (reflect.Value, bool) {
	for {
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface:
			if rv.IsNil() {
				return reflect.Value{}, true
			}
			rv = rv.Elem()
		default:
			return rv, false
		}
	}
}

// tryIsZero calls IsZero() if the value or its address implements it.
func tryIsZero(rv reflect.Value) (bool, bool) {
	if !rv.IsValid() {
		return true, true
	}
	if iz, ok := rv.Interface().(isZeroer); ok {
		return iz.IsZero(), true
	}
	if rv.CanAddr() {
		if iz, ok := rv.Addr().Interface().(isZeroer); ok {
			return iz.IsZero(), true
		}
	}
	return false, false
}
