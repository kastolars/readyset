package lockorder

import (
	"sync"
	"testing"
	"time"
)

// within fails the test if fn has not returned after d, which is how a
// deadlock shows up.
func within(t *testing.T, d time.Duration, what string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("%s: still running after %v, probably deadlocked", what, d)
	}
}

func TestOrderIsStableAndUnique(t *testing.T) {
	var a, b RWMutex
	first := a.order()
	if again := a.order(); first == 0 || first != again {
		t.Errorf("order() = %d then %d, want the same non-zero number", first, again)
	}
	if a.order() == b.order() {
		t.Error("two mutexes share an order")
	}
}

// The same mutex passed twice is locked once. Locking it twice for reading
// deadlocks when a writer arrives in between.
func TestSameMutexIsLockedOnce(t *testing.T) {
	for _, write := range []bool{false, true} {
		var m RWMutex
		within(t, 5*time.Second, "locking a mutex with itself", func() {
			Lock(&m, &m, write)
			Unlock(&m, &m, write)
			if !m.TryLock() { // fails if Unlock left it held
				t.Error("the mutex is still held after Unlock")
			} else {
				m.Unlock()
			}
		})
	}
}

func TestLockModes(t *testing.T) {
	var a, b RWMutex

	Lock(&a, &b, true) // writes a, reads b
	if a.TryRLock() {
		t.Error("a is write-locked but a read lock succeeded")
	}
	if !b.TryRLock() {
		t.Error("b is only read-locked but a read lock failed")
	} else {
		b.RUnlock()
	}
	if b.TryLock() {
		t.Error("b is read-locked but a write lock succeeded")
	}
	Unlock(&a, &b, true)

	Lock(&a, &b, false) // reads both
	if !a.TryRLock() || !b.TryRLock() {
		t.Error("reading both should still allow other readers")
	} else {
		a.RUnlock()
		b.RUnlock()
	}
	Unlock(&a, &b, false)

	if !a.TryLock() || !b.TryLock() {
		t.Error("Unlock left a mutex held")
	}
}

// Writers locking the same pair from opposite directions is the classic
// deadlock; the fixed order prevents it.
func TestOppositeDirectionsDoNotDeadlock(t *testing.T) {
	var a, b RWMutex
	within(t, 30*time.Second, "crossed locking", func() {
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(2)
			go func() {
				defer wg.Done()
				for j := 0; j < 2000; j++ {
					Lock(&a, &b, true)
					Unlock(&a, &b, true)
				}
			}()
			go func() {
				defer wg.Done()
				for j := 0; j < 2000; j++ {
					Lock(&b, &a, true)
					Unlock(&b, &a, true)
				}
			}()
		}
		wg.Wait()
	})
}

// Locking a mutex with itself for reading takes one read lock, not two: a second
// RLock behind a waiting writer is the deadlock the same-mutex case exists to
// avoid, and Lock followed by Unlock cannot tell one lock from two.
func TestSameMutexReadLockedOnce(t *testing.T) {
	var m RWMutex
	Lock(&m, &m, false)
	m.RUnlock() // exactly one read lock should be held
	if !m.TryLock() {
		t.Error("Lock(m, m, false) took more than one read lock")
	} else {
		m.Unlock()
	}
}

// Many goroutines ask a mutex that has never been used for its order at once;
// they must all be given the same number, or two sets could be locked in
// different orders by different goroutines.
func TestOrderConcurrentFirstUse(t *testing.T) {
	for i := 0; i < 2000; i++ {
		var m RWMutex
		const n = 8
		got := make([]uint64, n)
		var start, wg sync.WaitGroup
		start.Add(1)
		for g := 0; g < n; g++ {
			wg.Add(1)
			go func() { defer wg.Done(); start.Wait(); got[g] = m.order() }()
		}
		start.Done()
		wg.Wait()
		for _, v := range got {
			if v == 0 || v != got[0] {
				t.Fatalf("iteration %d: goroutines were given different orders: %v", i, got)
			}
		}
	}
}
