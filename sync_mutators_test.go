package readyset

import (
	"sync"
	"testing"
	"time"
)

// Each method that changes the set runs on several goroutines alongside readers
// of the same set. A method that holds only a read lock while it writes is a
// data race, which -race reports; none of the other tests exercises Pop, Clear,
// RemoveFunc or UnmarshalJSON concurrently.
func TestSyncEveryWriterExcludesReaders(t *testing.T) {
	other := NewSync(2, 3)
	writers := map[string]func(s *SyncSet[int]){
		"Add":           func(s *SyncSet[int]) { s.Add(1, 2, 3, 4) },
		"Remove":        func(s *SyncSet[int]) { s.Add(1, 2, 3, 4); s.Remove(1, 2) },
		"AddSet":        func(s *SyncSet[int]) { s.AddSet(NewSync(1, 2, 3, 4)) },
		"RemoveSet":     func(s *SyncSet[int]) { s.Add(1, 2, 3, 4); s.RemoveSet(other) },
		"RetainSet":     func(s *SyncSet[int]) { s.Add(1, 2, 3, 4); s.RetainSet(other) },
		"RemoveFunc":    func(s *SyncSet[int]) { s.Add(1, 2, 3, 4); s.RemoveFunc(func(e int) bool { return e > 1 }) },
		"Pop":           func(s *SyncSet[int]) { s.Add(1, 2, 3, 4); s.Pop() },
		"Clear":         func(s *SyncSet[int]) { s.Add(1, 2, 3, 4); s.Clear() },
		"UnmarshalJSON": func(s *SyncSet[int]) { _ = s.UnmarshalJSON([]byte("[1,2,3,4]")) },
		"Do":            func(s *SyncSet[int]) { s.Do(func(set *Set[int]) { set.Add(1, 2, 3, 4); set.Remove(1) }) },
	}
	for name, write := range writers {
		t.Run(name, func(t *testing.T) {
			s := NewSync(1, 2, 3, 4)
			within(t, 30*time.Second, name+" with readers", func() {
				var wg sync.WaitGroup
				for g := 0; g < 4; g++ {
					wg.Add(2)
					go func() {
						defer wg.Done()
						for i := 0; i < 300; i++ {
							write(s)
						}
					}()
					go func() {
						defer wg.Done()
						for i := 0; i < 300; i++ {
							s.Contains(i)
							s.Len()
							s.Slice()
							s.Equal(other)
							s.Snapshot()
						}
					}()
				}
				wg.Wait()
			})
		})
	}
}
