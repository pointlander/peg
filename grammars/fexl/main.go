// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"log"
)

func main() {
	buffer, err := ioutil.ReadFile("doc/try.fxl")
	if err != nil {
		log.Fatal(err)
	}

	fexl := &Fexl{Buffer: string(buffer)}
	fexl.Init()

	if err := fexl.Parse(); err != nil {
		log.Fatal(err)
	}
	fexl.PrintSyntaxTree()
}
