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

	"github.com/pointlander/peg/tree"
)

//go:generate -command build go run build.go
//go:generate build buildinfo
//go:generate build peg

var (
	inline   = flag.Bool("inline", false, "parse rule inlining")
	_switch  = flag.Bool("switch", false, "replace if-else if-else like blocks with switch blocks")
	print    = flag.Bool("print", false, "directly dump the syntax tree")
	syntax   = flag.Bool("syntax", false, "print out the syntax tree")
	noast    = flag.Bool("noast", false, "disable AST")
	strict   = flag.Bool("strict", false, "treat compiler warnings as errors")
	filename = flag.String("output", "", "specify name of output file")
	showVersion = flag.Bool("version", false, "print the version and exit")
)

// whether running with -version should
// show the last time `build.go buildinfo` was ran 
const Show_BUILDTIME = false

func main() {
	runtime.GOMAXPROCS(2)
	flag.Parse()

	if *showVersion {
		if IS_TAGGED {
			fmt.Println("version:",VERSION)
		} else {
			fmt.Printf("version: %s-%s\n",VERSION,COMMIT)
		}
		if Show_BUILDTIME {fmt.Println("time:",BUILDTIME)}
		return
	}
	
	if flag.NArg() != 1 {
		flag.Usage()
		log.Fatalf("FILE: the peg file to compile")
	}
	file := flag.Arg(0)

	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	p := &Peg{Tree: tree.New(*inline, *_switch, *noast), Buffer: string(buffer)}
	p.Init(Pretty(true), Size(1<<15))
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

	if *filename == "" {
		*filename = file + ".go"
	}
	out, err := os.OpenFile(*filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("%v: %v\n", *filename, err)
		return
	}
	defer out.Close()

	p.Strict = *strict
	if err = p.Compile(*filename, os.Args, out); err != nil {
		log.Fatal(err)
	}
}
