//go:build go1.23

package keyed

import "iter"

// All returns an iterator over the elements of the set, in no particular
// order. It is available when building with Go 1.23 or later.
//
// The set must not be modified while iterating over it.
func (s Set[K, V]) All() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, e := range s.items {
			if !yield(e) {
				return
			}
		}
	}
}
