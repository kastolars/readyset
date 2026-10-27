package readyset

import (
	"encoding/json"
	"reflect"
	"slices"
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

func TestSyncBasics(t *testing.T) {
	s := NewSync(3, 1, 2, 2)
	if s.Len() != 3 || s.IsEmpty() {
		t.Fatalf("Len() = %d, IsEmpty() = %v", s.Len(), s.IsEmpty())
	}

	s.Add(4, 5)
	s.Remove(1, 99)
	if got, want := sorted(s.Snapshot()), []int{2, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Errorf("after Add and Remove = %v, want %v", got, want)
	}

	if !s.Contains(2) || s.Contains(1) {
		t.Error("Contains is wrong")
	}
	if !s.ContainsAll(2, 3) || s.ContainsAll(2, 9) || !s.ContainsAll() {
		t.Error("ContainsAll is wrong")
	}
	if !s.ContainsAny(9, 3) || s.ContainsAny(8, 9) || s.ContainsAny() {
		t.Error("ContainsAny is wrong")
	}

	if n := s.RemoveFunc(func(e int) bool { return e > 3 }); n != 2 {
		t.Errorf("RemoveFunc = %d, want 2", n)
	}
	if got, want := sorted(s.Filter(func(e int) bool { return e == 2 }).Snapshot()), []int{2}; !reflect.DeepEqual(got, want) {
		t.Errorf("Filter = %v, want %v", got, want)
	}

	slice := s.Slice()
	slices.Sort(slice)
	if want := []int{2, 3}; !reflect.DeepEqual(slice, want) {
		t.Errorf("Slice = %v, want %v", slice, want)
	}
	if got, want := s.String(), "{2, 3}"; got != want {
		t.Errorf("String = %q, want %q", got, want)
	}

	popped := NewSync[int]()
	for {
		e, ok := s.Pop()
		if !ok {
			break
		}
		popped.Add(e)
	}
	if !s.IsEmpty() || !popped.Equal(NewSync(2, 3)) {
		t.Errorf("draining with Pop left %v, popped %v", s, popped)
	}

	s.Add(1)
	s.Clear()
	if !s.IsEmpty() {
		t.Errorf("Clear left %v", s)
	}
}

func TestSyncZeroValue(t *testing.T) {
	var s SyncSet[string]
	if s.Len() != 0 || s.Contains("a") || !s.IsEmpty() {
		t.Fatal("zero value is not empty")
	}
	s.Add("a")
	if !s.Contains("a") {
		t.Error("Add on the zero value did not take effect")
	}

	var other SyncSet[string]
	s.AddSet(&other) // must not panic
	other.AddSet(&s)
	if !other.Equal(&s) {
		t.Errorf("AddSet into the zero value = %v, want %v", &other, &s)
	}
}

// Clone and Snapshot are independent of the set they came from.
func TestSyncCopiesAreIndependent(t *testing.T) {
	s := NewSync(1, 2)
	clone, snap := s.Clone(), s.Snapshot()
	s.Add(3)
	clone.Add(4)
	snap.Add(5)

	if got, want := sorted(s.Snapshot()), []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("original = %v, want %v", got, want)
	}
	if got, want := sorted(clone.Snapshot()), []int{1, 2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("Clone = %v, want %v", got, want)
	}
	if got, want := sorted(snap), []int{1, 2, 5}; !reflect.DeepEqual(got, want) {
		t.Errorf("Snapshot = %v, want %v", got, want)
	}
}

// Every SyncSet operation gives the same answer as the plain Set operation.
func TestSyncMatchesSet(t *testing.T) {
	err := quick.Check(func(x, y []byte) bool {
		sa, sb := NewSync(x...), NewSync(y...)
		a, b := New(x...), New(y...)

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
		setlaws.Check(t, NewSync[byte], x, y, z)
		return true
	}, &quick.Config{MaxCount: 500})
	if err != nil {
		t.Fatal(err)
	}
}

// Combining a set with itself locks it once; locking it twice for reading could
// deadlock behind a waiting writer.
func TestSyncWithItself(t *testing.T) {
	within(t, 5*time.Second, "operations with the same set", func() {
		s := NewSync(1, 2, 3)
		if !s.Union(s).Equal(s) || !s.Intersection(s).Equal(s) || !s.Difference(s).IsEmpty() {
			t.Error("a set combined with itself gave the wrong result")
		}
		if !s.IsSubset(s) || s.IsProperSubset(s) || !s.Equal(s) {
			t.Error("a set compared with itself gave the wrong answer")
		}
		s.AddSet(s)
		s.RetainSet(s)
		if got, want := sorted(s.Snapshot()), []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
			t.Errorf("after AddSet and RetainSet with itself = %v, want %v", got, want)
		}
		s.RemoveSet(s)
		if !s.IsEmpty() {
			t.Errorf("RemoveSet with itself left %v", s)
		}
	})
}

func TestSyncDoIsAtomic(t *testing.T) {
	s := NewSync[int]()
	var added atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			s.Do(func(set *Set[int]) {
				if !set.Contains(v) {
					set.Add(v)
					added.Add(1)
				}
			})
		}(i % 10)
	}
	wg.Wait()

	if added.Load() != 10 || s.Len() != 10 {
		t.Errorf("added %d values, Len() = %d; want exactly 10 of each", added.Load(), s.Len())
	}
}

func TestSyncAllIteratesOverACopy(t *testing.T) {
	within(t, 5*time.Second, "changing a set inside its own loop", func() {
		s := NewSync(1, 2, 3)
		seen := 0
		for e := range s.All() {
			seen++
			s.Add(e + 10) // would deadlock if the loop held the lock
			s.Remove(e)
		}
		if seen != 3 {
			t.Errorf("the loop saw %d elements, want the 3 that were there when it started", seen)
		}
		if got, want := sorted(s.Snapshot()), []int{11, 12, 13}; !reflect.DeepEqual(got, want) {
			t.Errorf("after the loop = %v, want %v", got, want)
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
		s := NewSync(1, 2, 3)
		for name, call := range map[string]func(){
			"RemoveFunc": func() { s.RemoveFunc(func(int) bool { panic("boom") }) },
			"Filter":     func() { s.Filter(func(int) bool { panic("boom") }) },
			"Do":         func() { s.Do(func(*Set[int]) { panic("boom") }) },
		} {
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("%s: expected a panic", name)
					}
				}()
				call()
			}()
			s.Add(4) // takes the write lock; hangs if the panic leaked it
			if s.Len() < 3 {
				t.Errorf("%s: Len() = %d", name, s.Len())
			}
		}
	})
}

func TestSyncJSONAndString(t *testing.T) {
	s := NewSync("b", "a")
	b, err := json.Marshal(s)
	if err != nil || string(b) != `["a","b"]` {
		t.Fatalf("Marshal = %s, %v; want [\"a\",\"b\"]", b, err)
	}

	var back SyncSet[string]
	if err := json.Unmarshal([]byte(`["z","y","y"]`), &back); err != nil {
		t.Fatal(err)
	}
	if got, want := back.String(), "{y, z}"; got != want {
		t.Errorf("Unmarshal gave %s, want %s", got, want)
	}

	if err := json.Unmarshal([]byte(`{"not":"an array"}`), &back); err == nil {
		t.Error("decoding an object succeeded, want an error")
	}
	if back.Len() != 2 { // a failed decode leaves the set as it was, and unlocked
		t.Errorf("after a failed decode Len() = %d, want 2", back.Len())
	}

	var holder struct{ Tags *SyncSet[string] }
	if err := json.Unmarshal([]byte(`{"Tags":["x"]}`), &holder); err != nil || !holder.Tags.Contains("x") {
		t.Errorf("as a struct field: %v, %v", holder.Tags, err)
	}
}

// Many goroutines use two sets in every combination, including crossed
// operations (a with b while b with a) and a set with itself, which are the
// shapes that deadlock a naive implementation. Run with -race.
func TestSyncConcurrentUse(t *testing.T) {
	a, b := NewSync(1, 2, 3), NewSync(3, 4, 5)

	within(t, 60*time.Second, "concurrent operations", func() {
		var wg sync.WaitGroup
		for g := 0; g < 40; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := 0; i < 150; i++ {
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
						a.Add(i % 50)
						b.Remove(i % 50)
					case 6:
						a.Contains(i)
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
