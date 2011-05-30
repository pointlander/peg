// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stl

import (
	"math"
	"strconv"
)

type Vector [3]float64

type Triangle struct {
	Normal   Vector
	Vertices [3]Vector
}

type StlFile struct {
	Name      string
	EndName   string
	Triangles []Triangle
	t         *Triangle
	v         *Vector
	ind       int
	vind      int
}

func (s *StlFile) Add() {
	var t Triangle
	s.Triangles = append(s.Triangles, t)
	s.t = &s.Triangles[len(s.Triangles)-1]
	s.v = &s.t.Normal
	s.ind = 0
	s.vind = 0
}

func (s *StlFile) Num(str string) {
	num, err := strconv.Atof64(str)
	if err != nil {
		num = math.NaN()
	}
	s.v[s.ind] = num
	s.ind++
}

func (s *StlFile) Vertex() {
	s.v = &s.t.Vertices[s.vind]
	s.ind = 0
	s.vind++
}
