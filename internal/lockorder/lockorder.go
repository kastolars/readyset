// Package lockorder locks two read-write mutexes together without risking a
// deadlock. It exists so that every concurrent set type in this module shares
// one implementation of that delicate step.
package lockorder

import (
	"sync"
	"sync/atomic"
)

// ids hands out the numbers that decide which of two mutexes is locked first.
var ids atomic.Uint64

// RWMutex is a [sync.RWMutex] that also knows its place in the global locking
// order. The zero value is an unlocked mutex. Like sync.RWMutex it must not be
// copied after first use.
type RWMutex struct {
	sync.RWMutex
	id atomic.Uint64 // assigned on first use; never zero afterwards
}

// order returns this mutex's position in the global locking order.
func (m *RWMutex) order() uint64 {
	if id := m.id.Load(); id != 0 {
		return id
	}
	m.id.CompareAndSwap(0, ids.Add(1))
	return m.id.Load()
}

func (m *RWMutex) lock(write bool) {
	if write {
		m.Lock()
	} else {
		m.RLock()
	}
}

func (m *RWMutex) unlock(write bool) {
	if write {
		m.Unlock()
	} else {
		m.RUnlock()
	}
}

// Lock locks m and other for an operation that reads both of them or, if write
// is true, modifies m and reads other. It always locks the two in the same
// order, whichever way round they are passed, so concurrent calls on the same
// pair cannot deadlock. If m and other are the same mutex it is locked once,
// because taking a read lock twice can deadlock behind a waiting writer.
//
// Pair every call with [Unlock], passing the same arguments.
func Lock(m, other *RWMutex, write bool) {
	if m == other {
		m.lock(write)
		return
	}
	first, second := m, other
	if first.order() > second.order() {
		first, second = second, first
	}
	first.lock(write && first == m)
	second.lock(write && second == m)
}

// Unlock undoes [Lock].
func Unlock(m, other *RWMutex, write bool) {
	if m == other {
		m.unlock(write)
		return
	}
	m.unlock(write)
	other.unlock(false)
}
