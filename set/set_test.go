// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set

import (
	"testing"
)

func TestAddBasic(t *testing.T) {
	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')
	if s.Len() != 1 {
		t.Fatal("size should be 1")
	}
	if !s.Has('b') {
		t.Fatal("set should have b")
	}
	if !s.Has('d') {
		t.Fatal("set should have d")
	}

	s.AddRange('g', 'i')
	if s.Len() != 2 {
		t.Log(s.Len())
		t.Fatal("size should be 2")
	}
	if !s.Has('h') {
		t.Fatal("set should have h")
	}

	s.AddRange('A', 'C')
	if s.Len() != 3 {
		t.Log(s.Len())
		t.Fatal("size should be 3")
	}
	if !s.Has('B') {
		t.Fatal("set should have B")
	}

	s.AddRange('A', 'z')
	if s.Len() != 1 {
		t.Log(s.Len())
		t.Fatal("size should be 1")
	}
	if !s.Has('B') {
		t.Fatal("set should have B")
	}
}
