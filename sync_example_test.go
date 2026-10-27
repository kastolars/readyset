package readyset_test

import (
	"fmt"
	"sync"

	"github.com/kastolars/readyset"
)

// A SyncSet can be shared between goroutines without any locking of your own.
func ExampleNewSync() {
	seen := readyset.NewSync[string]()

	var wg sync.WaitGroup
	for _, word := range []string{"b", "a", "b", "c", "a"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seen.Add(word)
		}()
	}
	wg.Wait()

	fmt.Println(seen.Len(), seen)
	// Output: 3 {a, b, c}
}

// Contains followed by Add is a race, because another goroutine can add the
// name in between. Do runs both steps under one lock.
func ExampleSyncSet_Do() {
	names := readyset.NewSync("ann")

	claim := func(name string) bool {
		claimed := false
		names.Do(func(set *readyset.Set[string]) {
			if !set.Contains(name) {
				set.Add(name)
				claimed = true
			}
		})
		return claimed
	}

	fmt.Println(claim("ann"), claim("bob"))
	// Output: false true
}

// Snapshot copies a SyncSet into a plain Set, for functions like Sorted.
func ExampleSyncSet_Snapshot() {
	s := readyset.NewSync(30, 10, 20)
	fmt.Println(readyset.Sorted(s.Snapshot()))
	// Output: [10 20 30]
}
