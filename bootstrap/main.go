// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/pointlander/peg/tree"
)

func main() {
	runtime.GOMAXPROCS(2)
	t := tree.New(true, true, false)

	/*package main

	  import "fmt"
	  import "math"
	  import "sort"
	  import "strconv"

	  type Peg Peg {
	   *Tree
	  }*/
	t.AddPackage("main")
	t.AddImport("github.com/pointlander/peg/tree")
	t.AddPeg("Peg")
	t.AddState(`
 *tree.Tree
`)

	addDot := t.AddDot
	addName := t.AddName
	addCharacter := t.AddCharacter
	addAction := t.AddAction

	addRule := func(name string, item func()) {
		t.AddRule(name)
		item()
		t.AddExpression()
	}

	addSequence := func(items ...func()) {
		sequence := false
		for _, item := range items {
			item()
			if sequence {
				t.AddSequence()
			} else {
				sequence = true
			}
		}
	}

	addAlternate := func(items ...func()) {
		alternate := false
		for _, item := range items {
			item()
			if alternate {
				t.AddAlternate()
			} else {
				alternate = true
			}
		}
	}

	addString := func(s string) {
		sequence := false
		for _, r := range s {
			t.AddCharacter(string(r))
			if sequence {
				t.AddSequence()
			} else {
				sequence = true
			}
		}
	}

	addRange := func(begin, end string) {
		addCharacter(begin)
		addCharacter(end)
		t.AddRange()
	}

	addStar := func(item func()) {
		item()
		t.AddStar()
	}

	addQuery := func(item func()) {
		item()
		t.AddQuery()
	}

	addPush := func(item func()) {
		item()
		t.AddPush()
	}

	addPeekNot := func(item func()) {
		item()
		t.AddPeekNot()
	}

	addPeekFor := func(item func()) {
		item()
		t.AddPeekFor()
	}

	/* Grammar <- Spacing { hdr; } Action* Definition* !. */
	addRule("Grammar", func() {
		addSequence(
			func() { addName("Spacing") },
			func() { addAction(`p.AddPackage("main")`) },
			func() { addAction(`p.AddImport("github.com/pointlander/peg/tree")`) },
			func() { addAction(`p.AddPeg("Peg")`) },
			func() { addAction(`p.AddState("*tree.Tree")`) },
			func() { addStar(func() { addName("Action") }) },
			func() { addStar(func() { addName("Definition") }) },
			func() { addPeekNot(func() { addDot() }) },
		)
	})

	/* Definition      <- Identifier                   { p.AddRule(text) }
	   LeftArrow Expression         { p.AddExpression() } &(Identifier LeftArrow / !.)*/
	addRule("Definition", func() {
		addSequence(
			func() { addName("Identifier") },
			func() { addAction(" p.AddRule(text) ") },
			func() { addName("LeftArrow") },
			func() { addName("Expression") },
			func() { addAction(" p.AddExpression() ") },
			func() {
				addPeekFor(func() {
					addAlternate(
						func() {
							addSequence(
								func() { addName("Identifier") },
								func() { addName("LeftArrow") },
							)
						},
						func() { addPeekNot(func() { addDot() }) },
					)
				})
			},
		)
	})

	/* Expression <- Sequence (Slash Sequence { p.AddAlternate() })* */
	addRule("Expression", func() {
		addSequence(
			func() { addName("Sequence") },
			func() {
				addStar(func() {
					addSequence(
						func() { addName("Slash") },
						func() { addName("Sequence") },
						func() { addAction(" p.AddAlternate() ") },
					)
				})
			},
		)
	})

	/* Sequence <- Prefix (Prefix { p.AddSequence() } )* */
	addRule("Sequence", func() {
		addSequence(
			func() { addName("Prefix") },
			func() {
				addStar(func() {
					addSequence(
						func() { addName("Prefix") },
						func() { addAction(" p.AddSequence() ") },
					)
				})
			},
		)
	})

	/* Prefix <- '!' Suffix { p.AddPeekNot() } / Suffix */
	addRule("Prefix", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addCharacter(`!`) },
					func() { addName("Suffix") },
					func() { addAction(" p.AddPeekNot() ") },
				)
			},
			func() { addName("Suffix") },
		)
	})

	/* Suffix          <- Primary (	Question	{ p.AddQuery() }
	  				/ Star		{ p.AddStar() }
	)? */
	addRule("Suffix", func() {
		addSequence(
			func() { addName("Primary") },
			func() {
				addQuery(func() {
					addAlternate(
						func() {
							addSequence(
								func() { addName("Question") },
								func() { addAction(" p.AddQuery() ") },
							)
						},
						func() {
							addSequence(
								func() { addName("Star") },
								func() { addAction(" p.AddStar() ") },
							)
						},
					)
				})
			},
		)
	})

	/* Primary         <- Identifier !LeftArrow        { p.AddName(text) }
	   / Open Expression Close
	   / Literal
	   / Class
	   / Dot                          { p.AddDot() }
	   / Action                       { p.AddAction(text) }
	   / Begin Expression End         { p.AddPush() }*/
	addRule("Primary", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addName("Identifier") },
					func() { addPeekNot(func() { t.AddName("LeftArrow") }) },
					func() { addAction(" p.AddName(text) ") },
				)
			},
			func() {
				addSequence(
					func() { addName("Open") },
					func() { addName("Expression") },
					func() { addName("Close") },
				)
			},
			func() { addName("Literal") },
			func() { addName("Class") },
			func() {
				addSequence(
					func() { addName("Dot") },
					func() { addAction(" p.AddDot() ") },
				)
			},
			func() {
				addSequence(
					func() { addName("Action") },
					func() { addAction(" p.AddAction(text) ") },
				)
			},
			func() {
				addSequence(
					func() { addName("Begin") },
					func() { addName("Expression") },
					func() { addName("End") },
					func() { addAction(" p.AddPush() ") },
				)
			},
		)
	})

	/* Identifier      <- < Ident Ident* > Spacing */
	addRule("Identifier", func() {
		addSequence(
			func() {
				addPush(func() {
					addSequence(
						func() { addName("Ident") },
						func() { addStar(func() { addName("Ident") }) },
					)
				})
			},
			func() { addName("Spacing") },
		)
	})

	/* Ident <- [A-Za-z] */
	addRule("Ident", func() {
		addAlternate(
			func() { addRange(`A`, `Z`) },
			func() { addRange(`a`, `z`) },
		)
	})

	/* Literal <- ['] !['] Char (!['] Char { p.AddSequence() } )* ['] Spacing */
	addRule("Literal", func() {
		addSequence(
			func() { addCharacter(`'`) },
			func() {
				addSequence(
					func() { addPeekNot(func() { addCharacter(`'`) }) },
					func() { addName("Char") },
				)
			},
			func() {
				addStar(func() {
					addSequence(
						func() { addPeekNot(func() { addCharacter(`'`) }) },
						func() { addName("Char") },
						func() { addAction(` p.AddSequence() `) },
					)
				})
			},
			func() { addCharacter(`'`) },
			func() { addName("Spacing") },
		)
	})

	/* Class  <- '[' Range (!']' Range { p.AddAlternate() })* ']' Spacing */
	addRule("Class", func() {
		addSequence(
			func() { addCharacter(`[`) },
			func() { addName("Range") },
			func() {
				addStar(func() {
					addSequence(
						func() { addPeekNot(func() { addCharacter(`]`) }) },
						func() { addName("Range") },
						func() { addAction(" p.AddAlternate() ") },
					)
				})
			},
			func() { addCharacter(`]`) },
			func() { addName("Spacing") },
		)
	})

	/* Range           <- Char '-' Char { p.AddRange() }
	   / Char */
	addRule("Range", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addName("Char") },
					func() { addCharacter(`-`) },
					func() { addName("Char") },
					func() { addAction(" p.AddRange() ") },
				)
			},
			func() { addName("Char") },
		)
	})

	/* Char	<- Escape
	/  '\\' "0x"<[0-9a-f]*>   { p.AddHexaCharacter(text) }
	/  '\\\\'                  { p.AddCharacter("\\") }
	/  !'\\' <.>                  { p.AddCharacter(text) } */
	addRule("Char", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter(`0`) },
					func() { addCharacter(`x`) },
					func() {
						addPush(func() {
							addStar(func() {
								addAlternate(
									func() { addRange(`0`, `9`) },
									func() { addRange(`a`, `f`) },
								)
							})
						})
					},
					func() { addAction(` p.AddHexaCharacter(text) `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter("\\") },
					func() { addAction(` p.AddCharacter("\\") `) },
				)
			},
			func() {
				addSequence(
					func() { addPeekNot(func() { addCharacter("\\") }) },
					func() { addPush(func() { addDot() }) },
					func() { addAction(` p.AddCharacter(text) `) },
				)
			},
		)
	})
	/* LeftArrow       <- '<-' Spacing */
	addRule("LeftArrow", func() {
		addSequence(
			func() { addString(`<-`) },
			func() { addName("Spacing") },
		)
	})

	/* Slash           <- '/' Spacing */
	addRule("Slash", func() {
		addSequence(
			func() { addCharacter(`/`) },
			func() { addName("Spacing") },
		)
	})

	/* Question        <- '?' Spacing */
	addRule("Question", func() {
		addSequence(
			func() { addCharacter(`?`) },
			func() { addName("Spacing") },
		)
	})

	/* Star            <- '*' Spacing */
	addRule("Star", func() {
		addSequence(
			func() { addCharacter(`*`) },
			func() { addName("Spacing") },
		)
	})

	/* Open            <- '(' Spacing */
	addRule("Open", func() {
		addSequence(
			func() { addCharacter(`(`) },
			func() { addName("Spacing") },
		)
	})

	/* Close           <- ')' Spacing */
	addRule("Close", func() {
		addSequence(
			func() { addCharacter(`)`) },
			func() { addName("Spacing") },
		)
	})

	/* Dot             <- '.' Spacing */
	addRule("Dot", func() {
		addSequence(
			func() { addCharacter(`.`) },
			func() { addName("Spacing") },
		)
	})

	addRule("Spacing", func() {
		addStar(func() {
			addAlternate(
				func() { addName("Space") },
				func() { addName("Comment") },
			)
		})
	})

	/* Comment         <- '#' (!EndOfLine .)* */
	addRule("Comment", func() {
		addSequence(
			func() { addCharacter(`#`) },
			func() {
				addStar(func() {
					addSequence(
						func() { addPeekNot(func() { addName("EndOfLine") }) },
						func() { addDot() },
					)
				})
			},
		)
	})

	/* Space           <- ' ' / '\t' / EndOfLine */
	addRule("Space", func() {
		addAlternate(
			func() { addCharacter(` `) },
			func() { addCharacter("\t") },
			func() { addName("EndOfLine") },
		)
	})

	/* EndOfLine       <- '\r\n' / '\n' / '\r' */
	addRule("EndOfLine", func() {
		addAlternate(
			func() { addString("\r\n") },
			func() { addCharacter("\n") },
			func() { addCharacter("\r") },
		)
	})

	/* Action		<- '{' < (![}].)* > '}' Spacing */
	addRule("Action", func() {
		addSequence(
			func() { addCharacter(`{`) },
			func() {
				addPush(func() {
					addStar(func() {
						addSequence(
							func() {
								addPeekNot(func() {
									addCharacter(`}`)
								})
							},
							func() { addDot() },
						)
					})
				})
			},
			func() { addCharacter(`}`) },
			func() { addName("Spacing") },
		)
	})

	/* Begin           <- '<' Spacing */
	addRule("Begin", func() {
		addSequence(
			func() { addCharacter(`<`) },
			func() { addName("Spacing") },
		)
	})

	/* End             <- '>' Spacing */
	addRule("End", func() {
		addSequence(
			func() { addCharacter(`>`) },
			func() { addName("Spacing") },
		)
	})

	filename := "bootstrap.peg.go"
	out, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("%v: %v\n", filename, err)
		return
	}
	defer out.Close()
	t.Compile(filename, os.Args, out)
}
