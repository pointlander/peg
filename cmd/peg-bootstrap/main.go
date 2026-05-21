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
	log.SetFlags(0)

	buffer, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	p := &Peg[uint32]{Tree: tree.New(false, false, false), Buffer: string(buffer)}
	if err := p.Init(Pretty[uint32](true), Size[uint32](1<<15)); err != nil {
		log.Fatal("Init:", err)
	}
	if err := p.Parse(); err != nil {
		log.Fatal("Parse:", err)
	}
	p.Execute()
	if err := p.Compile("boot.peg.go", os.Args, os.Stdout); err != nil {
		log.Fatal("Compile:", err)
	}
}
