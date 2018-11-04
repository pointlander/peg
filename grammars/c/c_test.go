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

func parseCBuffer(buffer string) (*C, error) {
	clang := &C{Buffer: buffer}
	clang.Init()
	err := clang.Parse()
	return clang, err
}

func parseC_4t(t *testing.T, src string) *C {
	c, err := parseCBuffer(src)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func noParseC_4t(t *testing.T, src string) {
	_, err := parseCBuffer(src)
	if err == nil {
		t.Fatal("Parsed what should not have parsed.")
	}
}

func TestCParsing_Expressions1(t *testing.T) {
	case1src :=
		`int a() {
		(es);
		1++;
		1+1;
		a+1;
		(a)+1;
		a->x;
		return 0;
}`
	parseC_4t(t, case1src)
}

func TestCParsing_Expressions2(t *testing.T) {
	parseC_4t(t,
		`int a() {
	if (a) { return (a); }

	return (0);
	return a+b;
	return (a+b);
	return (a)+0;
}`)

	parseC_4t(t, `int a() { return (a)+0; }`)
}

func TestCParsing_Expressions3(t *testing.T) {
	parseC_4t(t,
		`int a() {
1+(a);
(a)++;
(es)++;
(es)||a;
(es)->a;
return (a)+(b);
return 0+(a);
}`)
}

func TestCParsing_Expressions4(t *testing.T) {
	parseC_4t(t, `int a(){1+(a);}`)
}
func TestCParsing_Expressions5(t *testing.T) {
	parseC_4t(t, `int a(){return (int)0;}`)
}
func TestCParsing_Expressions6(t *testing.T) {
	parseC_4t(t, `int a(){return (in)0;}`)
}
func TestCParsing_Expressions7(t *testing.T) {
	parseC_4t(t, `int a()
{ return (0); }`)
}
func TestCParsing_Cast0(t *testing.T) {
	parseC_4t(t, `int a(){(cast)0;}`)
}
func TestCParsing_Cast1(t *testing.T) {
	parseC_4t(t, `int a(){(m*)(rsp);}`)
	parseC_4t(t, `int a(){(struct m*)(rsp);}`)
}

func TestCParsing_Empty(t *testing.T) {
	parseC_4t(t, `/** empty is valid. */  `)
}
func TestCParsing_EmptyStruct(t *testing.T) {
	parseC_4t(t, `struct empty{};`)
	parseC_4t(t, `struct {} empty;`)
	parseC_4t(t, `struct empty {} empty;`)
}
func TestCParsing_EmptyEmbeddedUnion(t *testing.T) {
	parseC_4t(t, `struct empty{
	union {
		int a;
		char b;
	};
};`)
}
func TestCParsing_ExtraSEMI(t *testing.T) {
	parseC_4t(t, `int func(){}
;
struct {} empty;
struct {} empty;;
int foo() {};
int foo() {};;
`)

	noParseC_4t(t, `struct empty{}`)
}
func TestCParsing_ExtraSEMI2(t *testing.T) {
	parseC_4t(t, `
struct a { int b; ; };
`)

	noParseC_4t(t, `struct empty{}`)
}

func TestCParsing_Escapes(t *testing.T) {
	parseC_4t(t, `
int f() {
	printf("%s", "\a\b\f\n\r\t\v");
	printf("\\");
	printf("\%");
	printf("\"");
	printf('\"'); // <- semantically wrong but syntactically valid.
}`)
}

func TestCParsing_Long(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping c parsing long test")
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
		} else if strings.HasSuffix(name, ".c") {
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

			clang := &C{Buffer: string(buffer)}
			clang.Init()
			if err := clang.Parse(); err != nil {
				log.Fatal(err)
			}
		}
	}
	walk("c/")
}
