// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"stl"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		name := os.Args[0]
		fmt.Printf("Usage: %v \"EXPRESSION\"\n", name)
		fmt.Printf("Example: %v \"( 1 - -3 ) / 3 + 2 * ( 3 + -4 ) + 3 %% 2^2\"\n         =2\n", name)
		os.Exit(1)
	}
	filename := os.Args[1]
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	stl := &stl.Stl{Buffer: string(data)}
	stl.Init()
	if err := stl.Parse(); err != nil {
		log.Fatal(err)
	}
	//	fmt.Printf("= %v\n", &stl.StlFile)
}
