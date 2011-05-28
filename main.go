// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"peg"
	"fmt"
	"io/ioutil"
	"runtime"
	"flag"
	"os"
	"path"
)

var inline = flag.Bool("inline", false, "parse rule inlining")
var _switch = flag.Bool("switch", false, "replace if-else if-else like blocks with switch blocks")

func gofmt(filename string) {
     gobin := os.Getenv("GOBIN")
     if gobin == "" {
     return
}
	p, err := os.StartProcess(path.Join(gobin, "gofmt"), []string{"gofmt", "-w", filename},
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err != nil {
		return
	}
	p.Wait(0)
}

func main() {
	runtime.GOMAXPROCS(2)
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "  FILE: the peg file to compile\n")
		os.Exit(1)
	}
	file := flag.Arg(0)

	buffer, error := ioutil.ReadFile(file)
	if error != nil {
		fmt.Printf("%v\n", error)
		return
	}
	p := &peg.Peg{Tree: peg.New(*inline, *_switch), Buffer: string(buffer)}
	p.Init()
	if p.Parse() {
		filename := file + ".go"
		fmt.Printf("Compiling...\n")
		p.Compile(filename)
		fmt.Printf("Formatting...\n")
		gofmt(filename)
		fmt.Printf("OK\n")

	} else {
		p.PrintError()
	}
}
