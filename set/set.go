// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set

import (
	"fmt"
	"math"
)

// Node is a node.
type Node struct {
	Forward  *Node
	Backward *Node
	Begin    rune
	End      rune
}

// Set is a set.
type Set struct {
	Head Node
	Tail Node
}

// NewSet returns a new set.
func NewSet() *Set {
	return &Set{
		Head: Node{
			Begin: math.MaxInt32,
		},
	}
}

// String returns the string of a set.
func (s *Set) String() string {
	codes, space := "[", ""
	node := s.Head.Forward
	for node.Forward != nil {
		for code := node.Begin; code <= node.End; code++ {
			codes += space + fmt.Sprintf("%v", code)
			space = " "
		}
		node = node.Forward
	}
	return codes + "]"
}

// Copy copies a set.
func (s *Set) Copy() *Set {
	set := NewSet()
	if s.Head.Forward == nil {
		return set
	}
	a, b := s.Head.Forward, &set.Head
	for a.Forward != nil {
		node := Node{
			Backward: b,
			Begin:    a.Begin,
			End:      a.End,
		}
		b.Forward = &node
		a = a.Forward
		b = b.Forward
	}
	b.Forward = &set.Tail
	set.Tail.Backward = b
	return set
}

// Add adds a symbol to the set.
func (s *Set) Add(a rune) {
	s.AddRange(a, a)
}

// AddRange adds to a set.
func (s *Set) AddRange(begin, end rune) {
	beginNode := &s.Head
	for beginNode.Forward != nil && begin > beginNode.Forward.End {
		beginNode = beginNode.Forward
	}
	endNode := &s.Tail
	for endNode.Backward != nil && end < endNode.Backward.Begin {
		endNode = endNode.Backward
	}
	if beginNode.Forward == nil && endNode.Backward == nil {
		node := Node{
			Begin: begin,
			End:   end,
		}
		node.Forward = endNode
		endNode.Backward = &node
		node.Backward = beginNode
		beginNode.Forward = &node
	} else if beginNode.Forward == endNode.Backward {
		if begin < beginNode.Forward.Begin {
			beginNode.Forward.Begin = begin
		}
		if end > beginNode.Forward.End {
			beginNode.Forward.End = end
		}
	} else if beginNode.Forward != nil && endNode.Backward == nil {
		node := Node{
			Begin: begin,
			End:   end,
		}
		node.Backward = beginNode
		node.Forward = beginNode.Forward
		beginNode.Forward.Backward = &node
		beginNode.Forward = &node
	} else if beginNode.Forward == nil && endNode.Backward != nil {
		node := Node{
			Begin: begin,
			End:   end,
		}
		node.Forward = endNode
		node.Backward = endNode.Backward
		endNode.Backward.Forward = &node
		endNode.Backward = &node
	} else if beginNode.Forward == endNode {
		node := Node{
			Begin: begin,
			End:   end,
		}
		node.Backward = beginNode
		node.Forward = beginNode.Forward
		beginNode.Forward.Backward = &node
		beginNode.Forward = &node
	} else if beginNode == endNode.Backward {
		node := Node{
			Begin: begin,
			End:   end,
		}
		node.Forward = endNode
		node.Backward = endNode.Backward
		endNode.Backward.Forward = &node
		endNode.Backward = &node
	} else {
		if begin < beginNode.Forward.Begin {
			beginNode.Forward.Begin = begin
		}
		if end > endNode.Backward.End {
			beginNode.Forward.End = end
		} else {
			beginNode.Forward.End = endNode.Backward.End
		}
		node := beginNode.Forward
		node.Forward = endNode
		endNode.Backward = node
	}
}

// Has tests if a set has a rune.
func (s *Set) Has(begin rune) bool {
	beginNode := &s.Head
	for beginNode.Forward != nil && begin > beginNode.Forward.End {
		beginNode = beginNode.Forward
	}
	if beginNode.Forward == nil {
		return false
	}
	return begin >= beginNode.Forward.Begin
}

// Complement computes the complement of a set.
func (s *Set) Complement(endSymbol rune) *Set {
	set := NewSet()
	if s.Len() == 0 {
		node := Node{
			Forward:  &set.Tail,
			Backward: &set.Head,
			Begin:    0,
			End:      endSymbol,
		}
		set.Head.Forward = &node
		set.Tail.Backward = &node
		return set
	}
	if s.Head.Forward.Begin == 0 && s.Head.Forward.End == endSymbol {
		return set
	}
	a, b := &s.Head, &set.Head
	pre := rune(0)
	if pre == a.Forward.Begin {
		a = a.Forward
		pre = a.End + 1
	}
	a = a.Forward
	for a.Forward != nil {
		node := Node{
			Backward: b,
			Begin:    pre,
			End:      a.Begin - 1,
		}
		if a.End == endSymbol {
			pre = endSymbol
		} else {
			pre = a.End + 1
		}
		b.Forward = &node
		a = a.Forward
		b = b.Forward
	}
	if pre < endSymbol {
		node := Node{
			Backward: b,
			Begin:    pre,
			End:      endSymbol,
		}
		b.Forward = &node
		b = b.Forward
	}
	b.Forward = &set.Tail
	set.Tail.Backward = b
	return set
}

// Union is the union of two sets.
func (s *Set) Union(a *Set) *Set {
	set := s.Copy()
	node := a.Head.Forward
	if node == nil {
		return set
	}
	for node.Forward != nil {
		set.AddRange(node.Begin, node.End)
		node = node.Forward
	}
	return set
}

// Intersects returns true if two sets intersect.
func (s *Set) Intersects(b *Set) bool {
	x := s.Head.Forward
	if x == nil {
		return false
	}
	for x.Forward != nil {
		y := b.Head.Forward
		if y == nil {
			return false
		}
		for y.Forward != nil {
			if y.Begin >= x.Begin && y.Begin <= x.End {
				return true
			} else if y.End >= x.Begin && y.End <= x.End {
				return true
			}
			y = y.Forward
		}
		x = x.Forward
	}
	x = b.Head.Forward
	if x == nil {
		return false
	}
	for x.Forward != nil {
		y := s.Head.Forward
		if y == nil {
			return false
		}
		for y.Forward != nil {
			if y.Begin >= x.Begin && y.Begin <= x.End {
				return true
			} else if y.End >= x.Begin && y.End <= x.End {
				return true
			}
			y = y.Forward
		}
		x = x.Forward
	}
	return false
}

// Equal returns true if two sets are equal.
func (s *Set) Equal(a *Set) bool {
	lens, lena := s.Len(), a.Len()
	if lens != lena {
		return false
	} else if lens == 0 && lena == 0 {
		return true
	}
	x, y := s.Head.Forward, a.Head.Forward
	for {
		if x.Begin != y.Begin || x.End != y.End {
			return false
		}
		x, y = x.Forward, y.Forward
		if x == nil && y == nil {
			break
		}
	}
	return true
}

// Len returns the size of the set.
func (s *Set) Len() int {
	size := 0
	if s.Head.Forward == nil {
		return size
	}
	beginNode := s.Head.Forward
	for beginNode.Forward != nil {
		size += int(beginNode.End) - int(beginNode.Begin) + 1
		beginNode = beginNode.Forward
	}
	return size
}
