package main

import (
	"bytes"
	"fmt"
	"os"
)

type Peg struct {
	*Tree

	Buffer   string
	Min, Max int
	rules    [33]func() bool
}

type parseError struct {
	p *Peg
}

func (p *Peg) Parse() os.Error {
	if p.rules[0]() {
		return nil
	}
	return &parseError{p}
}

func (e *parseError) String() string {
	buf := new(bytes.Buffer)
	line := 1
	character := 0
	for i, c := range e.p.Buffer[0:] {
		if c == '\n' {
			line++
			character = 0
		} else {
			character++
		}
		if i == e.p.Min {
			if e.p.Min != e.p.Max {
				fmt.Fprintf(buf, "parse error after line %v character %v\n", line, character)
			} else {
				break
			}
		} else if i == e.p.Max {
			break
		}
	}
	fmt.Fprintf(buf, "parse error: unexpected ")
	if e.p.Max >= len(e.p.Buffer) {
		fmt.Fprintf(buf, "end of file found\n")
	} else {
		fmt.Fprintf(buf, "'%c' at line %v character %v\n", e.p.Buffer[e.p.Max], line, character)
	}
	return buf.String()
}
func (p *Peg) Init() {
	var position int
	actions := [...]func(buffer string, begin, end int){
		/* 0 Grammar */
		func(buffer string, begin, end int) {
			p.AddPackage(buffer[begin:end])
		},
		/* 1 Grammar */
		func(buffer string, begin, end int) {
			p.AddPeg(buffer[begin:end])
		},
		/* 2 Grammar */
		func(buffer string, begin, end int) {
			p.AddState(buffer[begin:end])
		},
		/* 3 Definition */
		func(buffer string, begin, end int) {
			p.AddRule(buffer[begin:end])
		},
		/* 4 Definition */
		func(buffer string, begin, end int) {
			p.AddExpression()
		},
		/* 5 Expression */
		func(buffer string, begin, end int) {
			p.AddAlternate()
		},
		/* 6 Expression */
		func(buffer string, begin, end int) {
			p.AddEmptyAlternate()
		},
		/* 7 Sequence */
		func(buffer string, begin, end int) {
			p.AddSequence()
		},
		/* 8 Prefix */
		func(buffer string, begin, end int) {
			p.AddPredicate(buffer[begin:end])
		},
		/* 9 Prefix */
		func(buffer string, begin, end int) {
			p.AddPeekFor()
		},
		/* 10 Prefix */
		func(buffer string, begin, end int) {
			p.AddPeekNot()
		},
		/* 11 Suffix */
		func(buffer string, begin, end int) {
			p.AddQuery()
		},
		/* 12 Suffix */
		func(buffer string, begin, end int) {
			p.AddStar()
		},
		/* 13 Suffix */
		func(buffer string, begin, end int) {
			p.AddPlus()
		},
		/* 14 Primary */
		func(buffer string, begin, end int) {
			p.AddCommit()
		},
		/* 15 Primary */
		func(buffer string, begin, end int) {
			p.AddName(buffer[begin:end])
		},
		/* 16 Primary */
		func(buffer string, begin, end int) {
			p.AddDot()
		},
		/* 17 Primary */
		func(buffer string, begin, end int) {
			p.AddAction(buffer[begin:end])
		},
		/* 18 Primary */
		func(buffer string, begin, end int) {
			p.AddBegin()
		},
		/* 19 Primary */
		func(buffer string, begin, end int) {
			p.AddEnd()
		},
		/* 20 Literal */
		func(buffer string, begin, end int) {
			p.AddSequence()
		},
		/* 21 Literal */
		func(buffer string, begin, end int) {
			p.AddSequence()
		},
		/* 22 Class */
		func(buffer string, begin, end int) {
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		},
		/* 23 Ranges */
		func(buffer string, begin, end int) {
			p.AddAlternate()
		},
		/* 24 Range */
		func(buffer string, begin, end int) {
			p.AddRange()
		},
		/* 25 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\a")
		},
		/* 26 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\b")
		},
		/* 27 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\x1B")
		},
		/* 28 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\f")
		},
		/* 29 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\n")
		},
		/* 30 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\r")
		},
		/* 31 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\t")
		},
		/* 32 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\v")
		},
		/* 33 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("'")
		},
		/* 34 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\"")
		},
		/* 35 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("[")
		},
		/* 36 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("]")
		},
		/* 37 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("-")
		},
		/* 38 Char */
		func(buffer string, begin, end int) {
			p.AddOctalCharacter(buffer[begin:end])
		},
		/* 39 Char */
		func(buffer string, begin, end int) {
			p.AddOctalCharacter(buffer[begin:end])
		},
		/* 40 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter("\\")
		},
		/* 41 Char */
		func(buffer string, begin, end int) {
			p.AddCharacter(buffer[begin:end])
		},
	}
	var thunkPosition, begin, end int
	thunks := make([]struct {
		action     uint8
		begin, end int
	}, 32)
	do := func(action uint8) {
		if thunkPosition == len(thunks) {
			newThunks := make([]struct {
				action     uint8
				begin, end int
			}, 2*len(thunks))
			copy(newThunks, thunks)
			thunks = newThunks
		}
		thunks[thunkPosition].action = action
		thunks[thunkPosition].begin = begin
		thunks[thunkPosition].end = end
		thunkPosition++
	}
	commit := func(thunkPosition0 int) bool {
		if thunkPosition0 == 0 {
			for thunk := 0; thunk < thunkPosition; thunk++ {
				actions[thunks[thunk].action](p.Buffer, thunks[thunk].begin, thunks[thunk].end)
			}
			p.Min = position
			thunkPosition = 0
			return true
		}
		return false
	}
	matchDot := func() bool {
		if position < len(p.Buffer) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	matchChar := func(c byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] == c) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	matchRange := func(lower byte, upper byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] >= lower) && (p.Buffer[position] <= upper) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	p.rules = [...]func() bool{
		/* 0 Grammar <- (Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') Spacing Identifier { p.AddPackage(buffer[begin:end]) } ('t' 'y' 'p' 'e') Spacing Identifier { p.AddPeg(buffer[begin:end]) } ('P' 'e' 'g') Spacing Action { p.AddState(buffer[begin:end]) } commit Definition+ EndOfFile) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[25]() {
				goto l0
			}
			if !matchChar('p') {
				goto l0
			}
			if !matchChar('a') {
				goto l0
			}
			if !matchChar('c') {
				goto l0
			}
			if !matchChar('k') {
				goto l0
			}
			if !matchChar('a') {
				goto l0
			}
			if !matchChar('g') {
				goto l0
			}
			if !matchChar('e') {
				goto l0
			}
			if !p.rules[25]() {
				goto l0
			}
			if !p.rules[7]() {
				goto l0
			}
			do(0)
			if !matchChar('t') {
				goto l0
			}
			if !matchChar('y') {
				goto l0
			}
			if !matchChar('p') {
				goto l0
			}
			if !matchChar('e') {
				goto l0
			}
			if !p.rules[25]() {
				goto l0
			}
			if !p.rules[7]() {
				goto l0
			}
			do(1)
			if !matchChar('P') {
				goto l0
			}
			if !matchChar('e') {
				goto l0
			}
			if !matchChar('g') {
				goto l0
			}
			if !p.rules[25]() {
				goto l0
			}
			if !p.rules[30]() {
				goto l0
			}
			do(2)
			if !(commit(thunkPosition0)) {
				goto l0
			}
			if !p.rules[1]() {
				goto l0
			}
		l1:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[1]() {
					goto l2
				}
				goto l1
			l2:
				position, thunkPosition = position2, thunkPosition2
			}
			if !p.rules[29]() {
				goto l0
			}
			return true
		l0:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 1 Definition <- (Identifier { p.AddRule(buffer[begin:end]) } LEFTARROW Expression { p.AddExpression() } &((Identifier LEFTARROW) / !.) commit) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[7]() {
				goto l3
			}
			do(3)
			if !p.rules[15]() {
				goto l3
			}
			if !p.rules[2]() {
				goto l3
			}
			do(4)
			{
				position4, thunkPosition4 := position, thunkPosition
				{
					position5, thunkPosition5 := position, thunkPosition
					if !p.rules[7]() {
						goto l6
					}
					if !p.rules[15]() {
						goto l6
					}
					goto l5
				l6:
					position, thunkPosition = position5, thunkPosition5
					{
						position7, thunkPosition7 := position, thunkPosition
						if !matchDot() {
							goto l7
						}
						goto l3
					l7:
						position, thunkPosition = position7, thunkPosition7
					}
				}
			l5:
				position, thunkPosition = position4, thunkPosition4
			}
			if !(commit(thunkPosition0)) {
				goto l3
			}
			return true
		l3:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 2 Expression <- ((Sequence (SLASH Sequence { p.AddAlternate() })* (SLASH { p.AddEmptyAlternate() })?) /) */
		func() bool {
			{
				position9, thunkPosition9 := position, thunkPosition
				if !p.rules[3]() {
					goto l10
				}
			l11:
				{
					position12, thunkPosition12 := position, thunkPosition
					if !p.rules[16]() {
						goto l12
					}
					if !p.rules[3]() {
						goto l12
					}
					do(5)
					goto l11
				l12:
					position, thunkPosition = position12, thunkPosition12
				}
				{
					position13, thunkPosition13 := position, thunkPosition
					if !p.rules[16]() {
						goto l13
					}
					do(6)
					goto l14
				l13:
					position, thunkPosition = position13, thunkPosition13
				}
			l14:
				goto l9
			l10:
				position, thunkPosition = position9, thunkPosition9
			}
		l9:
			return true
		},
		/* 3 Sequence <- (Prefix (Prefix { p.AddSequence() })*) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[4]() {
				goto l15
			}
		l16:
			{
				position17, thunkPosition17 := position, thunkPosition
				if !p.rules[4]() {
					goto l17
				}
				do(7)
				goto l16
			l17:
				position, thunkPosition = position17, thunkPosition17
			}
			return true
		l15:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 4 Prefix <- ((AND Action { p.AddPredicate(buffer[begin:end]) }) / (AND Suffix { p.AddPeekFor() }) / (NOT Suffix { p.AddPeekNot() }) / Suffix) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position19, thunkPosition19 := position, thunkPosition
				if !p.rules[17]() {
					goto l20
				}
				if !p.rules[30]() {
					goto l20
				}
				do(8)
				goto l19
			l20:
				position, thunkPosition = position19, thunkPosition19
				if !p.rules[17]() {
					goto l21
				}
				if !p.rules[5]() {
					goto l21
				}
				do(9)
				goto l19
			l21:
				position, thunkPosition = position19, thunkPosition19
				if !p.rules[18]() {
					goto l22
				}
				if !p.rules[5]() {
					goto l22
				}
				do(10)
				goto l19
			l22:
				position, thunkPosition = position19, thunkPosition19
				if !p.rules[5]() {
					goto l18
				}
			}
		l19:
			return true
		l18:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 5 Suffix <- (Primary ((QUESTION { p.AddQuery() }) / (STAR { p.AddStar() }) / (PLUS { p.AddPlus() }))?) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[6]() {
				goto l23
			}
			{
				position24, thunkPosition24 := position, thunkPosition
				{
					position26, thunkPosition26 := position, thunkPosition
					if !p.rules[19]() {
						goto l27
					}
					do(11)
					goto l26
				l27:
					position, thunkPosition = position26, thunkPosition26
					if !p.rules[20]() {
						goto l28
					}
					do(12)
					goto l26
				l28:
					position, thunkPosition = position26, thunkPosition26
					if !p.rules[21]() {
						goto l24
					}
					do(13)
				}
			l26:
				goto l25
			l24:
				position, thunkPosition = position24, thunkPosition24
			}
		l25:
			return true
		l23:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 6 Primary <- (('c' 'o' 'm' 'm' 'i' 't' Spacing { p.AddCommit() }) / (Identifier !LEFTARROW { p.AddName(buffer[begin:end]) }) / (OPEN Expression CLOSE) / Literal / Class / (DOT { p.AddDot() }) / (Action { p.AddAction(buffer[begin:end]) }) / (BEGIN { p.AddBegin() }) / (END { p.AddEnd() })) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position30, thunkPosition30 := position, thunkPosition
				if !matchChar('c') {
					goto l31
				}
				if !matchChar('o') {
					goto l31
				}
				if !matchChar('m') {
					goto l31
				}
				if !matchChar('m') {
					goto l31
				}
				if !matchChar('i') {
					goto l31
				}
				if !matchChar('t') {
					goto l31
				}
				if !p.rules[25]() {
					goto l31
				}
				do(14)
				goto l30
			l31:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[7]() {
					goto l32
				}
				{
					position33, thunkPosition33 := position, thunkPosition
					if !p.rules[15]() {
						goto l33
					}
					goto l32
				l33:
					position, thunkPosition = position33, thunkPosition33
				}
				do(15)
				goto l30
			l32:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[22]() {
					goto l34
				}
				if !p.rules[2]() {
					goto l34
				}
				if !p.rules[23]() {
					goto l34
				}
				goto l30
			l34:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[10]() {
					goto l35
				}
				goto l30
			l35:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[11]() {
					goto l36
				}
				goto l30
			l36:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[24]() {
					goto l37
				}
				do(16)
				goto l30
			l37:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[30]() {
					goto l38
				}
				do(17)
				goto l30
			l38:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[31]() {
					goto l39
				}
				do(18)
				goto l30
			l39:
				position, thunkPosition = position30, thunkPosition30
				if !p.rules[32]() {
					goto l29
				}
				do(19)
			}
		l30:
			return true
		l29:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 7 Identifier <- (< IdentStart IdentCont* > Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[8]() {
				goto l40
			}
		l41:
			{
				position42, thunkPosition42 := position, thunkPosition
				if !p.rules[9]() {
					goto l42
				}
				goto l41
			l42:
				position, thunkPosition = position42, thunkPosition42
			}
			end = position
			if !p.rules[25]() {
				goto l40
			}
			return true
		l40:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 IdentStart <- ([a-z] / [A-Z] / '_') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position44, thunkPosition44 := position, thunkPosition
				if !matchRange('a', 'z') {
					goto l45
				}
				goto l44
			l45:
				position, thunkPosition = position44, thunkPosition44
				if !matchRange('A', 'Z') {
					goto l46
				}
				goto l44
			l46:
				position, thunkPosition = position44, thunkPosition44
				if !matchChar('_') {
					goto l43
				}
			}
		l44:
			return true
		l43:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 9 IdentCont <- (IdentStart / [0-9]) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position48, thunkPosition48 := position, thunkPosition
				if !p.rules[8]() {
					goto l49
				}
				goto l48
			l49:
				position, thunkPosition = position48, thunkPosition48
				if !matchRange('0', '9') {
					goto l47
				}
			}
		l48:
			return true
		l47:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 10 Literal <- (('\'' (!'\'' Char)? (!'\'' Char { p.AddSequence() })* '\'' Spacing) / ('"' (!'"' Char)? (!'"' Char { p.AddSequence() })* '"' Spacing)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position51, thunkPosition51 := position, thunkPosition
				if !matchChar('\'') {
					goto l52
				}
				{
					position53, thunkPosition53 := position, thunkPosition
					{
						position55, thunkPosition55 := position, thunkPosition
						if !matchChar('\'') {
							goto l55
						}
						goto l53
					l55:
						position, thunkPosition = position55, thunkPosition55
					}
					if !p.rules[14]() {
						goto l53
					}
					goto l54
				l53:
					position, thunkPosition = position53, thunkPosition53
				}
			l54:
			l56:
				{
					position57, thunkPosition57 := position, thunkPosition
					{
						position58, thunkPosition58 := position, thunkPosition
						if !matchChar('\'') {
							goto l58
						}
						goto l57
					l58:
						position, thunkPosition = position58, thunkPosition58
					}
					if !p.rules[14]() {
						goto l57
					}
					do(20)
					goto l56
				l57:
					position, thunkPosition = position57, thunkPosition57
				}
				if !matchChar('\'') {
					goto l52
				}
				if !p.rules[25]() {
					goto l52
				}
				goto l51
			l52:
				position, thunkPosition = position51, thunkPosition51
				if !matchChar('"') {
					goto l50
				}
				{
					position59, thunkPosition59 := position, thunkPosition
					{
						position61, thunkPosition61 := position, thunkPosition
						if !matchChar('"') {
							goto l61
						}
						goto l59
					l61:
						position, thunkPosition = position61, thunkPosition61
					}
					if !p.rules[14]() {
						goto l59
					}
					goto l60
				l59:
					position, thunkPosition = position59, thunkPosition59
				}
			l60:
			l62:
				{
					position63, thunkPosition63 := position, thunkPosition
					{
						position64, thunkPosition64 := position, thunkPosition
						if !matchChar('"') {
							goto l64
						}
						goto l63
					l64:
						position, thunkPosition = position64, thunkPosition64
					}
					if !p.rules[14]() {
						goto l63
					}
					do(21)
					goto l62
				l63:
					position, thunkPosition = position63, thunkPosition63
				}
				if !matchChar('"') {
					goto l50
				}
				if !p.rules[25]() {
					goto l50
				}
			}
		l51:
			return true
		l50:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 11 Class <- ('[' (('^' Ranges { p.AddPeekNot(); p.AddDot(); p.AddSequence() }) / Ranges)? ']' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('[') {
				goto l65
			}
			{
				position66, thunkPosition66 := position, thunkPosition
				{
					position68, thunkPosition68 := position, thunkPosition
					if !matchChar('^') {
						goto l69
					}
					if !p.rules[12]() {
						goto l69
					}
					do(22)
					goto l68
				l69:
					position, thunkPosition = position68, thunkPosition68
					if !p.rules[12]() {
						goto l66
					}
				}
			l68:
				goto l67
			l66:
				position, thunkPosition = position66, thunkPosition66
			}
		l67:
			if !matchChar(']') {
				goto l65
			}
			if !p.rules[25]() {
				goto l65
			}
			return true
		l65:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 12 Ranges <- (!']' Range (!']' Range { p.AddAlternate() })*) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position71, thunkPosition71 := position, thunkPosition
				if !matchChar(']') {
					goto l71
				}
				goto l70
			l71:
				position, thunkPosition = position71, thunkPosition71
			}
			if !p.rules[13]() {
				goto l70
			}
		l72:
			{
				position73, thunkPosition73 := position, thunkPosition
				{
					position74, thunkPosition74 := position, thunkPosition
					if !matchChar(']') {
						goto l74
					}
					goto l73
				l74:
					position, thunkPosition = position74, thunkPosition74
				}
				if !p.rules[13]() {
					goto l73
				}
				do(23)
				goto l72
			l73:
				position, thunkPosition = position73, thunkPosition73
			}
			return true
		l70:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 13 Range <- ((Char '-' Char { p.AddRange() }) / Char) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position76, thunkPosition76 := position, thunkPosition
				if !p.rules[14]() {
					goto l77
				}
				if !matchChar('-') {
					goto l77
				}
				if !p.rules[14]() {
					goto l77
				}
				do(24)
				goto l76
			l77:
				position, thunkPosition = position76, thunkPosition76
				if !p.rules[14]() {
					goto l75
				}
			}
		l76:
			return true
		l75:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 14 Char <- (('\\' 'a' { p.AddCharacter("\a") }) / ('\\' 'b' { p.AddCharacter("\b") }) / ('\\' 'e' { p.AddCharacter("\x1B") }) / ('\\' 'f' { p.AddCharacter("\f") }) / ('\\' 'n' { p.AddCharacter("\n") }) / ('\\' 'r' { p.AddCharacter("\r") }) / ('\\' 't' { p.AddCharacter("\t") }) / ('\\' 'v' { p.AddCharacter("\v") }) / ('\\' '\'' { p.AddCharacter("'") }) / ('\\' '"' { p.AddCharacter("\"") }) / ('\\' '[' { p.AddCharacter("[") }) / ('\\' ']' { p.AddCharacter("]") }) / ('\\' '-' { p.AddCharacter("-") }) / ('\\' < [0-3] [0-7] [0-7] > { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' < [0-7] [0-7]? > { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' '\\' { p.AddCharacter("\\") }) / (!'\\' < . > { p.AddCharacter(buffer[begin:end]) })) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position79, thunkPosition79 := position, thunkPosition
				if !matchChar('\\') {
					goto l80
				}
				if !matchChar('a') {
					goto l80
				}
				do(25)
				goto l79
			l80:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l81
				}
				if !matchChar('b') {
					goto l81
				}
				do(26)
				goto l79
			l81:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l82
				}
				if !matchChar('e') {
					goto l82
				}
				do(27)
				goto l79
			l82:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l83
				}
				if !matchChar('f') {
					goto l83
				}
				do(28)
				goto l79
			l83:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l84
				}
				if !matchChar('n') {
					goto l84
				}
				do(29)
				goto l79
			l84:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l85
				}
				if !matchChar('r') {
					goto l85
				}
				do(30)
				goto l79
			l85:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l86
				}
				if !matchChar('t') {
					goto l86
				}
				do(31)
				goto l79
			l86:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l87
				}
				if !matchChar('v') {
					goto l87
				}
				do(32)
				goto l79
			l87:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l88
				}
				if !matchChar('\'') {
					goto l88
				}
				do(33)
				goto l79
			l88:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l89
				}
				if !matchChar('"') {
					goto l89
				}
				do(34)
				goto l79
			l89:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l90
				}
				if !matchChar('[') {
					goto l90
				}
				do(35)
				goto l79
			l90:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l91
				}
				if !matchChar(']') {
					goto l91
				}
				do(36)
				goto l79
			l91:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l92
				}
				if !matchChar('-') {
					goto l92
				}
				do(37)
				goto l79
			l92:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l93
				}
				begin = position
				if !matchRange('0', '3') {
					goto l93
				}
				if !matchRange('0', '7') {
					goto l93
				}
				if !matchRange('0', '7') {
					goto l93
				}
				end = position
				do(38)
				goto l79
			l93:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l94
				}
				begin = position
				if !matchRange('0', '7') {
					goto l94
				}
				{
					position95, thunkPosition95 := position, thunkPosition
					if !matchRange('0', '7') {
						goto l95
					}
					goto l96
				l95:
					position, thunkPosition = position95, thunkPosition95
				}
			l96:
				end = position
				do(39)
				goto l79
			l94:
				position, thunkPosition = position79, thunkPosition79
				if !matchChar('\\') {
					goto l97
				}
				if !matchChar('\\') {
					goto l97
				}
				do(40)
				goto l79
			l97:
				position, thunkPosition = position79, thunkPosition79
				{
					position98, thunkPosition98 := position, thunkPosition
					if !matchChar('\\') {
						goto l98
					}
					goto l78
				l98:
					position, thunkPosition = position98, thunkPosition98
				}
				begin = position
				if !matchDot() {
					goto l78
				}
				end = position
				do(41)
			}
		l79:
			return true
		l78:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 LEFTARROW <- ('<' '-' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l99
			}
			if !matchChar('-') {
				goto l99
			}
			if !p.rules[25]() {
				goto l99
			}
			return true
		l99:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 SLASH <- ('/' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('/') {
				goto l100
			}
			if !p.rules[25]() {
				goto l100
			}
			return true
		l100:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 17 AND <- ('&' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('&') {
				goto l101
			}
			if !p.rules[25]() {
				goto l101
			}
			return true
		l101:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 NOT <- ('!' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('!') {
				goto l102
			}
			if !p.rules[25]() {
				goto l102
			}
			return true
		l102:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 19 QUESTION <- ('?' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('?') {
				goto l103
			}
			if !p.rules[25]() {
				goto l103
			}
			return true
		l103:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 20 STAR <- ('*' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('*') {
				goto l104
			}
			if !p.rules[25]() {
				goto l104
			}
			return true
		l104:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 21 PLUS <- ('+' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('+') {
				goto l105
			}
			if !p.rules[25]() {
				goto l105
			}
			return true
		l105:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 22 OPEN <- ('(' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('(') {
				goto l106
			}
			if !p.rules[25]() {
				goto l106
			}
			return true
		l106:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 23 CLOSE <- (')' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar(')') {
				goto l107
			}
			if !p.rules[25]() {
				goto l107
			}
			return true
		l107:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 24 DOT <- ('.' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('.') {
				goto l108
			}
			if !p.rules[25]() {
				goto l108
			}
			return true
		l108:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 25 Spacing <- (Space / Comment)* */
		func() bool {
		l110:
			{
				position111, thunkPosition111 := position, thunkPosition
				{
					position112, thunkPosition112 := position, thunkPosition
					if !p.rules[27]() {
						goto l113
					}
					goto l112
				l113:
					position, thunkPosition = position112, thunkPosition112
					if !p.rules[26]() {
						goto l111
					}
				}
			l112:
				goto l110
			l111:
				position, thunkPosition = position111, thunkPosition111
			}
			return true
		},
		/* 26 Comment <- ('#' (!EndOfLine .)* EndOfLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('#') {
				goto l114
			}
		l115:
			{
				position116, thunkPosition116 := position, thunkPosition
				{
					position117, thunkPosition117 := position, thunkPosition
					if !p.rules[28]() {
						goto l117
					}
					goto l116
				l117:
					position, thunkPosition = position117, thunkPosition117
				}
				if !matchDot() {
					goto l116
				}
				goto l115
			l116:
				position, thunkPosition = position116, thunkPosition116
			}
			if !p.rules[28]() {
				goto l114
			}
			return true
		l114:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 27 Space <- (' ' / '\t' / EndOfLine) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position119, thunkPosition119 := position, thunkPosition
				if !matchChar(' ') {
					goto l120
				}
				goto l119
			l120:
				position, thunkPosition = position119, thunkPosition119
				if !matchChar('\t') {
					goto l121
				}
				goto l119
			l121:
				position, thunkPosition = position119, thunkPosition119
				if !p.rules[28]() {
					goto l118
				}
			}
		l119:
			return true
		l118:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 EndOfLine <- (('\r' '\n') / '\n' / '\r') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position123, thunkPosition123 := position, thunkPosition
				if !matchChar('\r') {
					goto l124
				}
				if !matchChar('\n') {
					goto l124
				}
				goto l123
			l124:
				position, thunkPosition = position123, thunkPosition123
				if !matchChar('\n') {
					goto l125
				}
				goto l123
			l125:
				position, thunkPosition = position123, thunkPosition123
				if !matchChar('\r') {
					goto l122
				}
			}
		l123:
			return true
		l122:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 29 EndOfFile <- !. */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position127, thunkPosition127 := position, thunkPosition
				if !matchDot() {
					goto l127
				}
				goto l126
			l127:
				position, thunkPosition = position127, thunkPosition127
			}
			return true
		l126:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 Action <- ('{' < (!'}' .)* > '}' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('{') {
				goto l128
			}
			begin = position
		l129:
			{
				position130, thunkPosition130 := position, thunkPosition
				{
					position131, thunkPosition131 := position, thunkPosition
					if !matchChar('}') {
						goto l131
					}
					goto l130
				l131:
					position, thunkPosition = position131, thunkPosition131
				}
				if !matchDot() {
					goto l130
				}
				goto l129
			l130:
				position, thunkPosition = position130, thunkPosition130
			}
			end = position
			if !matchChar('}') {
				goto l128
			}
			if !p.rules[25]() {
				goto l128
			}
			return true
		l128:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 BEGIN <- ('<' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l132
			}
			if !p.rules[25]() {
				goto l132
			}
			return true
		l132:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 32 END <- ('>' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('>') {
				goto l133
			}
			if !p.rules[25]() {
				goto l133
			}
			return true
		l133:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
	}
}
