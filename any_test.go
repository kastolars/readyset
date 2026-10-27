package readyset

import "testing"

// A Set[any] can hold elements of mixed kinds, which print in a stable order.
func TestStringMixedKinds(t *testing.T) {
	got := New[any](2, "a", 1.5, nil, true).String()
	if want := "{<nil>, true, 2, 1.5, a}"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
