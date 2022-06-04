package hashable

import (
	"testing"
)

type testHashable[T comparable] struct {
	a, b int
}

func (th testHashable[T]) Hash() int {
	return th.a + th.b
}

func TestNewHashable(t *testing.T) {
	th := testHashable[int]{1, 2}
	NewHashable[int](th)
}

func TestAdd(t *testing.T) {
	t.Skip()
	th := testHashable[int]{1, 2}

	s := NewHashable[int]()

	s.Add(th)
	if !s.Contains(th) {
		t.Error()
	}
}

func TestSetRemove(t *testing.T) {
	t.Skip()
	th := testHashable[int]{1, 2}

	s := NewHashable[int](th)
	s.Remove(th)
	if s.Contains(th) {
		t.Error()
	}
}

func TestContains(t *testing.T) {
	t.Skip()

	th := testHashable[int]{1, 2}

	s := NewHashable[int](th)
	if !s.Contains(th) {
		t.Error()
	}
}

func TestIter(t *testing.T) {
	th1 := testHashable[int]{1, 2}
	th2 := testHashable[int]{1, 1}

	s := NewHashable[int](th1, th2)
	if len(s.Iter()) != len(s.set) {
		t.Error()
	}
	for _, val := range s.Iter() {
		if !s.Contains(val) {
			t.Error()
			break
		}
	}
}

func TestUnion(t *testing.T) {
	th1 := testHashable[int]{1, 2}
	th2 := testHashable[int]{1, 1}
	th3 := testHashable[int]{2, 2}
	s1 := NewHashable[int](th1, th2)
	s2 := NewHashable[int](th3)
	s3 := s1.Union(s2)

	expected := NewHashable[int](th1, th2, th3)

	if !s3.Equal(expected) {
		t.Error()
	}
}

func TestIntersection(t *testing.T) {
	th1 := testHashable[int]{1, 2}
	th2 := testHashable[int]{1, 1}
	th3 := testHashable[int]{2, 2}
	s1 := NewHashable[int](th1, th2)
	s2 := NewHashable[int](th1, th3)
	s3 := s1.Intersection(s2)

	if s3.Len() != 1 {
		t.Error()
	}

	expected := NewHashable[int](th1)

	if !s3.Equal(expected) {
		t.Error()
	}

	s4 := s1.Intersection(expected)

	if !s4.Equal(expected) {
		t.Error()
	}
}

func TestDifference(t *testing.T) {
	th1 := testHashable[int]{1, 2}
	th2 := testHashable[int]{1, 1}
	th3 := testHashable[int]{2, 2}
	s1 := NewHashable[int](th1, th2)
	s2 := NewHashable[int](th1, th3)
	s3 := s1.Difference(s2)

	expected := NewHashable[int](th2)

	if !s3.Equal(expected) {
		t.Error()
	}
}
