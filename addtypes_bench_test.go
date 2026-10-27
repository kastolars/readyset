package readyset

import "testing"

type benchPoint struct {
	X, Y int
	Name string
}

// BenchmarkAddTypes checks that Add stays cheap for the element kinds that
// cost the most to compare with themselves.
func BenchmarkAddTypes(b *testing.B) {
	b.Run("int", func(b *testing.B) {
		var s Set[int]
		for i := 0; i < b.N; i++ {
			s.Add(i % 10_000)
		}
	})
	b.Run("string", func(b *testing.B) {
		keys := make([]string, 10_000)
		for i := range keys {
			keys[i] = "element-number-" + string(rune('a'+i%26)) + string(rune('a'+i/26%26)) + string(rune('a'+i/676%26))
		}
		var s Set[string]
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			s.Add(keys[i%len(keys)])
		}
	})
	b.Run("struct", func(b *testing.B) {
		var s Set[benchPoint]
		for i := 0; i < b.N; i++ {
			s.Add(benchPoint{X: i % 10_000, Y: 1, Name: "point"})
		}
	})
}
