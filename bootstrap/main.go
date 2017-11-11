// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)
	t := New(true, true, false)

	/*package main

	  import "fmt"
	  import "math"
	  import "sort"
	  import "strconv"

	  type Peg Peg {
	   *Tree
	  }*/
	t.AddPackage("main")
	t.AddPeg("Peg")
	t.AddState(`
 *Tree
`)

	addDot := t.AddDot
	addName := t.AddName
	addCharacter := t.AddCharacter
	addDoubleCharacter := t.AddDoubleCharacter
	addHexaCharacter := t.AddHexaCharacter
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

	addDoubleRange := func(begin, end string) {
		addCharacter(begin)
		addCharacter(end)
		t.AddDoubleRange()
	}

	addStar := func(item func()) {
		item()
		t.AddStar()
	}

	addPlus := func(item func()) {
		item()
		t.AddPlus()
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

	/* Grammar         <- Spacing 'package' MustSpacing Identifier      { p.AddPackage(text) }
	   Import*
	   'type' MustSpacing Identifier         { p.AddPeg(text) }
	   'Peg' Spacing Action              { p.AddState(text) }
	   Definition+ EndOfFile */
	addRule("Grammar", func() {
		addSequence(
			func() { addName("Spacing") },
			func() { addString("package") },
			func() { addName("MustSpacing") },
			func() { addName("Identifier") },
			func() { addAction(" p.AddPackage(text) ") },
			func() { addStar(func() { addName("Import") }) },
			func() { addString("type") },
			func() { addName("MustSpacing") },
			func() { addName("Identifier") },
			func() { addAction(" p.AddPeg(text) ") },
			func() { addString("Peg") },
			func() { addName("Spacing") },
			func() { addName("Action") },
			func() { addAction(" p.AddState(text) ") },
			func() { addPlus(func() { addName("Definition") }) },
			func() { addName("EndOfFile") },
		)
	})

	/* Import          <- 'import' Spacing ["] < [a-zA-Z_/.\-]+ > ["] Spacing { p.AddImport(text) } */
	addRule("Import", func() {
		addSequence(
			func() { addString("import") },
			func() { addName("Spacing") },
			func() { addCharacter(`"`) },
			func() {
				addPush(func() {
					addPlus(func() {
						addAlternate(
							func() { addRange(`a`, `z`) },
							func() { addRange(`A`, `Z`) },
							func() { addCharacter(`_`) },
							func() { addCharacter(`/`) },
							func() { addCharacter(`.`) },
							func() { addCharacter(`-`) },
						)
					})
				})
			},
			func() { addCharacter(`"`) },
			func() { addName("Spacing") },
			func() { addAction(" p.AddImport(text) ") },
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

	/* Expression      <- Sequence (Slash Sequence     { p.AddAlternate() }
	           )* (Slash           { p.AddNil(); p.AddAlternate() }
	              )?
	/ { p.AddNil() } */
	addRule("Expression", func() {
		addAlternate(
			func() {
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
					func() {
						addQuery(func() {
							addSequence(
								func() { addName("Slash") },
								func() { addAction(" p.AddNil(); p.AddAlternate() ") },
							)
						})
					},
				)
			},
			func() { addAction(" p.AddNil() ") },
		)
	})

	/* Sequence        <- Prefix (Prefix               { p.AddSequence() }
	   )* */
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

	/* Prefix          <- And Action                   { p.AddPredicate(text) }
	   / Not Action                   { p.AddStateChange(text) }
	   / And Suffix                   { p.AddPeekFor() }
	   / Not Suffix                   { p.AddPeekNot() }
	   /     Suffix */
	addRule("Prefix", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addName("And") },
					func() { addName("Action") },
					func() { addAction(" p.AddPredicate(text) ") },
				)
			},
			func() {
				addSequence(
					func() { addName("Not") },
					func() { addName("Action") },
					func() { addAction(" p.AddStateChange(text) ") },
				)
			},
			func() {
				addSequence(
					func() { addName("And") },
					func() { addName("Suffix") },
					func() { addAction(" p.AddPeekFor() ") },
				)
			},
			func() {
				addSequence(
					func() { addName("Not") },
					func() { addName("Suffix") },
					func() { addAction(" p.AddPeekNot() ") },
				)
			},
			func() { addName("Suffix") },
		)
	})

	/* Suffix          <- Primary (Question            { p.AddQuery() }
	   / Star             { p.AddStar() }
	   / Plus             { p.AddPlus() }
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
						func() {
							addSequence(
								func() { addName("Plus") },
								func() { addAction(" p.AddPlus() ") },
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

	/* Identifier      <- < IdentStart IdentCont* > Spacing */
	addRule("Identifier", func() {
		addSequence(
			func() {
				addPush(func() {
					addSequence(
						func() { addName("IdentStart") },
						func() { addStar(func() { addName("IdentCont") }) },
					)
				})
			},
			func() { addName("Spacing") },
		)
	})

	/* IdentStart      <- [[a-z_]] */
	addRule("IdentStart", func() {
		addAlternate(
			func() { addDoubleRange(`a`, `z`) },
			func() { addCharacter(`_`) },
		)
	})

	/* IdentCont       <- IdentStart / [0-9] */
	addRule("IdentCont", func() {
		addAlternate(
			func() { addName("IdentStart") },
			func() { addRange(`0`, `9`) },
		)
	})

	/* Literal         <- ['] (!['] Char)? (!['] Char          { p.AddSequence() }
	                     )* ['] Spacing
	   / ["] (!["] DoubleChar)? (!["] DoubleChar          { p.AddSequence() }
	                            )* ["] Spacing */
	addRule("Literal", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addCharacter(`'`) },
					func() {
						addQuery(func() {
							addSequence(
								func() { addPeekNot(func() { addCharacter(`'`) }) },
								func() { addName("Char") },
							)
						})
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
			},
			func() {
				addSequence(
					func() { addCharacter(`"`) },
					func() {
						addQuery(func() {
							addSequence(
								func() { addPeekNot(func() { addCharacter(`"`) }) },
								func() { addName("DoubleChar") },
							)
						})
					},
					func() {
						addStar(func() {
							addSequence(
								func() { addPeekNot(func() { addCharacter(`"`) }) },
								func() { addName("DoubleChar") },
								func() { addAction(` p.AddSequence() `) },
							)
						})
					},
					func() { addCharacter(`"`) },
					func() { addName("Spacing") },
				)
			},
		)
	})

	/* Class  <- ( '[[' ( '^' DoubleRanges              { p.AddPeekNot(); p.AddDot(); p.AddSequence() }
	          / DoubleRanges )?
	     ']]'
	   / '[' ( '^' Ranges                     { p.AddPeekNot(); p.AddDot(); p.AddSequence() }
	         / Ranges )?
	     ']' )
	   Spacing */
	addRule("Class", func() {
		addSequence(
			func() {
				addAlternate(
					func() {
						addSequence(
							func() { addString(`[[`) },
							func() {
								addQuery(func() {
									addAlternate(
										func() {
											addSequence(
												func() { addCharacter(`^`) },
												func() { addName("DoubleRanges") },
												func() { addAction(` p.AddPeekNot(); p.AddDot(); p.AddSequence() `) },
											)
										},
										func() { addName("DoubleRanges") },
									)
								})
							},
							func() { addString(`]]`) },
						)
					},
					func() {
						addSequence(
							func() { addCharacter(`[`) },
							func() {
								addQuery(func() {
									addAlternate(
										func() {
											addSequence(
												func() { addCharacter(`^`) },
												func() { addName("Ranges") },
												func() { addAction(` p.AddPeekNot(); p.AddDot(); p.AddSequence() `) },
											)
										},
										func() { addName("Ranges") },
									)
								})
							},
							func() { addCharacter(`]`) },
						)
					},
				)
			},
			func() { addName("Spacing") },
		)
	})

	/* Ranges          <- !']' Range (!']' Range  { p.AddAlternate() }
	   )* */
	addRule("Ranges", func() {
		addSequence(
			func() { addPeekNot(func() { addCharacter(`]`) }) },
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
		)
	})

	/* DoubleRanges          <- !']]' DoubleRange (!']]' DoubleRange  { p.AddAlternate() }
	   )* */
	addRule("DoubleRanges", func() {
		addSequence(
			func() { addPeekNot(func() { addString(`]]`) }) },
			func() { addName("DoubleRange") },
			func() {
				addStar(func() {
					addSequence(
						func() { addPeekNot(func() { addString(`]]`) }) },
						func() { addName("DoubleRange") },
						func() { addAction(" p.AddAlternate() ") },
					)
				})
			},
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

	/* DoubleRange      <- Char '-' Char { p.AddDoubleRange() }
	   / DoubleChar */
	addRule("DoubleRange", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addName("Char") },
					func() { addCharacter(`-`) },
					func() { addName("Char") },
					func() { addAction(" p.AddDoubleRange() ") },
				)
			},
			func() { addName("DoubleChar") },
		)
	})

	/* Char            <- Escape
	   / !'\\' <.>                  { p.AddCharacter(text) } */
	addRule("Char", func() {
		addAlternate(
			func() { addName("Escape") },
			func() {
				addSequence(
					func() { addPeekNot(func() { addCharacter("\\") }) },
					func() { addPush(func() { addDot() }) },
					func() { addAction(` p.AddCharacter(text) `) },
				)
			},
		)
	})

	/* DoubleChar      <- Escape
	   / <[a-zA-Z]>                 { p.AddDoubleCharacter(text) }
	   / !'\\' <.>                  { p.AddCharacter(text) } */
	addRule("DoubleChar", func() {
		addAlternate(
			func() { addName("Escape") },
			func() {
				addSequence(
					func() {
						addPush(func() {
							addAlternate(
								func() { addRange(`a`, `z`) },
								func() { addRange(`A`, `Z`) },
							)
						})
					},
					func() { addAction(` p.AddDoubleCharacter(text) `) },
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

	/* Escape            <- "\\a"                      { p.AddCharacter("\a") }   # bell
		                      / "\\b"                      { p.AddCharacter("\b") }   # bs
	                              / "\\e"                      { p.AddCharacter("\x1B") } # esc
	                              / "\\f"                      { p.AddCharacter("\f") }   # ff
	                              / "\\n"                      { p.AddCharacter("\n") }   # nl
	                              / "\\r"                      { p.AddCharacter("\r") }   # cr
	                              / "\\t"                      { p.AddCharacter("\t") }   # ht
	                              / "\\v"                      { p.AddCharacter("\v") }   # vt
	                              / "\\'"                      { p.AddCharacter("'") }
	                              / '\\"'                      { p.AddCharacter("\"") }
	                              / '\\['                      { p.AddCharacter("[") }
	                              / '\\]'                      { p.AddCharacter("]") }
	                              / '\\-'                      { p.AddCharacter("-") }
				      / '\\' "0x"<[0-9a-fA-F]+>    { p.AddHexaCharacter(text) }
	                              / '\\' <[0-3][0-7][0-7]>     { p.AddOctalCharacter(text) }
	                              / '\\' <[0-7][0-7]?>         { p.AddOctalCharacter(text) }
	                              / '\\\\'                     { p.AddCharacter("\\") } */
	addRule("Escape", func() {
		addAlternate(
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`a`) },
					func() { addAction(` p.AddCharacter("\a") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`b`) },
					func() { addAction(` p.AddCharacter("\b") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`e`) },
					func() { addAction(` p.AddCharacter("\x1B") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`f`) },
					func() { addAction(` p.AddCharacter("\f") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`n`) },
					func() { addAction(` p.AddCharacter("\n") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`r`) },
					func() { addAction(` p.AddCharacter("\r") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`t`) },
					func() { addAction(` p.AddCharacter("\t") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addDoubleCharacter(`v`) },
					func() { addAction(` p.AddCharacter("\v") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter(`'`) },
					func() { addAction(` p.AddCharacter("'") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter(`"`) },
					func() { addAction(` p.AddCharacter("\"") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter(`[`) },
					func() { addAction(` p.AddCharacter("[") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter(`]`) },
					func() { addAction(` p.AddCharacter("]") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter(`-`) },
					func() { addAction(` p.AddCharacter("-") `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() {
						addSequence(
							func() { addCharacter(`0`) },
							func() { addDoubleCharacter(`x`) },
						)
					},
					func() {
						addPush(func() {
							addPlus(func() {
								addAlternate(
									func() { addRange(`0`, `9`) },
									func() { addRange(`a`, `f`) },
									func() { addRange(`A`, `F`) },
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
					func() {
						addPush(func() {
							addSequence(
								func() { addRange(`0`, `3`) },
								func() { addRange(`0`, `7`) },
								func() { addRange(`0`, `7`) },
							)
						})
					},
					func() { addAction(` p.AddOctalCharacter(text) `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() {
						addPush(func() {
							addSequence(
								func() { addRange(`0`, `7`) },
								func() { addQuery(func() { addRange(`0`, `7`) }) },
							)
						})
					},
					func() { addAction(` p.AddOctalCharacter(text) `) },
				)
			},
			func() {
				addSequence(
					func() { addCharacter("\\") },
					func() { addCharacter("\\") },
					func() { addAction(` p.AddCharacter("\\") `) },
				)
			},
		)
	})

	/* LeftArrow       <- ('<-' / '\0x2190') Spacing */
	addRule("LeftArrow", func() {
		addSequence(
			func() {
				addAlternate(
					func() { addString(`<-`) },
					func() { addHexaCharacter("2190") },
				)
			},
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

	/* And             <- '&' Spacing */
	addRule("And", func() {
		addSequence(
			func() { addCharacter(`&`) },
			func() { addName("Spacing") },
		)
	})

	/* Not             <- '!' Spacing */
	addRule("Not", func() {
		addSequence(
			func() { addCharacter(`!`) },
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

	/* Plus            <- '+' Spacing */
	addRule("Plus", func() {
		addSequence(
			func() { addCharacter(`+`) },
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

	/* SpaceComment         <- (Space / Comment) */
	addRule("SpaceComment", func() {
		addAlternate(
			func() { addName("Space") },
			func() { addName("Comment") },
		)
	})

	/* Spacing         <- SpaceComment* */
	addRule("Spacing", func() {
		addStar(func() { addName("SpaceComment") })
	})

	/* MustSpacing     <- SpaceComment+ */
	addRule("MustSpacing", func() {
		addPlus(func() { t.AddName("SpaceComment") })
	})

	/* Comment         <- '#' (!EndOfLine .)* EndOfLine */
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
			func() { addName("EndOfLine") },
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

	/* EndOfFile       <- !. */
	addRule("EndOfFile", func() {
		addPeekNot(func() { addDot() })
	})

	/* Action		<- '{' < ActionBody* > '}' Spacing */
	addRule("Action", func() {
		addSequence(
			func() { addCharacter(`{`) },
			func() {
				addPush(func() {
					addStar(func() { addName("ActionBody") })
				})
			},
			func() { addCharacter(`}`) },
			func() { addName("Spacing") },
		)
	})

	/* ActionBody	<- [^{}] / '{' ActionBody* '}' */
	addRule("ActionBody", func() {
		addAlternate(
			func() {
				addSequence(
					func() {
						addPeekNot(func() {
							addAlternate(
								func() { addCharacter(`{`) },
								func() { addCharacter(`}`) },
							)
						})
					},
					func() { addDot() },
				)
			},
			func() {
				addSequence(
					func() { addCharacter(`{`) },
					func() { addStar(func() { addName("ActionBody") }) },
					func() { addCharacter(`}`) },
				)
			},
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
	out, error := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if error != nil {
		fmt.Printf("%v: %v\n", filename, error)
		return
	}
	defer out.Close()
	t.Compile(filename, os.Args, out)
}
