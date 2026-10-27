package readyset

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kastolars/readyset/internal/jsonenc"
	"github.com/kastolars/readyset/internal/order"
)

// MarshalJSON implements [json.Marshaler]. It encodes the set as a JSON array
// of its elements, in the same stable order as [Set.String], so a set encodes
// the same way whatever its map iteration order. An empty set encodes as [].
func (s Set[T]) MarshalJSON() ([]byte, error) {
	elems := s.Slice()
	order.Sort(elems)
	return jsonenc.Marshal(elems)
}

// UnmarshalJSON implements [json.Unmarshaler]. It decodes a JSON array into
// the set, replacing its previous contents; duplicate elements in the array
// are collapsed. A JSON null leaves the set unchanged. If decoding fails the
// set is also left unchanged, including when T is an interface type and the
// array holds an element that cannot be in a set, such as a nested array.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	var elems []T
	if err := json.Unmarshal(data, &elems); err != nil {
		return err
	}
	fresh, err := fromElems(elems)
	if err != nil {
		return err
	}
	s.Clear()
	s.AddSet(fresh)
	return nil
}

// fromElems makes a set of elems, turning the panic that adding an unhashable
// element causes into an error, so that a failed decode changes nothing.
func fromElems[T comparable](elems []T) (fresh Set[T], err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("readyset: cannot decode set: %v", r)
		}
	}()
	return New(elems...), nil
}
