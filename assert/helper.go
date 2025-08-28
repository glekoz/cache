package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expected T) { // reduces boilerplate
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v, want: %v", actual, expected)
	}
}
