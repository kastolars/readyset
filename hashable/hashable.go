package hashable

type Hashable[T comparable] interface {
	Hash() T
}

type HashableSet[T comparable] struct {
	set map[T]Hashable[T]
}

func NewHashable[T comparable](elements ...Hashable[T]) HashableSet[T] {
	set := make(map[T]Hashable[T])
	for _, val := range elements {
		set[val.Hash()] = val
	}
	return HashableSet[T]{set}
}

func (s *HashableSet[T]) Add(element Hashable[T]) {
	s.set[element.Hash()] = element
}

func (s *HashableSet[T]) Remove(element Hashable[T]) {
	delete(s.set, element.Hash())
}

func (s HashableSet[T]) Contains(element Hashable[T]) bool {
	_, ok := s.set[element.Hash()]
	return ok
}

func (s HashableSet[T]) Len() int {
	return len(s.set)
}

func (s1 HashableSet[T]) Equal(s2 HashableSet[T]) bool {
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

func (s HashableSet[T]) Iter() []Hashable[T] {
	elements := make([]Hashable[T], 0, s.Len())
	for _, val := range s.set {
		elements = append(elements, val)
	}
	return elements
}

func (s1 HashableSet[T]) Union(s2 HashableSet[T]) HashableSet[T] {
	s3 := NewHashable[T]()

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

func (s1 HashableSet[T]) Intersection(s2 HashableSet[T]) HashableSet[T] {
	s3 := NewHashable[T]()

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

func (s1 HashableSet[T]) Difference(s2 HashableSet[T]) HashableSet[T] {
	s3 := NewHashable[T]()

	for _, val := range s1.Iter() {
		if !s2.Contains(val) {
			s3.Add(val)
		}
	}

	return s3
}
