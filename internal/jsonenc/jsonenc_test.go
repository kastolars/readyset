package jsonenc

import (
	"testing"
)

func TestMarshal(t *testing.T) {
	got, err := Marshal([]string{"<a&b>", "x"})
	if err != nil || string(got) != `["<a&b>","x"]` {
		t.Errorf("Marshal = %s, %v; want the array with no escaping and no trailing newline", got, err)
	}
	if got, err := Marshal([]int{}); err != nil || string(got) != `[]` {
		t.Errorf("Marshal(empty) = %s, %v; want []", got, err)
	}
	if _, err := Marshal(make(chan int)); err == nil {
		t.Error("Marshal of a channel succeeded, want an error")
	}
}
