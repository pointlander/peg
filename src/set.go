package main

/* Used to represent character classes. */
type set struct {
	predicates node
	elements [32]uint8
}

func (s *set) copy() *set {
	t := &set{}
	copy(t.elements[0:], s.elements[0:])
	return t
}

func (s *set) add(element uint8) {
	s.elements[element >> 3] |= (1 << (element & 7))
}

func (s *set) has(element uint8) bool {
	return s.elements[element >> 3] & (1 << (element & 7)) != 0
}

func (s *set) complement() {
	for i := range s.elements {
		s.elements[i] = ^s.elements[i]
	}
}

func (s *set) union(t *set) {
	for i, element := range t.elements {
		s.elements[i] |= element
	}
}

func (s *set) intersection(t *set) {
	for i, element := range t.elements {
		s.elements[i] &= element
	}
}

func (s *set) intersects(t *set) bool {
	for i, element := range s.elements {
		if (t.elements[i] & element) != 0 {
			return true
		}
	}
	return false
}

func (s *set) len() (length int) {
	for element := 0; element < 256; element++ {
		if s.has(uint8(element)) {
			length++
		}
	}
	return
}
