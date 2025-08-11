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

func TestCheckAlwaysSucceeds(t *testing.T) {
	pegHeader := `
package main
type Test Peg {}
`

	testCases := []struct {
		name           string
		testRule       string
		expectedResult bool
	}{
		{
			name:           "Character expression does not always succeed (TypeChar)",
			testRule:       `A <- 'a'`,
			expectedResult: false,
		},
		{
			name:           "Star expression always succeed (TypeStar)",
			testRule:       `A <- 'a'*`,
			expectedResult: true,
		},
		{
			name:           "Dot expression does not always succeed (TypeDot)",
			testRule:       `A <- .`,
			expectedResult: false,
		},
		{
			name:           "Range expression does not always succeed (TypeRange)",
			testRule:       `A <- [a-z]`,
			expectedResult: false,
		},
		{
			name:           "String expression does not always succeed (TypeString)",
			testRule:       `A <- "abc"`,
			expectedResult: false,
		},
		{
			name:           "Predicate expression does not always succeed (TypePredicate)",
			testRule:       `A <- &{ true } 'a'*`,
			expectedResult: false,
		},
		{
			name:           "StateChange expression does not always succeed (TypeStateChange)",
			testRule:       `A <- !{ false } 'a'*`,
			expectedResult: false,
		},
		{
			name:           "Action expression does not always succeed (TypeAction)",
			testRule:       `A <- { } 'a'*`,
			expectedResult: true,
		},
		{
			name:           "Space expression does not always succeed (TypeSpace)",
			testRule:       `A <- ' '`,
			expectedResult: false,
		},
		{
			name:           "PeekFor expression does not always succeed (TypePeekFor)",
			testRule:       `A <- &'a'`,
			expectedResult: false,
		},
		{
			name:           "PeekNot expression does not always succeed (TypePeekNot)",
			testRule:       `A <- !'a'`,
			expectedResult: false,
		},
		{
			name:           "Plus expression does not always succeed (TypePlus)",
			testRule:       `A <- 'a'+`,
			expectedResult: false,
		},
		{
			name:           "Push expression does not always succeed (TypePush)",
			testRule:       `A <- <'a'*>`,
			expectedResult: true,
		},
		{
			name:           "Nil expression always succeeds (TypeNil)",
			testRule:       `A <- `,
			expectedResult: true,
		},
		{
			name:           "Optional expression always succeeds (TypeQuery)",
			testRule:       `A <- 'b'?`,
			expectedResult: true,
		},
		{
			name:           "Nested star expression always succeeds",
			testRule:       `A <- ('a' / 'b')*`,
			expectedResult: true,
		},
		{
			name:           "Sequence with star always succeeds",
			testRule:       `A <- 'a'* 'b'*`,
			expectedResult: true,
		},
		{
			name:           "Sequence with non-star does not always succeed",
			testRule:       `A <- 'a'* 'b'`,
			expectedResult: false,
		},
		{
			name:           "Alternate with star always succeeds",
			testRule:       `A <- 'a' / 'b'*`,
			expectedResult: true,
		},
		{
			name:           "Alternate without star does not always succeed",
			testRule:       `A <- 'a' / 'b'`,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sourceCode := pegHeader + tc.testRule

			p := &Peg[uint32]{Tree: tree.New(false, true, true), Buffer: sourceCode}
			_ = p.Init(Size[uint32](1 << 15))
			if err := p.Parse(); err != nil {
				t.Fatal(err)
			}
			p.Execute()
			buf := &bytes.Buffer{}
			_ = p.Compile("", []string{"peg"}, buf)

			if len(p.RuleNames) == 0 {
				t.Fatal("No rules found in the parsed tree")
			}
			rule := p.RuleNames[0]
			actualResult := rule.CheckAlwaysSucceeds(p.Tree)
			if actualResult != tc.expectedResult {
				t.Errorf("Rule [%s]: expected CheckAlwaysSucceeds() = %v, got %v",
					tc.name, tc.expectedResult, actualResult)
			}
		})
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
