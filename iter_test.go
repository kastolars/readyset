//go:build go1.23

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
