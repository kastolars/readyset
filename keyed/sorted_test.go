//go:build go1.21

package keyed

import (
	"reflect"
	"testing"
)

func TestSortedOrdersByKey(t *testing.T) {
	s := newPeople(person{3, "c"}, person{1, "a"}, person{2, "b"})

	var got []string
	for _, p := range Sorted(s) {
		got = append(got, p.Name)
	}
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// The order follows the key, not the element: here the names sort one way and
// the keys the other.
func TestSortedIgnoresElementOrder(t *testing.T) {
	byName := New(func(p person) string { return p.Name }, person{1, "z"}, person{2, "a"})

	got := Sorted(byName)
	if len(got) != 2 || got[0].Name != "a" || got[1].Name != "z" {
		t.Errorf("got %v, want elements ordered a, z", got)
	}
}

func TestSortedEmpty(t *testing.T) {
	var zero Set[int, person]
	for name, got := range map[string][]person{"zero value": Sorted(zero), "empty": Sorted(newPeople())} {
		if got == nil || len(got) != 0 {
			t.Errorf("%s: got %#v, want an empty non-nil slice", name, got)
		}
	}
}
