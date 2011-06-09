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
	buf, line, character := new(bytes.Buffer), 1, 0

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
		/* 0 */
		func(buffer string, begin, end int) {
			p.AddAlternate()
		},
		/* 1 */
		func(buffer string, begin, end int) {
			p.AddPackage(buffer[begin:end])
		},
		/* 2 */
		func(buffer string, begin, end int) {
			p.AddPeg(buffer[begin:end])
		},
		/* 3 */
		func(buffer string, begin, end int) {
			p.AddState(buffer[begin:end])
		},
		/* 4 */
		func(buffer string, begin, end int) {
			p.AddPredicate(buffer[begin:end])
		},
		/* 5 */
		func(buffer string, begin, end int) {
			p.AddPeekFor()
		},
		/* 6 */
		func(buffer string, begin, end int) {
			p.AddPeekNot()
		},
		/* 7 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\a")
		},
		/* 8 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\b")
		},
		/* 9 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\x1B")
		},
		/* 10 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\f")
		},
		/* 11 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\n")
		},
		/* 12 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\r")
		},
		/* 13 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\t")
		},
		/* 14 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\v")
		},
		/* 15 */
		func(buffer string, begin, end int) {
			p.AddCharacter("'")
		},
		/* 16 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\"")
		},
		/* 17 */
		func(buffer string, begin, end int) {
			p.AddCharacter("[")
		},
		/* 18 */
		func(buffer string, begin, end int) {
			p.AddCharacter("]")
		},
		/* 19 */
		func(buffer string, begin, end int) {
			p.AddCharacter("-")
		},
		/* 20 */
		func(buffer string, begin, end int) {
			p.AddOctalCharacter(buffer[begin:end])
		},
		/* 21 */
		func(buffer string, begin, end int) {
			p.AddOctalCharacter(buffer[begin:end])
		},
		/* 22 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\\")
		},
		/* 23 */
		func(buffer string, begin, end int) {
			p.AddCharacter(buffer[begin:end])
		},
		/* 24 */
		func(buffer string, begin, end int) {
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		},
		/* 25 */
		func(buffer string, begin, end int) {
			p.AddSequence()
		},
		/* 26 */
		func(buffer string, begin, end int) {
			p.AddSequence()
		},
		/* 27 */
		func(buffer string, begin, end int) {
			p.AddAlternate()
		},
		/* 28 */
		func(buffer string, begin, end int) {
			p.AddNil()
			p.AddAlternate()
		},
		/* 29 */
		func(buffer string, begin, end int) {
			p.AddNil()
		},
		/* 30 */
		func(buffer string, begin, end int) {
			p.AddQuery()
		},
		/* 31 */
		func(buffer string, begin, end int) {
			p.AddStar()
		},
		/* 32 */
		func(buffer string, begin, end int) {
			p.AddPlus()
		},
		/* 33 */
		func(buffer string, begin, end int) {
			p.AddRule(buffer[begin:end])
		},
		/* 34 */
		func(buffer string, begin, end int) {
			p.AddExpression()
		},
		/* 35 */
		func(buffer string, begin, end int) {
			p.AddRange()
		},
		/* 36 */
		func(buffer string, begin, end int) {
			p.AddCommit()
		},
		/* 37 */
		func(buffer string, begin, end int) {
			p.AddName(buffer[begin:end])
		},
		/* 38 */
		func(buffer string, begin, end int) {
			p.AddDot()
		},
		/* 39 */
		func(buffer string, begin, end int) {
			p.AddAction(buffer[begin:end])
		},
		/* 40 */
		func(buffer string, begin, end int) {
			p.AddBegin()
		},
		/* 41 */
		func(buffer string, begin, end int) {
			p.AddEnd()
		},
		/* 42 */
		func(buffer string, begin, end int) {
			p.AddSequence()
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
			do(1)
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
			do(2)
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
			do(3)
			if !(commit(thunkPosition0)) {
				goto l0
			}
			if !p.rules[7]() {
				goto l0
			}
			do(33)
			if !p.rules[15]() {
				goto l0
			}
			if !p.rules[2]() {
				goto l0
			}
			do(34)
			{
				position3, thunkPosition3 := position, thunkPosition
				{
					position4, thunkPosition4 := position, thunkPosition
					if !p.rules[7]() {
						goto l5
					}
					if !p.rules[15]() {
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
				do(33)
				if !p.rules[15]() {
					goto l2
				}
				if !p.rules[2]() {
					goto l2
				}
				do(34)
				{
					position7, thunkPosition7 := position, thunkPosition
					{
						position8, thunkPosition8 := position, thunkPosition
						if !p.rules[7]() {
							goto l9
						}
						if !p.rules[15]() {
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
		/* 2 Expression <- ((Sequence (SLASH Sequence { p.AddAlternate() })* (SLASH { p.AddNil(); p.AddAlternate() })?) / { p.AddNil() }) */
		func() bool {
			{
				position14, thunkPosition14 := position, thunkPosition
				if !p.rules[3]() {
					goto l15
				}
			l16:
				{
					position17, thunkPosition17 := position, thunkPosition
					if !p.rules[16]() {
						goto l17
					}
					if !p.rules[3]() {
						goto l17
					}
					do(27)
					goto l16
				l17:
					position, thunkPosition = position17, thunkPosition17
				}
				{
					position18, thunkPosition18 := position, thunkPosition
					if !p.rules[16]() {
						goto l18
					}
					do(28)
					goto l19
				l18:
					position, thunkPosition = position18, thunkPosition18
				}
			l19:
				goto l14
			l15:
				position, thunkPosition = position14, thunkPosition14
				do(29)
			}
		l14:
			return true
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
				do(42)
				goto l21
			l22:
				position, thunkPosition = position22, thunkPosition22
			}
			return true
		l20:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 4 Prefix <- ((AND Action { p.AddPredicate(buffer[begin:end]) }) / ((&('!') (NOT Suffix { p.AddPeekNot() })) | (&('&') (AND Suffix { p.AddPeekFor() })) | (&('"' | '\'' | '(' | '.' | '<' | '>' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{') Suffix))) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position24, thunkPosition24 := position, thunkPosition
				if !p.rules[17]() {
					goto l25
				}
				if !p.rules[30]() {
					goto l25
				}
				do(4)
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
						if !p.rules[25]() {
							goto l23
						}
						if !p.rules[5]() {
							goto l23
						}
						do(6)
						break
					case '&':
						if !p.rules[17]() {
							goto l23
						}
						if !p.rules[5]() {
							goto l23
						}
						do(5)
						break
					default:
						if !p.rules[5]() {
							goto l23
						}
						break
					}
				}

			}
		l24:
			return true
		l23:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 5 Suffix <- (Primary ((&('+') (PLUS { p.AddPlus() })) | (&('*') (STAR { p.AddStar() })) | (&('?') (QUESTION { p.AddQuery() })))?) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position28, thunkPosition28 := position, thunkPosition
				if !matchChar('c') {
					goto l29
				}
				if !matchChar('o') {
					goto l29
				}
				if !matchChar('m') {
					goto l29
				}
				if !matchChar('m') {
					goto l29
				}
				if !matchChar('i') {
					goto l29
				}
				if !matchChar('t') {
					goto l29
				}
				if !p.rules[25]() {
					goto l29
				}
				do(36)
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
						if !p.rules[25]() {
							goto l27
						}
						do(41)
						break
					case '<':
						if !matchChar('<') {
							goto l27
						}
						if !p.rules[25]() {
							goto l27
						}
						do(40)
						break
					case '{':
						if !p.rules[30]() {
							goto l27
						}
						do(39)
						break
					case '.':
						if !matchChar('.') {
							goto l27
						}
						if !p.rules[25]() {
							goto l27
						}
						do(38)
						break
					case '[':
						if !matchChar('[') {
							goto l27
						}
						{
							position31, thunkPosition31 := position, thunkPosition
							{
								position33, thunkPosition33 := position, thunkPosition
								if !matchChar('^') {
									goto l34
								}
								if !p.rules[12]() {
									goto l34
								}
								do(24)
								goto l33
							l34:
								position, thunkPosition = position33, thunkPosition33
								if !p.rules[12]() {
									goto l31
								}
							}
						l33:
							goto l32
						l31:
							position, thunkPosition = position31, thunkPosition31
						}
					l32:
						if !matchChar(']') {
							goto l27
						}
						if !p.rules[25]() {
							goto l27
						}
						break
					case '"', '\'':
						{
							position35, thunkPosition35 := position, thunkPosition
							if !matchChar('\'') {
								goto l36
							}
							{
								position37, thunkPosition37 := position, thunkPosition
								{
									position39, thunkPosition39 := position, thunkPosition
									if !matchChar('\'') {
										goto l39
									}
									goto l37
								l39:
									position, thunkPosition = position39, thunkPosition39
								}
								if !p.rules[14]() {
									goto l37
								}
								goto l38
							l37:
								position, thunkPosition = position37, thunkPosition37
							}
						l38:
						l40:
							{
								position41, thunkPosition41 := position, thunkPosition
								{
									position42, thunkPosition42 := position, thunkPosition
									if !matchChar('\'') {
										goto l42
									}
									goto l41
								l42:
									position, thunkPosition = position42, thunkPosition42
								}
								if !p.rules[14]() {
									goto l41
								}
								do(25)
								goto l40
							l41:
								position, thunkPosition = position41, thunkPosition41
							}
							if !matchChar('\'') {
								goto l36
							}
							if !p.rules[25]() {
								goto l36
							}
							goto l35
						l36:
							position, thunkPosition = position35, thunkPosition35
							if !matchChar('"') {
								goto l27
							}
							{
								position43, thunkPosition43 := position, thunkPosition
								{
									position45, thunkPosition45 := position, thunkPosition
									if !matchChar('"') {
										goto l45
									}
									goto l43
								l45:
									position, thunkPosition = position45, thunkPosition45
								}
								if !p.rules[14]() {
									goto l43
								}
								goto l44
							l43:
								position, thunkPosition = position43, thunkPosition43
							}
						l44:
						l46:
							{
								position47, thunkPosition47 := position, thunkPosition
								{
									position48, thunkPosition48 := position, thunkPosition
									if !matchChar('"') {
										goto l48
									}
									goto l47
								l48:
									position, thunkPosition = position48, thunkPosition48
								}
								if !p.rules[14]() {
									goto l47
								}
								do(26)
								goto l46
							l47:
								position, thunkPosition = position47, thunkPosition47
							}
							if !matchChar('"') {
								goto l27
							}
							if !p.rules[25]() {
								goto l27
							}
						}
					l35:
						break
					case '(':
						if !matchChar('(') {
							goto l27
						}
						if !p.rules[25]() {
							goto l27
						}
						if !p.rules[2]() {
							goto l27
						}
						if !matchChar(')') {
							goto l27
						}
						if !p.rules[25]() {
							goto l27
						}
						break
					default:
						if !p.rules[7]() {
							goto l27
						}
						{
							position49, thunkPosition49 := position, thunkPosition
							if !p.rules[15]() {
								goto l49
							}
							goto l27
						l49:
							position, thunkPosition = position49, thunkPosition49
						}
						do(37)
						break
					}
				}

			}
		l28:
			{
				position50, thunkPosition50 := position, thunkPosition
				{
					if position == len(p.Buffer) {
						goto l50
					}
					switch p.Buffer[position] {
					case '+':
						if !matchChar('+') {
							goto l50
						}
						if !p.rules[25]() {
							goto l50
						}
						do(32)
						break
					case '*':
						if !matchChar('*') {
							goto l50
						}
						if !p.rules[25]() {
							goto l50
						}
						do(31)
						break
					default:
						if !matchChar('?') {
							goto l50
						}
						if !p.rules[25]() {
							goto l50
						}
						do(30)
						break
					}
				}

				goto l51
			l50:
				position, thunkPosition = position50, thunkPosition50
			}
		l51:
			return true
		l27:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 6 Primary <- (('c' 'o' 'm' 'm' 'i' 't' Spacing { p.AddCommit() }) / ((&('>') (END { p.AddEnd() })) | (&('<') (BEGIN { p.AddBegin() })) | (&('{') (Action { p.AddAction(buffer[begin:end]) })) | (&('.') (DOT { p.AddDot() })) | (&('[') Class) | (&('"' | '\'') Literal) | (&('(') (OPEN Expression CLOSE)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (Identifier !LEFTARROW { p.AddName(buffer[begin:end]) })))) */
		nil,
		/* 7 Identifier <- (< IdentStart IdentCont* > Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			begin = position
			if !p.rules[8]() {
				goto l54
			}
		l55:
			{
				position56, thunkPosition56 := position, thunkPosition
				{
					position57, thunkPosition57 := position, thunkPosition
					if !p.rules[8]() {
						goto l58
					}
					goto l57
				l58:
					position, thunkPosition = position57, thunkPosition57
					if !matchRange('0', '9') {
						goto l56
					}
				}
			l57:
				goto l55
			l56:
				position, thunkPosition = position56, thunkPosition56
			}
			end = position
			if !p.rules[25]() {
				goto l54
			}
			return true
		l54:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 8 IdentStart <- ((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				if position == len(p.Buffer) {
					goto l59
				}
				switch p.Buffer[position] {
				case '_':
					if !matchChar('_') {
						goto l59
					}
					break
				case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
					if !matchRange('A', 'Z') {
						goto l59
					}
					break
				default:
					if !matchRange('a', 'z') {
						goto l59
					}
					break
				}
			}

			return true
		l59:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 9 IdentCont <- (IdentStart / [0-9]) */
		nil,
		/* 10 Literal <- (('\'' (!'\'' Char)? (!'\'' Char { p.AddSequence() })* '\'' Spacing) / ('"' (!'"' Char)? (!'"' Char { p.AddSequence() })* '"' Spacing)) */
		nil,
		/* 11 Class <- ('[' (('^' Ranges { p.AddPeekNot(); p.AddDot(); p.AddSequence() }) / Ranges)? ']' Spacing) */
		nil,
		/* 12 Ranges <- (!']' Range (!']' Range { p.AddAlternate() })*) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position65, thunkPosition65 := position, thunkPosition
				if !matchChar(']') {
					goto l65
				}
				goto l64
			l65:
				position, thunkPosition = position65, thunkPosition65
			}
			if !p.rules[13]() {
				goto l64
			}
		l66:
			{
				position67, thunkPosition67 := position, thunkPosition
				{
					position68, thunkPosition68 := position, thunkPosition
					if !matchChar(']') {
						goto l68
					}
					goto l67
				l68:
					position, thunkPosition = position68, thunkPosition68
				}
				if !p.rules[13]() {
					goto l67
				}
				do(0)
				goto l66
			l67:
				position, thunkPosition = position67, thunkPosition67
			}
			return true
		l64:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 13 Range <- ((Char '-' Char { p.AddRange() }) / Char) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position70, thunkPosition70 := position, thunkPosition
				if !p.rules[14]() {
					goto l71
				}
				if !matchChar('-') {
					goto l71
				}
				if !p.rules[14]() {
					goto l71
				}
				do(35)
				goto l70
			l71:
				position, thunkPosition = position70, thunkPosition70
				if !p.rules[14]() {
					goto l69
				}
			}
		l70:
			return true
		l69:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 14 Char <- (('\\' 'a' { p.AddCharacter("\a") }) / ('\\' 'b' { p.AddCharacter("\b") }) / ('\\' 'e' { p.AddCharacter("\x1B") }) / ('\\' 'f' { p.AddCharacter("\f") }) / ('\\' 'n' { p.AddCharacter("\n") }) / ('\\' 'r' { p.AddCharacter("\r") }) / ('\\' 't' { p.AddCharacter("\t") }) / ('\\' 'v' { p.AddCharacter("\v") }) / ('\\' '\'' { p.AddCharacter("'") }) / ('\\' '"' { p.AddCharacter("\"") }) / ('\\' '[' { p.AddCharacter("[") }) / ('\\' ']' { p.AddCharacter("]") }) / ('\\' '-' { p.AddCharacter("-") }) / ('\\' < [0-3] [0-7] [0-7] > { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' < [0-7] [0-7]? > { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' '\\' { p.AddCharacter("\\") }) / (!'\\' < . > { p.AddCharacter(buffer[begin:end]) })) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position73, thunkPosition73 := position, thunkPosition
				if !matchChar('\\') {
					goto l74
				}
				if !matchChar('a') {
					goto l74
				}
				do(7)
				goto l73
			l74:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l75
				}
				if !matchChar('b') {
					goto l75
				}
				do(8)
				goto l73
			l75:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l76
				}
				if !matchChar('e') {
					goto l76
				}
				do(9)
				goto l73
			l76:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l77
				}
				if !matchChar('f') {
					goto l77
				}
				do(10)
				goto l73
			l77:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l78
				}
				if !matchChar('n') {
					goto l78
				}
				do(11)
				goto l73
			l78:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l79
				}
				if !matchChar('r') {
					goto l79
				}
				do(12)
				goto l73
			l79:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l80
				}
				if !matchChar('t') {
					goto l80
				}
				do(13)
				goto l73
			l80:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l81
				}
				if !matchChar('v') {
					goto l81
				}
				do(14)
				goto l73
			l81:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l82
				}
				if !matchChar('\'') {
					goto l82
				}
				do(15)
				goto l73
			l82:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l83
				}
				if !matchChar('"') {
					goto l83
				}
				do(16)
				goto l73
			l83:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l84
				}
				if !matchChar('[') {
					goto l84
				}
				do(17)
				goto l73
			l84:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l85
				}
				if !matchChar(']') {
					goto l85
				}
				do(18)
				goto l73
			l85:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l86
				}
				if !matchChar('-') {
					goto l86
				}
				do(19)
				goto l73
			l86:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l87
				}
				begin = position
				if !matchRange('0', '3') {
					goto l87
				}
				if !matchRange('0', '7') {
					goto l87
				}
				if !matchRange('0', '7') {
					goto l87
				}
				end = position
				do(20)
				goto l73
			l87:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l88
				}
				begin = position
				if !matchRange('0', '7') {
					goto l88
				}
				{
					position89, thunkPosition89 := position, thunkPosition
					if !matchRange('0', '7') {
						goto l89
					}
					goto l90
				l89:
					position, thunkPosition = position89, thunkPosition89
				}
			l90:
				end = position
				do(21)
				goto l73
			l88:
				position, thunkPosition = position73, thunkPosition73
				if !matchChar('\\') {
					goto l91
				}
				if !matchChar('\\') {
					goto l91
				}
				do(22)
				goto l73
			l91:
				position, thunkPosition = position73, thunkPosition73
				{
					position92, thunkPosition92 := position, thunkPosition
					if !matchChar('\\') {
						goto l92
					}
					goto l72
				l92:
					position, thunkPosition = position92, thunkPosition92
				}
				begin = position
				if !matchDot() {
					goto l72
				}
				end = position
				do(23)
			}
		l73:
			return true
		l72:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 15 LEFTARROW <- ('<' '-' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('<') {
				goto l93
			}
			if !matchChar('-') {
				goto l93
			}
			if !p.rules[25]() {
				goto l93
			}
			return true
		l93:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 16 SLASH <- ('/' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('/') {
				goto l94
			}
			if !p.rules[25]() {
				goto l94
			}
			return true
		l94:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 17 AND <- ('&' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('&') {
				goto l95
			}
			if !p.rules[25]() {
				goto l95
			}
			return true
		l95:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 18 NOT <- ('!' Spacing) */
		nil,
		/* 19 QUESTION <- ('?' Spacing) */
		nil,
		/* 20 STAR <- ('*' Spacing) */
		nil,
		/* 21 PLUS <- ('+' Spacing) */
		nil,
		/* 22 OPEN <- ('(' Spacing) */
		nil,
		/* 23 CLOSE <- (')' Spacing) */
		nil,
		/* 24 DOT <- ('.' Spacing) */
		nil,
		/* 25 Spacing <- (Space / Comment)* */
		func() bool {
		l104:
			{
				position105, thunkPosition105 := position, thunkPosition
				{
					position106, thunkPosition106 := position, thunkPosition
					{
						if position == len(p.Buffer) {
							goto l107
						}
						switch p.Buffer[position] {
						case '\t':
							if !matchChar('\t') {
								goto l107
							}
							break
						case ' ':
							if !matchChar(' ') {
								goto l107
							}
							break
						default:
							if !p.rules[28]() {
								goto l107
							}
							break
						}
					}

					goto l106
				l107:
					position, thunkPosition = position106, thunkPosition106
					if !matchChar('#') {
						goto l105
					}
				l109:
					{
						position110, thunkPosition110 := position, thunkPosition
						{
							position111, thunkPosition111 := position, thunkPosition
							if !p.rules[28]() {
								goto l111
							}
							goto l110
						l111:
							position, thunkPosition = position111, thunkPosition111
						}
						if !matchDot() {
							goto l110
						}
						goto l109
					l110:
						position, thunkPosition = position110, thunkPosition110
					}
					if !p.rules[28]() {
						goto l105
					}
				}
			l106:
				goto l104
			l105:
				position, thunkPosition = position105, thunkPosition105
			}
			return true
		},
		/* 26 Comment <- ('#' (!EndOfLine .)* EndOfLine) */
		nil,
		/* 27 Space <- ((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine)) */
		nil,
		/* 28 EndOfLine <- (('\r' '\n') / '\n' / '\r') */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			{
				position115, thunkPosition115 := position, thunkPosition
				if !matchChar('\r') {
					goto l116
				}
				if !matchChar('\n') {
					goto l116
				}
				goto l115
			l116:
				position, thunkPosition = position115, thunkPosition115
				if !matchChar('\n') {
					goto l117
				}
				goto l115
			l117:
				position, thunkPosition = position115, thunkPosition115
				if !matchChar('\r') {
					goto l114
				}
			}
		l115:
			return true
		l114:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 29 EndOfFile <- !. */
		nil,
		/* 30 Action <- ('{' < (!'}' .)* > '}' Spacing) */
		func() bool {
			position0, thunkPosition0 := position, thunkPosition
			if !matchChar('{') {
				goto l119
			}
			begin = position
		l120:
			{
				position121, thunkPosition121 := position, thunkPosition
				{
					position122, thunkPosition122 := position, thunkPosition
					if !matchChar('}') {
						goto l122
					}
					goto l121
				l122:
					position, thunkPosition = position122, thunkPosition122
				}
				if !matchDot() {
					goto l121
				}
				goto l120
			l121:
				position, thunkPosition = position121, thunkPosition121
			}
			end = position
			if !matchChar('}') {
				goto l119
			}
			if !p.rules[25]() {
				goto l119
			}
			return true
		l119:
			position, thunkPosition = position0, thunkPosition0
			return false
		},
		/* 31 BEGIN <- ('<' Spacing) */
		nil,
		/* 32 END <- ('>' Spacing) */
		nil,
	}
}
