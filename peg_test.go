package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestCorrect(t *testing.T) {
	buffer := `package p
type T Peg {}
Grammar <- !.
`
	p := &Peg{Tree: New(false, false, false), Buffer: buffer}
	p.Init()
	err := p.Parse()
	if err != nil {
		t.Error(err)
	}
}

func TestNoSpacePackage(t *testing.T) {
	buffer := `packagenospace
type T Peg {}
Grammar <- !.
`
	p := &Peg{Tree: New(false, false, false), Buffer: buffer}
	p.Init()
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
	p := &Peg{Tree: New(false, false, false), Buffer: buffer}
	p.Init()
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

	p := &Peg{Tree: New(true, true, false), Buffer: string(buffer)}
	p.Init()
	if err = p.Parse(); err != nil {
		t.Error(err)
	}

	p.Execute()

	out := &bytes.Buffer{}
	p.Compile("peg.peg.go", []string{"bootstrap/bootstrap"}, out)

	bootstrap, err := ioutil.ReadFile("bootstrap.peg.go")
	if err != nil {
		t.Error(err)
	}

	if len(out.Bytes()) != len(bootstrap) {
		t.Error("code generated from peg.peg is not the same as bootstrap.peg.go")
		return
	}

	for i, v := range out.Bytes() {
		if v != bootstrap[i] {
			t.Error("code generated from peg.peg is not the same as bootstrap.peg.go")
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
		p := &Peg{Tree: New(false, false, false), Buffer: buffer}
		p.Init()
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
		p.strict = true
		if err = p.Compile(f.Name(), []string{"peg"}, out); err == nil {
			t.Fatalf("#%d: expected warning error", i)
		}
		p.strict = false
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
			p := &Peg{Tree: New(true, true, false), Buffer: string(peg)}
			p.Init()
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

		p := &Peg{Tree: New(true, true, false), Buffer: string(input)}
		p.Init()
		pegs[i] = p
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, peg := range pegs {
			b.StopTimer()
			peg.Reset()
			b.StartTimer()
			if err := peg.Parse(); err != nil {
				b.Error(err)
			}
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

		p := &Peg{Tree: New(true, true, false), Buffer: string(input)}
		p.Init()
		pegs[i] = p
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, peg := range pegs {
			peg.Reset()
			if err := peg.Parse(); err != nil {
				b.Error(err)
			}
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
			peg := &Peg{Tree: New(true, true, false), Buffer: string(str)}
			peg.Init()
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
			peg := &Peg{Tree: New(true, true, false), Buffer: string(str)}
			peg.Init()
			peg.Reset()
			if err := peg.Parse(); err != nil {
				b.Error(err)
			}
		}
	}
}
