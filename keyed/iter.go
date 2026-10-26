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

// Collect returns a new set, identified by key, of the elements yielded by
// seq. As with [New], if several elements share a key the last one wins, and
// Collect panics if key is nil. It is available when building with Go 1.23 or
// later.
func Collect[K comparable, V any](key func(V) K, seq iter.Seq[V]) Set[K, V] {
	s := New(key)
	for e := range seq {
		s.Add(e)
	}
	return s
}
