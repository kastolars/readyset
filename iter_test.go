package readyset

import (
	"reflect"
	"slices"
	"testing"
)

func TestAll(t *testing.T) {
	s := New(3, 1, 2)
	got := slices.Sorted(s.All())
	if want := []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("All() = %v, want %v", got, want)
	}

	var zero Set[int]
	if n := len(slices.Collect(zero.All())); n != 0 {
		t.Errorf("zero value All() yielded %d elements", n)
	}
}

func TestAllStopsEarly(t *testing.T) {
	s := New(1, 2, 3, 4, 5)
	calls := 0
	s.All()(func(int) bool {
		calls++
		return false
	})
	if calls != 1 {
		t.Errorf("yield called %d times after returning false, want 1", calls)
	}
}

func TestCollect(t *testing.T) {
	got := Collect(slices.Values([]int{3, 1, 3, 2}))
	if want := []int{1, 2, 3}; !reflect.DeepEqual(sorted(got), want) {
		t.Errorf("Collect = %v, want %v", sorted(got), want)
	}

	empty := Collect(slices.Values([]int(nil)))
	empty.Add(1) // usable even when nothing was collected
	if !empty.Contains(1) {
		t.Error("Collect of an empty sequence is not usable")
	}

	// Collect and All round-trip.
	s := New(4, 5, 6)
	if !Collect(s.All()).Equal(s) {
		t.Error("Collect(s.All()) != s")
	}
}
