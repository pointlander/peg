// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)
	t := New(true, true)

	/*package main

	  type Peg Peg {
	   *Tree
	  }*/
	t.AddPackage("main")
	t.AddPeg("Peg")
	t.AddState(`
 *Tree
`)

	/* Grammar         <- Spacing 'package' Spacing Identifier      { p.AddPackage(buffer[begin:end]) }
	   'type' Spacing Identifier         { p.AddPeg(buffer[begin:end]) }
	   'Peg' Spacing Action              { p.AddState(buffer[begin:end]) }
	   commit
	   Definition+ EndOfFile */
	t.AddRule("Grammar")
	t.AddName("Spacing")
	t.AddString("package")
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddName("Identifier")
	t.AddSequence()
	t.AddAction(" p.AddPackage(buffer[begin:end]) ")
	t.AddSequence()
	t.AddString("type")
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddName("Identifier")
	t.AddSequence()
	t.AddAction(" p.AddPeg(buffer[begin:end]) ")
	t.AddSequence()
	t.AddString("Peg")
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddName("Action")
	t.AddSequence()
	t.AddAction(" p.AddState(buffer[begin:end]) ")
	t.AddSequence()
	t.AddCommit()
	t.AddSequence()
	t.AddName("Definition")
	t.AddPlus()
	t.AddSequence()
	t.AddName("EndOfFile")
	t.AddSequence()
	t.AddExpression()

	/* Definition      <- Identifier                   { p.AddRule(buffer[begin:end]) }
	   LEFTARROW Expression         { p.AddExpression() } &(Identifier LEFTARROW / !.) commit */
	t.AddRule("Definition")
	t.AddName("Identifier")
	t.AddAction(" p.AddRule(buffer[begin:end]) ")
	t.AddSequence()
	t.AddName("LEFTARROW")
	t.AddSequence()
	t.AddName("Expression")
	t.AddSequence()
	t.AddAction(" p.AddExpression() ")
	t.AddSequence()
	t.AddName("Identifier")
	t.AddName("LEFTARROW")
	t.AddSequence()
	t.AddDot()
	t.AddPeekNot()
	t.AddAlternate()
	t.AddPeekFor()
	t.AddSequence()
	t.AddCommit()
	t.AddSequence()
	t.AddExpression()

	/* Expression      <- Sequence (SLASH Sequence     { p.AddAlternate() }
	           )* (SLASH           { p.AddEmptyAlternate() }
	              )?
	/ */
	t.AddRule("Expression")
	t.AddName("Sequence")
	t.AddName("SLASH")
	t.AddName("Sequence")
	t.AddSequence()
	t.AddAction(" p.AddAlternate() ")
	t.AddSequence()
	t.AddStar()
	t.AddSequence()
	t.AddName("SLASH")
	t.AddAction(" p.AddEmptyAlternate() ")
	t.AddSequence()
	t.AddQuery()
	t.AddSequence()
	t.AddEmptyAlternate()
	t.AddExpression()

	/* Sequence        <- Prefix (Prefix               { p.AddSequence() }
	   )* */
	t.AddRule("Sequence")
	t.AddName("Prefix")
	t.AddName("Prefix")
	t.AddAction(" p.AddSequence() ")
	t.AddSequence()
	t.AddStar()
	t.AddSequence()
	t.AddExpression()

	/* Prefix          <- AND Action                   { p.AddPredicate(buffer[begin:end]) }
	   / AND Suffix                   { p.AddPeekFor() }
	   / NOT Suffix                   { p.AddPeekNot() }
	   /     Suffix */
	t.AddRule("Prefix")
	t.AddName("AND")
	t.AddName("Action")
	t.AddSequence()
	t.AddAction(" p.AddPredicate(buffer[begin:end]) ")
	t.AddSequence()
	t.AddName("AND")
	t.AddName("Suffix")
	t.AddSequence()
	t.AddAction(" p.AddPeekFor() ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("NOT")
	t.AddName("Suffix")
	t.AddSequence()
	t.AddAction(" p.AddPeekNot() ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("Suffix")
	t.AddAlternate()
	t.AddExpression()

	/* Suffix          <- Primary (QUESTION            { p.AddQuery() }
	   / STAR             { p.AddStar() }
	   / PLUS             { p.AddPlus() }
	 )? */
	t.AddRule("Suffix")
	t.AddName("Primary")
	t.AddName("QUESTION")
	t.AddAction(" p.AddQuery() ")
	t.AddSequence()
	t.AddName("STAR")
	t.AddAction(" p.AddStar() ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("PLUS")
	t.AddAction(" p.AddPlus() ")
	t.AddSequence()
	t.AddAlternate()
	t.AddQuery()
	t.AddSequence()
	t.AddExpression()

	/* Primary         <- 'commit' Spacing             { p.AddCommit() }
	   / Identifier !LEFTARROW        { p.AddName(buffer[begin:end]) }
	   / OPEN Expression CLOSE
	   / Literal                      { p.AddString(buffer[begin:end]) }
	   / Class                        { p.AddClass(buffer[begin:end]) }
	   / DOT                          { p.AddDot() }
	   / Action                       { p.AddAction(buffer[begin:end]) }
	   / BEGIN                        { p.AddBegin() }
	   / END                          { p.AddEnd() } */
	t.AddRule("Primary")
	t.AddString("commit")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddAction(" p.AddCommit() ")
	t.AddSequence()
	t.AddName("Identifier")
	t.AddName("LEFTARROW")
	t.AddPeekNot()
	t.AddSequence()
	t.AddAction(" p.AddName(buffer[begin:end]) ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("OPEN")
	t.AddName("Expression")
	t.AddSequence()
	t.AddName("CLOSE")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("Literal")
	t.AddAction(" p.AddString(buffer[begin:end]) ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("Class")
	t.AddAction(" p.AddClass(buffer[begin:end]) ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("DOT")
	t.AddAction(" p.AddDot() ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("Action")
	t.AddAction(" p.AddAction(buffer[begin:end]) ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("BEGIN")
	t.AddAction(" p.AddBegin() ")
	t.AddSequence()
	t.AddAlternate()
	t.AddName("END")
	t.AddAction(" p.AddEnd() ")
	t.AddSequence()
	t.AddAlternate()
	t.AddExpression()

	/* Identifier      <- < IdentStart IdentCont* > Spacing */
	t.AddRule("Identifier")
	t.AddBegin()
	t.AddName("IdentStart")
	t.AddSequence()
	t.AddName("IdentCont")
	t.AddStar()
	t.AddSequence()
	t.AddEnd()
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* IdentStart      <- [a-zA-Z_] */
	t.AddRule("IdentStart")
	t.AddClass("a-zA-Z_")
	t.AddExpression()

	/* IdentCont       <- IdentStart / [0-9] */
	t.AddRule("IdentCont")
	t.AddName("IdentStart")
	t.AddClass("0-9")
	t.AddAlternate()
	t.AddExpression()

	/* Literal         <- ['] < (!['] Char )* > ['] Spacing
	   / ["] < (!["] Char )* > ["] Spacing */
	t.AddRule("Literal")
	t.AddClass("'")
	t.AddBegin()
	t.AddSequence()
	t.AddClass("'")
	t.AddPeekNot()
	t.AddName("Char")
	t.AddSequence()
	t.AddStar()
	t.AddSequence()
	t.AddEnd()
	t.AddSequence()
	t.AddClass("'")
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddClass(`"`)
	t.AddBegin()
	t.AddSequence()
	t.AddClass(`"`)
	t.AddPeekNot()
	t.AddName("Char")
	t.AddSequence()
	t.AddStar()
	t.AddSequence()
	t.AddEnd()
	t.AddSequence()
	t.AddClass(`"`)
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddAlternate()
	t.AddExpression()

	/* Class           <- '[' < (!']' Range)* > ']' Spacing */
	t.AddRule("Class")
	t.AddString("[")
	t.AddBegin()
	t.AddSequence()
	t.AddString("]")
	t.AddPeekNot()
	t.AddName("Range")
	t.AddSequence()
	t.AddStar()
	t.AddSequence()
	t.AddEnd()
	t.AddSequence()
	t.AddString("]")
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* Range           <- Char '-' Char / Char */
	t.AddRule("Range")
	t.AddName("Char")
	t.AddString("-")
	t.AddSequence()
	t.AddName("Char")
	t.AddSequence()
	t.AddName("Char")
	t.AddAlternate()
	t.AddExpression()

	/* Char            <- '\\' [abefnrtv'"\[\]\\]
	   / '\\' [0-3][0-7][0-7]
	   / '\\' [0-7][0-7]?
	   / '\\' '-'
	   / !'\\' . */
	t.AddRule("Char")
	t.AddString(`\\`)
	t.AddClass(`abefnrtv'"\[\]\\`)
	t.AddSequence()
	t.AddString(`\\`)
	t.AddClass("0-3")
	t.AddSequence()
	t.AddClass("0-7")
	t.AddSequence()
	t.AddClass("0-7")
	t.AddSequence()
	t.AddAlternate()
	t.AddString(`\\`)
	t.AddClass("0-7")
	t.AddSequence()
	t.AddClass("0-7")
	t.AddQuery()
	t.AddSequence()
	t.AddAlternate()
	t.AddString(`\\`)
	t.AddString("-")
	t.AddSequence()
	t.AddAlternate()
	t.AddString(`\\`)
	t.AddPeekNot()
	t.AddDot()
	t.AddSequence()
	t.AddAlternate()
	t.AddExpression()

	/* LEFTARROW       <- '<-' Spacing */
	t.AddRule("LEFTARROW")
	t.AddString("<-")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* SLASH           <- '/' Spacing */
	t.AddRule("SLASH")
	t.AddString("/")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* AND             <- '&' Spacing */
	t.AddRule("AND")
	t.AddString("&")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* NOT             <- '!' Spacing */
	t.AddRule("NOT")
	t.AddString("!")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* QUESTION        <- '?' Spacing */
	t.AddRule("QUESTION")
	t.AddString("?")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* STAR            <- '*' Spacing */
	t.AddRule("STAR")
	t.AddString("*")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* PLUS            <- '+' Spacing */
	t.AddRule("PLUS")
	t.AddString("+")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* OPEN            <- '(' Spacing */
	t.AddRule("OPEN")
	t.AddString("(")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* CLOSE           <- ')' Spacing */
	t.AddRule("CLOSE")
	t.AddString(")")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* DOT             <- '.' Spacing */
	t.AddRule("DOT")
	t.AddString(".")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* Spacing         <- (Space / Comment)* */
	t.AddRule("Spacing")
	t.AddName("Space")
	t.AddName("Comment")
	t.AddAlternate()
	t.AddStar()
	t.AddExpression()

	/* Comment         <- '#' (!EndOfLine .)* EndOfLine */
	t.AddRule("Comment")
	t.AddString("#")
	t.AddName("EndOfLine")
	t.AddPeekNot()
	t.AddDot()
	t.AddSequence()
	t.AddStar()
	t.AddSequence()
	t.AddName("EndOfLine")
	t.AddSequence()
	t.AddExpression()

	/* Space           <- ' ' / '\t' / EndOfLine */
	t.AddRule("Space")
	t.AddString(" ")
	t.AddString(`\t`)
	t.AddAlternate()
	t.AddName("EndOfLine")
	t.AddAlternate()
	t.AddExpression()

	/* EndOfLine       <- '\r\n' / '\n' / '\r' */
	t.AddRule("EndOfLine")
	t.AddString(`\r\n`)
	t.AddString(`\n`)
	t.AddAlternate()
	t.AddString(`\r`)
	t.AddAlternate()
	t.AddExpression()

	/* EndOfFile       <- !. */
	t.AddRule("EndOfFile")
	t.AddDot()
	t.AddPeekNot()
	t.AddExpression()

	/* Action          <- '{' < [^}]* > '}' Spacing */
	t.AddRule("Action")
	t.AddString("{")
	t.AddBegin()
	t.AddSequence()
	t.AddClass("^}")
	t.AddStar()
	t.AddSequence()
	t.AddEnd()
	t.AddSequence()
	t.AddString("}")
	t.AddSequence()
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* BEGIN           <- '<' Spacing */
	t.AddRule("BEGIN")
	t.AddString("<")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	/* END             <- '>' Spacing */
	t.AddRule("END")
	t.AddString(">")
	t.AddName("Spacing")
	t.AddSequence()
	t.AddExpression()

	t.Compile("bootstrap.peg.go")
}
