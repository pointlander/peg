// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
)

func main() {
	expression := ""
	long := &Long{Buffer: "\"" + expression + "\""}
	long.Init()
	for c := 0; c < 100000; c++ {
		if err := long.Parse(); err != nil {
			fmt.Printf("%v\n", c)
			log.Fatal(err)
		}
		long.Reset()
		expression = expression + "X"
		long.Buffer = "\"" + expression + "\""
	}
}
