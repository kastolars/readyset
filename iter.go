package readyset

import "iter"

// All returns an iterator over the elements of the set, in no particular
// order.
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

// Collect returns a new set of the elements yielded by seq.
func Collect[T comparable](seq iter.Seq[T]) Set[T] {
	s := Set[T]{set: make(map[T]struct{})}
	for e := range seq {
		s.set[e] = struct{}{}
	}
	return s
}
