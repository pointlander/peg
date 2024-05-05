// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set

import (
	"math"
	"testing"
)

func TestAddBasic(t *testing.T) {
	s := Set{
		Head: Node{
			Begin: math.MaxInt32,
		},
	}
	s.Add('a', 'c')
	s.Add('c', 'e')
	if s.Size() != 1 {
		t.Fatal("size should be 1")
	}
	if !s.Has('b') {
		t.Fatal("set should have b")
	}
	if !s.Has('d') {
		t.Fatal("set should have d")
	}

	s.Add('g', 'i')
	if s.Size() != 2 {
		t.Log(s.Size())
		t.Fatal("size should be 2")
	}
	if !s.Has('h') {
		t.Fatal("set should have h")
	}

	s.Add('A', 'C')
	if s.Size() != 3 {
		t.Log(s.Size())
		t.Fatal("size should be 3")
	}
	if !s.Has('B') {
		t.Fatal("set should have B")
	}

	s.Add('A', 'z')
	if s.Size() != 1 {
		t.Log(s.Size())
		t.Fatal("size should be 1")
	}
	if !s.Has('B') {
		t.Fatal("set should have B")
	}
}
