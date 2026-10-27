package readyset

import (
	"bytes"
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

// A set holds only elements that are equal to themselves. A NaN would otherwise
// get in but never be found, removed or compared, and every set law would fail
// for it.
func TestNaNIsIgnored(t *testing.T) {
	nan := math.NaN()

	s := New(nan, 1.0, nan)
	if s.Len() != 1 || s.Contains(nan) || !s.Contains(1) {
		t.Errorf("New(NaN, 1, NaN): Len() = %d, want just {1}", s.Len())
	}
	s.Add(nan)
	if s.Len() != 1 {
		t.Errorf("Add(NaN) changed Len() to %d, want 1", s.Len())
	}

	var zero Set[float64]
	zero.Add(nan)
	if !zero.IsEmpty() {
		t.Errorf("Add(NaN) to the zero value gave %v, want an empty set", zero)
	}

	// A value that contains a NaN is not equal to itself either.
	type reading struct{ Value float64 }
	r := New(reading{nan}, reading{2})
	if r.Len() != 1 || !r.Contains(reading{2}) {
		t.Errorf("a struct holding NaN got into the set: Len() = %d", r.Len())
	}
}

func TestNaNLaws(t *testing.T) {
	nan := math.NaN()
	s := New(nan, 1.0, 2.0)

	if !s.Equal(s) || !s.IsSubset(s) || !s.IsSuperset(s) || s.IsProperSubset(s) {
		t.Error("a set is not equal to, or a subset of, itself")
	}
	if got := s.Union(s).Len(); got != 2 {
		t.Errorf("s.Union(s).Len() = %d, want 2", got)
	}
	if !s.Difference(s).IsEmpty() || !s.SymmetricDifference(s).IsEmpty() {
		t.Error("s minus itself is not empty")
	}

	s.AddSet(s)
	if s.Len() != 2 {
		t.Errorf("s.AddSet(s) changed Len() to %d, want 2", s.Len())
	}
	s.RetainSet(s)
	s.RemoveSet(s)
	if !s.IsEmpty() {
		t.Errorf("s.RemoveSet(s) left %v", s)
	}
}

// With a NaN stuck in the set, Pop would never finish draining and RemoveFunc
// would report removals that did not happen.
func TestNaNCannotBreakMutators(t *testing.T) {
	s := New(math.NaN(), 1.0)
	for i := 0; ; i++ {
		if _, ok := s.Pop(); !ok {
			break
		}
		if i > 10 {
			t.Fatal("Pop never finished draining the set")
		}
	}

	s = New(math.NaN(), 1.0, 2.0)
	if n := s.RemoveFunc(func(float64) bool { return true }); n != 2 || !s.IsEmpty() {
		t.Errorf("RemoveFunc = %d, left %v; want 2 and an empty set", n, s)
	}

	s = New(math.NaN(), 1.0)
	s.Clear()
	if !s.IsEmpty() {
		t.Errorf("Clear left %v", s)
	}
}

func TestNaNInCollect(t *testing.T) {
	got := Collect(func(yield func(float64) bool) {
		for _, f := range []float64{1, math.NaN(), 1, 2} {
			if !yield(f) {
				return
			}
		}
	})
	if want := []float64{1, 2}; !reflect.DeepEqual(Sorted(got), want) {
		t.Errorf("Collect = %v, want %v", Sorted(got), want)
	}
}

// Decoding into a Set[any] used to panic on an array or object element and
// leave the set half-changed.
func TestUnmarshalJSONUnhashableElement(t *testing.T) {
	s := New[any](1.0, 2.0)
	err := json.Unmarshal([]byte(`[7, [8]]`), &s)
	if err == nil {
		t.Fatal("decoding a nested array into a Set[any] succeeded, want an error")
	}
	if !s.Contains(1.0) || !s.Contains(2.0) || s.Len() != 2 {
		t.Errorf("a failed decode changed the set to %v", s)
	}

	if err := json.Unmarshal([]byte(`[7, "x"]`), &s); err != nil || s.Len() != 2 || !s.Contains("x") {
		t.Errorf("a valid decode gave %v, %v", s, err)
	}
}

// MarshalJSON must leave HTML escaping to the encoder, as a plain slice does.
func TestMarshalJSONHonoursEscapeHTML(t *testing.T) {
	s := New("<a&b>")

	got, err := json.Marshal(s)
	want, _ := json.Marshal([]string{"<a&b>"})
	if err != nil || !bytes.Equal(got, want) {
		t.Errorf("json.Marshal = %s, %v; want %s", got, err, want)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "[\"<a&b>\"]\n"; got != want {
		t.Errorf("Encode with SetEscapeHTML(false) = %q, want %q", got, want)
	}
}
