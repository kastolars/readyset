package keyed_test

import (
	"fmt"

	"github.com/kastolars/readyset/keyed"
)

// Checking for a key and then adding is a race, because another goroutine can
// add the key in between. Do runs both steps under one lock.
func ExampleSyncSet_Do() {
	users := keyed.NewSync(byEmail, User{Email: "a@x.io", Name: "Ann"})

	register := func(u User) bool {
		registered := false
		users.Do(func(set keyed.Set[string, User]) {
			if !set.ContainsKey(u.Email) {
				set.Add(u)
				registered = true
			}
		})
		return registered
	}

	fmt.Println(register(User{Email: "a@x.io", Name: "Impostor"}), register(User{Email: "b@x.io", Name: "Bob"}))
	fmt.Println(names(users.Snapshot()))
	// Output:
	// false true
	// [Ann Bob]
}
