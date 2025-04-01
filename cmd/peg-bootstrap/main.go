// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build bootstrap
// +build bootstrap

package main

import (
	"io"
	"log"
	"os"

	"github.com/pointlander/peg/tree"
)

func main() {
	buffer, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	p := &Peg[uint32]{Tree: tree.New(false, false, false), Buffer: string(buffer)}
	p.Init(Pretty[uint32](true), Size[uint32](1<<15))
	if err := p.Parse(); err != nil {
		log.Fatal(err)
	}
	p.Execute()
	p.Compile("boot.peg.go", os.Args, os.Stdout)
}
