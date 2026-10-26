//go:build go1.21

package keyed_test

import (
	"fmt"

	"github.com/kastolars/readyset/keyed"
)

func ExampleSorted() {
	users := keyed.New(byEmail,
		User{Email: "c@x.io", Name: "Cy"},
		User{Email: "a@x.io", Name: "Ann"},
		User{Email: "b@x.io", Name: "Bob"},
	)
	// Elements come back in ascending order of their keys (the emails).
	for _, u := range keyed.Sorted(users) {
		fmt.Println(u.Name)
	}
	// Output:
	// Ann
	// Bob
	// Cy
}
