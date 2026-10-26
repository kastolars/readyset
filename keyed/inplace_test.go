package keyed

import (
	"reflect"
	"testing"
	"testing/quick"
)

// For a key in both sets, AddSet behaves like Add: the argument's element wins.
func TestAddSetReplacesEqualKeys(t *testing.T) {
	s := newPeople(person{1, "old"}, person{2, "keep"})
	other := newPeople(person{1, "new"}, person{3, "added"})
	s.AddSet(other)

	if got, want := ids(s), []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("after AddSet ids = %v, want %v", got, want)
	}
	if p, _ := s.Get(1); p.Name != "new" {
		t.Errorf("Get(1) = %q, want the element from the argument", p.Name)
	}
	if other.Len() != 2 {
		t.Errorf("AddSet modified its argument: Len() = %d, want 2", other.Len())
	}
}

func TestRemoveSet(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
	other := newPeople(person{2, "different name"}, person{9, "z"})
	s.RemoveSet(other)

	if got, want := ids(s), []int{1, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("after RemoveSet ids = %v, want %v", got, want)
	}
	if other.Len() != 2 {
		t.Errorf("RemoveSet modified its argument: Len() = %d, want 2", other.Len())
	}
}

func TestInPlaceOnZeroValue(t *testing.T) {
	var s Set[int, person]
	s.RemoveSet(newPeople(person{1, "a"})) // must not panic
	s.AddSet(newPeople())                  // adds nothing, so must not panic

	mustPanic(t, "must be created with New", func() { s.AddSet(newPeople(person{1, "a"})) })
}

func TestInPlaceWithItself(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"})
	s.AddSet(s)
	if got, want := ids(s), []int{1, 2}; !reflect.DeepEqual(got, want) {
		t.Errorf("s.AddSet(s) ids = %v, want %v", got, want)
	}
	s.RemoveSet(s)
	if !s.IsEmpty() {
		t.Errorf("s.RemoveSet(s) left %d elements, want 0", s.Len())
	}
}

// AddSet and RemoveSet select the same keys as Union and Difference.
func TestInPlaceMatchesOperations(t *testing.T) {
	err := quick.Check(func(x, y []byte) bool {
		a, b := newBytes(x...), newBytes(y...)

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
