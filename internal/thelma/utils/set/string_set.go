package set

type StringSet interface {
	// Add adds the element(s) to the set
	Add(elements ...string)
	// Remove removes the element(s) from the set, if they are included
	Remove(elements ...string)
	// Exists returns true if the element exists in the set
	Exists(element string) bool
	// Elements returns the elements of the set as a slice of strings
	Elements() []string
	// Difference returns a new set containing the elements that are in this set but not in s
	Difference(s StringSet) StringSet
	// Size returns the size of the set
	Size() int
	// Empty returns true if this set is empty
	Empty() bool
}

type stringSet struct {
	set map[string]interface{}
}

// NewStringSet returns a new StringSet
func NewStringSet(elements ...string) StringSet {
	s := &stringSet{
		set: make(map[string]interface{}),
	}
	s.Add(elements...)
	return s
}

// Add adds the element(s) to the set
func (s *stringSet) Add(elements ...string) {
	for _, e := range elements {
		s.set[e] = struct{}{}
	}
}

// Remove removes the element(s) from the set, if they are included
func (s *stringSet) Remove(elements ...string) {
	for _, e := range elements {
		delete(s.set, e)
	}
}

// Exists returns true if the element exists in the set
func (s *stringSet) Exists(element string) bool {
	_, exists := s.set[element]
	return exists
}

// Difference returns a new set containing the elements that are in this set but not in other
func (s *stringSet) Difference(other StringSet) StringSet {
	diff := NewStringSet()
	for e := range s.set {
		if !other.Exists(e) {
			diff.Add(e)
		}
	}
	return diff
}

// Elements returns the elements of the set as a slice of strings
func (s *stringSet) Elements() []string {
	elements := make([]string, s.Size())

	i := 0
	for e := range s.set {
		elements[i] = e
		i++
	}

	return elements
}

// Size returns the size of the set
func (s *stringSet) Size() int {
	return len(s.set)
}

// Empty returns true if the set is empty
func (s *stringSet) Empty() bool {
	return s.Size() == 0
}
