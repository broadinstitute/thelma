package set

type Set[K comparable] interface {
	// Add adds the element(s) to the set
	Add(elements ...K)
	// Remove removes the element(s) from the set, if they are included
	Remove(elements ...K)
	// Exists returns true if the element exists in the set
	Exists(element K) bool
	// Elements returns the elements of the set as a slice of strings
	Elements() []K
	// Difference returns a new set containing the elements that are in this set but not in s
	Difference(s Set[K]) Set[K]
	// Size returns the size of the set
	Size() int
	// Empty returns true if this set is empty
	Empty() bool
}

type set[K comparable] struct {
	set map[K]interface{}
}

// NewSet returns a new Set for the given comparable type
func NewSet[K comparable](elements ...K) Set[K] {
	s := &set[K]{
		set: make(map[K]interface{}),
	}
	s.Add(elements...)
	return s
}

// Add adds the element(s) to the set
func (s *set[K]) Add(elements ...K) {
	for _, e := range elements {
		s.set[e] = struct{}{}
	}
}

// Remove removes the element(s) from the set, if they are included
func (s *set[K]) Remove(elements ...K) {
	for _, e := range elements {
		delete(s.set, e)
	}
}

// Exists returns true if the element exists in the set
func (s *set[K]) Exists(element K) bool {
	_, exists := s.set[element]
	return exists
}

// Difference returns a new set containing the elements that are in this set but not in other
func (s *set[K]) Difference(other Set[K]) Set[K] {
	diff := NewSet[K]()
	for e := range s.set {
		if !other.Exists(e) {
			diff.Add(e)
		}
	}
	return diff
}

// Elements returns the elements of the set as a slice of strings
func (s *set[K]) Elements() []K {
	elements := make([]K, s.Size())

	i := 0
	for e := range s.set {
		elements[i] = e
		i++
	}

	return elements
}

// Size returns the size of the set
func (s *set[K]) Size() int {
	return len(s.set)
}

// Empty returns true if the set is empty
func (s *set[K]) Empty() bool {
	return s.Size() == 0
}
