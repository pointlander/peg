package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/pointlander/peg/tree"
)

func TestCorrect(t *testing.T) {
	buffer := `package p
type T Peg {}
Grammar <- !.
`
	p := &Peg[uint32]{Tree: tree.New(false, false, false), Buffer: buffer}
	_ = p.Init()
	err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	p = &Peg[uint32]{Tree: tree.New(false, false, false), Buffer: buffer}
	_ = p.Init(Size[uint32](1 << 15))
	err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoSpacePackage(t *testing.T) {
	buffer := `packagenospace
type T Peg {}
Grammar <- !.
`
	p := &Peg[uint32]{Tree: tree.New(false, false, false), Buffer: buffer}
	_ = p.Init(Size[uint32](1 << 15))
	err := p.Parse()
	if err == nil {
		t.Fatal("packagenospace was parsed without error")
	}
}

func TestNoSpaceType(t *testing.T) {
	buffer := `
package p
typenospace Peg {}
Grammar <- !.
`
	p := &Peg[uint32]{Tree: tree.New(false, false, false), Buffer: buffer}
	_ = p.Init(Size[uint32](1 << 15))
	err := p.Parse()
	if err == nil {
		t.Fatal("typenospace was parsed without error")
	}
}

func TestSame(t *testing.T) {
	buffer, err := os.ReadFile("peg.peg")
	if err != nil {
		t.Fatal(err)
	}

	p := &Peg[uint32]{Tree: tree.New(true, true, false), Buffer: string(buffer)}
	_ = p.Init(Size[uint32](1 << 15))
	if err = p.Parse(); err != nil {
		t.Fatal(err)
	}

	p.Execute()

	out := &bytes.Buffer{}
	_ = p.Compile("peg.peg.go", []string{"./peg", "-inline", "-switch", "peg.peg"}, out)

	bootstrap, err := os.ReadFile("peg.peg.go")
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Bytes()) != len(bootstrap) {
		t.Fatal("code generated from peg.peg is not the same as .go")
	}

	for i, v := range out.Bytes() {
		if v != bootstrap[i] {
			t.Fatal("code generated from peg.peg is not the same as .go")
		}
	}
}

func TestStrict(t *testing.T) {
	tt := []string{
		// rule used but not defined
		`
package main
type test Peg {}
Begin <- begin !.
`,
		// rule defined but not used
		`
package main
type test Peg {}
Begin <- .
unused <- 'unused'
`,
		// left recursive rule
		`package main
type test Peg {}
Begin <- Begin 'x'
`,
	}

	for i, buffer := range tt {
		p := &Peg[uint32]{Tree: tree.New(false, false, false), Buffer: buffer}
		_ = p.Init(Size[uint32](1 << 15))
		if err := p.Parse(); err != nil {
			t.Fatal(err)
		}
		p.Execute()

		tempDir := t.TempDir()

		out := &bytes.Buffer{}
		p.Strict = true
		if err := p.Compile(tempDir, []string{"peg"}, out); err == nil {
			t.Fatalf("#%d: expected warning error", i)
		}
		p.Strict = false
		if err := p.Compile(tempDir, []string{"peg"}, out); err != nil {
			t.Fatalf("#%d: unexpected error (%v)", i, err)
		}
	}
}

func TestCJKCharacter(t *testing.T) {
	buffer := `
package main

type DiceExprParser Peg {
}

Expr <- 'CJK' / '汉字' / 'test'
`
	p := &Peg[uint32]{Tree: tree.New(false, true, false), Buffer: buffer}
	_ = p.Init(Size[uint32](1 << 15))
	err := p.Parse()
	if err != nil {
		t.Fatal("cjk character test failed")
	}
}

var pegFileContents = func(files []string) []string {
	contents := make([]string, len(files))
	for i, file := range files {
		input, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		contents[i] = string(input)
	}
	return contents
}([]string{
	"peg.peg",
	"grammars/c/c.peg",
	"grammars/calculator/calculator.peg",
	"grammars/fexl/fexl.peg",
	"grammars/java/java_1_7.peg",
})

func BenchmarkInitOnly(b *testing.B) {
	for b.Loop() {
		for _, peg := range pegFileContents {
			p := &Peg[uint32]{Tree: tree.New(true, true, false), Buffer: peg}
			_ = p.Init(Size[uint32](1 << 15))
		}
	}
}

func BenchmarkParse(b *testing.B) {
	pegs := make([]*Peg[uint32], len(pegFileContents))
	for i, content := range pegFileContents {
		p := &Peg[uint32]{Tree: tree.New(true, true, false), Buffer: content}
		_ = p.Init(Size[uint32](1 << 15))
		pegs[i] = p
	}

	for b.Loop() {
		for _, peg := range pegs {
			if err := peg.Parse(); err != nil {
				b.Fatal(err)
			}
			b.StopTimer()
			peg.Reset()
			b.StartTimer()
		}
	}
}

func BenchmarkParseAndReset(b *testing.B) {
	pegs := make([]*Peg[uint32], len(pegFileContents))
	for i, content := range pegFileContents {
		p := &Peg[uint32]{Tree: tree.New(true, true, false), Buffer: content}
		_ = p.Init(Size[uint32](1 << 15))
		pegs[i] = p
	}

	for b.Loop() {
		for _, peg := range pegs {
			if err := peg.Parse(); err != nil {
				b.Fatal(err)
			}
			peg.Reset()
		}
	}
}

func BenchmarkInitAndParse(b *testing.B) {
	for b.Loop() {
		for _, peg := range pegFileContents {
			p := &Peg[uint32]{Tree: tree.New(true, true, false), Buffer: peg}
			_ = p.Init(Size[uint32](1 << 15))
			if err := p.Parse(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkInitParseAndReset(b *testing.B) {
	for b.Loop() {
		for _, peg := range pegFileContents {
			p := &Peg[uint32]{Tree: tree.New(true, true, false), Buffer: peg}
			_ = p.Init(Size[uint32](1 << 15))
			if err := p.Parse(); err != nil {
				b.Fatal(err)
			}
			p.Reset()
		}
	}
}
