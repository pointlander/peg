// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stl

import (
	"io/ioutil"
	"os"
	"testing"
)

func read(t *testing.T, filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(data)
}

func load(t *testing.T, filename string) (stl *Stl) {
	stl = &Stl{Buffer: read(t, filename)}
	stl.Init()
	return
}

func parse(t *testing.T, filename string) (stl *Stl, err os.Error) {
	stl = load(t, filename)
	err = stl.Parse()
	return
}

func TestParse(t *testing.T) {
	tests := []string{
		"test.stl",
	}
	for _, test := range tests {
		if _, err := parse(t, test); err != nil {
			t.Errorf("Parse: %v", err)
		}
	}
}

func TestFail(t *testing.T) {
	tests := []string{
		"test_fail.stl",
	}
	for _, test := range tests {
		if _, err := parse(t, test); err == nil {
			t.Errorf("Unexpected 'successful' parse of %s", test)
		}
	}
}
