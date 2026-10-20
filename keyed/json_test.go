package keyed

import (
	"encoding/json"
	"strings"
	"testing"
)

type roster struct {
	People Set[int, person] `json:"people"`
}

func TestMarshalJSON(t *testing.T) {
	var zero Set[int, person]
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"zero value", zero, `[]`},
		{"empty", newPeople(), `[]`},
		{"ordered by key", newPeople(person{10, "j"}, person{2, "b"}), `[{"ID":2,"Name":"b"},{"ID":10,"Name":"j"}]`},
		{"struct field", roster{People: newPeople(person{1, "a"})}, `{"people":[{"ID":1,"Name":"a"}]}`},
		{"zero struct field", roster{}, `{"people":[]}`},
	}
	for _, c := range cases {
		got, err := json.Marshal(c.v)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if string(got) != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	s := newPeople(person{99, "old"})
	err := json.Unmarshal([]byte(`[{"ID":2,"Name":"b"},{"ID":1,"Name":"a"}]`), &s)
	if err != nil {
		t.Fatal(err)
	}
	// The previous contents are replaced and the key function still works.
	if got := ids(s); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("ids = %v, want [1 2]", got)
	}
	if !s.Contains(person{ID: 2}) {
		t.Error("key function lost after decoding")
	}
}

func TestUnmarshalJSONDuplicateKeysLastWins(t *testing.T) {
	s := newPeople()
	if err := json.Unmarshal([]byte(`[{"ID":1,"Name":"first"},{"ID":1,"Name":"last"}]`), &s); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Get(1); got.Name != "last" || s.Len() != 1 {
		t.Errorf("got %v with Len %d", got, s.Len())
	}
}

func TestUnmarshalJSONNullIsNoOp(t *testing.T) {
	s := newPeople(person{1, "a"})
	if err := json.Unmarshal([]byte(`null`), &s); err != nil || s.Len() != 1 {
		t.Errorf("null: Len %d, err %v", s.Len(), err)
	}

	// Even a set that cannot decode accepts null, as encoding/json expects.
	var zero Set[int, person]
	if err := zero.UnmarshalJSON([]byte(" null ")); err != nil {
		t.Errorf("null into zero value: %v", err)
	}
}

func TestUnmarshalJSONRequiresNew(t *testing.T) {
	var zero Set[int, person]
	err := json.Unmarshal([]byte(`[{"ID":1}]`), &zero)
	if err == nil || !strings.Contains(err.Error(), "not created with New") {
		t.Errorf("got %v, want an error mentioning New", err)
	}

	// A struct field must be initialised first for the same reason.
	var r roster
	if err := json.Unmarshal([]byte(`{"people":[{"ID":1}]}`), &r); err == nil {
		t.Error("decoding into an uninitialised field should fail")
	}
	r = roster{People: newPeople()}
	if err := json.Unmarshal([]byte(`{"people":[{"ID":1,"Name":"a"}]}`), &r); err != nil || r.People.Len() != 1 {
		t.Errorf("initialised field: Len %d, err %v", r.People.Len(), err)
	}
}

func TestUnmarshalJSONErrorLeavesSetUnchanged(t *testing.T) {
	for _, in := range []string{`{"a":1}`, `[{"ID":"x"}]`, `[`, `"str"`} {
		s := newPeople(person{9, "keep"})
		if err := json.Unmarshal([]byte(in), &s); err == nil {
			t.Errorf("decoding %s: expected an error", in)
		}
		if got := ids(s); len(got) != 1 || got[0] != 9 {
			t.Errorf("decoding %s changed the set to %v", in, got)
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	in := newPeople(person{3, "c"}, person{1, "a"}, person{2, "b"})
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	out := newPeople()
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if !out.Equal(in) || out.String() != in.String() {
		t.Errorf("round trip changed the set: %v -> %v", in, out)
	}
}
