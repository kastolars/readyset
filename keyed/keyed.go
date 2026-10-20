// Package keyed provides a generic Set for elements that are identified by a
// key rather than by their own value.
//
// Use it when the element type is not comparable (it contains slices or maps),
// or when only part of an element decides whether two elements are the same,
// such as a record's ID. A key function maps each element to a comparable key,
// and two elements are the same element exactly when their keys are equal:
//
//	byEmail := keyed.New(func(u User) string { return u.Email })
//
// Because key equality is element equality, adding an element whose key is
// already present replaces the element stored under that key.
//
// If your elements are already comparable and you want whole-value equality,
// use the [github.com/kastolars/readyset] package instead.
package keyed

// Set is an unordered collection of elements with unique keys.
//
// A Set must be created with [New]; the zero value cannot add, remove or look
// up elements. Copying a Set copies a reference to the same underlying storage;
// use [Set.Clone] for an independent copy.
//
// Operations between two sets compare keys only, so both sets must use the
// same key function. A set produced by an operation uses the key function of
// the receiver.
//
// A Set is not safe for concurrent use.
type Set[K comparable, V any] struct {
	key   func(V) K
	items map[K]V
}

// New returns a Set that identifies elements with key, containing elements.
// If several elements share a key, the last one wins. New panics if key is nil.
func New[K comparable, V any](key func(V) K, elements ...V) Set[K, V] {
	if key == nil {
		panic("keyed: New called with a nil key function")
	}
	s := Set[K, V]{key: key, items: make(map[K]V, len(elements))}
	s.Add(elements...)
	return s
}

// like returns an empty set with the same key function as s.
func (s Set[K, V]) like(size int) Set[K, V] {
	return Set[K, V]{key: s.key, items: make(map[K]V, size)}
}

// keyOf returns the key of e, with a clear failure for a set not made by New.
func (s Set[K, V]) keyOf(e V) K {
	if s.key == nil {
		panic("keyed: Set must be created with New")
	}
	return s.key(e)
}

// Add inserts elements into the set. An element whose key is already present
// replaces the element stored under that key.
func (s Set[K, V]) Add(elements ...V) {
	for _, e := range elements {
		s.items[s.keyOf(e)] = e
	}
}

// Remove deletes the elements stored under the keys of elements.
// Keys not present are ignored.
func (s Set[K, V]) Remove(elements ...V) {
	for _, e := range elements {
		delete(s.items, s.keyOf(e))
	}
}

// RemoveKey deletes the elements stored under keys. Keys not present are
// ignored. Unlike [Set.Remove] it needs no element, so removing an element you
// only know the key of does not require building one.
func (s Set[K, V]) RemoveKey(keys ...K) {
	for _, k := range keys {
		delete(s.items, k)
	}
}

// Clear removes every element from the set.
func (s Set[K, V]) Clear() {
	for k := range s.items {
		delete(s.items, k)
	}
}

// Contains reports whether the set holds an element with the same key as e.
func (s Set[K, V]) Contains(e V) bool {
	_, ok := s.items[s.keyOf(e)]
	return ok
}

// ContainsKey reports whether the set holds an element stored under key.
func (s Set[K, V]) ContainsKey(key K) bool {
	_, ok := s.items[key]
	return ok
}

// Get returns the element stored under key. The boolean reports whether key
// was present; if it was not, the element is the zero value of V.
func (s Set[K, V]) Get(key K) (V, bool) {
	e, ok := s.items[key]
	return e, ok
}

// Keys returns the keys of the elements in the set as a new slice, in no
// particular order. The order can differ between calls.
func (s Set[K, V]) Keys() []K {
	keys := make([]K, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}
	return keys
}

// Len returns the number of elements in the set.
func (s Set[K, V]) Len() int {
	return len(s.items)
}

// IsEmpty reports whether the set has no elements.
func (s Set[K, V]) IsEmpty() bool {
	return len(s.items) == 0
}

// Equal reports whether s and other contain the same keys.
func (s Set[K, V]) Equal(other Set[K, V]) bool {
	return s.Len() == other.Len() && s.IsSubset(other)
}

// IsSubset reports whether every key of s is also in other.
func (s Set[K, V]) IsSubset(other Set[K, V]) bool {
	if s.Len() > other.Len() {
		return false
	}
	for k := range s.items {
		if _, ok := other.items[k]; !ok {
			return false
		}
	}
	return true
}

// IsSuperset reports whether every key of other is also in s.
func (s Set[K, V]) IsSuperset(other Set[K, V]) bool {
	return other.IsSubset(s)
}

// IsDisjoint reports whether s and other have no keys in common.
func (s Set[K, V]) IsDisjoint(other Set[K, V]) bool {
	small, large := s, other
	if small.Len() > large.Len() {
		small, large = large, small
	}
	for k := range small.items {
		if _, ok := large.items[k]; ok {
			return false
		}
	}
	return true
}

// Slice returns the elements of the set as a new slice, in no particular
// order. The order can differ between calls.
func (s Set[K, V]) Slice() []V {
	elements := make([]V, 0, len(s.items))
	for _, e := range s.items {
		elements = append(elements, e)
	}
	return elements
}

// Clone returns an independent copy of the set.
func (s Set[K, V]) Clone() Set[K, V] {
	c := s.like(len(s.items))
	for k, e := range s.items {
		c.items[k] = e
	}
	return c
}

// Union returns a new set of the keys that are in s, in other, or in both.
// For a key in both, the element from s is kept.
func (s Set[K, V]) Union(other Set[K, V]) Set[K, V] {
	u := s.like(len(s.items) + len(other.items))
	for k, e := range other.items {
		u.items[k] = e
	}
	for k, e := range s.items {
		u.items[k] = e
	}
	return u
}

// Intersection returns a new set of the keys that are in both s and other.
// The elements come from s.
func (s Set[K, V]) Intersection(other Set[K, V]) Set[K, V] {
	i := s.like(0)
	if len(s.items) <= len(other.items) {
		for k, e := range s.items {
			if _, ok := other.items[k]; ok {
				i.items[k] = e
			}
		}
	} else {
		// Walk the smaller set, but still take the element from s.
		for k := range other.items {
			if e, ok := s.items[k]; ok {
				i.items[k] = e
			}
		}
	}
	return i
}

// Difference returns a new set of the keys that are in s but not in other.
func (s Set[K, V]) Difference(other Set[K, V]) Set[K, V] {
	d := s.like(0)
	for k, e := range s.items {
		if _, ok := other.items[k]; !ok {
			d.items[k] = e
		}
	}
	return d
}

// SymmetricDifference returns a new set of the keys that are in exactly one of
// s and other.
func (s Set[K, V]) SymmetricDifference(other Set[K, V]) Set[K, V] {
	d := s.like(0)
	for k, e := range s.items {
		if _, ok := other.items[k]; !ok {
			d.items[k] = e
		}
	}
	for k, e := range other.items {
		if _, ok := s.items[k]; !ok {
			d.items[k] = e
		}
	}
	return d
}
