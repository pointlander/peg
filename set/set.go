// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set

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

// Add adds to a set
func (s *Set) Add(begin, end rune) {
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

// Size returns the size of the set
func (s *Set) Size() int {
	size := 0
	beginNode := &s.Head
	for beginNode.Forward != nil {
		beginNode = beginNode.Forward
		size++
	}
	return size - 1
}

// Has tests if a set has a rune
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
