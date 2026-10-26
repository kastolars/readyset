//go:build go1.21

package keyed

import (
	"cmp"
	"slices"
)

// Sorted returns the elements of s as a new slice in ascending order of their
// keys. It works when the key type is ordered: integers, floats and strings.
//
// Sorted is a function rather than a method because only an ordered key type
// can be sorted, and a method cannot add that constraint to [Set]. It is
// available when building with Go 1.21 or later.
func Sorted[K cmp.Ordered, V any](s Set[K, V]) []V {
	keys := s.Keys()
	slices.Sort(keys)

	elements := make([]V, len(keys))
	for i, k := range keys {
		elements[i] = s.items[k]
	}
	return elements
}
