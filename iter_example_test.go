package readyset_test

import (
	"fmt"
	"slices"

	"github.com/kastolars/readyset"
)

func ExampleCollect() {
	words := []string{"b", "a", "b", "c"}
	unique := readyset.Collect(slices.Values(words))
	fmt.Println(unique)
	// Output: {a, b, c}
}
