package keyed

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

type person struct {
	ID   int
	Name string
}

func personID(p person) int { return p.ID }

func newPeople(ps ...person) Set[int, person] { return New(personID, ps...) }

// names returns the Name of every element, sorted, so tests can see which
// element was kept for a key.
func names(s Set[int, person]) []string {
	out := []string{}
	for _, p := range s.Slice() {
		out = append(out, p.Name)
	}
	sort.Strings(out)
	return out
}

// ids returns the sorted IDs in s.
func ids(s Set[int, person]) []int {
	out := []int{}
	for _, p := range s.Slice() {
		out = append(out, p.ID)
	}
	sort.Ints(out)
	return out
}

func mustPanic(t *testing.T, wantMsg string, fn func()) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q, got none", wantMsg)
		}
		if msg, _ := r.(string); !strings.Contains(msg, wantMsg) {
			t.Fatalf("panic = %v, want it to contain %q", r, wantMsg)
		}
	}()
	fn()
}

func TestNew(t *testing.T) {
	if s := newPeople(); !s.IsEmpty() {
		t.Errorf("New() has %d elements, want 0", s.Len())
	}
	s := newPeople(person{1, "a"}, person{2, "b"})
	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
}

func TestNewNilKeyPanics(t *testing.T) {
	mustPanic(t, "nil key function", func() { New[int, person](nil) })
}

func TestZeroValue(t *testing.T) {
	var s Set[int, person]

	// Reads that do not need a key function are fine.
	if s.Len() != 0 || !s.IsEmpty() || len(s.Slice()) != 0 {
		t.Error("zero value should read as an empty set")
	}
	if _, ok := s.Get(1); ok {
		t.Error("Get on zero value found an element")
	}
	var other Set[int, person]
	if !s.Equal(other) || !s.IsDisjoint(other) {
		t.Error("two zero values should be equal and disjoint")
	}

	// Anything that needs the key function fails with a clear message.
	mustPanic(t, "must be created with New", func() { s.Add(person{1, "a"}) })
	mustPanic(t, "must be created with New", func() { s.Contains(person{1, "a"}) })
	mustPanic(t, "must be created with New", func() { s.Remove(person{1, "a"}) })
}

// Elements with the same key are the same element: adding replaces.
func TestAddReplacesEqualKey(t *testing.T) {
	s := newPeople(person{1, "old"})
	s.Add(person{1, "new"}, person{2, "other"})

	if s.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", s.Len())
	}
	got, ok := s.Get(1)
	if !ok || got.Name != "new" {
		t.Errorf("Get(1) = %v, %v; want the replacement element", got, ok)
	}
}

// New with duplicate keys keeps the last one, matching Add.
func TestNewLastDuplicateWins(t *testing.T) {
	s := newPeople(person{1, "first"}, person{1, "last"})
	if got, _ := s.Get(1); got.Name != "last" {
		t.Errorf("Get(1).Name = %q, want %q", got.Name, "last")
	}
}

func TestContainsComparesKeysOnly(t *testing.T) {
	s := newPeople(person{1, "ann"})
	if !s.Contains(person{1, "a completely different name"}) {
		t.Error("Contains should match on key alone")
	}
	if s.Contains(person{2, "ann"}) {
		t.Error("Contains matched a different key")
	}
}

func TestRemoveByKey(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"})
	s.Remove(person{1, "ignored name"}, person{99, "absent"})
	if got := ids(s); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("after Remove = %v, want [2]", got)
	}
}

func TestGet(t *testing.T) {
	s := newPeople(person{7, "grace"})
	if p, ok := s.Get(7); !ok || p.Name != "grace" {
		t.Errorf("Get(7) = %v, %v", p, ok)
	}
	if _, ok := s.Get(8); ok {
		t.Error("Get(8) found an element")
	}
}

// Slice-bearing structs are not comparable, so they cannot be members of a
// readyset.Set, but they work here through a key.
func TestNonComparableElements(t *testing.T) {
	type doc struct {
		Name string
		Tags []string
	}
	s := New(func(d doc) string { return d.Name }, doc{"a", []string{"x"}}, doc{"b", nil})
	if !s.Contains(doc{Name: "a"}) || s.Len() != 2 {
		t.Errorf("unexpected contents: %v", s.Slice())
	}
}

func TestClearAndClone(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"})
	c := s.Clone()
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Clear left elements behind")
	}
	if got := ids(c); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("clone changed through original: %v", got)
	}
	c.Add(person{3, "c"})
	if s.Contains(person{3, ""}) {
		t.Error("original changed through clone")
	}
	s.Add(person{4, "d"}) // a cleared set is still usable
	if got := ids(s); !reflect.DeepEqual(got, []int{4}) {
		t.Errorf("Add after Clear = %v, want [4]", got)
	}
}

// Union keeps the receiver's element for a key that is in both sets.
func TestUnionKeepsReceiverElement(t *testing.T) {
	a := newPeople(person{1, "from-a"}, person{2, "only-a"})
	b := newPeople(person{1, "from-b"}, person{3, "only-b"})

	if got := names(a.Union(b)); !reflect.DeepEqual(got, []string{"from-a", "only-a", "only-b"}) {
		t.Errorf("a.Union(b) = %v", got)
	}
	if got := names(b.Union(a)); !reflect.DeepEqual(got, []string{"from-b", "only-a", "only-b"}) {
		t.Errorf("b.Union(a) = %v", got)
	}
}

// Intersection takes elements from the receiver whichever set is smaller.
func TestIntersectionKeepsReceiverElement(t *testing.T) {
	small := newPeople(person{1, "small"})
	large := newPeople(person{1, "large"}, person{2, "x"}, person{3, "y"})

	if got := names(small.Intersection(large)); !reflect.DeepEqual(got, []string{"small"}) {
		t.Errorf("small.Intersection(large) = %v, want [small]", got)
	}
	if got := names(large.Intersection(small)); !reflect.DeepEqual(got, []string{"large"}) {
		t.Errorf("large.Intersection(small) = %v, want [large]", got)
	}
}

func TestEqualComparesKeysOnly(t *testing.T) {
	a := newPeople(person{1, "a"}, person{2, "b"})
	b := newPeople(person{2, "different"}, person{1, "names"})
	if !a.Equal(b) {
		t.Error("sets with the same keys should be Equal")
	}
	if a.Equal(newPeople(person{1, "a"})) {
		t.Error("sets with different sizes are not Equal")
	}
	if a.Equal(newPeople(person{1, "a"}, person{3, "c"})) {
		t.Error("sets with different keys are not Equal")
	}
}

func TestBinaryOperations(t *testing.T) {
	mk := func(keys ...int) Set[int, person] {
		ps := make([]person, len(keys))
		for i, k := range keys {
			ps[i] = person{ID: k}
		}
		return newPeople(ps...)
	}
	cases := []struct {
		name                       string
		a, b                       []int
		union, inter, diff, sym    []int
		disjoint, subset, superset bool
	}{
		{"both empty", nil, nil, []int{}, []int{}, []int{}, []int{}, true, true, true},
		{"a empty", nil, []int{1, 2}, []int{1, 2}, []int{}, []int{}, []int{1, 2}, true, true, false},
		{"b empty", []int{1, 2}, nil, []int{1, 2}, []int{}, []int{1, 2}, []int{1, 2}, true, false, true},
		{"identical", []int{1, 2}, []int{1, 2}, []int{1, 2}, []int{1, 2}, []int{}, []int{}, false, true, true},
		{"disjoint", []int{2, 4}, []int{1, 3}, []int{1, 2, 3, 4}, []int{}, []int{2, 4}, []int{1, 2, 3, 4}, true, false, false},
		{"overlap", []int{1, 2, 3}, []int{2, 3, 5}, []int{1, 2, 3, 5}, []int{2, 3}, []int{1}, []int{1, 5}, false, false, false},
		{"a inside b", []int{2, 3}, []int{1, 2, 3, 4}, []int{1, 2, 3, 4}, []int{2, 3}, []int{}, []int{1, 4}, false, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, b := mk(c.a...), mk(c.b...)
			check := func(op string, got Set[int, person], want []int) {
				t.Helper()
				if g := ids(got); !reflect.DeepEqual(g, want) {
					t.Errorf("%s = %v, want %v", op, g, want)
				}
			}
			check("Union", a.Union(b), c.union)
			check("Intersection", a.Intersection(b), c.inter)
			check("Difference", a.Difference(b), c.diff)
			check("SymmetricDifference", a.SymmetricDifference(b), c.sym)
			if got := a.IsDisjoint(b); got != c.disjoint {
				t.Errorf("IsDisjoint = %v, want %v", got, c.disjoint)
			}
			if got := a.IsSubset(b); got != c.subset {
				t.Errorf("IsSubset = %v, want %v", got, c.subset)
			}
			if got := a.IsSuperset(b); got != c.superset {
				t.Errorf("IsSuperset = %v, want %v", got, c.superset)
			}
			if !reflect.DeepEqual(ids(a), sortedCopy(c.a)) || !reflect.DeepEqual(ids(b), sortedCopy(c.b)) {
				t.Error("an operand was modified")
			}
		})
	}
}

func sortedCopy(in []int) []int {
	out := append([]int{}, in...)
	sort.Ints(out)
	return out
}

func TestRemoveKey(t *testing.T) {
	s := newPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
	s.RemoveKey(1, 3, 99) // 99 is not present; that is not an error
	if got := ids(s); !reflect.DeepEqual(got, []int{2}) {
		t.Errorf("after RemoveKey = %v, want [2]", got)
	}
	s.RemoveKey() // no-op
	if s.Len() != 1 {
		t.Errorf("RemoveKey() with no keys changed the set: %v", ids(s))
	}
}

func TestContainsKey(t *testing.T) {
	s := newPeople(person{7, "grace"})
	if !s.ContainsKey(7) || s.ContainsKey(8) {
		t.Error("ContainsKey gave the wrong answer")
	}
}

func TestKeys(t *testing.T) {
	s := newPeople(person{3, "c"}, person{1, "a"}, person{2, "b"})
	got := s.Keys()
	sort.Ints(got)
	if want := []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
	if n := len(newPeople().Keys()); n != 0 {
		t.Errorf("Keys() of an empty set has %d keys", n)
	}
}

// The by-key methods never call the key function, so they work on a zero value.
func TestKeyMethodsOnZeroValue(t *testing.T) {
	var s Set[int, person]
	s.RemoveKey(1) // must not panic
	if s.ContainsKey(1) || len(s.Keys()) != 0 {
		t.Error("zero value should be empty")
	}
}
