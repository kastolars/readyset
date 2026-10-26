package readyset

import (
	"reflect"
	"testing"
	"testing/quick"
)

func TestAddSet(t *testing.T) {
	s := New(1, 2)
	other := New(2, 3)
	s.AddSet(other)

	if got, want := sorted(s), []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("after AddSet = %v, want %v", got, want)
	}
	if got, want := sorted(other), []int{2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("AddSet modified its argument: %v, want %v", got, want)
	}
}

func TestRemoveSet(t *testing.T) {
	s := New(1, 2, 3)
	other := New(2, 3, 9)
	s.RemoveSet(other)

	if got, want := sorted(s), []int{1}; !reflect.DeepEqual(got, want) {
		t.Errorf("after RemoveSet = %v, want %v", got, want)
	}
	if got, want := sorted(other), []int{2, 3, 9}; !reflect.DeepEqual(got, want) {
		t.Errorf("RemoveSet modified its argument: %v, want %v", got, want)
	}
}

func TestInPlaceOnZeroValue(t *testing.T) {
	var s Set[int]
	s.RemoveSet(New(1))  // must not panic
	s.AddSet(Set[int]{}) // must not panic
	if !s.IsEmpty() {
		t.Fatal("zero value changed by empty operations")
	}

	s.AddSet(New(1, 2))
	if got, want := sorted(s), []int{1, 2}; !reflect.DeepEqual(got, want) {
		t.Errorf("AddSet on zero value = %v, want %v", got, want)
	}
}

func TestInPlaceWithItself(t *testing.T) {
	s := New(1, 2, 3)
	s.AddSet(s)
	if got, want := sorted(s), []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("s.AddSet(s) = %v, want %v", got, want)
	}
	s.RemoveSet(s)
	if !s.IsEmpty() {
		t.Errorf("s.RemoveSet(s) left %v, want an empty set", sorted(s))
	}
}

// AddSet and RemoveSet are the in-place forms of Union and Difference.
func TestInPlaceMatchesOperations(t *testing.T) {
	err := quick.Check(func(x, y []byte) bool {
		a, b := New(x...), New(y...)

		added := a.Clone()
		added.AddSet(b)
		removed := a.Clone()
		removed.RemoveSet(b)

		return added.Equal(a.Union(b)) && removed.Equal(a.Difference(b))
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
