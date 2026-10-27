// Package readyset provides generic sets for Go, with the usual set logic:
// union, intersection, difference, symmetric difference, and subset, superset
// and disjointness checks.
//
// There are two kinds of set. [Set] holds any comparable type, and two
// elements are the same element when they are equal:
//
//	a := readyset.New(1, 2, 3)
//	b := readyset.New(3, 4)
//	fmt.Println(a.Union(b))        // {1, 2, 3, 4}
//	fmt.Println(a.Intersection(b)) // {3}
//
// For sets of records identified by a key, such as users by email or rows by
// ID, including records that cannot be compared with ==, use
// [github.com/kastolars/readyset/keyed]. Two elements there are the same
// element when their keys are equal.
//
// Union, difference and intersection also have in-place forms: [Set.AddSet],
// [Set.RemoveSet] and [Set.RetainSet] update a set instead of returning a new
// one. [Set.Filter], [Set.RemoveFunc] and [Set.Pop] select, delete and drain
// elements, and [Sorted] and [Collect] convert to ordered slices and from
// iterators.
//
// A [Set] is not safe for concurrent use. [SyncSet], created with [NewSync], is
// the same set behind a lock, for goroutines that share it.
//
// Sets print with their elements in a stable order, and encode to and decode
// from JSON as arrays. Iteration order ([Set.Slice]) is random, as with Go
// maps.
package readyset
