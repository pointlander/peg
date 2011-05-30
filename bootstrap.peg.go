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
	rules    [32]func() bool
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
			p.AddString(buffer[begin:end])
		},
		/* 17 Primary */
		func(buffer string, begin, end int) {
			p.AddClass(buffer[begin:end])
		},
		/* 18 Primary */
		func(buffer string, begin, end int) {
			p.AddDot()
		},
		/* 19 Primary */
		func(buffer string, begin, end int) {
			p.AddAction(buffer[begin:end])
		},
		/* 20 Primary */
		func(buffer string, begin, end int) {
			p.AddBegin()
		},
		/* 21 Primary */
		func(buffer string, begin, end int) {
			p.AddEnd()
		},
	}
	var thunkPosition, begin, end int
	thunks := make([]struct {
		action     uint8
		begin, end int
	},32)
	do := func(action uint8) {
		if thunkPosition == len(thunks) {
			newThunks := make([]struct {
				action     uint8
				begin, end int
			},2*len(thunks))
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
	matchString := func(s string) bool {
		length := len(s)
		next := position + length
		if (next <= len(p.Buffer)) && (p.Buffer[position:next] == s) {
			position = next
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	classes := [...][32]uint8{
		[32]uint8{0, 0, 0, 0, 132, 0, 0, 0, 0, 0, 0, 56, 102, 64, 84, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[32]uint8{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 223, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		[32]uint8{0, 0, 0, 0, 0, 0, 255, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[32]uint8{0, 0, 0, 0, 0, 0, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[32]uint8{0, 0, 0, 0, 0, 0, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[32]uint8{0, 0, 0, 0, 0, 0, 0, 0, 254, 255, 255, 135, 254, 255, 255, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[32]uint8{0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[32]uint8{0, 0, 0, 0, 128, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	matchClass := func(class uint) bool {
		if (position < len(p.Buffer)) &&
			((classes[class][p.Buffer[position]>>3] & (1 << (p.Buffer[position] & 7))) != 0) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	p.rules = [...]func() bool{
		/* 0 Grammar <- (Spacing 'package' Spacing Identifier { p.AddPackage(buffer[begin:end]) } 'type' Spacing Identifier { p.AddPeg(buffer[begin:end]) } 'Peg' Spacing Action { p.AddState(buffer[begin:end]) } commit Definition+ EndOfFile) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[24]() {
				goto l0
			}
			if !matchString("package") {
				goto l0
			}
			if !p.rules[24]() {
				goto l0
			}
			if !p.rules[7]() {
				goto l0
			}
			do(0)
			if !matchString("type") {
				goto l0
			}
			if !p.rules[24]() {
				goto l0
			}
			if !p.rules[7]() {
				goto l0
			}
			do(1)
			if !matchString("Peg") {
				goto l0
			}
			if !p.rules[24]() {
				goto l0
			}
			if !p.rules[29]() {
				goto l0
			}
			do(2)
			if !(commit(thunkPosition0)) {
				goto l0
			}
			if !p.rules[7]() {
				goto l0
			}
			do(3)
			if !p.rules[14]() {
				goto l0
			}
			if !p.rules[2]() {
				goto l0
			}
			do(4)
			{
				position3, thunkPosition3 := position, thunkPosition
				{
					position4, thunkPosition4 := position, thunkPosition
					if !p.rules[7]() {
						goto l5
					}
					if !p.rules[14]() {
						goto l5
					}
					goto l4
				l5:
					position, thunkPosition = position4, thunkPosition4
					{
						position6, thunkPosition6 := position, thunkPosition
						if !matchDot() {
							goto l6
						}
						goto l0
					l6:
						position, thunkPosition = position6, thunkPosition6
					}
				}
			l4:
				position, thunkPosition = position3, thunkPosition3
			}
			if !(commit(thunkPosition0)) {
				goto l0
			}
		l1:
			{
				position2, thunkPosition2 := position, thunkPosition
				if !p.rules[7]() {
					goto l2
				}
				do(3)
				if !p.rules[14]() {
					goto l2
				}
				if !p.rules[2]() {
					goto l2
				}
				do(4)
				{
					position7, thunkPosition7 := position, thunkPosition
					{
						position8, thunkPosition8 := position, thunkPosition
						if !p.rules[7]() {
							goto l9
						}
						if !p.rules[14]() {
							goto l9
						}
						goto l8
					l9:
						position, thunkPosition = position8, thunkPosition8
						{
							position10, thunkPosition10 := position, thunkPosition
							if !matchDot() {
								goto l10
							}
							goto l2
						l10:
							position, thunkPosition = position10, thunkPosition10
						}
					}
				l8:
					position, thunkPosition = position7, thunkPosition7
				}
				if !(commit(thunkPosition0)) {
					goto l2
				}
				goto l1
			l2:
				position, thunkPosition = position2, thunkPosition2
			}
			{
				position11, thunkPosition11 := position, thunkPosition
				if !matchDot() {
					goto l11
				}
				goto l0
			l11:
				position, thunkPosition = position11, thunkPosition11
			}
			return true
		l0:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 1 Definition <- (Identifier { p.AddRule(buffer[begin:end]) } LEFTARROW Expression { p.AddExpression() } &((Identifier LEFTARROW) / !.) commit) */
		nil,
		/* 2 Expression <- ((Sequence (SLASH Sequence { p.AddAlternate() })* (SLASH { p.AddEmptyAlternate() })?) /) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position14, thunkPosition14 := position, thunkPosition
				if !p.rules[3]() {
					goto l15
				}
			l16:
				{
					position17, thunkPosition17 := position, thunkPosition
					if !p.rules[15]() {
						goto l17
					}
					if !p.rules[3]() {
						goto l17
					}
					do(5)
					goto l16
				l17:
					position, thunkPosition = position17, thunkPosition17
				}
				{
					position18, thunkPosition18 := position, thunkPosition
					if !p.rules[15]() {
						goto l18
					}
					do(6)
					goto l19
				l18:
					position, thunkPosition = position18, thunkPosition18
				}
			l19:
				goto l14
			l15:
				position, thunkPosition = position14, thunkPosition14
			}
		l14:
			return true
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 3 Sequence <- (Prefix (Prefix { p.AddSequence() })*) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !p.rules[4]() {
				goto l20
			}
		l21:
			{
				position22, thunkPosition22 := position, thunkPosition
				if !p.rules[4]() {
					goto l22
				}
				do(7)
				goto l21
			l22:
				position, thunkPosition = position22, thunkPosition22
			}
			return true
		l20:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 4 Prefix <- ((AND Action { p.AddPredicate(buffer[begin:end]) }) / ((&[!] (NOT Suffix { p.AddPeekNot() })) | (&[&] (AND Suffix { p.AddPeekFor() })) | (&[\"\'(.<>A-\[_a-{] Suffix))) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position24, thunkPosition24 := position, thunkPosition
				if !p.rules[16]() {
					goto l25
				}
				if !p.rules[29]() {
					goto l25
				}
				do(8)
				goto l24
			l25:
				position, thunkPosition = position24, thunkPosition24
				{
					if position == len(p.Buffer) {
						goto l23
					}
					switch p.Buffer[position] {
					case '!':
						if !matchChar('!') {
							goto l23
						}
						if !p.rules[24]() {
							goto l23
						}
						if !p.rules[5]() {
							goto l23
						}
						do(10)
					case '&':
						if !p.rules[16]() {
							goto l23
						}
						if !p.rules[5]() {
							goto l23
						}
						do(9)
					default:
						if !p.rules[5]() {
							goto l23
						}
					}
				}

			}
		l24:
			return true
		l23:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 5 Suffix <- (Primary ((&[+] (PLUS { p.AddPlus() })) | (&[*] (STAR { p.AddStar() })) | (&[?] (QUESTION { p.AddQuery() })))?) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position28, thunkPosition28 := position, thunkPosition
				if !matchString("commit") {
					goto l29
				}
				if !p.rules[24]() {
					goto l29
				}
				do(14)
				goto l28
			l29:
				position, thunkPosition = position28, thunkPosition28
				{
					if position == len(p.Buffer) {
						goto l27
					}
					switch p.Buffer[position] {
					case '>':
						if !matchChar('>') {
							goto l27
						}
						if !p.rules[24]() {
							goto l27
						}
						do(21)
					case '<':
						if !matchChar('<') {
							goto l27
						}
						if !p.rules[24]() {
							goto l27
						}
						do(20)
					case '{':
						if !p.rules[29]() {
							goto l27
						}
						do(19)
					case '.':
						if !matchChar('.') {
							goto l27
						}
						if !p.rules[24]() {
							goto l27
						}
						do(18)
					case '[':
						if !matchChar('[') {
							goto l27
						}
						begin = position
					l31:
						{
							position32, thunkPosition32 := position, thunkPosition
							{
								position33, thunkPosition33 := position, thunkPosition
								if !matchChar(']') {
									goto l33
								}
								goto l32
							l33:
								position, thunkPosition = position33, thunkPosition33
							}
							{
								position34, thunkPosition34 := position, thunkPosition
								if !p.rules[13]() {
									goto l35
								}
								if !matchChar('-') {
									goto l35
								}
								if !p.rules[13]() {
									goto l35
								}
								goto l34
							l35:
								position, thunkPosition = position34, thunkPosition34
								if !p.rules[13]() {
									goto l32
								}
							}
						l34:
							goto l31
						l32:
							position, thunkPosition = position32, thunkPosition32
						}
						end = position
						if !matchChar(']') {
							goto l27
						}
						if !p.rules[24]() {
							goto l27
						}
						do(17)
					case '"', '\'':
						{
							position36, thunkPosition36 := position, thunkPosition
							if !matchClass(7) {
								goto l37
							}
							begin = position
						l38:
							{
								position39, thunkPosition39 := position, thunkPosition
								{
									position40, thunkPosition40 := position, thunkPosition
									if !matchClass(7) {
										goto l40
									}
									goto l39
								l40:
									position, thunkPosition = position40, thunkPosition40
								}
								if !p.rules[13]() {
									goto l39
								}
								goto l38
							l39:
								position, thunkPosition = position39, thunkPosition39
							}
							end = position
							if !matchClass(7) {
								goto l37
							}
							if !p.rules[24]() {
								goto l37
							}
							goto l36
						l37:
							position, thunkPosition = position36, thunkPosition36
							if !matchClass(6) {
								goto l27
							}
							begin = position
						l41:
							{
								position42, thunkPosition42 := position, thunkPosition
								{
									position43, thunkPosition43 := position, thunkPosition
									if !matchClass(6) {
										goto l43
									}
									goto l42
								l43:
									position, thunkPosition = position43, thunkPosition43
								}
								if !p.rules[13]() {
									goto l42
								}
								goto l41
							l42:
								position, thunkPosition = position42, thunkPosition42
							}
							end = position
							if !matchClass(6) {
								goto l27
							}
							if !p.rules[24]() {
								goto l27
							}
						}
					l36:
						do(16)
					case '(':
						if !matchChar('(') {
							goto l27
						}
						if !p.rules[24]() {
							goto l27
						}
						if !p.rules[2]() {
							goto l27
						}
						if !matchChar(')') {
							goto l27
						}
						if !p.rules[24]() {
							goto l27
						}
					default:
						if !p.rules[7]() {
							goto l27
						}
						{
							position44, thunkPosition44 := position, thunkPosition
							if !p.rules[14]() {
								goto l44
							}
							goto l27
						l44:
							position, thunkPosition = position44, thunkPosition44
						}
						do(15)
					}
				}

			}
		l28:
			{
				position45, thunkPosition45 := position, thunkPosition
				{
					if position == len(p.Buffer) {
						goto l45
					}
					switch p.Buffer[position] {
					case '+':
						if !matchChar('+') {
							goto l45
						}
						if !p.rules[24]() {
							goto l45
						}
						do(13)
					case '*':
						if !matchChar('*') {
							goto l45
						}
						if !p.rules[24]() {
							goto l45
						}
						do(12)
					default:
						if !matchChar('?') {
							goto l45
						}
						if !p.rules[24]() {
							goto l45
						}
						do(11)
					}
				}

				goto l46
			l45:
				position, thunkPosition = position45, thunkPosition45
			}
		l46:
			return true
		l27:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 6 Primary <- (('commit' Spacing { p.AddCommit() }) / ((&[>] (END { p.AddEnd() })) | (&[<] (BEGIN { p.AddBegin() })) | (&[{] (Action { p.AddAction(buffer[begin:end]) })) | (&[.] (DOT { p.AddDot() })) | (&[\[] (Class { p.AddClass(buffer[begin:end]) })) | (&[\"\'] (Literal { p.AddString(buffer[begin:end]) })) | (&[(] (OPEN Expression CLOSE)) | (&[A-Z_a-z] (Identifier !LEFTARROW { p.AddName(buffer[begin:end]) })))) */
		nil,
		/* 7 Identifier <- (< IdentStart IdentCont* > Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[8]() {
				goto l49
			}
		l50:
			{
				position51, thunkPosition51 := position, thunkPosition
				{
					position52, thunkPosition52 := position, thunkPosition
					if !p.rules[8]() {
						goto l53
					}
					goto l52
				l53:
					position, thunkPosition = position52, thunkPosition52
					if !matchClass(2) {
						goto l51
					}
				}
			l52:
				goto l50
			l51:
				position, thunkPosition = position51, thunkPosition51
			}
			end = position
			if !p.rules[24]() {
				goto l49
			}
			return true
		l49:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 IdentStart <- [a-zA-Z_] */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchClass(5) {
				goto l54
			}
			return true
		l54:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 9 IdentCont <- (IdentStart / [0-9]) */
		nil,
		/* 10 Literal <- ((['] < (!['] Char)* > ['] Spacing) / (["] < (!["] Char)* > ["] Spacing)) */
		nil,
		/* 11 Class <- ('[' < (!']' Range)* > ']' Spacing) */
		nil,
		/* 12 Range <- ((Char '-' Char) / Char) */
		nil,
		/* 13 Char <- (('\\' [abefnrtv'"\[\]\\]) / ('\\' [0-3] [0-7] [0-7]) / ('\\' [0-7] [0-7]?) / ('\\' '-') / (!'\\' .)) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position60, thunkPosition60 := position, thunkPosition
				if !matchChar('\\') {
					goto l61
				}
				if !matchClass(0) {
					goto l61
				}
				goto l60
			l61:
				position, thunkPosition = position60, thunkPosition60
				if !matchChar('\\') {
					goto l62
				}
				if !matchClass(4) {
					goto l62
				}
				if !matchClass(3) {
					goto l62
				}
				if !matchClass(3) {
					goto l62
				}
				goto l60
			l62:
				position, thunkPosition = position60, thunkPosition60
				if !matchChar('\\') {
					goto l63
				}
				if !matchClass(3) {
					goto l63
				}
				{
					position64, thunkPosition64 := position, thunkPosition
					if !matchClass(3) {
						goto l64
					}
					goto l65
				l64:
					position, thunkPosition = position64, thunkPosition64
				}
			l65:
				goto l60
			l63:
				position, thunkPosition = position60, thunkPosition60
				if !matchChar('\\') {
					goto l66
				}
				if !matchChar('-') {
					goto l66
				}
				goto l60
			l66:
				position, thunkPosition = position60, thunkPosition60
				{
					position67, thunkPosition67 := position, thunkPosition
					if !matchChar('\\') {
						goto l67
					}
					goto l59
				l67:
					position, thunkPosition = position67, thunkPosition67
				}
				if !matchDot() {
					goto l59
				}
			}
		l60:
			return true
		l59:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 14 LEFTARROW <- ('<-' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchString("<-") {
				goto l68
			}
			if !p.rules[24]() {
				goto l68
			}
			return true
		l68:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 SLASH <- ('/' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('/') {
				goto l69
			}
			if !p.rules[24]() {
				goto l69
			}
			return true
		l69:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 AND <- ('&' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('&') {
				goto l70
			}
			if !p.rules[24]() {
				goto l70
			}
			return true
		l70:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 17 NOT <- ('!' Spacing) */
		nil,
		/* 18 QUESTION <- ('?' Spacing) */
		nil,
		/* 19 STAR <- ('*' Spacing) */
		nil,
		/* 20 PLUS <- ('+' Spacing) */
		nil,
		/* 21 OPEN <- ('(' Spacing) */
		nil,
		/* 22 CLOSE <- (')' Spacing) */
		nil,
		/* 23 DOT <- ('.' Spacing) */
		nil,
		/* 24 Spacing <- (Space / Comment)* */
		func() bool {
		l79:
			{
				position80, thunkPosition80 := position, thunkPosition
				{
					position81, thunkPosition81 := position, thunkPosition
					{
						if position == len(p.Buffer) {
							goto l82
						}
						switch p.Buffer[position] {
						case '\t':
							if !matchChar('\t') {
								goto l82
							}
						case ' ':
							if !matchChar(' ') {
								goto l82
							}
						default:
							if !p.rules[27]() {
								goto l82
							}
						}
					}

					goto l81
				l82:
					position, thunkPosition = position81, thunkPosition81
					if !matchChar('#') {
						goto l80
					}
				l84:
					{
						position85, thunkPosition85 := position, thunkPosition
						{
							position86, thunkPosition86 := position, thunkPosition
							if !p.rules[27]() {
								goto l86
							}
							goto l85
						l86:
							position, thunkPosition = position86, thunkPosition86
						}
						if !matchDot() {
							goto l85
						}
						goto l84
					l85:
						position, thunkPosition = position85, thunkPosition85
					}
					if !p.rules[27]() {
						goto l80
					}
				}
			l81:
				goto l79
			l80:
				position, thunkPosition = position80, thunkPosition80
			}
			return true
		},
		/* 25 Comment <- ('#' (!EndOfLine .)* EndOfLine) */
		nil,
		/* 26 Space <- ((&[\t] '\t') | (&[ ] ' ') | (&[\n\r] EndOfLine)) */
		nil,
		/* 27 EndOfLine <- ('\r\n' / '\n' / '\r') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position90, thunkPosition90 := position, thunkPosition
				if !matchString("\r\n") {
					goto l91
				}
				goto l90
			l91:
				position, thunkPosition = position90, thunkPosition90
				if !matchChar('\n') {
					goto l92
				}
				goto l90
			l92:
				position, thunkPosition = position90, thunkPosition90
				if !matchChar('\r') {
					goto l89
				}
			}
		l90:
			return true
		l89:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 28 EndOfFile <- !. */
		nil,
		/* 29 Action <- ('{' < [^}]* > '}' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('{') {
				goto l94
			}
			begin = position
		l95:
			{
				position96, thunkPosition96 := position, thunkPosition
				if !matchClass(1) {
					goto l96
				}
				goto l95
			l96:
				position, thunkPosition = position96, thunkPosition96
			}
			end = position
			if !matchChar('}') {
				goto l94
			}
			if !p.rules[24]() {
				goto l94
			}
			return true
		l94:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 30 BEGIN <- ('<' Spacing) */
		nil,
		/* 31 END <- ('>' Spacing) */
		nil,
	}
}
