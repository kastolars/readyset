package readyset

import (
	"math"
	"reflect"
	"testing"
)

func TestSorted(t *testing.T) {
	if got, want := Sorted(New(10, 2, -3, 1)), []int{-3, 1, 2, 10}; !reflect.DeepEqual(got, want) {
		t.Errorf("ints: got %v, want %v", got, want)
	}
	if got, want := Sorted(New("b", "a", "B")), []string{"B", "a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("strings: got %v, want %v", got, want)
	}

	type id int // a defined type with an ordered underlying type
	if got, want := Sorted(New(id(3), id(1))), []id{1, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("defined type: got %v, want %v", got, want)
	}
}

func TestSortedEmpty(t *testing.T) {
	var zero Set[int]
	for name, got := range map[string][]int{"zero value": Sorted(zero), "empty": Sorted(New[int]())} {
		if got == nil || len(got) != 0 {
			t.Errorf("%s: got %#v, want an empty non-nil slice", name, got)
		}
	}
}

// A NaN is not equal to itself, so it never gets into a set.
func TestSortedIgnoresNaN(t *testing.T) {
	got := Sorted(New(2.5, math.NaN(), -1))
	if want := []float64{-1, 2.5}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSortedLeavesSetAlone(t *testing.T) {
	s := New(3, 1, 2)
	got := Sorted(s)
	got[0] = 99 // the result is a copy

	if s.Len() != 3 || !s.Contains(1) || s.Contains(99) {
		t.Errorf("Sorted shared storage with the set: %v", s)
	}
}
