package readyset

import (
	"fmt"
	"testing"
)

type point struct{ X, Y int }

func TestString(t *testing.T) {
	var zero Set[int]
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"zero value", zero.String(), "{}"},
		{"empty", New[int]().String(), "{}"},
		{"one element", New("a").String(), "{a}"},
		{"ints sort numerically", New(10, 2, 1, -3).String(), "{-3, 1, 2, 10}"},
		{"strings sort lexically", New("b", "a", "B").String(), "{B, a, b}"},
		{"structs sort by printed form", New(point{2, 1}, point{1, 9}, point{1, 2}).String(), "{{1 2}, {1 9}, {2 1}}"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestStringWorksWithFmt(t *testing.T) {
	s := New(3, 1, 2)
	const want = "{1, 2, 3}"

	for name, got := range map[string]string{
		"Sprint":     fmt.Sprint(s),
		"Sprintf %v": fmt.Sprintf("%v", s),
		//lint:ignore S1025 this test exercises the %s verb on purpose
		"Sprintf %s":    fmt.Sprintf("%s", s),
		"pointer %v":    fmt.Sprintf("%v", &s),
		"nested struct": fmt.Sprint(struct{ Tags Set[int] }{s}),
	} {
		if name == "nested struct" {
			if got != "{"+want+"}" {
				t.Errorf("%s: got %q, want %q", name, got, "{"+want+"}")
			}
			continue
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

// Map iteration order is random, so the output being identical every time is
// what makes String usable in logs and test expectations.
func TestStringIsDeterministic(t *testing.T) {
	s := New[int]()
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
