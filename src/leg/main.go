// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"time"
)

var (
	inline = flag.Bool("inline", false, "parse rule inlining")
	_switch = flag.Bool("switch", false, "replace if-else if-else like blocks with switch blocks")
	syntax = flag.Bool("syntax", false, "print out the syntax tree")
	highlight = flag.Bool("highlight", false, "test the syntax highlighter")
	test = flag.Bool("test", false, "test the LEG parser performance")
	print = flag.Bool("print", false, "directly dump the syntax tree")
)

func main() {
	runtime.GOMAXPROCS(2)
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		log.Fatalf("FILE: the leg file to compile")
	}
	file := flag.Arg(0)

	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	if *test {
		iterations, p := 1000, &Leg{Tree: New(*inline, *_switch), Buffer: string(buffer)}
		p.Init()
		start := time.Now()
		for i := 0; i < iterations; i++ {
			p.Parse()
			p.Reset()
		}
		total := float64(time.Since(start).Nanoseconds()) / float64(1000)
		fmt.Printf("time: %v us\n", total / float64(iterations))
		return
	}

	p := &Leg{Tree: New(*inline, *_switch), Buffer: string(buffer)}
	p.Init()
	if err := p.Parse(); err != nil {
		log.Fatal(err)
	}

	p.Execute()

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
	p.Compile(filename)
}
