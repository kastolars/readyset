package keyed

import (
	"testing"
	"testing/quick"

	"github.com/kastolars/readyset/internal/setlaws"
)

// newBytes builds a set whose key is the byte itself, so the algebra can be
// checked with the same inputs as the readyset package.
func newBytes(elements ...byte) Set[byte, byte] {
	return New(func(b byte) byte { return b }, elements...)
}

func TestLaws(t *testing.T) {
	err := quick.Check(func(x, y, z []byte) bool {
		setlaws.Check(t, newBytes, x, y, z)
		return true
	}, &quick.Config{MaxCount: 2000})
	if err != nil {
		t.Fatal(err)
	}
}

func FuzzLaws(f *testing.F) {
	f.Add([]byte{1, 2, 3}, []byte{3, 4}, []byte{})
	f.Add([]byte{}, []byte{}, []byte{})
	f.Add([]byte{1, 1, 1}, []byte{1}, []byte{1, 2})
	f.Fuzz(func(t *testing.T, x, y, z []byte) {
		setlaws.Check(t, newBytes, x, y, z)
	})
}
