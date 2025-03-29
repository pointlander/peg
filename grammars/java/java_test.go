// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate ../../peg -switch -inline java_1_7.peg

package java

import (
	"io/fs"
	"os"
	"path/filepath"
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
	err := java.Init()
	if err != nil {
		t.Fatal(err)
	}

	if err := java.Parse(); err != nil {
		t.Fatal(err)
	}
}

func TestJavaFiles(t *testing.T) {
	err := filepath.Walk(".", func(path string, _ fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".java" {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			java := &Java{Buffer: string(b)}
			err = java.Init()
			if err != nil {
				t.Fatal(err)
			}
			if err := java.Parse(); err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
