// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"runtime"
)

var (
	inline = flag.Bool("inline", false, "parse rule inlining")
	_switch = flag.Bool("switch", false, "replace if-else if-else like blocks with switch blocks")
	syntax = flag.Bool("syntax", false, "print out the syntax tree")
	highlight = flag.Bool("highlight", false, "test the syntax highlighter")
)

func main() {
	runtime.GOMAXPROCS(2)
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		log.Fatalf("FILE: the peg file to compile")
	}
	file := flag.Arg(0)

	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	p := &Peg{Tree: New(*inline, *_switch), Buffer: string(buffer)}
	p.Init()
	if err := p.Parse(); err != nil {
		log.Fatal(err)
	}

	if *syntax {
		p.PrintSyntaxTree()
	}
	if *highlight {
		p.Highlighter()
	}
	filename := file + ".go"
	p.Compile(filename)
}
