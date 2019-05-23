package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pointlander/peg/tree"
)

func TestCorrect(t *testing.T) {
	buffer := `package p
type T Peg {}
Grammar <- !.
`
	p := &Peg{Tree: tree.New(false, false, false), Buffer: buffer}
	p.Init()
	err := p.Parse()
	if err != nil {
		t.Error(err)
	}

	p = &Peg{Tree: tree.New(false, false, false), Buffer: buffer}
	p.Init(Size(1<<15))
	err = p.Parse()
	if err != nil {
		t.Error(err)
	}
}

func TestNoSpacePackage(t *testing.T) {
	buffer := `packagenospace
type T Peg {}
Grammar <- !.
`
	p := &Peg{Tree: tree.New(false, false, false), Buffer: buffer}
	p.Init(Size(1<<15))
	err := p.Parse()
	if err == nil {
		t.Error("packagenospace was parsed without error")
	}
}

func TestNoSpaceType(t *testing.T) {
	buffer := `
package p
typenospace Peg {}
Grammar <- !.
`
	p := &Peg{Tree: tree.New(false, false, false), Buffer: buffer}
	p.Init(Size(1<<15))
	err := p.Parse()
	if err == nil {
		t.Error("typenospace was parsed without error")
	}
}

func TestSame(t *testing.T) {
	buffer, err := ioutil.ReadFile("peg.peg")
	if err != nil {
		t.Error(err)
	}

	p := &Peg{Tree: tree.New(true, true, false), Buffer: string(buffer)}
	p.Init(Size(1<<15))
	if err = p.Parse(); err != nil {
		t.Error(err)
	}

	p.Execute()

	out := &bytes.Buffer{}
	p.Compile("peg.peg.go", []string{"./peg", "-inline", "-switch", "peg.peg"}, out)

	bootstrap, err := ioutil.ReadFile("peg.peg.go")
	if err != nil {
		t.Error(err)
	}

	if len(out.Bytes()) != len(bootstrap) {
		t.Error("code generated from peg.peg is not the same as .go")
		return
	}

	for i, v := range out.Bytes() {
		if v != bootstrap[i] {
			t.Error("code generated from peg.peg is not the same as .go")
			return
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
		p := &Peg{Tree: tree.New(false, false, false), Buffer: buffer}
		p.Init(Size(1<<15))
		if err := p.Parse(); err != nil {
			t.Fatal(err)
		}
		p.Execute()

		f, err := ioutil.TempFile("", "peg")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			os.Remove(f.Name())
			f.Close()
		}()
		out := &bytes.Buffer{}
		p.Strict = true
		if err = p.Compile(f.Name(), []string{"peg"}, out); err == nil {
			t.Fatalf("#%d: expected warning error", i)
		}
		p.Strict = false
		if err = p.Compile(f.Name(), []string{"peg"}, out); err != nil {
			t.Fatalf("#%d: unexpected error (%v)", i, err)
		}
	}
}

var files = [...]string{
	"peg.peg",
	"grammars/c/c.peg",
	"grammars/calculator/calculator.peg",
	"grammars/fexl/fexl.peg",
	"grammars/java/java_1_7.peg",
}

func BenchmarkInitOnly(b *testing.B) {
	pegs := []string{}
	for _, file := range files {
		input, err := ioutil.ReadFile(file)
		if err != nil {
			b.Error(err)
		}
		pegs = append(pegs, string(input))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, peg := range pegs {
			p := &Peg{Tree: tree.New(true, true, false), Buffer: peg}
			p.Init(Size(1<<15))
		}
	}
}

func BenchmarkParse(b *testing.B) {
	pegs := make([]*Peg, len(files))
	for i, file := range files {
		input, err := ioutil.ReadFile(file)
		if err != nil {
			b.Error(err)
		}

		p := &Peg{Tree: tree.New(true, true, false), Buffer: string(input)}
		p.Init(Size(1<<15))
		pegs[i] = p
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, peg := range pegs {
			if err := peg.Parse(); err != nil {
				b.Error(err)
			}
			b.StopTimer()
			peg.Reset()
			b.StartTimer()
		}
	}
}

func BenchmarkResetAndParse(b *testing.B) {
	pegs := make([]*Peg, len(files))
	for i, file := range files {
		input, err := ioutil.ReadFile(file)
		if err != nil {
			b.Error(err)
		}

		p := &Peg{Tree: tree.New(true, true, false), Buffer: string(input)}
		p.Init(Size(1<<15))
		pegs[i] = p
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, peg := range pegs {
			if err := peg.Parse(); err != nil {
				b.Error(err)
			}
			peg.Reset()
		}
	}
}

func BenchmarkInitAndParse(b *testing.B) {
	strs := []string{}
	for _, file := range files {
		input, err := ioutil.ReadFile(file)
		if err != nil {
			b.Error(err)
		}
		strs = append(strs, string(input))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, str := range strs {
			peg := &Peg{Tree: tree.New(true, true, false), Buffer: str}
			peg.Init(Size(1<<15))
			if err := peg.Parse(); err != nil {
				b.Error(err)
			}
		}
	}
}

func BenchmarkInitResetAndParse(b *testing.B) {
	strs := []string{}
	for _, file := range files {
		input, err := ioutil.ReadFile(file)
		if err != nil {
			b.Error(err)
		}
		strs = append(strs, string(input))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, str := range strs {
			peg := &Peg{Tree: tree.New(true, true, false), Buffer: str}
			peg.Init(Size(1<<15))
			if err := peg.Parse(); err != nil {
				b.Error(err)
			}
			peg.Reset()
		}
	}
}
