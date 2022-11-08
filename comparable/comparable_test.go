package comparable

import "testing"

func TestNew(t *testing.T) {
	_ = New[int]()

	_ = New(1, 2, 3, 4)
}

func TestEqual(t *testing.T) {
	s1 := New(1, 2, 3, 4)
	s2 := New(1, 2, 3, 4)

	if !s1.Equal(s2) {
		t.Error()
	}

	s3 := New(1, 2, 3)

	if s1.Equal(s3) {
		t.Error()
	}

	s4 := New(1, 2, 3, 5)

	if s1.Equal(s4) {
		t.Error()
	}

}

func TestSetAdd(t *testing.T) {
	s := New[int]()
	s.Add(1)
	if !s.Contains(1) {
		t.Error()
	}
}

func TestSetRemove(t *testing.T) {
	s := New(1)
	s.Remove(1)
	if s.Contains(1) {
		t.Error()
	}
}

func TestContains(t *testing.T) {
	s := New(45, 82)
	if !s.Contains(45) {
		t.Error()
	}
}

func TestIter(t *testing.T) {
	s := New(2, 4, 6, 8, 10, 12)
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
	s1 := New(2, 4, 6)
	s2 := New(1, 3, 5)
	s3 := s1.Union(s2)

	expected := New(1, 2, 3, 4, 5, 6)

	if !s3.Equal(expected) {
		t.Error()
	}
}

func TestIntersection(t *testing.T) {
	s1 := New(2, 4, 6)
	s2 := New(1, 4, 5)
	s3 := s1.Intersection(s2)

	if s3.Len() != 1 {
		t.Error()
	}

	expected := New(4)

	if !s3.Equal(expected) {
		t.Error()
	}

	s4 := s1.Intersection(expected)

	if !s4.Equal(expected) {
		t.Error()
	}
}

func TestDifference(t *testing.T) {
	s1 := New(1, 2, 3)
	s2 := New(2, 3, 5)
	s3 := s1.Difference(s2)

	expected := New(1)

	if !s3.Equal(expected) {
		t.Error()
	}
}

func TestIsDisjoint(t *testing.T) {
	s1 := New(2, 4, 6)
	s2 := New(1, 3, 5)

	if !s1.IsDisjoint(s2) {
		t.Error()
	}
}
