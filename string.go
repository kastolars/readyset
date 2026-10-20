package readyset

import (
	"fmt"
	"strings"

	"github.com/kastolars/readyset/internal/order"
)

// String returns the elements in braces, separated by commas, in a stable
// order: numbers numerically, strings lexically, and any other value by its
// printed form. It implements [fmt.Stringer], so a set prints readably with
// fmt.Println and the %v and %s verbs.
func (s Set[T]) String() string {
	elems := s.Slice()
	order.Sort(elems)

	var b strings.Builder
	b.WriteByte('{')
	for i, e := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, e)
	}
	b.WriteByte('}')
	return b.String()
}
