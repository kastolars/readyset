package keyed

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

type reading struct {
	At   float64
	Note string
}

func newReadings(rs ...reading) Set[float64, reading] {
	return New(func(r reading) float64 { return r.At }, rs...)
}

// An element whose key is a NaN is ignored. It used to get in, and then Sorted,
// String and MarshalJSON looked it up by key, missed, and printed a zero value
// in its place.
func TestNaNKeyIsIgnored(t *testing.T) {
	nan := math.NaN()
	s := newReadings(reading{nan, "lost"}, reading{1, "kept"}, reading{nan, "lost too"})

	if s.Len() != 1 || !s.ContainsKey(1) {
		t.Fatalf("Len() = %d, want just the element with key 1", s.Len())
	}
	if got := Sorted(s); !reflect.DeepEqual(got, []reading{{1, "kept"}}) {
		t.Errorf("Sorted = %v", got)
	}
	if got, want := s.String(), "{{1 kept}}"; got != want {
		t.Errorf("String = %q, want %q", got, want)
	}
	b, err := json.Marshal(s)
	if err != nil || string(b) != `[{"At":1,"Note":"kept"}]` {
		t.Errorf("MarshalJSON = %s, %v", b, err)
	}

	s.Add(reading{nan, "also lost"})
	if s.Len() != 1 {
		t.Errorf("Add changed Len() to %d, want 1", s.Len())
	}
}

func TestNaNKeyCannotBreakMutators(t *testing.T) {
	nan := math.NaN()

	s := newReadings(reading{nan, "x"}, reading{1, "y"})
	if !s.Equal(s) || !s.IsSubset(s) {
		t.Error("a set is not equal to, or a subset of, itself")
	}
	for i := 0; ; i++ {
		if _, ok := s.Pop(); !ok {
			break
		}
		if i > 10 {
			t.Fatal("Pop never finished draining the set")
		}
	}

	s = newReadings(reading{nan, "x"}, reading{1, "y"}, reading{2, "z"})
	if n := s.RemoveFunc(func(reading) bool { return true }); n != 2 || !s.IsEmpty() {
		t.Errorf("RemoveFunc = %d, left %d elements; want 2 and none", n, s.Len())
	}
}

func TestNaNKeyInCollect(t *testing.T) {
	key := func(r reading) float64 { return r.At }
	got := Collect(key, func(yield func(reading) bool) {
		for _, r := range []reading{{math.NaN(), "a"}, {1, "b"}} {
			if !yield(r) {
				return
			}
		}
	})
	if got.Len() != 1 || !got.ContainsKey(1) {
		t.Errorf("Collect kept %d elements, want just the one with key 1", got.Len())
	}
}

// Combining the zero value with a real set gives a usable set that uses the real
// set's key function, instead of one that panics on the first Add.
func TestZeroValueReceiverOperationsAreUsable(t *testing.T) {
	var zero Set[int, person]
	other := newPeople(person{1, "a"})

	for name, got := range map[string]Set[int, person]{
		"Union":               zero.Union(other),
		"SymmetricDifference": zero.SymmetricDifference(other),
	} {
		got.Add(person{2, "b"}) // panicked: no key function
		if got.Len() != 2 || !got.ContainsKey(1) || !got.ContainsKey(2) {
			t.Errorf("%s: ids = %v, want [1 2]", name, ids(got))
		}
	}
}

func TestUnmarshalJSONUnhashableKey(t *testing.T) {
	s := New(func(v any) any { return v }, any(1.0), any(2.0))
	if err := json.Unmarshal([]byte(`[7, [8]]`), &s); err == nil {
		t.Fatal("decoding an unhashable key succeeded, want an error")
	}
	if s.Len() != 2 || !s.ContainsKey(1.0) || !s.ContainsKey(2.0) {
		t.Errorf("a failed decode changed the set: Len() = %d", s.Len())
	}
}
