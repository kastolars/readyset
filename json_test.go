package readyset

import (
	"encoding/json"
	"strings"
	"testing"
	"testing/quick"
)

type config struct {
	Tags Set[string] `json:"tags"`
}

func TestMarshalJSON(t *testing.T) {
	var zero Set[int]
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"zero value", zero, `[]`},
		{"empty", New[int](), `[]`},
		{"ints sort numerically", New(10, 2, 1), `[1,2,10]`},
		{"strings", New("b", "a"), `["a","b"]`},
		{"structs", New(point{2, 1}, point{1, 2}), `[{"X":1,"Y":2},{"X":2,"Y":1}]`},
		{"pointer to set", func() any { s := New(2, 1); return &s }(), `[1,2]`},
		{"struct field", config{Tags: New("b", "a")}, `{"tags":["a","b"]}`},
		{"zero struct field", config{}, `{"tags":[]}`},
		{"map value", map[string]Set[int]{"k": New(2, 1)}, `{"k":[1,2]}`},
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

func TestMarshalJSONIsDeterministic(t *testing.T) {
	s := New[int]()
	for i := 0; i < 200; i++ {
		s.Add(i * 7 % 101)
	}
	first, _ := json.Marshal(s)
	for i := 0; i < 50; i++ {
		if got, _ := json.Marshal(s); string(got) != string(first) {
			t.Fatalf("encoding changed between calls:\n%s\n%s", first, got)
		}
	}
}

func TestMarshalJSONError(t *testing.T) {
	// A channel is comparable but cannot be encoded as JSON.
	_, err := New(make(chan int)).MarshalJSON()
	if err == nil {
		t.Fatal("expected an error encoding a set of channels")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var s Set[int]
	if err := json.Unmarshal([]byte(`[3, 1, 2, 2]`), &s); err != nil {
		t.Fatal(err)
	}
	if want := New(1, 2, 3); !s.Equal(want) {
		t.Errorf("got %v, want %v", s, want)
	}

	// Decoding replaces the previous contents.
	if err := json.Unmarshal([]byte(`[7]`), &s); err != nil {
		t.Fatal(err)
	}
	if want := New(7); !s.Equal(want) {
		t.Errorf("after second decode got %v, want %v", s, want)
	}

	if err := json.Unmarshal([]byte(`[]`), &s); err != nil || !s.IsEmpty() {
		t.Errorf("decoding [] gave %v, err %v; want an empty set", s, err)
	}
}

func TestUnmarshalJSONNullIsNoOp(t *testing.T) {
	s := New(9)
	if err := json.Unmarshal([]byte(`null`), &s); err != nil {
		t.Fatal(err)
	}
	if err := s.UnmarshalJSON([]byte(" null\n")); err != nil {
		t.Fatal(err)
	}
	if want := New(9); !s.Equal(want) {
		t.Errorf("null changed the set to %v", s)
	}
}

func TestUnmarshalJSONErrorLeavesSetUnchanged(t *testing.T) {
	for _, in := range []string{`{"a":1}`, `["x"]`, `[1,`, `"str"`} {
		s := New(9)
		if err := json.Unmarshal([]byte(in), &s); err == nil {
			t.Errorf("decoding %s: expected an error", in)
		}
		if want := New(9); !s.Equal(want) {
			t.Errorf("decoding %s changed the set to %v", in, s)
		}
	}
}

func TestUnmarshalJSONStructField(t *testing.T) {
	var c config
	if err := json.Unmarshal([]byte(`{"tags":["b","a","a"]}`), &c); err != nil {
		t.Fatal(err)
	}
	if want := New("a", "b"); !c.Tags.Equal(want) {
		t.Errorf("got %v, want %v", c.Tags, want)
	}

	var missing config
	if err := json.Unmarshal([]byte(`{}`), &missing); err != nil || !missing.Tags.IsEmpty() {
		t.Errorf("missing field: %v, err %v", missing.Tags, err)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	err := quick.Check(func(xs []int) bool {
		in := New(xs...)
		b, err := json.Marshal(in)
		if err != nil {
			return false
		}
		var out Set[int]
		return json.Unmarshal(b, &out) == nil && out.Equal(in)
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	pts := New(point{1, 2}, point{3, 4})
	b, err := json.Marshal(pts)
	if err != nil {
		t.Fatal(err)
	}
	var back Set[point]
	if err := json.Unmarshal(b, &back); err != nil || !back.Equal(pts) {
		t.Errorf("struct elements: got %v, err %v (json %s)", back, err, b)
	}
}

// Like encoding/json, HTML-sensitive characters are escaped on output and the
// original element is restored on input.
func TestMarshalJSONEscapesLikeEncodingJSON(t *testing.T) {
	in := New("<a&b>")
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(string(b), "<>&") {
		t.Errorf("got %s; want <, > and & escaped", b)
	}
	var out Set[string]
	if err := json.Unmarshal(b, &out); err != nil || !out.Equal(in) {
		t.Errorf("round trip: got %v, err %v", out, err)
	}
}
