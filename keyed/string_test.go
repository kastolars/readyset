package keyed

import (
	"fmt"
	"testing"
)

func TestString(t *testing.T) {
	var zero Set[int, person]
	if got := zero.String(); got != "{}" {
		t.Errorf("zero value = %q, want {}", got)
	}
	if got := newPeople().String(); got != "{}" {
		t.Errorf("empty = %q, want {}", got)
	}

	// Ordered by key, numerically, not by the printed form of the element.
	s := newPeople(person{10, "j"}, person{2, "b"}, person{1, "z"})
	if got, want := s.String(), "{{1 z}, {2 b}, {10 j}}"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringWorksWithFmt(t *testing.T) {
	s := newPeople(person{2, "b"}, person{1, "a"})
	const want = "{{1 a}, {2 b}}"
	for name, got := range map[string]string{
		"Sprint":     fmt.Sprint(s),
		"Sprintf %v": fmt.Sprintf("%v", s),
		"pointer":    fmt.Sprintf("%v", &s),
	} {
		if got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func TestStringIsDeterministic(t *testing.T) {
	s := New(func(n int) int { return n })
	for i := 0; i < 200; i++ {
		s.Add(i * 7 % 101)
	}
	first := s.String()
	for i := 0; i < 50; i++ {
		if got := s.String(); got != first {
			t.Fatalf("String changed between calls:\n%s\n%s", first, got)
		}
	}
}
