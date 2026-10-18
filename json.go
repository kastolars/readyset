package readyset

import (
	"bytes"
	"encoding/json"

	"github.com/kastolars/readyset/internal/order"
)

// MarshalJSON implements [json.Marshaler]. It encodes the set as a JSON array
// of its elements, in the same stable order as [Set.String], so equal sets
// always encode to the same bytes. An empty set encodes as [].
func (s Set[T]) MarshalJSON() ([]byte, error) {
	elems := s.Slice()
	order.Sort(elems)
	return json.Marshal(elems)
}

// UnmarshalJSON implements [json.Unmarshaler]. It decodes a JSON array into
// the set, replacing its previous contents; duplicate elements in the array
// are collapsed. A JSON null leaves the set unchanged. If decoding fails the
// set is also left unchanged.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	var elems []T
	if err := json.Unmarshal(data, &elems); err != nil {
		return err
	}
	s.Clear()
	s.Add(elems...)
	return nil
}
