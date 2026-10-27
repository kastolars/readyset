package keyed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kastolars/readyset/internal/jsonenc"
)

// MarshalJSON implements [json.Marshaler]. It encodes the set as a JSON array
// of its elements, ordered by key, so a set encodes the same way whatever its
// map iteration order. An empty set encodes as [].
func (s Set[K, V]) MarshalJSON() ([]byte, error) {
	return jsonenc.Marshal(s.ordered())
}

// UnmarshalJSON implements [json.Unmarshaler]. It decodes a JSON array into
// the set, replacing its previous contents; if several elements share a key,
// the last one wins. A JSON null leaves the set unchanged. If decoding fails
// the set is also left unchanged, including when a key function returns a value
// that cannot be a map key.
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
	fresh, err := s.fromElems(elems)
	if err != nil {
		return err
	}
	s.Clear()
	s.AddSet(fresh)
	return nil
}

// fromElems makes a set of elems with the key function of s, turning a panic while
// adding them into an error, so that a failed decode changes nothing.
func (s Set[K, V]) fromElems(elems []V) (fresh Set[K, V], err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("keyed: cannot decode set: %v", r)
		}
	}()
	fresh = s.like(len(elems))
	fresh.Add(elems...)
	return fresh, nil
}
