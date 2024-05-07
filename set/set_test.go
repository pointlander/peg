// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package set

import (
	"math"
	"testing"
)

func TestString(t *testing.T) {
	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')

	if s.String() != "[97 98 99 100 101]" {
		t.Fatal("string is broken")
	}
}

func TestCopy(t *testing.T) {
	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')

	cp := s.Copy()
	if !cp.Equal(s) {
		t.Fatal("cp should be a copy of s")
	}
}

func TestAdd(t *testing.T) {
	s := NewSet()
	s.Add('a')

	if s.Len() != 1 {
		t.Fatal("length should be 1", s.Len())
	}

	if !s.Has('a') {
		t.Fatal("set should have a")
	}

	s.Add('c')
	s.Add('e')
	s.Add('A')
}

func TestAddRange(t *testing.T) {
	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')
	if s.Len() != 5 {
		t.Fatal("size should be 5")
	}
	if !s.Has('b') {
		t.Fatal("set should have b")
	}
	if !s.Has('d') {
		t.Fatal("set should have d")
	}

	s.AddRange('g', 'i')
	if s.Len() != 8 {
		t.Log(s.Len())
		t.Fatal("size should be 7")
	}
	if !s.Has('h') {
		t.Fatal("set should have h")
	}

	s.AddRange('A', 'C')
	if s.Len() != 11 {
		t.Log(s.Len())
		t.Fatal("size should be 10")
	}
	if !s.Has('B') {
		t.Fatal("set should have B")
	}

	s.AddRange('A', 'z')
	if s.Len() != 'z'-'A'+1 {
		t.Log(s.Len())
		t.Fatalf("size should be %d", 'z'-'A'+1)
	}
	if !s.Has('B') {
		t.Fatal("set should have B")
	}
}

func TestHas(t *testing.T) {
	r := NewSet()
	r.AddRange('a', 'c')

	if !r.Has('b') {
		t.Fatal("set should have b")
	}

	if r.Has('d') {
		t.Fatal("set should not have d")
	}
}

func TestComplement(t *testing.T) {
	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')
	s.AddRange('g', 'i')
	s.AddRange('A', 'C')
	c1 := s.Complement(rune(math.MaxInt32))
	c2 := c1.Complement(rune(math.MaxInt32))
	if !s.Equal(c2) {
		t.Fatal("sets should be equal")
	}
}

func TestUnion(t *testing.T) {
	r := NewSet()
	r.AddRange('a', 'c')
	r.AddRange('c', 'e')

	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')
	s.AddRange('g', 'i')
	s.AddRange('A', 'C')

	z := NewSet()
	z.AddRange('g', 'i')
	z.AddRange('A', 'C')

	z = r.Union(z)

	if !z.Equal(s) {
		t.Fatal("sets should be equal")
	}
}

func TestIntersects(t *testing.T) {
	r := NewSet()
	r.AddRange('a', 'c')

	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')
	s.AddRange('g', 'i')
	s.AddRange('A', 'C')

	if !r.Intersects(s) {
		t.Fatal("sets should intersect")
	}

	z := NewSet()
	z.Add('z')

	if z.Intersects(s) {
		t.Fatal("sets should not intersect")
	}
}

func TestEqual(t *testing.T) {
	r := NewSet()
	r.AddRange('a', 'c')
	r.AddRange('c', 'e')
	r.AddRange('g', 'i')

	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')
	s.AddRange('g', 'i')
	s.AddRange('A', 'C')

	if r.Equal(s) {
		t.Fatal("sets should not be equal")
	}

	r.AddRange('A', 'C')

	if !r.Equal(s) {
		t.Fatal("sets should be equal")
	}
}

func TestLen(t *testing.T) {
	r := NewSet()
	r.AddRange('a', 'c')
	r.AddRange('c', 'e')
	r.AddRange('g', 'i')

	s := NewSet()
	s.AddRange('a', 'c')
	s.AddRange('c', 'e')
	s.AddRange('g', 'i')
	s.AddRange('A', 'C')

	if r.Len() == s.Len() {
		t.Fatal("sets should not be equal in length")
	}

	r.AddRange('A', 'C')

	if r.Len() != s.Len() {
		t.Fatal("sets should be equal in length")
	}
}
