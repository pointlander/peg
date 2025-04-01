// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate ../../peg -switch -inline c.peg

package c

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func parseCBuffer(buffer string) (*C[uint32], error) {
	clang := &C[uint32]{Buffer: buffer}
	err := clang.Init()
	if err != nil {
		return nil, err
	}
	err = clang.Parse()
	return clang, err
}

func parseC4t(t *testing.T, src string) *C[uint32] {
	c, err := parseCBuffer(src)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func noParseC4t(t *testing.T, src string) {
	_, err := parseCBuffer(src)
	if err == nil {
		t.Fatal("Parsed what should not have parsed.")
	}
}

func TestCParsing_Expressions1(t *testing.T) {
	case1src := `int a() {
		(es);
		1++;
		1+1;
		a+1;
		(a)+1;
		a->x;
		return 0;
}`
	parseC4t(t, case1src)
}

func TestCParsing_Expressions2(t *testing.T) {
	parseC4t(t,
		`int a() {
	if (a) { return (a); }

	return (0);
	return a+b;
	return (a+b);
	return (a)+0;
}`)

	parseC4t(t, `int a() { return (a)+0; }`)
}

func TestCParsing_Expressions3(t *testing.T) {
	parseC4t(t,
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
	parseC4t(t, `int a(){1+(a);}`)
}

func TestCParsing_Expressions5(t *testing.T) {
	parseC4t(t, `int a(){return (int)0;}`)
}

func TestCParsing_Expressions6(t *testing.T) {
	parseC4t(t, `int a(){return (in)0;}`)
}

func TestCParsing_Expressions7(t *testing.T) {
	parseC4t(t, `int a()
{ return (0); }`)
}

func TestCParsing_Cast0(t *testing.T) {
	parseC4t(t, `int a(){(cast)0;}`)
}

func TestCParsing_Cast1(t *testing.T) {
	parseC4t(t, `int a(){(m*)(rsp);}`)
	parseC4t(t, `int a(){(struct m*)(rsp);}`)
}

func TestCParsing_Empty(t *testing.T) {
	parseC4t(t, `/** empty is valid. */  `)
}

func TestCParsing_EmptyStruct(t *testing.T) {
	parseC4t(t, `struct empty{};`)
	parseC4t(t, `struct {} empty;`)
	parseC4t(t, `struct empty {} empty;`)
}

func TestCParsing_EmptyEmbeddedUnion(t *testing.T) {
	parseC4t(t, `struct empty{
	union {
		int a;
		char b;
	};
};`)
}

func TestCParsing_ExtraSEMI(t *testing.T) {
	parseC4t(t, `int func(){}
;
struct {} empty;
struct {} empty;;
int foo() {};
int foo() {};;
`)

	noParseC4t(t, `struct empty{}`)
}

func TestCParsing_ExtraSEMI2(t *testing.T) {
	parseC4t(t, `
struct a { int b; ; };
`)

	noParseC4t(t, `struct empty{}`)
}

func TestCParsing_Escapes(t *testing.T) {
	parseC4t(t, `
int f() {
	printf("%s", "\a\b\f\n\r\t\v");
	printf("\\");
	printf("\%");
	printf("\"");
	printf('\"'); // <- semantically wrong but syntactically valid.
}`)
}

func TestCFiles(t *testing.T) {
	// TODO: find  appropriate c files.
	err := filepath.Walk(".", func(path string, _ fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if filepath.Ext(path) == ".c" {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			clang := &C[uint32]{Buffer: string(b)}
			err = clang.Init()
			if err != nil {
				t.Fatal(err)
			}
			if err := clang.Parse(); err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCParsing_WideString(t *testing.T) {
	parseC4t(t, `wchar_t *msg = L"Hello";`)
}
