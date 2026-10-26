package readyset

import (
	"reflect"
	"testing"
	"testing/quick"
)

func TestRetainSet(t *testing.T) {
	s := New(1, 2, 3)
	other := New(2, 3, 9)
	s.RetainSet(other)

	if got, want := sorted(s), []int{2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("after RetainSet = %v, want %v", got, want)
	}
	if got, want := sorted(other), []int{2, 3, 9}; !reflect.DeepEqual(got, want) {
		t.Errorf("RetainSet modified its argument: %v, want %v", got, want)
	}

	var zero Set[int]
	zero.RetainSet(New(1)) // must not panic
	s.RetainSet(s)
	if got, want := sorted(s), []int{2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("s.RetainSet(s) = %v, want %v", got, want)
	}
	s.RetainSet(Set[int]{})
	if !s.IsEmpty() {
		t.Errorf("RetainSet with an empty set left %v", sorted(s))
	}
}

// RetainSet is the in-place form of Intersection.
func TestRetainSetMatchesIntersection(t *testing.T) {
	err := quick.Check(func(x, y []byte) bool {
		a, b := New(x...), New(y...)
		retained := a.Clone()
		retained.RetainSet(b)
		return retained.Equal(a.Intersection(b))
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveFunc(t *testing.T) {
	s := New(1, 2, 3, 4, 5)
	if n := s.RemoveFunc(func(e int) bool { return e%2 == 0 }); n != 2 {
		t.Errorf("RemoveFunc = %d, want 2", n)
	}
	if got, want := sorted(s), []int{1, 3, 5}; !reflect.DeepEqual(got, want) {
		t.Errorf("after RemoveFunc = %v, want %v", got, want)
	}

	var zero Set[int]
	if n := zero.RemoveFunc(func(int) bool { return true }); n != 0 {
		t.Errorf("RemoveFunc on zero value = %d, want 0", n)
	}
}

func TestPop(t *testing.T) {
	s := New(1, 2, 3)
	seen := New[int]()
	for {
		e, ok := s.Pop()
		if !ok {
			break
		}
		if seen.Contains(e) {
			t.Fatalf("Pop returned %d twice", e)
		}
		seen.Add(e)
	}
	if !s.IsEmpty() || !seen.Equal(New(1, 2, 3)) {
		t.Errorf("draining left s=%v, popped %v", sorted(s), sorted(seen))
	}

	var zero Set[string]
	if e, ok := zero.Pop(); ok || e != "" {
		t.Errorf("Pop on zero value = %q, %v, want \"\", false", e, ok)
	}
}

func TestContainsAllAny(t *testing.T) {
	s := New(1, 2, 3)
	for _, tc := range []struct {
		name     string
		elements []int
		all, any bool
	}{
		{"none", nil, true, false},
		{"all present", []int{1, 3}, true, true},
		{"some present", []int{1, 9}, false, true},
		{"none present", []int{8, 9}, false, false},
	} {
		if got := s.ContainsAll(tc.elements...); got != tc.all {
			t.Errorf("%s: ContainsAll = %v, want %v", tc.name, got, tc.all)
		}
		if got := s.ContainsAny(tc.elements...); got != tc.any {
			t.Errorf("%s: ContainsAny = %v, want %v", tc.name, got, tc.any)
		}
	}
	var zero Set[int]
	if zero.ContainsAll(1) || zero.ContainsAny(1) || !zero.ContainsAll() {
		t.Error("zero value gave a wrong answer")
	}
}

func TestFilter(t *testing.T) {
	s := New(1, 2, 3, 4)
	even := s.Filter(func(e int) bool { return e%2 == 0 })

	if got, want := sorted(even), []int{2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("Filter = %v, want %v", got, want)
	}
	if s.Len() != 4 {
		t.Errorf("Filter modified the receiver: Len() = %d, want 4", s.Len())
	}
	even.Add(10) // the result is independent of s
	if s.Contains(10) {
		t.Error("Filter shares storage with the receiver")
	}

	var zero Set[int]
	none := zero.Filter(func(int) bool { return true })
	none.Add(1) // the result is usable even when empty
	if !none.Contains(1) {
		t.Error("Filter of an empty set is not usable")
	}
}
