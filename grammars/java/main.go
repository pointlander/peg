// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("%v FILE\n", os.Args[0])
		os.Exit(1)
	}

	var walk func(name string)
	walk = func(name string) {
		fileInfo, err := os.Stat(name)
		if err != nil {
			log.Fatal(err)
		}

		if fileInfo.Mode() & (os.ModeNamedPipe | os.ModeSocket | os.ModeDevice) != 0 {
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
	walk(os.Args[1])
}
