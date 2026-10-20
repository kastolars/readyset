package readyset

// Set is an unordered collection of unique elements.
//
// The zero value is an empty set that is ready to use. Methods that modify a
// set have pointer receivers, so call them on an addressable variable. Copying
// a non-empty Set copies a reference to the same underlying storage; use
// [Set.Clone] for an independent copy.
//
// A Set is not safe for concurrent use.
type Set[T comparable] struct {
	set map[T]struct{}
}

// New returns a Set containing elements. Duplicates are ignored.
func New[T comparable](elements ...T) Set[T] {
	s := Set[T]{set: make(map[T]struct{}, len(elements))}
	for _, e := range elements {
		s.set[e] = struct{}{}
	}
	return s
}

// Add inserts elements into the set. Elements already present are ignored.
func (s *Set[T]) Add(elements ...T) {
	if s.set == nil {
		s.set = make(map[T]struct{}, len(elements))
	}
	for _, e := range elements {
		s.set[e] = struct{}{}
	}
}

// Remove deletes elements from the set. Elements not present are ignored.
func (s *Set[T]) Remove(elements ...T) {
	for _, e := range elements {
		delete(s.set, e)
	}
}

// Clear removes every element from the set.
func (s *Set[T]) Clear() {
	for e := range s.set {
		delete(s.set, e)
	}
}

// Contains reports whether element is in the set.
func (s Set[T]) Contains(element T) bool {
	_, ok := s.set[element]
	return ok
}

// Len returns the number of elements in the set.
func (s Set[T]) Len() int {
	return len(s.set)
}

// IsEmpty reports whether the set has no elements.
func (s Set[T]) IsEmpty() bool {
	return len(s.set) == 0
}

// Equal reports whether s and other contain exactly the same elements.
func (s Set[T]) Equal(other Set[T]) bool {
	return s.Len() == other.Len() && s.IsSubset(other)
}

// IsSubset reports whether every element of s is also in other.
// The empty set is a subset of every set, and every set is a subset of itself.
func (s Set[T]) IsSubset(other Set[T]) bool {
	if s.Len() > other.Len() {
		return false
	}
	for e := range s.set {
		if !other.Contains(e) {
			return false
		}
	}
	return true
}

// IsSuperset reports whether every element of other is also in s.
func (s Set[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}

// IsDisjoint reports whether s and other have no elements in common.
func (s Set[T]) IsDisjoint(other Set[T]) bool {
	small, large := s, other
	if small.Len() > large.Len() {
		small, large = large, small
	}
	for e := range small.set {
		if large.Contains(e) {
			return false
		}
	}
	return true
}

// Slice returns the elements of the set as a new slice, in no particular
// order. The order can differ between calls.
func (s Set[T]) Slice() []T {
	elements := make([]T, 0, len(s.set))
	for e := range s.set {
		elements = append(elements, e)
	}
	return elements
}

// Clone returns an independent copy of the set.
func (s Set[T]) Clone() Set[T] {
	c := Set[T]{set: make(map[T]struct{}, len(s.set))}
	for e := range s.set {
		c.set[e] = struct{}{}
	}
	return c
}

// Union returns a new set of the elements that are in s, in other, or in both.
func (s Set[T]) Union(other Set[T]) Set[T] {
	u := Set[T]{set: make(map[T]struct{}, len(s.set)+len(other.set))}
	for e := range s.set {
		u.set[e] = struct{}{}
	}
	for e := range other.set {
		u.set[e] = struct{}{}
	}
	return u
}

// Intersection returns a new set of the elements that are in both s and other.
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	small, large := s, other
	if small.Len() > large.Len() {
		small, large = large, small
	}
	i := Set[T]{set: make(map[T]struct{})}
	for e := range small.set {
		if large.Contains(e) {
			i.set[e] = struct{}{}
		}
	}
	return i
}

// Difference returns a new set of the elements that are in s but not in other.
func (s Set[T]) Difference(other Set[T]) Set[T] {
	d := Set[T]{set: make(map[T]struct{})}
	for e := range s.set {
		if !other.Contains(e) {
			d.set[e] = struct{}{}
		}
	}
	return d
}

// SymmetricDifference returns a new set of the elements that are in exactly
// one of s and other.
func (s Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	d := Set[T]{set: make(map[T]struct{})}
	for e := range s.set {
		if !other.Contains(e) {
			d.set[e] = struct{}{}
		}
	}
	for e := range other.set {
		if !s.Contains(e) {
			d.set[e] = struct{}{}
		}
	}
	return d
}
