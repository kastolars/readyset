package readyset

import (
	"iter"

	"github.com/kastolars/readyset/internal/lockorder"
)

// SyncSet is a [Set] that is safe for concurrent use by multiple goroutines.
//
// Use it when goroutines share a set and you want the locking done for you.
// A plain [Set] is faster and has no restrictions, so prefer it when only one
// goroutine touches the set at a time, or when you already hold a lock of your
// own around it.
//
// The zero value is an empty set that is ready to use. A SyncSet contains a
// lock, so it must be used through a pointer and must not be copied after first
// use; go vet reports copies. Use [NewSync] to create one with elements.
//
// Every method is atomic by itself, but a sequence of calls is not. This is a
// race, because another goroutine can add the element between the two calls:
//
//	if !s.Contains(x) { // wrong
//		s.Add(x)
//	}
//
// Use [SyncSet.Do] to run several steps under one lock.
//
// Operations that take another SyncSet lock both sets, always in the same
// order, so two goroutines running a.Union(b) and b.Union(a) cannot deadlock,
// and a.Union(a) is fine. Functions you pass in (to [SyncSet.Filter],
// [SyncSet.RemoveFunc] and [SyncSet.Do]) run while the lock is held, so they
// must not call methods on the same set.
type SyncSet[T comparable] struct {
	mu  lockorder.RWMutex
	set Set[T]
}

// NewSync returns a SyncSet containing elements. Duplicates are ignored.
func NewSync[T comparable](elements ...T) *SyncSet[T] {
	return &SyncSet[T]{set: New(elements...)}
}

// Add inserts elements into the set. Elements already present are ignored.
func (s *SyncSet[T]) Add(elements ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Add(elements...)
}

// Remove deletes elements from the set. Elements not present are ignored.
func (s *SyncSet[T]) Remove(elements ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Remove(elements...)
}

// AddSet inserts every element of other into the set. It is the in-place
// counterpart of [SyncSet.Union], and other is not modified. See [Set.AddSet].
func (s *SyncSet[T]) AddSet(other *SyncSet[T]) {
	lockorder.Lock(&s.mu, &other.mu, true)
	defer lockorder.Unlock(&s.mu, &other.mu, true)
	s.set.AddSet(other.set)
}

// RemoveSet deletes every element of other from the set. It is the in-place
// counterpart of [SyncSet.Difference], and other is not modified.
func (s *SyncSet[T]) RemoveSet(other *SyncSet[T]) {
	lockorder.Lock(&s.mu, &other.mu, true)
	defer lockorder.Unlock(&s.mu, &other.mu, true)
	s.set.RemoveSet(other.set)
}

// RetainSet deletes every element that is not also in other. It is the
// in-place counterpart of [SyncSet.Intersection], and other is not modified.
func (s *SyncSet[T]) RetainSet(other *SyncSet[T]) {
	lockorder.Lock(&s.mu, &other.mu, true)
	defer lockorder.Unlock(&s.mu, &other.mu, true)
	s.set.RetainSet(other.set)
}

// RemoveFunc deletes every element for which remove returns true and reports
// how many it deleted. remove runs while the set is locked, so it must not call
// methods on this set.
func (s *SyncSet[T]) RemoveFunc(remove func(T) bool) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.RemoveFunc(remove)
}

// Pop removes and returns an arbitrary element of the set. The boolean is false
// if the set was empty. See [Set.Pop].
func (s *SyncSet[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.Pop()
}

// Clear removes every element from the set.
func (s *SyncSet[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Clear()
}

// Do calls fn with the underlying [Set] while holding the set's write lock, and
// returns when fn returns. Use it for a sequence of steps that must not be
// interleaved with other goroutines, such as adding an element only if it is
// absent:
//
//	var added bool
//	s.Do(func(set *readyset.Set[string]) {
//		if !set.Contains(name) {
//			set.Add(name)
//			added = true
//		}
//	})
//
// fn must not call methods on s, which would deadlock, and must not keep the
// *Set after it returns.
func (s *SyncSet[T]) Do(fn func(*Set[T])) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.set)
}

// Contains reports whether element is in the set.
func (s *SyncSet[T]) Contains(element T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Contains(element)
}

// ContainsAll reports whether every one of elements is in the set. It is true
// when elements is empty.
func (s *SyncSet[T]) ContainsAll(elements ...T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.ContainsAll(elements...)
}

// ContainsAny reports whether at least one of elements is in the set. It is
// false when elements is empty.
func (s *SyncSet[T]) ContainsAny(elements ...T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.ContainsAny(elements...)
}

// Len returns the number of elements in the set.
func (s *SyncSet[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Len()
}

// IsEmpty reports whether the set has no elements.
func (s *SyncSet[T]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.IsEmpty()
}

// Equal reports whether s and other contain exactly the same elements.
func (s *SyncSet[T]) Equal(other *SyncSet[T]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.Equal(other.set)
}

// IsSubset reports whether every element of s is also in other.
func (s *SyncSet[T]) IsSubset(other *SyncSet[T]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsSubset(other.set)
}

// IsSuperset reports whether every element of other is also in s.
func (s *SyncSet[T]) IsSuperset(other *SyncSet[T]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsSuperset(other.set)
}

// IsProperSubset reports whether s is a subset of other and the two are not
// equal.
func (s *SyncSet[T]) IsProperSubset(other *SyncSet[T]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsProperSubset(other.set)
}

// IsProperSuperset reports whether s is a superset of other and the two are not
// equal.
func (s *SyncSet[T]) IsProperSuperset(other *SyncSet[T]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsProperSuperset(other.set)
}

// IsDisjoint reports whether s and other have no elements in common.
func (s *SyncSet[T]) IsDisjoint(other *SyncSet[T]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsDisjoint(other.set)
}

// Slice returns the elements of the set as a new slice, in no particular
// order. The order can differ between calls.
func (s *SyncSet[T]) Slice() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Slice()
}

// All returns an iterator over the elements of the set, in no particular
// order. The elements are copied when iteration starts, so the loop body may
// freely change this set (or any other goroutine may) without affecting or
// deadlocking the loop, and the loop does not see those changes.
func (s *SyncSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, e := range s.Slice() {
			if !yield(e) {
				return
			}
		}
	}
}

// Clone returns an independent copy of the set.
func (s *SyncSet[T]) Clone() *SyncSet[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &SyncSet[T]{set: s.set.Clone()}
}

// Snapshot returns the elements as a plain, independent [Set], for use with
// [Sorted] or in code that does not need locking. Later changes to either set
// do not affect the other.
func (s *SyncSet[T]) Snapshot() Set[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Clone()
}

// Filter returns a new set of the elements for which keep returns true. keep
// runs while the set is locked, so it must not call methods on this set.
func (s *SyncSet[T]) Filter(keep func(T) bool) *SyncSet[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &SyncSet[T]{set: s.set.Filter(keep)}
}

// Union returns a new set of the elements that are in s, in other, or in both.
func (s *SyncSet[T]) Union(other *SyncSet[T]) *SyncSet[T] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[T]{set: s.set.Union(other.set)}
}

// Intersection returns a new set of the elements that are in both s and other.
func (s *SyncSet[T]) Intersection(other *SyncSet[T]) *SyncSet[T] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[T]{set: s.set.Intersection(other.set)}
}

// Difference returns a new set of the elements that are in s but not in other.
func (s *SyncSet[T]) Difference(other *SyncSet[T]) *SyncSet[T] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[T]{set: s.set.Difference(other.set)}
}

// SymmetricDifference returns a new set of the elements that are in exactly
// one of s and other.
func (s *SyncSet[T]) SymmetricDifference(other *SyncSet[T]) *SyncSet[T] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[T]{set: s.set.SymmetricDifference(other.set)}
}

// String returns the elements in braces, in a stable order. See [Set.String].
func (s *SyncSet[T]) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.String()
}

// MarshalJSON encodes the set as a JSON array in a stable order. See
// [Set.MarshalJSON].
func (s *SyncSet[T]) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.MarshalJSON()
}

// UnmarshalJSON decodes a JSON array into the set, replacing its previous
// contents. See [Set.UnmarshalJSON].
func (s *SyncSet[T]) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.UnmarshalJSON(data)
}
