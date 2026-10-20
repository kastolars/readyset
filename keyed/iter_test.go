//go:build go1.23

package keyed

import (
	"reflect"
	"slices"
	"sort"
	"testing"
)

func TestAll(t *testing.T) {
	s := newPeople(person{2, "b"}, person{1, "a"})
	got := []string{}
	for _, p := range slices.Collect(s.All()) {
		got = append(got, p.Name)
	}
	sort.Strings(got)
	if want := []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("All() = %v, want %v", got, want)
	}

	var zero Set[int, person]
	if n := len(slices.Collect(zero.All())); n != 0 {
		t.Errorf("zero value All() yielded %d elements", n)
	}
}

func TestAllStopsEarly(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
	calls := 0
	s.All()(func(person) bool {
		calls++
		return false
	})
	if calls != 1 {
		t.Errorf("yield called %d times after returning false, want 1", calls)
	}
}
