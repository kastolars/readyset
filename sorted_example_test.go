//go:build go1.21

package readyset_test

import (
	"fmt"

	"github.com/kastolars/readyset"
)

func ExampleSorted() {
	// Slice returns elements in random order; Sorted gives a stable one.
	ids := readyset.New(30, 10, 20, 10)
	fmt.Println(readyset.Sorted(ids))
	// Output: [10 20 30]
}
