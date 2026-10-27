package keyed

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
// A SyncSet must be created with [NewSync]; like a [Set], the zero value cannot
// add, remove or look up elements. It contains a lock, so it must be used
// through a pointer and must not be copied after first use; go vet reports
// copies.
//
// Every method is atomic by itself, but a sequence of calls is not. This is a
// race, because another goroutine can add the key between the two calls:
//
//	if !s.ContainsKey(k) { // wrong
//		s.Add(v)
//	}
//
// Use [SyncSet.Do] to run several steps under one lock.
//
// Operations that take another SyncSet lock both sets, always in the same
// order, so two goroutines running a.Union(b) and b.Union(a) cannot deadlock,
// and a.Union(a) is fine. Functions you pass in (to [SyncSet.Filter],
// [SyncSet.RemoveFunc] and [SyncSet.Do]) run while the lock is held, so they
// must not call methods on the same set.
type SyncSet[K comparable, V any] struct {
	mu  lockorder.RWMutex
	set Set[K, V]
}

// NewSync returns a SyncSet that identifies elements with key, containing
// elements. If several elements share a key, the last one wins. NewSync panics
// if key is nil.
func NewSync[K comparable, V any](key func(V) K, elements ...V) *SyncSet[K, V] {
	return &SyncSet[K, V]{set: New(key, elements...)}
}

// Add inserts elements into the set. An element whose key is already present
// replaces the element stored under that key.
func (s *SyncSet[K, V]) Add(elements ...V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Add(elements...)
}

// Remove deletes the elements stored under the keys of elements.
// Keys not present are ignored.
func (s *SyncSet[K, V]) Remove(elements ...V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Remove(elements...)
}

// RemoveKey deletes the elements stored under keys. Keys not present are
// ignored.
func (s *SyncSet[K, V]) RemoveKey(keys ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.RemoveKey(keys...)
}

// AddSet inserts every element of other into the set. For a key present in
// both sets the element from other replaces the one in s. See [Set.AddSet].
func (s *SyncSet[K, V]) AddSet(other *SyncSet[K, V]) {
	lockorder.Lock(&s.mu, &other.mu, true)
	defer lockorder.Unlock(&s.mu, &other.mu, true)
	s.set.AddSet(other.set)
}

// RemoveSet deletes the elements stored under the keys of other. It is the
// in-place counterpart of [SyncSet.Difference], and other is not modified.
func (s *SyncSet[K, V]) RemoveSet(other *SyncSet[K, V]) {
	lockorder.Lock(&s.mu, &other.mu, true)
	defer lockorder.Unlock(&s.mu, &other.mu, true)
	s.set.RemoveSet(other.set)
}

// RetainSet deletes the elements whose keys are not also in other. It is the
// in-place counterpart of [SyncSet.Intersection], and other is not modified.
func (s *SyncSet[K, V]) RetainSet(other *SyncSet[K, V]) {
	lockorder.Lock(&s.mu, &other.mu, true)
	defer lockorder.Unlock(&s.mu, &other.mu, true)
	s.set.RetainSet(other.set)
}

// RemoveFunc deletes every element for which remove returns true and reports
// how many it deleted. remove runs while the set is locked, so it must not call
// methods on this set.
func (s *SyncSet[K, V]) RemoveFunc(remove func(V) bool) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.RemoveFunc(remove)
}

// Pop removes and returns an arbitrary element of the set. The boolean is false
// if the set was empty. See [Set.Pop].
func (s *SyncSet[K, V]) Pop() (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.Pop()
}

// Clear removes every element from the set.
func (s *SyncSet[K, V]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Clear()
}

// Do calls fn with the underlying [Set] while holding the set's write lock, and
// returns when fn returns. Use it for a sequence of steps that must not be
// interleaved with other goroutines, such as adding an element only if its key
// is absent:
//
//	var added bool
//	users.Do(func(set keyed.Set[string, User]) {
//		if !set.ContainsKey(u.Email) {
//			set.Add(u)
//			added = true
//		}
//	})
//
// fn must not call methods on s, which would deadlock, and must not keep the
// Set after it returns.
func (s *SyncSet[K, V]) Do(fn func(Set[K, V])) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.set)
}

// Contains reports whether the set holds an element with the same key as e.
func (s *SyncSet[K, V]) Contains(e V) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Contains(e)
}

// ContainsAll reports whether the set holds an element with the same key as
// each of elements. It is true when elements is empty.
func (s *SyncSet[K, V]) ContainsAll(elements ...V) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.ContainsAll(elements...)
}

// ContainsAny reports whether the set holds an element with the same key as at
// least one of elements. It is false when elements is empty.
func (s *SyncSet[K, V]) ContainsAny(elements ...V) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.ContainsAny(elements...)
}

// ContainsKey reports whether the set holds an element stored under key.
func (s *SyncSet[K, V]) ContainsKey(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.ContainsKey(key)
}

// Get returns the element stored under key. The boolean reports whether key
// was present; if it was not, the element is the zero value of V.
func (s *SyncSet[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Get(key)
}

// Keys returns the keys of the elements in the set as a new slice, in no
// particular order.
func (s *SyncSet[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Keys()
}

// Len returns the number of elements in the set.
func (s *SyncSet[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Len()
}

// IsEmpty reports whether the set has no elements.
func (s *SyncSet[K, V]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.IsEmpty()
}

// Equal reports whether s and other contain the same keys.
func (s *SyncSet[K, V]) Equal(other *SyncSet[K, V]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.Equal(other.set)
}

// IsSubset reports whether every key of s is also in other.
func (s *SyncSet[K, V]) IsSubset(other *SyncSet[K, V]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsSubset(other.set)
}

// IsSuperset reports whether every key of other is also in s.
func (s *SyncSet[K, V]) IsSuperset(other *SyncSet[K, V]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsSuperset(other.set)
}

// IsProperSubset reports whether s is a subset of other and the two are not
// equal.
func (s *SyncSet[K, V]) IsProperSubset(other *SyncSet[K, V]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsProperSubset(other.set)
}

// IsProperSuperset reports whether s is a superset of other and the two are not
// equal.
func (s *SyncSet[K, V]) IsProperSuperset(other *SyncSet[K, V]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsProperSuperset(other.set)
}

// IsDisjoint reports whether s and other have no keys in common.
func (s *SyncSet[K, V]) IsDisjoint(other *SyncSet[K, V]) bool {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return s.set.IsDisjoint(other.set)
}

// Slice returns the elements of the set as a new slice, in no particular
// order. The order can differ between calls.
func (s *SyncSet[K, V]) Slice() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Slice()
}

// All returns an iterator over the elements of the set, in no particular
// order. The elements are copied when iteration starts, so the loop body may
// freely change this set (or any other goroutine may) without affecting or
// deadlocking the loop, and the loop does not see those changes.
func (s *SyncSet[K, V]) All() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, e := range s.Slice() {
			if !yield(e) {
				return
			}
		}
	}
}

// Clone returns an independent copy of the set.
func (s *SyncSet[K, V]) Clone() *SyncSet[K, V] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &SyncSet[K, V]{set: s.set.Clone()}
}

// Snapshot returns the elements as a plain, independent [Set], for use with
// [Sorted] or in code that does not need locking. Later changes to either set
// do not affect the other.
func (s *SyncSet[K, V]) Snapshot() Set[K, V] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Clone()
}

// Filter returns a new set of the elements for which keep returns true. It uses
// the key function of s. keep runs while the set is locked, so it must not call
// methods on this set.
func (s *SyncSet[K, V]) Filter(keep func(V) bool) *SyncSet[K, V] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &SyncSet[K, V]{set: s.set.Filter(keep)}
}

// Union returns a new set of the keys that are in s, in other, or in both.
// For a key in both, the element from s is kept.
func (s *SyncSet[K, V]) Union(other *SyncSet[K, V]) *SyncSet[K, V] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[K, V]{set: s.set.Union(other.set)}
}

// Intersection returns a new set of the keys that are in both s and other.
// The elements come from s.
func (s *SyncSet[K, V]) Intersection(other *SyncSet[K, V]) *SyncSet[K, V] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[K, V]{set: s.set.Intersection(other.set)}
}

// Difference returns a new set of the keys that are in s but not in other.
func (s *SyncSet[K, V]) Difference(other *SyncSet[K, V]) *SyncSet[K, V] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[K, V]{set: s.set.Difference(other.set)}
}

// SymmetricDifference returns a new set of the keys that are in exactly one of
// s and other.
func (s *SyncSet[K, V]) SymmetricDifference(other *SyncSet[K, V]) *SyncSet[K, V] {
	lockorder.Lock(&s.mu, &other.mu, false)
	defer lockorder.Unlock(&s.mu, &other.mu, false)
	return &SyncSet[K, V]{set: s.set.SymmetricDifference(other.set)}
}

// String returns the elements in braces, ordered by key. See [Set.String].
func (s *SyncSet[K, V]) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.String()
}

// MarshalJSON encodes the set as a JSON array ordered by key. See
// [Set.MarshalJSON].
func (s *SyncSet[K, V]) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.MarshalJSON()
}

// UnmarshalJSON decodes a JSON array into the set. The set must have been
// created with [NewSync], because it needs its key function. See
// [Set.UnmarshalJSON].
func (s *SyncSet[K, V]) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set.UnmarshalJSON(data)
}
