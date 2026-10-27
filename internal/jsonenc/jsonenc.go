// Package jsonenc encodes the JSON arrays that sets marshal to.
package jsonenc

import (
	"bytes"
	"encoding/json"
)

// Marshal returns the JSON encoding of v without escaping HTML characters. A
// type's MarshalJSON should not escape them itself: the encoder that called it
// escapes the result when its own SetEscapeHTML option says to, and json.Marshal
// does, so a Marshaler that escapes up front ignores SetEscapeHTML(false).
func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}
