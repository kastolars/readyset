package order

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func TestSortNumbers(t *testing.T) {
	ints := []int{10, 2, -5, 1, 100, 0}
	Sort(ints)
	if want := []int{-5, 0, 1, 2, 10, 100}; !reflect.DeepEqual(ints, want) {
		t.Errorf("ints = %v, want %v", ints, want)
	}

	uints := []uint8{200, 3, 20}
	Sort(uints)
	if want := []uint8{3, 20, 200}; !reflect.DeepEqual(uints, want) {
		t.Errorf("uint8s = %v, want %v", uints, want)
	}

	floats := []float64{2.5, math.NaN(), -1, math.Inf(1), 0}
	Sort(floats)
	got := fmt.Sprint(floats)
	if want := "[NaN -1 0 2.5 +Inf]"; got != want {
		t.Errorf("floats = %s, want %s", got, want)
	}
}

func TestSortStringsAndBools(t *testing.T) {
	strs := []string{"b", "B", "a", "", "10", "2"}
	Sort(strs)
	if want := []string{"", "10", "2", "B", "a", "b"}; !reflect.DeepEqual(strs, want) {
		t.Errorf("strings = %v, want %v", strs, want)
	}

	bools := []bool{true, false, true}
	Sort(bools)
	if want := []bool{false, true, true}; !reflect.DeepEqual(bools, want) {
		t.Errorf("bools = %v, want %v", bools, want)
	}
}

func TestSortShortSlices(t *testing.T) {
	Sort[int](nil)
	Sort([]int{})
	one := []string{"x"}
	Sort(one)
	if one[0] != "x" {
		t.Error("single-element slice changed")
	}
}

type point struct{ X, Y int }

func TestSortOtherKinds(t *testing.T) {
	pts := []point{{2, 1}, {1, 9}, {1, 2}}
	Sort(pts)
	// Ordered by printed form: {1 2} < {1 9} < {2 1}.
	if want := []point{{1, 2}, {1, 9}, {2, 1}}; !reflect.DeepEqual(pts, want) {
		t.Errorf("points = %v, want %v", pts, want)
	}
}

// Strings that print identically must still sort the same way every time, so
// values are tie-broken by their Go-syntax form.
func TestSortTieBreak(t *testing.T) {
	type pair struct{ A, B string }
	// Both print as "{a b c}" with fmt.Sprint.
	x, y := pair{"a b", "c"}, pair{"a", "b c"}
	// In Go syntax the first differs from the second at a space (0x20) versus
	// a quote (0x22), so x sorts first whichever way round the input is.
	for _, in := range [][]pair{{x, y}, {y, x}} {
		Sort(in)
		if in[0] != x || in[1] != y {
			t.Errorf("tie-break order = %#v, want [%#v %#v]", in, x, y)
		}
	}
}

func TestSortMixedKindsIsOrderIndependent(t *testing.T) {
	vals := []any{nil, true, int8(-3), 2, uint(7), 1.5, "a", "B", point{1, 2}, []int{9}}
	want := append([]any(nil), vals...)
	Sort(want)

	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 200; i++ {
		got := append([]any(nil), vals...)
		rng.Shuffle(len(got), func(a, b int) { got[a], got[b] = got[b], got[a] })
		Sort(got)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("order depends on input order:\n got %v\nwant %v", got, want)
		}
	}

	// Grouped by kind: nil, bool, ints, uints, floats, then the rest.
	if got := fmt.Sprint(want[:7]); got != "[<nil> true -3 2 7 1.5 [9]]" {
		t.Errorf("kind grouping = %s", got)
	}
}

// Values that print the same and have the same number, such as an int and an
// int64, tie on everything but their type, so the type breaks the tie. Without
// that the order followed the random order of the input.
func TestSortTiesAreBrokenByType(t *testing.T) {
	reference := []any{int8(1), int(1), int64(1), uint8(1), uint(1), float32(1), float64(1), "1"}
	var first []string
	for i := 0; i < 200; i++ {
		in := append([]any(nil), reference...)
		rand.Shuffle(len(in), func(a, b int) { in[a], in[b] = in[b], in[a] })
		Sort(in)

		got := make([]string, len(in))
		for j, v := range in {
			got[j] = fmt.Sprintf("%T:%v", v, v)
		}
		if first == nil {
			first = got
		} else if !reflect.DeepEqual(first, got) {
			t.Fatalf("order depends on the input order:\n%v\n%v", first, got)
		}
	}
}
