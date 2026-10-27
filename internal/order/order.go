// Package order sorts values of any type into a stable, human-friendly order,
// so that printed and encoded sets do not depend on map iteration order.
package order

import (
	"fmt"
	"reflect"
	"sort"
)

// Sort sorts s in place.
//
// Integers, unsigned integers and floating-point numbers sort numerically
// (NaN first), and strings sort lexically. Every other value sorts by its
// default text form, as printed by fmt.Sprint. Values that still tie, such as
// an int and an int64 holding the same number, are ordered by their Go-syntax
// form and then their type name. Values of different kinds, such as in a slice
// of any, are grouped by kind, so the result does not depend on the input
// order.
func Sort[T any](s []T) {
	if len(s) < 2 {
		return
	}
	es := make([]entry[T], len(s))
	for i, v := range s {
		es[i] = newEntry(v)
	}
	sort.Slice(es, func(i, j int) bool { return less(es[i], es[j]) })
	for i := range es {
		s[i] = es[i].val
	}
}

// entry is a value together with the precomputed key it is sorted by.
type entry[T any] struct {
	val   T
	kind  reflect.Kind // the group the value sorts in
	i     int64
	u     uint64
	f     float64
	str   string // the value itself for strings, otherwise its fmt.Sprint form
	gostr string // the Go-syntax form, used only to break ties
	typ   string // the dynamic type's name, the last tie-break
}

func newEntry[T any](v T) entry[T] {
	e := entry[T]{val: v}
	rv := reflect.ValueOf(any(v))
	if !rv.IsValid() { // a nil interface
		e.str = "<nil>"
		return e
	}
	e.typ = rv.Type().String()
	switch k := rv.Kind(); k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.kind, e.i = reflect.Int, rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		e.kind, e.u = reflect.Uint, rv.Uint()
	case reflect.Float32, reflect.Float64:
		e.kind, e.f = reflect.Float64, rv.Float()
	case reflect.String:
		e.kind, e.str = reflect.String, rv.String()
	default:
		e.kind = k
		e.str = fmt.Sprint(v)
		e.gostr = fmt.Sprintf("%#v", v)
	}
	return e
}

func less[T any](a, b entry[T]) bool {
	if a.kind != b.kind {
		return a.kind < b.kind
	}
	switch a.kind {
	case reflect.Int:
		if a.i != b.i {
			return a.i < b.i
		}
	case reflect.Uint:
		if a.u != b.u {
			return a.u < b.u
		}
	case reflect.Float64:
		aNaN, bNaN := a.f != a.f, b.f != b.f
		if aNaN != bNaN { // a NaN sorts before everything except another NaN
			return aNaN
		}
		if !aNaN && a.f != b.f {
			return a.f < b.f
		}
	default:
		if a.str != b.str {
			return a.str < b.str
		}
	}
	if a.gostr != b.gostr {
		return a.gostr < b.gostr
	}
	return a.typ < b.typ
}
