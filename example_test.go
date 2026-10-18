package readyset_test

import (
	"fmt"
	"sort"

	"github.com/kastolars/readyset"
)

// show prints the elements in ascending order so example output is stable.
func show(s readyset.Set[int]) {
	e := s.Slice()
	sort.Ints(e)
	fmt.Println(e)
}

func ExampleNew() {
	s := readyset.New(3, 1, 2, 3)
	fmt.Println(s.Len(), s.Contains(2), s.Contains(9))
	// Output: 3 true false
}

func ExampleSet_Add() {
	var s readyset.Set[string] // the zero value is ready to use
	s.Add("a", "b")
	s.Add("b", "c")
	fmt.Println(s.Len())
	// Output: 3
}

func ExampleSet_Remove() {
	s := readyset.New(1, 2, 3)
	s.Remove(1, 9) // 9 is not present; that is not an error
	show(s)
	// Output: [2 3]
}

func ExampleSet_Union() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(3, 4)
	show(a.Union(b))
	// Output: [1 2 3 4]
}

func ExampleSet_Intersection() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(2, 3, 4)
	show(a.Intersection(b))
	// Output: [2 3]
}

func ExampleSet_Difference() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(2, 3, 4)
	show(a.Difference(b))
	// Output: [1]
}

func ExampleSet_SymmetricDifference() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(2, 3, 4)
	show(a.SymmetricDifference(b))
	// Output: [1 4]
}

func ExampleSet_IsSubset() {
	a := readyset.New(2, 3)
	b := readyset.New(1, 2, 3, 4)
	fmt.Println(a.IsSubset(b), b.IsSubset(a))
	// Output: true false
}

func ExampleSet_IsDisjoint() {
	odds := readyset.New(1, 3, 5)
	evens := readyset.New(2, 4, 6)
	fmt.Println(odds.IsDisjoint(evens))
	// Output: true
}

func ExampleSet_Clone() {
	a := readyset.New(1, 2)
	b := a.Clone()
	b.Add(3)
	fmt.Println(a.Len(), b.Len())
	// Output: 2 3
}
