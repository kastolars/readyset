// Package readyset provides generic sets for Go, with the usual set logic:
// union, intersection, difference, symmetric difference, and subset, superset
// and disjointness checks.
//
// [Set] holds any comparable type, and two elements are the same element when
// they are equal:
//
//	a := readyset.New(1, 2, 3)
//	b := readyset.New(3, 4)
//	fmt.Println(a.Union(b))        // {1, 2, 3, 4}
//	fmt.Println(a.Intersection(b)) // {3}
//
// Union, difference and intersection also have in-place forms: [Set.AddSet],
// [Set.RemoveSet] and [Set.RetainSet] update a set instead of returning a new one. [Set.Filter], [Set.RemoveFunc] and [Set.Pop] select, delete and
// drain elements, and [Sorted] and [Collect] convert to and from ordered slices
// and iterators on the Go versions that support them.
//
// For elements that are not comparable, or that are identified by only part of
// their value such as an ID, use [github.com/kastolars/readyset/keyed].
//
// Sets print with their elements in a stable order, and encode to and decode
// from JSON as arrays. Iteration order ([Set.Slice]) is random, as with Go
// maps.
package readyset
