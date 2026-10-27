# ReadySet

Generic sets for Go, with the usual set logic: union, intersection, difference, symmetric difference, subset and superset checks.

Two packages, for two situations:

| Package | Use it when | Elements are the same when |
|---|---|---|
| [`readyset`](.) | your elements are comparable (`int`, `string`, structs of comparable fields, ...) | their values are equal |
| [`readyset/keyed`](keyed) | elements are not comparable (they hold slices or maps), or only part of them, like an ID, defines identity | their keys are equal |

Requires Go 1.23 or later.

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
a.ContainsAll(2, 3)  // true: every one is present
a.ContainsAny(8, 9)  // false: none is present

a.Union(b)               // {2, 3, 4, 5, 6}
a.Intersection(b)        // {3}
a.Difference(b)          // {2, 5, 6}
a.SymmetricDifference(b) // {2, 4, 5, 6}
a.IsSubset(b)            // false
a.IsProperSubset(b)      // false: a proper subset is a subset that is not equal
a.IsDisjoint(b)          // false
a.Equal(b)               // false

a.AddSet(b)              // in place: a is now the union, {2, 3, 4, 5, 6}
a.RemoveSet(b)           // in place: a is now the difference, {2, 5, 6}

c := a.Clone()
c.RetainSet(b)           // in place: c is now the intersection, {}

evens := a.Filter(func(n int) bool { return n%2 == 0 }) // a new set, {2, 6}
c.AddSet(a)
c.RemoveFunc(func(n int) bool { return n > 4 })         // in place; returns how many it removed (2)
e, ok := c.Pop()                                        // remove and return an arbitrary element

readyset.Sorted(a)                  // []int{2, 5, 6}, ascending
readyset.Collect(slices.Values(xs)) // a set from any iter.Seq

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

users.ContainsKey("ann@example.com") // true
users.RemoveKey("ann@example.com")   // remove by key alone, no User needed
users.Keys()                         // []string{}, in no particular order
```

Because key equality is element equality, adding an element whose key is already present **replaces** the stored element, and `AddSet` follows the same rule. Operations between two sets compare keys only and expect the same key function on both. `Union` and `Intersection` return the receiver's elements. `keyed.Sorted(users)` returns the elements in ascending key order. `keyed` has the same `RetainSet`, `RemoveFunc`, `Filter`, `Pop`, `ContainsAll`, `ContainsAny` and proper-subset methods, plus `keyed.Collect(key, seq)`.

## Printing and JSON

Both packages print readably and encode to JSON as arrays, in a stable order (numbers numerically, strings lexically, `keyed` sets by key), so output never depends on map iteration order.

```go
fmt.Println(readyset.New(10, 2, 1)) // {1, 2, 10}

b, _ := json.Marshal(readyset.New("b", "a")) // ["a","b"]

var tags readyset.Set[string]
_ = json.Unmarshal([]byte(`["go","sets"]`), &tags) // works as a struct field too
```

A `keyed` set needs its key function to decode, so create it with `keyed.New` before calling `json.Unmarshal`.

## Sharing a set between goroutines

A plain `Set` is not safe for concurrent use. When goroutines share a set, use `SyncSet`, which does the locking for you. Both packages have one: `readyset.NewSync(...)` and `keyed.NewSync(key, ...)`.

```go
seen := readyset.NewSync[string]() // a *SyncSet; the zero value works too

go func() { seen.Add("a") }()
seen.Contains("a")
```

- **Each method is atomic, a sequence of methods is not.** `if !seen.Contains(x) { seen.Add(x) }` is a race, because another goroutine can add `x` in between. `Do` runs several steps under one lock:

  ```go
  var added bool
  seen.Do(func(set *readyset.Set[string]) {
      if !set.Contains("a") {
          set.Add("a")
          added = true
      }
  })
  ```

- **Operations between two sets lock both in a fixed order.** `a.Union(b)` and `b.Union(a)` running at once cannot deadlock, and `a.Union(a)` is fine.
- **Functions you pass in run under the lock.** The callbacks of `Filter`, `RemoveFunc` and `Do` must not call methods on the same set.
- **`All()` iterates over a copy** taken when the loop starts, so the loop body can change the set.
- **`Snapshot()` returns an independent plain `Set`**, for `Sorted` or for code that needs no locking.
- **Use a pointer and never copy one.** A `SyncSet` contains a lock, and `go vet` reports copies.

The lock has a cost, so a plain `Set` stays the default: use it when one goroutine owns the set or you already hold a mutex. Uncontended, the lock adds about 2 ns to a lookup (4.6 to 6.6 ns) and about 6 ns to an `Add` (4.4 to 10 ns). When many cores read the same set at once they contend on the lock, and a lookup averaged about 50 ns with 18 cores. For a set that is read by many goroutines and rarely changes, share an immutable `Snapshot()` instead. The numbers come from `go test -bench SyncCost` on an Apple M5 Pro.

## Recipes

```go
// Unique values from a slice. Slice() is in random order; Sorted is not.
unique := readyset.Sorted(readyset.New(ids...))

// What changed between two snapshots.
added := current.Difference(previous)
removed := previous.Difference(current)

// Does the caller hold every permission the action needs?
if granted.IsSuperset(required) {
    // ...
}

// Merge several sources into one set, then drop the blocked values.
var all readyset.Set[string] // the zero value is ready to use
for _, src := range sources {
    all.AddSet(src)
}
all.RemoveSet(blocked)

// Drain a worklist: process each item once, in no particular order.
for {
    item, ok := todo.Pop()
    if !ok {
        break
    }
    // ... may call todo.Add(...) for newly discovered work
}

// One record per ID, in ID order. Later records replace earlier ones.
byID := keyed.New(func(r Record) int { return r.ID }, records...)
latest := keyed.Sorted(byID)
```

## Behaviour to know about

- **Iteration order is random.** Go randomises map iteration, so `Slice()` and `All()` return elements in no particular order. Printing and JSON are sorted for you; sort the result yourself if you need a stable order from `Slice()`.
- **Copies share storage.** Assigning a non-empty set to another variable shares the underlying map. Use `Clone()` for an independent copy.
- **Result sets are sized for speed.** `Union`, `Intersection` and `Difference` allocate their result for the largest it could be, which avoids regrowing the map. If you keep a small result computed from a very large input, `Clone()` it to release the extra memory.
- **`Set` is not safe for concurrent use.** Use a `SyncSet` (see above), or guard the set with your own mutex.
- **Set operations return new sets.** `Union`, `Intersection` and the rest never modify their operands. `AddSet`, `RemoveSet` and `RetainSet` are the in-place forms of `Union`, `Difference` and `Intersection`.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate. `go test -race ./...` and `gofmt` must pass; CI also runs `staticcheck`.

## License

[MIT](LICENSE)
