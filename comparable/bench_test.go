package comparable

import "testing"

// build returns a set of n ints starting at offset.
func build(n, offset int) Set[int] {
	s := New[int]()
	for i := 0; i < n; i++ {
		s.Add(offset + i)
	}
	return s
}

func benchSizes(b *testing.B, name string, fn func(a, c Set[int]) any) {
	for _, n := range []int{100, 10_000, 100_000} {
		// Half of each set overlaps with the other.
		a, c := build(n, 0), build(n, n/2)
		b.Run(name+"/"+itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sink = fn(a, c)
			}
		})
	}
}

var sink any

func itoa(n int) string {
	switch n {
	case 100:
		return "100"
	case 10_000:
		return "10k"
	default:
		return "100k"
	}
}

func BenchmarkUnion(b *testing.B) {
	benchSizes(b, "Union", func(a, c Set[int]) any { return a.Union(c) })
}

func BenchmarkIntersection(b *testing.B) {
	benchSizes(b, "Intersection", func(a, c Set[int]) any { return a.Intersection(c) })
}

func BenchmarkDifference(b *testing.B) {
	benchSizes(b, "Difference", func(a, c Set[int]) any { return a.Difference(c) })
}

// IsDisjoint on overlapping sets should be able to stop at the first shared element.
func BenchmarkIsDisjoint(b *testing.B) {
	benchSizes(b, "IsDisjoint", func(a, c Set[int]) any { return a.IsDisjoint(c) })
}
