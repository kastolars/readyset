package keyed

import (
	"fmt"
	"strings"

	"github.com/kastolars/readyset/internal/order"
)

// ordered returns the elements sorted by key, in the stable order used for
// printing and encoding.
func (s Set[K, V]) ordered() []V {
	keys := make([]K, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}
	order.Sort(keys)

	elems := make([]V, len(keys))
	for i, k := range keys {
		elems[i] = s.items[k]
	}
	return elems
}

// String returns the elements in braces, separated by commas, ordered by key:
// numeric keys numerically, string keys lexically, and any other key by its
// printed form. It implements [fmt.Stringer].
func (s Set[K, V]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	for i, e := range s.ordered() {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, e)
	}
	b.WriteByte('}')
	return b.String()
}
