# ReadySet

Generic sets for Go, with the usual set logic: union, intersection, difference, symmetric difference, subset and superset checks.

Two packages, for two situations:

| Package | Use it when | Elements are the same when |
|---|---|---|
| [`readyset`](.) | your elements are comparable (`int`, `string`, structs of comparable fields, ...) | their values are equal |
| [`readyset/keyed`](keyed) | elements are not comparable (they hold slices or maps), or only part of them, like an ID, defines identity | their keys are equal |

Requires Go 1.18 or later. The `All()` iterator methods are available with Go 1.23 or later.

## Installation

```bash
go get github.com/kastolars/readyset
```

## `readyset`

```go
import "github.com/kastolars/readyset"

a := readyset.New(1, 2, 3)
b := readyset.New(3, 4)

a.Add(5, 6)          // variadic
a.Remove(1)
a.Contains(2)        // true

a.Union(b)               // {2, 3, 4, 5, 6}
a.Intersection(b)        // {3}
a.Difference(b)          // {2, 5, 6}
a.SymmetricDifference(b) // {2, 4, 5, 6}
a.IsSubset(b)            // false
a.IsDisjoint(b)          // false
a.Equal(b)               // false

var zero readyset.Set[string] // the zero value is an empty, usable set
zero.Add("x")
```

## `readyset/keyed`

A key function maps each element to a comparable key. Two elements are the same element exactly when their keys are equal.

```go
import "github.com/kastolars/readyset/keyed"

type User struct {
    Email string
    Name  string
    Tags  []string // not comparable, so User cannot go in a readyset.Set
}

users := keyed.New(func(u User) string { return u.Email })
users.Add(User{Email: "ann@example.com", Name: "Ann"})

users.Contains(User{Email: "ann@example.com"}) // true: only the key is compared
u, ok := users.Get("ann@example.com")
```

Because key equality is element equality, adding an element whose key is already present **replaces** the stored element. Operations between two sets compare keys only and expect the same key function on both. `Union` and `Intersection` return the receiver's elements.

## Printing and JSON

Both packages print readably and encode to JSON as arrays, in a stable order (numbers numerically, strings lexically, `keyed` sets by key), so output never depends on map iteration order.

```go
fmt.Println(readyset.New(10, 2, 1)) // {1, 2, 10}

b, _ := json.Marshal(readyset.New("b", "a")) // ["a","b"]

var tags readyset.Set[string]
_ = json.Unmarshal([]byte(`["go","sets"]`), &tags) // works as a struct field too
```

A `keyed` set needs its key function to decode, so create it with `keyed.New` before calling `json.Unmarshal`.

## Behaviour to know about

- **Iteration order is random.** Go randomises map iteration, so `Slice()` and `All()` return elements in no particular order. Printing and JSON are sorted for you; sort the result yourself if you need a stable order from `Slice()`.
- **Copies share storage.** Assigning a non-empty set to another variable shares the underlying map. Use `Clone()` for an independent copy.
- **Not safe for concurrent use.** Guard a set with your own mutex if goroutines share it.
- **Operations never modify their operands.** `Union`, `Intersection` and the rest return new sets.

## Alternatives

- [`deckarep/golang-set`](https://pkg.go.dev/github.com/deckarep/golang-set/v2): mature, with thread-safe and unsafe variants.
- [`hashicorp/go-set`](https://pkg.go.dev/github.com/hashicorp/go-set/v2): sets, hash sets and tree sets behind a common interface.
- A set type for the standard library has been proposed (see [golang/go#47331](https://golang.org/issue/47331) and the follow-up proposals). If one ships, prefer it.

ReadySet is deliberately small: no thread safety, no ordered sets, no shared collection interface.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate. `go test -race ./...` and `gofmt` must pass; CI also runs `staticcheck`.

## License

[MIT](LICENSE)
