package keyed

import (
	"sync"
	"testing"
	"time"
)

// See the readyset test of the same name.
func TestSyncEveryWriterExcludesReaders(t *testing.T) {
	other := newSyncBytes(2, 3)
	writers := map[string]func(s *SyncSet[byte, byte]){
		"Add":           func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4) },
		"Remove":        func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4); s.Remove(1, 2) },
		"RemoveKey":     func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4); s.RemoveKey(1, 2) },
		"AddSet":        func(s *SyncSet[byte, byte]) { s.AddSet(newSyncBytes(1, 2, 3, 4)) },
		"RemoveSet":     func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4); s.RemoveSet(other) },
		"RetainSet":     func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4); s.RetainSet(other) },
		"RemoveFunc":    func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4); s.RemoveFunc(func(e byte) bool { return e > 1 }) },
		"Pop":           func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4); s.Pop() },
		"Clear":         func(s *SyncSet[byte, byte]) { s.Add(1, 2, 3, 4); s.Clear() },
		"UnmarshalJSON": func(s *SyncSet[byte, byte]) { _ = s.UnmarshalJSON([]byte("[1,2,3,4]")) },
		"Do":            func(s *SyncSet[byte, byte]) { s.Do(func(set Set[byte, byte]) { set.Add(1, 2, 3, 4); set.Remove(1) }) },
	}
	for name, write := range writers {
		t.Run(name, func(t *testing.T) {
			s := newSyncBytes(1, 2, 3, 4)
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
							s.ContainsKey(byte(i))
							s.Len()
							s.Slice()
							s.Keys()
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
