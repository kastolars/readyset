package readyset_test

import (
	"encoding/json"
	"fmt"

	"github.com/kastolars/readyset"
)

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
	fmt.Println(s)
	// Output: {2, 3}
}

func ExampleSet_AddSet() {
	all := readyset.New(1, 2)
	all.AddSet(readyset.New(2, 3))
	fmt.Println(all)
	// Output: {1, 2, 3}
}

func ExampleSet_RemoveSet() {
	all := readyset.New(1, 2, 3)
	all.RemoveSet(readyset.New(2, 9))
	fmt.Println(all)
	// Output: {1, 3}
}

func ExampleSet_Union() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(3, 4)
	fmt.Println(a.Union(b))
	// Output: {1, 2, 3, 4}
}

func ExampleSet_Intersection() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(2, 3, 4)
	fmt.Println(a.Intersection(b))
	// Output: {2, 3}
}

func ExampleSet_Difference() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(2, 3, 4)
	fmt.Println(a.Difference(b))
	// Output: {1}
}

func ExampleSet_SymmetricDifference() {
	a := readyset.New(1, 2, 3)
	b := readyset.New(2, 3, 4)
	fmt.Println(a.SymmetricDifference(b))
	// Output: {1, 4}
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

func ExampleSet_String() {
	// Numbers print in numeric order, whatever order they were added in.
	s := readyset.New(10, 2, 1)
	fmt.Println(s)
	// Output: {1, 2, 10}
}

func ExampleSet_MarshalJSON() {
	b, _ := json.Marshal(readyset.New("b", "a"))
	fmt.Println(string(b))
	// Output: ["a","b"]
}

func ExampleSet_UnmarshalJSON() {
	type Config struct {
		Tags readyset.Set[string] `json:"tags"`
	}
	var c Config
	err := json.Unmarshal([]byte(`{"tags":["go","sets","go"]}`), &c)
	fmt.Println(err, c.Tags.Len(), c.Tags.Contains("sets"))
	// Output: <nil> 2 true
}

func ExampleSet_RetainSet() {
	granted := readyset.New("read", "write", "admin")
	granted.RetainSet(readyset.New("read", "write", "delete"))
	fmt.Println(granted)
	// Output: {read, write}
}

func ExampleSet_RemoveFunc() {
	s := readyset.New(1, 2, 3, 4, 5, 6)
	removed := s.RemoveFunc(func(n int) bool { return n%2 == 0 })
	fmt.Println(removed, s)
	// Output: 3 {1, 3, 5}
}

func ExampleSet_Filter() {
	s := readyset.New(1, 2, 3, 4, 5, 6)
	fmt.Println(s.Filter(func(n int) bool { return n > 4 }))
	// Output: {5, 6}
}

func ExampleSet_Pop() {
	// Pop drains a worklist: process each item once, in no particular order.
	todo := readyset.New("a", "b", "c")
	n := 0
	for {
		_, ok := todo.Pop()
		if !ok {
			break
		}
		n++
	}
	fmt.Println(n, todo.IsEmpty())
	// Output: 3 true
}

func ExampleSet_ContainsAll() {
	s := readyset.New(1, 2, 3)
	fmt.Println(s.ContainsAll(1, 2), s.ContainsAll(1, 9))
	// Output: true false
}

func ExampleSet_ContainsAny() {
	s := readyset.New(1, 2, 3)
	fmt.Println(s.ContainsAny(9, 3), s.ContainsAny(8, 9))
	// Output: true false
}

func ExampleSet_IsProperSubset() {
	a, b := readyset.New(1, 2), readyset.New(1, 2, 3)
	fmt.Println(a.IsSubset(a), a.IsProperSubset(a), a.IsProperSubset(b))
	// Output: true false true
}
