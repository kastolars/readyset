package keyed_test

import (
	"fmt"
	"sort"

	"github.com/kastolars/readyset/keyed"
)

// User is the element type used by the examples. Its Tags field makes it
// non-comparable, so it could not be a member of a readyset.Set.
type User struct {
	Email string
	Name  string
	Tags  []string
}

// byEmail identifies a User by their email address.
func byEmail(u User) string { return u.Email }

func names(s keyed.Set[string, User]) []string {
	out := []string{}
	for _, u := range s.Slice() {
		out = append(out, u.Name)
	}
	sort.Strings(out)
	return out
}

func ExampleNew() {
	users := keyed.New(byEmail,
		User{Email: "a@x.io", Name: "Ann"},
		User{Email: "b@x.io", Name: "Bob"},
	)
	fmt.Println(users.Len(), users.Contains(User{Email: "a@x.io"}))
	// Output: 2 true
}

// Elements with equal keys are the same element, so adding one replaces the
// other.
func ExampleSet_Add() {
	users := keyed.New(byEmail, User{Email: "a@x.io", Name: "Ann"})
	users.Add(User{Email: "a@x.io", Name: "Ann Lee"})

	u, _ := users.Get("a@x.io")
	fmt.Println(users.Len(), u.Name)
	// Output: 1 Ann Lee
}

func ExampleSet_Union() {
	staff := keyed.New(byEmail,
		User{Email: "a@x.io", Name: "Ann"},
		User{Email: "b@x.io", Name: "Bob"},
	)
	admins := keyed.New(byEmail,
		User{Email: "b@x.io", Name: "Bob (admin)"},
		User{Email: "c@x.io", Name: "Cy"},
	)
	// For a key in both sets, the receiver's element is kept.
	fmt.Println(names(staff.Union(admins)))
	// Output: [Ann Bob Cy]
}

func ExampleSet_Intersection() {
	staff := keyed.New(byEmail,
		User{Email: "a@x.io", Name: "Ann"},
		User{Email: "b@x.io", Name: "Bob"},
	)
	admins := keyed.New(byEmail,
		User{Email: "b@x.io", Name: "Bob (admin)"},
		User{Email: "c@x.io", Name: "Cy"},
	)
	fmt.Println(names(staff.Intersection(admins)))
	// Output: [Bob]
}

func ExampleSet_Difference() {
	staff := keyed.New(byEmail,
		User{Email: "a@x.io", Name: "Ann"},
		User{Email: "b@x.io", Name: "Bob"},
	)
	admins := keyed.New(byEmail, User{Email: "b@x.io"})
	fmt.Println(names(staff.Difference(admins)))
	// Output: [Ann]
}
