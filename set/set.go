// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set

import (
	"fmt"
	"math"
)

// Node is a node
type Node struct {
	Forward  *Node
	Backward *Node
	Begin    rune
	End      rune
}

// Set is a set
type Set struct {
	Head Node
	Tail Node
}

// NewSet returns a new set
func NewSet() Set {
	return Set{
		Head: Node{
			Begin: math.MaxInt32,
		},
	}
}

// String returns the string of a set
func (s Set) String() string {
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

// Copy copies a set
func (s Set) Copy() Set {
	set := NewSet()
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

// Add adds a symbol to the set
func (s *Set) Add(a rune) {
	s.AddRange(a, a)
}

// AddRange adds to a set
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
		return
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

// Has tests if a set has a rune
func (s Set) Has(begin rune) bool {
	beginNode := &s.Head
	for beginNode.Forward != nil && begin > beginNode.Forward.End {
		beginNode = beginNode.Forward
	}
	if beginNode.Forward == nil {
		return false
	}
	return begin >= beginNode.Forward.Begin
}

// Complement computes the complement of a set
func (s Set) Complement() Set {
	set := NewSet()
	if s.Len() == 0 {
		node := Node{
			Forward:  &set.Head,
			Backward: &set.Tail,
			Begin:    0,
			End:      rune(math.MaxInt32),
		}
		set.Head.Forward = &node
		set.Tail.Backward = &node
		return set
	}
	if s.Len() == 1 && s.Head.Forward.Begin == 0 && s.Head.Forward.End == rune(math.MaxInt32) {
		return set
	}
	a, b := s.Head.Forward, &set.Head
	if a.Begin == 0 && s.Tail.Backward.End == rune(math.MaxInt32) {
		pre := a.End + 1
		a = a.Forward
		for a.Forward != nil {
			node := Node{
				Backward: b,
				Begin:    pre,
				End:      a.Begin - 1,
			}
			pre = a.End + 1
			b.Forward = &node
			a = a.Forward
			b = b.Forward
		}
		b.Forward = &set.Tail
		set.Tail.Backward = b
	} else {
		pre := rune(0)
		for a.Forward != nil {
			node := Node{
				Backward: b,
				Begin:    pre,
				End:      a.Begin - 1,
			}
			pre = a.End + 1
			b.Forward = &node
			a = a.Forward
			b = b.Forward
		}
		node := Node{
			Backward: b,
			Begin:    pre,
			End:      rune(math.MaxInt32),
		}
		b.Forward = &node
		b = b.Forward
		b.Forward = &set.Tail
		set.Tail.Backward = b
	}
	return set
}

// Union is the union of two sets
func (s Set) Union(a Set) Set {
	set := s.Copy()
	node := a.Head.Forward
	for node.Forward != nil {
		set.AddRange(node.Begin, node.End)
		node = node.Forward
	}
	return set
}

// Intersects returns true if two sets intersect
func (a Set) Intersects(b Set) bool {
	x := &a.Head
	for x.Forward != nil {
		y := &b.Head
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

// Equal returns true if two sets are equal
func (s Set) Equal(a Set) bool {
	lens, lena := s.Len(), a.Len()
	if lens != lena {
		return false
	} else if lens == 0 && lena == 0 {
		return true
	}
	x, y := s.Head.Forward, a.Head.Forward
	for {
		if x.Begin != y.Begin || x.End != y.End {
			fmt.Println(x.Begin, x.End, y.Begin, y.End)
			return false
		}
		x, y = x.Forward, y.Forward
		if x == nil && y == nil {
			break
		}
	}
	return true
}

// Len returns the size of the set
func (s Set) Len() int {
	size := 0
	beginNode := &s.Head
	for beginNode.Forward != nil {
		beginNode = beginNode.Forward
		size++
	}
	return size - 1
}
