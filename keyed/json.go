package keyed

import (
	"bytes"
	"encoding/json"
	"errors"
)

// MarshalJSON implements [json.Marshaler]. It encodes the set as a JSON array
// of its elements, ordered by key, so equal sets always encode to the same
// bytes. An empty set encodes as [].
func (s Set[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ordered())
}

// UnmarshalJSON implements [json.Unmarshaler]. It decodes a JSON array into
// the set, replacing its previous contents; if several elements share a key,
// the last one wins. A JSON null leaves the set unchanged. If decoding fails
// the set is also left unchanged.
//
// The set needs a key function to decode, so it must have been created with
// [New]. To decode into a struct field, initialise the field first:
//
//	cfg := Config{Users: keyed.New(func(u User) string { return u.Email })}
//	err := json.Unmarshal(data, &cfg)
//
// Decoding into the zero value returns an error.
func (s *Set[K, V]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}
	if s.key == nil {
		return errors.New("keyed: cannot unmarshal into a Set that was not created with New")
	}
	var elems []V
	if err := json.Unmarshal(data, &elems); err != nil {
		return err
	}
	s.Clear()
	s.Add(elems...)
	return nil
}
