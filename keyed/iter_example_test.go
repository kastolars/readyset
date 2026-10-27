package keyed_test

import (
	"fmt"
	"slices"

	"github.com/kastolars/readyset/keyed"
)

func ExampleCollect() {
	rows := []User{
		{Email: "a@x.io", Name: "Ann"},
		{Email: "b@x.io", Name: "Bob"},
		{Email: "a@x.io", Name: "Ann Lee"}, // same key: the last one wins
	}
	users := keyed.Collect(byEmail, slices.Values(rows))
	fmt.Println(names(users))
	// Output: [Ann Lee Bob]
}
