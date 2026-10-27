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

// Collect returns a new set of the elements yielded by seq. As with [New], an
// element that is not equal to itself, such as a NaN, is ignored.
func Collect[T comparable](seq iter.Seq[T]) Set[T] {
	var s Set[T]
	for e := range seq {
		s.Add(e)
	}
	if s.set == nil {
		s.set = make(map[T]struct{})
	}
	return s
}
