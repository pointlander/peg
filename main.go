// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pointlander/peg/tree"
)

//go:generate ./bootstrap.bash
//go:generate ./generate-grammars.bash

var (
	Version = "dev"

	inline      = flag.Bool("inline", false, "parse rule inlining")
	switchFlag  = flag.Bool("switch", false, "replace if-else if-else like blocks with switch blocks")
	printFlag   = flag.Bool("print", false, "directly dump the syntax tree")
	syntax      = flag.Bool("syntax", false, "print out the syntax tree")
	noast       = flag.Bool("noast", false, "disable AST")
	strict      = flag.Bool("strict", false, "treat compiler warnings as errors")
	outputFile  = flag.String("output", "", "output to `FILE` (\"-\" for stdout)")
	showVersion = flag.Bool("version", false, "print the version and exit")
)

// main is the entry point for the PEG compiler.
func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println("version:", Version)
		return
	}

	err := parse(
		func(p *Peg[uint32], out io.Writer) error {
			if *printFlag {
				p.Print()
			}
			if *syntax {
				p.PrintSyntaxTree()
			}

			p.Strict = *strict
			if err := p.Compile(*outputFile, os.Args, out); err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		if *strict {
			log.Fatal(err)
		}
		fmt.Fprintln(os.Stderr, "warning:", err)
	}
}

// getIO returns input and output streams based on command-line flags.
func getIO() (in io.ReadCloser, out io.WriteCloser, err error) {
	in, out = os.Stdin, os.Stdout

	if flag.NArg() > 0 && flag.Arg(0) != "-" {
		in, err = os.Open(flag.Arg(0))
		if err != nil {
			return nil, nil, err
		}
		if *outputFile == "" {
			*outputFile = flag.Arg(0) + ".go"
		}
	}

	if *outputFile != "" && *outputFile != "-" {
		out, err = os.OpenFile(*outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			if in != nil && in != os.Stdin {
				err := in.Close()
				if err != nil {
					panic(err)
				}
			}
			return nil, nil, err
		}
	}

	return in, out, nil
}

// parse reads input, parses, executes, and compiles the PEG grammar.
func parse(compile func(*Peg[uint32], io.Writer) error) error {
	in, out, err := getIO()
	if err != nil {
		return err
	}
	defer func() {
		if in != nil && in != os.Stdin {
			err := in.Close()
			if err != nil {
				panic(err)
			}
		}
		if out != nil && out != os.Stdout {
			err := out.Close()
			if err != nil {
				panic(err)
			}
		}
	}()

	buffer, err := io.ReadAll(in)
	if err != nil {
		return err
	}

	p := &Peg[uint32]{Tree: tree.New(*inline, *switchFlag, *noast), Buffer: string(buffer)}
	_ = p.Init(Pretty[uint32](true), Size[uint32](1<<15))
	if err = p.Parse(); err != nil {
		return err
	}

	p.Execute()

	return compile(p, out)
}
