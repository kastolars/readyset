package comparable

import (
	"reflect"
	"sort"
	"testing"
)

// sorted returns the elements of s in ascending order, for stable comparison.
func sorted(s Set[int]) []int {
	out := s.Slice()
	sort.Ints(out)
	return out
}

func TestNew(t *testing.T) {
	if got := New[int](); !got.IsEmpty() {
		t.Errorf("New() has %d elements, want 0", got.Len())
	}
	if got := sorted(New(3, 1, 2, 3, 1)); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("New(3, 1, 2, 3, 1) = %v, want [1 2 3]", got)
	}
}

func TestZeroValue(t *testing.T) {
	var s Set[int]

	if s.Len() != 0 || !s.IsEmpty() || s.Contains(1) {
		t.Fatal("zero value is not an empty set")
	}
	if len(s.Slice()) != 0 {
		t.Error("zero value Slice() is not empty")
	}
	s.Remove(1) // must not panic
	s.Clear()   // must not panic

	var other Set[int]
	if !s.Equal(other) || !s.IsSubset(other) || !s.IsDisjoint(other) {
		t.Error("two zero values should be equal, subsets of each other, and disjoint")
	}
	if !s.Union(other).IsEmpty() || !s.Intersection(other).IsEmpty() ||
		!s.Difference(other).IsEmpty() || !s.SymmetricDifference(other).IsEmpty() {
		t.Error("set operations on zero values should be empty")
	}
	if !s.Clone().IsEmpty() {
		t.Error("Clone of zero value should be empty")
	}

	s.Add(1, 2) // must not panic
	if got := sorted(s); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("after Add(1, 2) = %v, want [1 2]", got)
	}
}

func TestAddRemoveContains(t *testing.T) {
	s := New[int]()
	s.Add(1)
	s.Add(2, 3, 3)
	s.Add() // no-op

	for _, e := range []int{1, 2, 3} {
		if !s.Contains(e) {
			t.Errorf("Contains(%d) = false after Add", e)
		}
	}
	if s.Len() != 3 {
		t.Errorf("Len() = %d, want 3", s.Len())
	}

	s.Remove(1, 99) // removing an absent element is a no-op
	if s.Contains(1) || s.Len() != 2 {
		t.Errorf("after Remove(1, 99): %v", sorted(s))
	}
}

func TestClear(t *testing.T) {
	s := New(1, 2, 3)
	s.Clear()
	if !s.IsEmpty() || s.Contains(1) {
		t.Errorf("after Clear: %v", sorted(s))
	}
	s.Add(4)
	if got := sorted(s); !reflect.DeepEqual(got, []int{4}) {
		t.Errorf("Add after Clear = %v, want [4]", got)
	}
}

func TestClone(t *testing.T) {
	orig := New(1, 2, 3)
	c := orig.Clone()
	if !c.Equal(orig) {
		t.Fatal("Clone is not equal to the original")
	}
	c.Add(4)
	orig.Remove(1)
	if got := sorted(orig); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("original changed through clone: %v", got)
	}
	if got := sorted(c); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Errorf("clone changed through original: %v", got)
	}
}

func TestSliceLength(t *testing.T) {
	s := New(2, 4, 6, 8, 10, 12)
	if len(s.Slice()) != s.Len() {
		t.Errorf("len(Slice()) = %d, Len() = %d", len(s.Slice()), s.Len())
	}
	for _, e := range s.Slice() {
		if !s.Contains(e) {
			t.Errorf("Slice() returned %d which is not in the set", e)
		}
	}
}

// pairs are shared by the binary-operation tests below.
var pairs = []struct {
	name string
	a, b []int
}{
	{"both empty", nil, nil},
	{"a empty", nil, []int{1, 2}},
	{"b empty", []int{1, 2}, nil},
	{"identical", []int{1, 2, 3}, []int{1, 2, 3}},
	{"disjoint", []int{2, 4, 6}, []int{1, 3, 5}},
	{"overlap", []int{1, 2, 3}, []int{2, 3, 5}},
	{"a inside b", []int{2, 3}, []int{1, 2, 3, 4}},
	{"b inside a", []int{1, 2, 3, 4}, []int{2, 3}},
}

func TestEqual(t *testing.T) {
	cases := []struct {
		a, b []int
		want bool
	}{
		{nil, nil, true},
		{[]int{1, 2, 3, 4}, []int{4, 3, 2, 1}, true},
		{[]int{1, 2, 3, 4}, []int{1, 2, 3}, false},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, false},
		{[]int{1, 2, 3, 4}, []int{1, 2, 3, 5}, false},
		{nil, []int{1}, false},
	}
	for _, c := range cases {
		if got := New(c.a...).Equal(New(c.b...)); got != c.want {
			t.Errorf("Equal(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestIsSubsetIsSuperset(t *testing.T) {
	cases := []struct {
		a, b          []int
		subset, super bool // a.IsSubset(b), a.IsSuperset(b)
	}{
		{nil, nil, true, true},
		{nil, []int{1}, true, false},
		{[]int{1}, nil, false, true},
		{[]int{1, 2}, []int{1, 2}, true, true},
		{[]int{2, 3}, []int{1, 2, 3, 4}, true, false},
		{[]int{1, 2, 3, 4}, []int{2, 3}, false, true},
		{[]int{1, 2}, []int{2, 3}, false, false},
	}
	for _, c := range cases {
		a, b := New(c.a...), New(c.b...)
		if got := a.IsSubset(b); got != c.subset {
			t.Errorf("%v.IsSubset(%v) = %v, want %v", c.a, c.b, got, c.subset)
		}
		if got := a.IsSuperset(b); got != c.super {
			t.Errorf("%v.IsSuperset(%v) = %v, want %v", c.a, c.b, got, c.super)
		}
	}
}

func TestBinaryOperations(t *testing.T) {
	// Expected results are computed independently of the implementation,
	// from plain slices, so a bug in Set cannot hide itself.
	in := func(xs []int, v int) bool {
		for _, x := range xs {
			if x == v {
				return true
			}
		}
		return false
	}
	universe := func(a, b []int) []int {
		seen := map[int]bool{}
		var out []int
		for _, v := range append(append([]int{}, a...), b...) {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
		sort.Ints(out)
		return out
	}
	want := func(a, b []int, keep func(inA, inB bool) bool) []int {
		out := []int{}
		for _, v := range universe(a, b) {
			if keep(in(a, v), in(b, v)) {
				out = append(out, v)
			}
		}
		return out
	}

	ops := []struct {
		name string
		got  func(a, b Set[int]) Set[int]
		keep func(inA, inB bool) bool
	}{
		{"Union", Set[int].Union, func(x, y bool) bool { return x || y }},
		{"Intersection", Set[int].Intersection, func(x, y bool) bool { return x && y }},
		{"Difference", Set[int].Difference, func(x, y bool) bool { return x && !y }},
		{"SymmetricDifference", Set[int].SymmetricDifference, func(x, y bool) bool { return x != y }},
	}

	for _, p := range pairs {
		for _, op := range ops {
			a, b := New(p.a...), New(p.b...)
			got := sorted(op.got(a, b))
			exp := want(p.a, p.b, op.keep)
			if !reflect.DeepEqual(got, exp) {
				t.Errorf("%s/%s: got %v, want %v", p.name, op.name, got, exp)
			}
			// Operands must not be modified.
			if !a.Equal(New(p.a...)) || !b.Equal(New(p.b...)) {
				t.Errorf("%s/%s modified an operand", p.name, op.name)
			}
		}

		a, b := New(p.a...), New(p.b...)
		wantDisjoint := len(want(p.a, p.b, func(x, y bool) bool { return x && y })) == 0
		if got := a.IsDisjoint(b); got != wantDisjoint {
			t.Errorf("%s/IsDisjoint: got %v, want %v", p.name, got, wantDisjoint)
		}
		if got := b.IsDisjoint(a); got != wantDisjoint {
			t.Errorf("%s/IsDisjoint (swapped): got %v, want %v", p.name, got, wantDisjoint)
		}
	}
}
