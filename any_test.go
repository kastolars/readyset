//go:build go1.20

package readyset

import "testing"

// Before Go 1.20, an interface type such as any did not satisfy comparable, so
// a Set[any] cannot be built with older toolchains.
func TestStringMixedKinds(t *testing.T) {
	got := New[any](2, "a", 1.5, nil, true).String()
	if want := "{<nil>, true, 2, 1.5, a}"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
