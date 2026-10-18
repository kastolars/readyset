package comparable

import (
	"testing"
	"testing/quick"

	"github.com/kastolars/readyset/internal/setlaws"
)

func TestLaws(t *testing.T) {
	err := quick.Check(func(x, y, z []byte) bool {
		setlaws.Check(t, New[byte], x, y, z)
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
		setlaws.Check(t, New[byte], x, y, z)
	})
}
