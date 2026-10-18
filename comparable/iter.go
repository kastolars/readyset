//go:build go1.23

package comparable

import "iter"

// All returns an iterator over the elements of the set, in no particular
// order. It is available when building with Go 1.23 or later.
//
// The set must not be modified while iterating over it.
func (s Set[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := range s.set {
			if !yield(e) {
				return
			}
		}
	}
}
