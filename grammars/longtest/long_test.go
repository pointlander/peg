// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate ../../peg -switch -inline long.peg

package longtest

import (
	"testing"
)

func TestLong(t *testing.T) {
	length := 100000
	if testing.Short() {
		length = 100
	}

	expression := ""
	long := &Long{Buffer: "\"" + expression + "\""}
	err := long.Init()
	if err != nil {
		t.Fatal(err)
	}
	for c := 0; c < length; c++ {
		if err := long.Parse(); err != nil {
			t.Fatal(err)
		}
		long.Reset()
		expression = expression + "X"
		long.Buffer = "\"" + expression + "\""
	}
}
