package readyset

// Set is an unordered collection of unique elements.
//
// The zero value is an empty set that is ready to use. Methods that modify a
// set have pointer receivers, so call them on an addressable variable. Copying
// a Set that holds or has held elements copies a reference to the same
// underlying storage; use [Set.Clone] for an independent copy.
//
// An element must be equal to itself to be in a set, so [Set.Add] ignores a
// NaN, or any value containing one: a set could never find or remove it again.
// If T is an interface type, adding or looking up an element whose dynamic type
// cannot be compared, such as a slice or a map, panics, as it does for a Go map
// key.
//
// A Set is not safe for concurrent use; use a [SyncSet] to share one between
// goroutines.
type Set[T comparable] struct {
	set map[T]struct{}
}

// New returns a Set containing elements. Duplicates are ignored, and so is any
// element that is not equal to itself, such as a NaN.
func New[T comparable](elements ...T) Set[T] {
	s := Set[T]{set: make(map[T]struct{}, len(elements))}
	s.Add(elements...)
	return s
}

// Add inserts elements into the set. Elements already present are ignored, and
// so is any element that is not equal to itself, such as a NaN.
func (s *Set[T]) Add(elements ...T) {
	if s.set == nil {
		s.set = make(map[T]struct{}, len(elements))
	}
	for _, e := range elements {
		// e != e only for a NaN or a value containing one.
		if e != e {
			continue
		}
		s.set[e] = struct{}{}
	}
}

// Remove deletes elements from the set. Elements not present are ignored.
func (s *Set[T]) Remove(elements ...T) {
	for _, e := range elements {
		delete(s.set, e)
	}
}

// AddSet inserts every element of other into the set. It is the in-place
// counterpart of [Set.Union]: afterwards s holds what s.Union(other) would have
// returned, and other is not modified.
func (s *Set[T]) AddSet(other Set[T]) {
	if s.set == nil {
		s.set = make(map[T]struct{}, len(other.set))
	}
	for e := range other.set {
		s.set[e] = struct{}{}
	}
}

// RemoveSet deletes every element of other from the set. It is the in-place
// counterpart of [Set.Difference]: afterwards s holds what s.Difference(other)
// would have returned, and other is not modified.
func (s *Set[T]) RemoveSet(other Set[T]) {
	for e := range other.set {
		delete(s.set, e)
	}
}

// RetainSet deletes every element that is not also in other. It is the
// in-place counterpart of [Set.Intersection]: afterwards s holds what
// s.Intersection(other) would have returned, and other is not modified.
func (s *Set[T]) RetainSet(other Set[T]) {
	for e := range s.set {
		if !other.Contains(e) {
			delete(s.set, e)
		}
	}
}

// RemoveFunc deletes every element for which remove returns true and reports
// how many it deleted. remove must not modify the set.
func (s *Set[T]) RemoveFunc(remove func(T) bool) int {
	n := 0
	for e := range s.set {
		if remove(e) {
			delete(s.set, e)
			n++
		}
	}
	return n
}

// Pop removes and returns an arbitrary element of the set. The boolean is false
// if the set was empty. Which element is returned is unspecified and can differ
// between calls, so use it to drain a worklist, not to pick a particular
// element.
func (s *Set[T]) Pop() (T, bool) {
	for e := range s.set {
		delete(s.set, e)
		return e, true
	}
	var zero T
	return zero, false
}

// Clear removes every element from the set.
func (s *Set[T]) Clear() {
	clear(s.set)
}

// Contains reports whether element is in the set.
func (s Set[T]) Contains(element T) bool {
	_, ok := s.set[element]
	return ok
}

// ContainsAll reports whether every one of elements is in the set. It is true
// when elements is empty.
func (s Set[T]) ContainsAll(elements ...T) bool {
	for _, e := range elements {
		if !s.Contains(e) {
			return false
		}
	}
	return true
}

// ContainsAny reports whether at least one of elements is in the set. It is
// false when elements is empty.
func (s Set[T]) ContainsAny(elements ...T) bool {
	for _, e := range elements {
		if s.Contains(e) {
			return true
		}
	}
	return false
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

// IsProperSubset reports whether s is a subset of other and the two are not
// equal, that is, other has at least one element that s lacks.
func (s Set[T]) IsProperSubset(other Set[T]) bool {
	return s.Len() < other.Len() && s.IsSubset(other)
}

// IsProperSuperset reports whether s is a superset of other and the two are not
// equal, that is, s has at least one element that other lacks.
func (s Set[T]) IsProperSuperset(other Set[T]) bool {
	return other.IsProperSubset(s)
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

// Filter returns a new set of the elements for which keep returns true.
func (s Set[T]) Filter(keep func(T) bool) Set[T] {
	f := Set[T]{set: make(map[T]struct{})}
	for e := range s.set {
		if keep(e) {
			f.set[e] = struct{}{}
		}
	}
	return f
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
	i := Set[T]{set: make(map[T]struct{}, len(small.set))}
	for e := range small.set {
		if large.Contains(e) {
			i.set[e] = struct{}{}
		}
	}
	return i
}

// Difference returns a new set of the elements that are in s but not in other.
func (s Set[T]) Difference(other Set[T]) Set[T] {
	d := Set[T]{set: make(map[T]struct{}, len(s.set))}
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
