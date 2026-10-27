package keyed

import (
	"encoding/json"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"testing/quick"
	"time"

	"github.com/kastolars/readyset/internal/setlaws"
)

// within fails the test if fn has not returned after d, which is how a
// deadlock shows up. It runs fn on its own goroutine.
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

func newSyncPeople(ps ...person) *SyncSet[int, person] { return NewSync(personID, ps...) }

// newSyncBytes builds a set whose key is the byte itself, like newBytes.
func newSyncBytes(elements ...byte) *SyncSet[byte, byte] {
	return NewSync(func(b byte) byte { return b }, elements...)
}

func TestSyncBasics(t *testing.T) {
	s := newSyncPeople(person{3, "c"}, person{1, "a"}, person{2, "b"})
	if s.Len() != 3 || s.IsEmpty() {
		t.Fatalf("Len() = %d, IsEmpty() = %v", s.Len(), s.IsEmpty())
	}

	// Adding an element with an existing key replaces it.
	s.Add(person{1, "new"}, person{4, "d"})
	if p, ok := s.Get(1); !ok || p.Name != "new" {
		t.Errorf("Get(1) = %v, %v; want the replacement", p, ok)
	}
	if _, ok := s.Get(99); ok {
		t.Error("Get(99) reported a missing key as present")
	}

	s.Remove(person{4, "ignored name"})
	s.RemoveKey(3, 99)
	if got, want := ids(s.Snapshot()), []int{1, 2}; !reflect.DeepEqual(got, want) {
		t.Errorf("after Remove and RemoveKey ids = %v, want %v", got, want)
	}

	if !s.Contains(person{ID: 2}) || s.Contains(person{ID: 3}) || !s.ContainsKey(1) || s.ContainsKey(3) {
		t.Error("Contains or ContainsKey is wrong")
	}
	if !s.ContainsAll(person{ID: 1}, person{ID: 2}) || s.ContainsAll(person{ID: 1}, person{ID: 9}) || !s.ContainsAll() {
		t.Error("ContainsAll is wrong")
	}
	if !s.ContainsAny(person{ID: 9}, person{ID: 2}) || s.ContainsAny(person{ID: 9}) || s.ContainsAny() {
		t.Error("ContainsAny is wrong")
	}

	keys := s.Keys()
	if len(keys) != 2 || len(s.Slice()) != 2 {
		t.Errorf("Keys() = %v, Slice() has %d elements; want 2 of each", keys, len(s.Slice()))
	}
	if got := s.Filter(func(p person) bool { return p.ID == 2 }); got.Len() != 1 || !got.ContainsKey(2) {
		t.Errorf("Filter kept %v", got)
	}
	if n := s.RemoveFunc(func(p person) bool { return p.ID == 1 }); n != 1 || s.Len() != 1 {
		t.Errorf("RemoveFunc = %d, Len() = %d; want 1 and 1", n, s.Len())
	}

	s.Add(person{5, "e"})
	popped := 0
	for {
		if _, ok := s.Pop(); !ok {
			break
		}
		popped++
	}
	if popped != 2 || !s.IsEmpty() {
		t.Errorf("Pop drained %d elements, IsEmpty() = %v; want 2 and true", popped, s.IsEmpty())
	}

	s.Add(person{1, "a"})
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Clear left elements behind")
	}
}

// A SyncSet not made with NewSync fails the way a plain Set does, and the lock
// is released, so the set stays usable.
func TestSyncZeroValue(t *testing.T) {
	within(t, 5*time.Second, "using the zero value", func() {
		var s SyncSet[int, person]
		if s.Len() != 0 || !s.IsEmpty() || s.ContainsKey(1) {
			t.Error("zero value is not empty")
		}
		mustPanic(t, "must be created with New", func() { s.Add(person{1, "a"}) })
		if s.Len() != 0 { // would hang if the panic left the write lock held
			t.Errorf("Len() = %d after the panic, want 0", s.Len())
		}
	})
	mustPanic(t, "nil key function", func() { NewSync[int, person](nil) })
}

func TestSyncCopiesAreIndependent(t *testing.T) {
	s := newSyncPeople(person{1, "a"})
	clone, snap := s.Clone(), s.Snapshot()
	s.Add(person{2, "b"})
	clone.Add(person{3, "c"})
	snap.Add(person{4, "d"})

	if got, want := ids(s.Snapshot()), []int{1, 2}; !reflect.DeepEqual(got, want) {
		t.Errorf("original ids = %v, want %v", got, want)
	}
	if got, want := ids(clone.Snapshot()), []int{1, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("Clone ids = %v, want %v", got, want)
	}
	if got, want := ids(snap), []int{1, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("Snapshot ids = %v, want %v", got, want)
	}
}

// Every SyncSet operation gives the same answer as the plain Set operation.
func TestSyncMatchesSet(t *testing.T) {
	err := quick.Check(func(x, y []byte) bool {
		sa, sb := newSyncBytes(x...), newSyncBytes(y...)
		a, b := newBytes(x...), newBytes(y...)

		same := sa.Union(sb).Snapshot().Equal(a.Union(b)) &&
			sa.Intersection(sb).Snapshot().Equal(a.Intersection(b)) &&
			sa.Difference(sb).Snapshot().Equal(a.Difference(b)) &&
			sa.SymmetricDifference(sb).Snapshot().Equal(a.SymmetricDifference(b)) &&
			sa.Equal(sb) == a.Equal(b) &&
			sa.IsSubset(sb) == a.IsSubset(b) &&
			sa.IsSuperset(sb) == a.IsSuperset(b) &&
			sa.IsProperSubset(sb) == a.IsProperSubset(b) &&
			sa.IsProperSuperset(sb) == a.IsProperSuperset(b) &&
			sa.IsDisjoint(sb) == a.IsDisjoint(b)

		added, removed, retained := sa.Clone(), sa.Clone(), sa.Clone()
		added.AddSet(sb)
		removed.RemoveSet(sb)
		retained.RetainSet(sb)

		return same &&
			added.Snapshot().Equal(a.Union(b)) &&
			removed.Snapshot().Equal(a.Difference(b)) &&
			retained.Snapshot().Equal(a.Intersection(b))
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSyncLaws(t *testing.T) {
	err := quick.Check(func(x, y, z []byte) bool {
		setlaws.Check(t, newSyncBytes, x, y, z)
		return true
	}, &quick.Config{MaxCount: 500})
	if err != nil {
		t.Fatal(err)
	}
}

// For a key in both sets, AddSet behaves like Add and the argument wins, and
// Union keeps the receiver's element, exactly as with a plain Set.
func TestSyncElementOwnership(t *testing.T) {
	s := newSyncPeople(person{1, "mine"})
	other := newSyncPeople(person{1, "theirs"})

	if p, _ := s.Union(other).Get(1); p.Name != "mine" {
		t.Errorf("Union kept %q, want the receiver's element", p.Name)
	}
	s.AddSet(other)
	if p, _ := s.Get(1); p.Name != "theirs" {
		t.Errorf("AddSet kept %q, want the argument's element", p.Name)
	}
}

func TestSyncWithItself(t *testing.T) {
	within(t, 5*time.Second, "operations with the same set", func() {
		s := newSyncPeople(person{1, "a"}, person{2, "b"})
		if !s.Union(s).Equal(s) || !s.Intersection(s).Equal(s) || !s.Difference(s).IsEmpty() {
			t.Error("a set combined with itself gave the wrong result")
		}
		s.AddSet(s)
		s.RetainSet(s)
		if s.Len() != 2 {
			t.Errorf("Len() = %d after AddSet and RetainSet with itself, want 2", s.Len())
		}
		s.RemoveSet(s)
		if !s.IsEmpty() {
			t.Errorf("RemoveSet with itself left %d elements", s.Len())
		}
	})
}

func TestSyncDoIsAtomic(t *testing.T) {
	s := newSyncPeople()
	var added atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s.Do(func(set Set[int, person]) {
				if !set.ContainsKey(id) {
					set.Add(person{id, "x"})
					added.Add(1)
				}
			})
		}(i % 10)
	}
	wg.Wait()

	if added.Load() != 10 || s.Len() != 10 {
		t.Errorf("added %d keys, Len() = %d; want exactly 10 of each", added.Load(), s.Len())
	}
}

func TestSyncAllIteratesOverACopy(t *testing.T) {
	within(t, 5*time.Second, "changing a set inside its own loop", func() {
		s := newSyncPeople(person{1, "a"}, person{2, "b"}, person{3, "c"})
		seen := 0
		for p := range s.All() {
			seen++
			s.Add(person{p.ID + 10, "new"}) // would deadlock if the loop held the lock
			s.RemoveKey(p.ID)
		}
		if seen != 3 {
			t.Errorf("the loop saw %d elements, want the 3 that were there when it started", seen)
		}
		if got, want := ids(s.Snapshot()), []int{11, 12, 13}; !reflect.DeepEqual(got, want) {
			t.Errorf("after the loop ids = %v, want %v", got, want)
		}

		calls := 0
		for range s.All() {
			calls++
			break
		}
		if calls != 1 {
			t.Errorf("loop body ran %d times after break, want 1", calls)
		}
	})
}

// A function passed in that panics must not leave the set locked.
func TestSyncPanicReleasesTheLock(t *testing.T) {
	within(t, 5*time.Second, "using a set after a panic", func() {
		s := newSyncPeople(person{1, "a"}, person{2, "b"})
		for name, call := range map[string]func(){
			"RemoveFunc": func() { s.RemoveFunc(func(person) bool { panic("boom") }) },
			"Filter":     func() { s.Filter(func(person) bool { panic("boom") }) },
			"Do":         func() { s.Do(func(Set[int, person]) { panic("boom") }) },
		} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("%s: expected a panic", name)
					}
				}()
				call()
			}()
			s.Add(person{3, "c"}) // takes the write lock; hangs if the panic leaked it
		}
	})
}

func TestSyncJSONAndString(t *testing.T) {
	s := newSyncPeople(person{2, "b"}, person{1, "a"})
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(s.Snapshot())
	if string(b) != string(want) {
		t.Errorf("Marshal = %s, want %s", b, want)
	}
	if got, want := s.String(), s.Snapshot().String(); got != want {
		t.Errorf("String = %q, want %q", got, want)
	}

	back := newSyncPeople()
	if err := json.Unmarshal(b, back); err != nil {
		t.Fatal(err)
	}
	if !back.Equal(s) {
		t.Errorf("round trip = %v, want %v", back, s)
	}

	var uncreated SyncSet[int, person]
	if err := json.Unmarshal(b, &uncreated); err == nil {
		t.Error("decoding into a set without a key function succeeded, want an error")
	}
	if uncreated.Len() != 0 { // returns, so the lock was released
		t.Errorf("Len() = %d after a failed decode, want 0", uncreated.Len())
	}
}

// See the readyset test of the same name: crossed operations and a set with
// itself are the shapes that deadlock a naive implementation. Run with -race.
func TestSyncConcurrentUse(t *testing.T) {
	a, b := newSyncBytes(1, 2, 3), newSyncBytes(3, 4, 5)

	within(t, 60*time.Second, "concurrent operations", func() {
		var wg sync.WaitGroup
		for g := 0; g < 40; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := 0; i < 150; i++ {
					n := byte(i % 50)
					switch (g + i) % 9 {
					case 0:
						a.Union(b)
					case 1:
						b.Union(a)
					case 2:
						a.AddSet(b)
					case 3:
						b.AddSet(a)
					case 4:
						a.Union(a)
						a.AddSet(a)
						a.Equal(a)
					case 5:
						a.Add(n)
						b.Remove(n)
					case 6:
						a.Contains(n)
						a.Slice()
						b.IsSubset(a)
						a.IsDisjoint(b)
					case 7:
						for e := range a.All() {
							b.Add(e)
						}
					default:
						a.RetainSet(b)
						b.RemoveSet(a)
						a.Add(1, 2, 3)
						b.Add(3, 4, 5)
					}
				}
			}(g)
		}
		wg.Wait()
	})
}
