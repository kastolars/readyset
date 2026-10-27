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

func TestCollect(t *testing.T) {
	key := func(p person) int { return p.ID }
	got := Collect(key, slices.Values([]person{{1, "old"}, {2, "b"}, {1, "new"}}))

	if want := []int{1, 2}; !reflect.DeepEqual(ids(got), want) {
		t.Errorf("Collect ids = %v, want %v", ids(got), want)
	}
	if p, _ := got.Get(1); p.Name != "new" {
		t.Errorf("Get(1) = %q, want the last element with that key", p.Name)
	}
	if !Collect(key, got.All()).Equal(got) {
		t.Error("Collect(key, s.All()) != s")
	}
	mustPanic(t, "nil key function", func() { Collect[int, person](nil, slices.Values([]person{})) })
}
