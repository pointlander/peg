// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build grammars
// +build grammars

package main

import (
	"os"
	"testing"
)

func TestFexl(t *testing.T) {
	buffer, err := os.ReadFile("doc/try.fxl")
	if err != nil {
		t.Fatal(err)
	}

	fexl := &Fexl{Buffer: string(buffer)}
	fexl.Init()

	if err := fexl.Parse(); err != nil {
		t.Fatal(err)
	}
}
