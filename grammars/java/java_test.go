// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build grammars

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

var example1 = `public class HelloWorld {
	public static void main(String[] args) {
		System.out.println("Hello, World");
	}
}
`

func TestBasic(t *testing.T) {
	java := &Java{Buffer: example1}
	java.Init()

	if err := java.Parse(); err != nil {
		t.Fatal(err)
	}
}

func TestJava(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping java parsing long test")
	}

	var walk func(name string)
	walk = func(name string) {
		fileInfo, err := os.Stat(name)
		if err != nil {
			log.Fatal(err)
		}

		if fileInfo.Mode()&(os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) != 0 {
			/* will lock up if opened */
		} else if fileInfo.IsDir() {
			fmt.Printf("directory %v\n", name)

			file, err := os.Open(name)
			if err != nil {
				log.Fatal(err)
			}

			files, err := file.Readdir(-1)
			if err != nil {
				log.Fatal(err)
			}
			file.Close()

			for _, f := range files {
				if !strings.HasSuffix(name, "/") {
					name += "/"
				}
				walk(name + f.Name())
			}
		} else if strings.HasSuffix(name, ".java") {
			fmt.Printf("parse %v\n", name)

			file, err := os.Open(name)
			if err != nil {
				log.Fatal(err)
			}

			buffer, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatal(err)
			}
			file.Close()

			java := &Java{Buffer: string(buffer)}
			java.Init()
			if err := java.Parse(); err != nil {
				log.Fatal(err)
			}
		}
	}
	walk("java/")
}
