package readyset

import (
	"cmp"
	"slices"
)

// Sorted returns the elements of s as a new slice in ascending order. It works
// for any ordered element type: integers, floats and strings. NaN, if present,
// sorts before every other float.
//
// Sorted is a function rather than a method because only an ordered element
// type can be sorted, and a method cannot add that constraint to [Set].
func Sorted[T cmp.Ordered](s Set[T]) []T {
	elements := s.Slice()
	slices.Sort(elements)
	return elements
}
