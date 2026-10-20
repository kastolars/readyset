package keyed_test

import (
	"encoding/json"
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

func ExampleSet_String() {
	users := keyed.New(byEmail,
		User{Email: "b@x.io", Name: "Bob"},
		User{Email: "a@x.io", Name: "Ann"},
	)
	// Elements print in key order.
	fmt.Println(users)
	// Output: {{a@x.io Ann []}, {b@x.io Bob []}}
}

func ExampleSet_MarshalJSON() {
	users := keyed.New(byEmail,
		User{Email: "b@x.io", Name: "Bob"},
		User{Email: "a@x.io", Name: "Ann"},
	)
	b, _ := json.Marshal(users)
	fmt.Println(string(b))
	// Output: [{"Email":"a@x.io","Name":"Ann","Tags":null},{"Email":"b@x.io","Name":"Bob","Tags":null}]
}

func ExampleSet_UnmarshalJSON() {
	type Config struct {
		Users keyed.Set[string, User] `json:"users"`
	}
	// The set needs its key function to decode, so create it with New first.
	cfg := Config{Users: keyed.New(byEmail)}
	err := json.Unmarshal([]byte(`{"users":[{"Email":"a@x.io","Name":"Ann"}]}`), &cfg)
	fmt.Println(err, cfg.Users.Len())
	// Output: <nil> 1
}

func ExampleSet_RemoveKey() {
	users := keyed.New(byEmail,
		User{Email: "a@x.io", Name: "Ann"},
		User{Email: "b@x.io", Name: "Bob"},
	)
	// Remove by key alone, without building a User.
	users.RemoveKey("a@x.io")
	fmt.Println(users.Len(), users.ContainsKey("a@x.io"), users.ContainsKey("b@x.io"))
	// Output: 1 false true
}
