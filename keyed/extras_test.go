package keyed

import (
	"reflect"
	"testing"
	"testing/quick"
)

func TestRetainSet(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
	other := newPeople(person{2, "different name"}, person{3, "c"}, person{9, "z"})
	s.RetainSet(other)

	if got, want := ids(s), []int{2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("after RetainSet ids = %v, want %v", got, want)
	}
	if p, _ := s.Get(2); p.Name != "b" {
		t.Errorf("Get(2) = %q, want the receiver's element", p.Name)
	}
	if other.Len() != 3 {
		t.Errorf("RetainSet modified its argument: Len() = %d, want 3", other.Len())
	}

	var zero Set[int, person]
	zero.RetainSet(other) // must not panic
	s.RetainSet(s)
	if s.Len() != 2 {
		t.Errorf("s.RetainSet(s) changed Len() to %d, want 2", s.Len())
	}
}

// RetainSet is the in-place form of Intersection.
func TestRetainSetMatchesIntersection(t *testing.T) {
	err := quick.Check(func(x, y []byte) bool {
		a, b := newBytes(x...), newBytes(y...)
		retained := a.Clone()
		retained.RetainSet(b)
		return retained.Equal(a.Intersection(b))
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveFunc(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
	if n := s.RemoveFunc(func(p person) bool { return p.ID != 2 }); n != 2 {
		t.Errorf("RemoveFunc = %d, want 2", n)
	}
	if got, want := ids(s), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("after RemoveFunc ids = %v, want %v", got, want)
	}

	var zero Set[int, person]
	if n := zero.RemoveFunc(func(person) bool { return true }); n != 0 {
		t.Errorf("RemoveFunc on zero value = %d, want 0", n)
	}
}

func TestPop(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
	var popped []int
	for {
		p, ok := s.Pop()
		if !ok {
			break
		}
		popped = append(popped, p.ID)
	}
	if !s.IsEmpty() || len(popped) != 3 {
		t.Errorf("draining left %d elements, popped %v", s.Len(), popped)
	}

	var zero Set[int, person]
	if p, ok := zero.Pop(); ok || p != (person{}) {
		t.Errorf("Pop on zero value = %v, %v, want zero value, false", p, ok)
	}
}

func TestContainsAllAny(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"})
	// Only the key matters, so the names here differ from the stored ones.
	present := []person{{1, "x"}, {2, "y"}}
	mixed := []person{{1, "x"}, {9, "z"}}
	absent := []person{{8, "x"}, {9, "z"}}

	for _, tc := range []struct {
		name     string
		elements []person
		all, any bool
	}{
		{"none", nil, true, false},
		{"all present", present, true, true},
		{"some present", mixed, false, true},
		{"none present", absent, false, false},
	} {
		if got := s.ContainsAll(tc.elements...); got != tc.all {
			t.Errorf("%s: ContainsAll = %v, want %v", tc.name, got, tc.all)
		}
		if got := s.ContainsAny(tc.elements...); got != tc.any {
			t.Errorf("%s: ContainsAny = %v, want %v", tc.name, got, tc.any)
		}
	}
}

func TestFilter(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
	got := s.Filter(func(p person) bool { return p.ID > 1 })

	if want := []int{2, 3}; !reflect.DeepEqual(ids(got), want) {
		t.Errorf("Filter ids = %v, want %v", ids(got), want)
	}
	if s.Len() != 3 {
		t.Errorf("Filter modified the receiver: Len() = %d, want 3", s.Len())
	}
	got.Add(person{10, "j"}) // uses the receiver's key function and is independent of s
	if s.ContainsKey(10) {
		t.Error("Filter shares storage with the receiver")
	}
}
