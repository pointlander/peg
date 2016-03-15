// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	inline    = flag.Bool("inline", false, "parse rule inlining")
	_switch   = flag.Bool("switch", false, "replace if-else if-else like blocks with switch blocks")
	syntax    = flag.Bool("syntax", false, "print out the syntax tree")
	highlight = flag.Bool("highlight", false, "test the syntax highlighter")
	ast       = flag.Bool("ast", false, "generate an AST")
	test      = flag.Bool("test", false, "test the PEG parser performance")
	print     = flag.Bool("print", false, "directly dump the syntax tree")
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

	if *test {
		iterations, p := 1000, &Peg{Tree: New(*inline, *_switch), Buffer: string(buffer)}
		p.Init()
		start := time.Now()
		for i := 0; i < iterations; i++ {
			p.Parse()
			p.Reset()
		}
		total := float64(time.Since(start).Nanoseconds()) / float64(1000)
		fmt.Printf("time: %v us\n", total/float64(iterations))
		return
	}

	p := &Peg{Tree: New(*inline, *_switch), Buffer: string(buffer), Pretty: true}
	p.Init()
	if err := p.Parse(); err != nil {
		log.Fatal(err)
	}

	p.Execute()

	if *ast {
		p.tokenTree.AST().Print(p.Buffer)
	}
	if *print {
		p.Print()
	}
	if *syntax {
		p.PrintSyntaxTree()
	}
	if *highlight {
		p.Highlighter()
	}

	filename := file + ".go"
	out, error := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if error != nil {
		fmt.Printf("%v: %v\n", filename, error)
		return
	}
	defer out.Close()
	p.Compile(filename, out)
}
