package set

type Hashable[T comparable] interface {
	Hash() T
}

type Set[T comparable] struct {
	set map[T]struct{}
}

func New[T comparable](elements ...T) Set[T] {
	set := make(map[T]struct{})
	for _, val := range elements {
		set[val] = struct{}{}
	}
	return Set[T]{set}
}

func (s *Set[comparable]) Add(element comparable) {
	s.set[element] = struct{}{}
}

func (s *Set[comparable]) Remove(element comparable) {
	delete(s.set, element)
}

func (s Set[comparable]) Contains(element comparable) bool {
	_, ok := s.set[element]
	return ok
}

func (s Set[comparable]) Len() int {
	return len(s.set)
}

func (s1 Set[comparable]) Equal(s2 Set[comparable]) bool {
	if s1.Len() != s2.Len() {
		return false
	}
	for val := range s1.set {
		if _, ok := s2.set[val]; !ok {
			return false
		}
	}
	return true
}

func (s *Set[comparable]) Iter() []comparable {
	keys := make([]comparable, 0, s.Len())
	for k := range s.set {
		keys = append(keys, k)
	}
	return keys
}

func (s1 Set[comparable]) Union(s2 Set[comparable]) Set[comparable] {
	s3 := New[comparable]()

	if s1.Len() > 0 {
		for _, val := range s1.Iter() {
			s3.Add(val)
		}
	}

	if s2.Len() > 0 {
		for _, val := range s2.Iter() {
			s3.Add(val)
		}
	}

	return s3
}

func (s1 Set[comparable]) Intersection(s2 Set[comparable]) Set[comparable] {
	s3 := New[comparable]()

	if s1.Len() <= s2.Len() {
		for _, val := range s1.Iter() {
			if s2.Contains(val) {
				s3.Add(val)
			}
		}
	} else {
		for _, val := range s2.Iter() {
			if s1.Contains(val) {
				s3.Add(val)
			}
		}
	}
	return s3

}

func (s1 *Set[comparable]) Difference(s2 *Set[comparable]) {}

func (s1 *Set[comparable]) Complement(s2 *Set[comparable]) {}
