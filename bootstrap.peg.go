package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
)

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	RuleGrammar
	RuleDefinition
	RuleExpression
	RuleSequence
	RulePrefix
	RuleSuffix
	RulePrimary
	RuleIdentifier
	RuleIdentStart
	RuleIdentCont
	RuleLiteral
	RuleClass
	RuleRanges
	RuleDoubleRanges
	RuleRange
	RuleDoubleRange
	RuleChar
	RuleDoubleChar
	RuleEscape
	RuleLeftArrow
	RuleSlash
	RuleAnd
	RuleNot
	RuleQuestion
	RuleStar
	RulePlus
	RuleOpen
	RuleClose
	RuleDot
	RuleSpacing
	RuleComment
	RuleSpace
	RuleEndOfLine
	RuleEndOfFile
	RuleAction
	RuleBegin
	RuleEnd
)

var Rul3s = [...]string{
	"Unknown",
	"Grammar",
	"Definition",
	"Expression",
	"Sequence",
	"Prefix",
	"Suffix",
	"Primary",
	"Identifier",
	"IdentStart",
	"IdentCont",
	"Literal",
	"Class",
	"Ranges",
	"DoubleRanges",
	"Range",
	"DoubleRange",
	"Char",
	"DoubleChar",
	"Escape",
	"LeftArrow",
	"Slash",
	"And",
	"Not",
	"Question",
	"Star",
	"Plus",
	"Open",
	"Close",
	"Dot",
	"Spacing",
	"Comment",
	"Space",
	"EndOfLine",
	"EndOfFile",
	"Action",
	"Begin",
	"End",
}

type TokenTree interface {
	sort.Interface
	Prepare()
	Add(rule Rule, begin, end, next int)
	Expand(index int) TokenTree
	Stack() []token32
	Tokens() <-chan token32
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	Rule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

type tokens16 struct {
	tree      []token16
	stackSize int32
}

func (t *tokens16) Len() int {
	return len(t.tree)
}

func (t *tokens16) Less(i, j int) bool {
	ii, jj := t.tree[i], t.tree[j]
	if ii.Rule != RuleUnknown {
		if jj.Rule == RuleUnknown {
			return true
		} else if ii.begin < jj.begin {
			return true
		} else if ii.begin == jj.begin {
			if ii.end == ii.begin || jj.end == jj.begin {
				if ii.next < jj.next {
					return true
				}
			} else if ii.end > jj.end {
				return true
			} else if ii.end == jj.end {
				if ii.next > jj.next {
					return true
				}
			}
		}
	}
	return false
}

func (t *tokens16) Swap(i, j int) {
	t.tree[i], t.tree[j] = t.tree[j], t.tree[i]
}

func (t *tokens16) Prepare() {
	sort.Sort(t)
	size := int(t.tree[0].next) + 1

	tree, stack, top := t.tree[0:size], make([]token16, size), -1
	for i, token := range tree {
		token.next = int16(i)
		for top >= 0 && token.begin >= stack[top].end {
			tree[stack[top].next].next, top = token.next, top-1
		}
		stack[top+1], top = token, top+1
	}

	for i, token := range stack {
		if token.isZero() {
			t.stackSize = int32(i)
			break
		}
	}

	t.tree = tree
}

func (t *tokens16) Add(rule Rule, begin, end, next int) {
	t.tree[next] = token16{Rule: rule, begin: int16(begin), end: int16(end), next: int16(next)}
}

func (t *tokens16) Stack() []token32 {
	if t.stackSize == 0 {
		t.Prepare()
	}
	return make([]token32, t.stackSize)
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- token32{Rule: v.Rule, begin: int32(v.begin), end: int32(v.end), next: int32(v.next)}
		}
		close(s)
	}()
	return s
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	Rule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

type tokens32 struct {
	tree      []token32
	stackSize int32
}

func (t *tokens32) Len() int {
	return len(t.tree)
}

func (t *tokens32) Less(i, j int) bool {
	ii, jj := t.tree[i], t.tree[j]
	if ii.Rule != RuleUnknown {
		if jj.Rule == RuleUnknown {
			return true
		} else if ii.begin < jj.begin {
			return true
		} else if ii.begin == jj.begin {
			if ii.end == ii.begin || jj.end == jj.begin {
				if ii.next < jj.next {
					return true
				}
			} else if ii.end > jj.end {
				return true
			} else if ii.end == jj.end {
				if ii.next > jj.next {
					return true
				}
			}
		}
	}
	return false
}

func (t *tokens32) Swap(i, j int) {
	t.tree[i], t.tree[j] = t.tree[j], t.tree[i]
}

func (t *tokens32) Prepare() {
	sort.Sort(t)
	size := int(t.tree[0].next) + 1

	tree, stack, top := t.tree[0:size], make([]token32, size), -1
	for i, token := range tree {
		token.next = int32(i)
		for top >= 0 && token.begin >= stack[top].end {
			tree[stack[top].next].next, top = token.next, top-1
		}
		stack[top+1], top = token, top+1
	}

	for i, token := range stack {
		if token.isZero() {
			t.stackSize = int32(i)
			break
		}
	}

	t.tree = tree
}

func (t *tokens32) Add(rule Rule, begin, end, next int) {
	t.tree[next] = token32{Rule: rule, begin: int32(begin), end: int32(end), next: int32(next)}
}

func (t *tokens32) Stack() []token32 {
	if t.stackSize == 0 {
		t.Prepare()
	}
	return make([]token32, t.stackSize)
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- token32{Rule: v.Rule, begin: int32(v.begin), end: int32(v.end), next: int32(v.next)}
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			e := &expanded[i]
			e.Rule, e.begin, e.end, e.next = v.Rule, int32(v.begin), int32(v.end), int32(v.next)
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

const END_SYMBOL byte = 0

type Peg struct {
	*Tree

	Buffer   string
	Min, Max int
	rules    [37]func() bool

	TokenTree
}

func (p *Peg) Add(rule Rule, begin, end, next int) {
	if tree := p.TokenTree.Expand(next); tree != nil {
		p.TokenTree = tree
	}
	p.TokenTree.Add(rule, begin, end, next)
}

func (p *Peg) Parse() os.Error {
	if p.rules[0]() {
		return nil
	}
	return &parseError{p}
}

type parseError struct {
	p *Peg
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

func (p *Peg) PrintSyntaxTree() {
	tokenTree := p.TokenTree
	stack, top, i := tokenTree.Stack(), -1, 0
	for token := range tokenTree.Tokens() {
		if top >= 0 && int(stack[top].next) == i {
			for top >= 0 && int(stack[top].next) == i {
				top--
			}
		}
		stack[top+1], top, i = token, top+1, i+1

		for c := 0; c < top; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(p.Buffer[token.begin:token.end]))
	}
}

func (p *Peg) Highlighter() {
	tokenTree := p.TokenTree
	stack, top, i, c := tokenTree.Stack(), -1, 0, 0
	for token := range tokenTree.Tokens() {
		pops := top
		if top >= 0 && int(stack[top].next) == i {
			for top >= 0 && int(stack[top].next) == i {
				top--
			}

			for c < int(stack[pops].end) {
				fmt.Printf("%v", c)
				for t := 0; t <= pops; t++ {
					if c >= int(stack[t].begin) && c < int(stack[t].end) {
						fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[stack[t].Rule])
					}
				}
				fmt.Printf("\n")
				c++
			}
		}
		stack[top+1], top, i = token, top+1, i+1
	}
}

func (p *Peg) Init() {
	position, tokenIndex := 0, 0
	p.TokenTree = &tokens16{tree: make([]token16, 65536)}

	if p.Buffer[len(p.Buffer)-1] != END_SYMBOL {
		p.Buffer = p.Buffer + string(END_SYMBOL)
	}

	actions := [...]func(buffer string, begin, end int){
		/* 0 */
		func(buffer string, begin, end int) {
			p.AddDoubleRange()
		},
		/* 1 */
		func(buffer string, begin, end int) {
			p.AddAlternate()
		},
		/* 2 */
		func(buffer string, begin, end int) {
			p.AddAlternate()
		},
		/* 3 */
		func(buffer string, begin, end int) {
			p.AddDoubleCharacter(buffer[begin:end])
		},
		/* 4 */
		func(buffer string, begin, end int) {
			p.AddCharacter(buffer[begin:end])
		},
		/* 5 */
		func(buffer string, begin, end int) {
			p.AddPackage(buffer[begin:end])
		},
		/* 6 */
		func(buffer string, begin, end int) {
			p.AddPeg(buffer[begin:end])
		},
		/* 7 */
		func(buffer string, begin, end int) {
			p.AddState(buffer[begin:end])
		},
		/* 8 */
		func(buffer string, begin, end int) {
			p.AddPredicate(buffer[begin:end])
		},
		/* 9 */
		func(buffer string, begin, end int) {
			p.AddPeekFor()
		},
		/* 10 */
		func(buffer string, begin, end int) {
			p.AddPeekNot()
		},
		/* 11 */
		func(buffer string, begin, end int) {
			p.AddCharacter(buffer[begin:end])
		},
		/* 12 */
		func(buffer string, begin, end int) {
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		},
		/* 13 */
		func(buffer string, begin, end int) {
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		},
		/* 14 */
		func(buffer string, begin, end int) {
			p.AddSequence()
		},
		/* 15 */
		func(buffer string, begin, end int) {
			p.AddSequence()
		},
		/* 16 */
		func(buffer string, begin, end int) {
			p.AddAlternate()
		},
		/* 17 */
		func(buffer string, begin, end int) {
			p.AddNil()
			p.AddAlternate()
		},
		/* 18 */
		func(buffer string, begin, end int) {
			p.AddNil()
		},
		/* 19 */
		func(buffer string, begin, end int) {
			p.AddQuery()
		},
		/* 20 */
		func(buffer string, begin, end int) {
			p.AddStar()
		},
		/* 21 */
		func(buffer string, begin, end int) {
			p.AddPlus()
		},
		/* 22 */
		func(buffer string, begin, end int) {
			p.AddRule(buffer[begin:end])
		},
		/* 23 */
		func(buffer string, begin, end int) {
			p.AddExpression()
		},
		/* 24 */
		func(buffer string, begin, end int) {
			p.AddRange()
		},
		/* 25 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\a")
		},
		/* 26 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\b")
		},
		/* 27 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\x1B")
		},
		/* 28 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\f")
		},
		/* 29 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\n")
		},
		/* 30 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\r")
		},
		/* 31 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\t")
		},
		/* 32 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\v")
		},
		/* 33 */
		func(buffer string, begin, end int) {
			p.AddCharacter("'")
		},
		/* 34 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\"")
		},
		/* 35 */
		func(buffer string, begin, end int) {
			p.AddCharacter("[")
		},
		/* 36 */
		func(buffer string, begin, end int) {
			p.AddCharacter("]")
		},
		/* 37 */
		func(buffer string, begin, end int) {
			p.AddCharacter("-")
		},
		/* 38 */
		func(buffer string, begin, end int) {
			p.AddOctalCharacter(buffer[begin:end])
		},
		/* 39 */
		func(buffer string, begin, end int) {
			p.AddOctalCharacter(buffer[begin:end])
		},
		/* 40 */
		func(buffer string, begin, end int) {
			p.AddCharacter("\\")
		},
		/* 41 */
		func(buffer string, begin, end int) {
			p.AddCommit()
		},
		/* 42 */
		func(buffer string, begin, end int) {
			p.AddName(buffer[begin:end])
		},
		/* 43 */
		func(buffer string, begin, end int) {
			p.AddDot()
		},
		/* 44 */
		func(buffer string, begin, end int) {
			p.AddAction(buffer[begin:end])
		},
		/* 45 */
		func(buffer string, begin, end int) {
			p.AddPush()
		},
		/* 46 */
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
		if p.Buffer[position] != END_SYMBOL {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	matchChar := func(c byte) bool {
		if p.Buffer[position] == c {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	matchRange := func(lower byte, upper byte) bool {
		if (p.Buffer[position] >= lower) && (p.Buffer[position] <= upper) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}

	p.rules = [...]func() bool{
		/* 0 Grammar <- <(Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') Spacing Identifier { p.AddPackage(buffer[begin:end]) } ('t' 'y' 'p' 'e') Spacing Identifier { p.AddPeg(buffer[begin:end]) } ('P' 'e' 'g') Spacing Action { p.AddState(buffer[begin:end]) } commit Definition+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0, thunkPosition0 := position, tokenIndex, thunkPosition
			{
				begin1 := position
				if !p.rules[29]() {
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
				if !p.rules[29]() {
					goto l0
				}
				if !p.rules[7]() {
					goto l0
				}
				do(5)
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
				if !p.rules[29]() {
					goto l0
				}
				if !p.rules[7]() {
					goto l0
				}
				do(6)
				if !matchChar('P') {
					goto l0
				}
				if !matchChar('e') {
					goto l0
				}
				if !matchChar('g') {
					goto l0
				}
				if !p.rules[29]() {
					goto l0
				}
				if !p.rules[34]() {
					goto l0
				}
				do(7)
				if !(commit(0)) {
					goto l0
				}
				{
					begin4 := position
					if !p.rules[7]() {
						goto l0
					}
					do(22)
					if !p.rules[19]() {
						goto l0
					}
					if !p.rules[2]() {
						goto l0
					}
					do(23)
					{
						position5, tokenIndex5, thunkPosition5 := position, tokenIndex, thunkPosition
						{
							position6, tokenIndex6, thunkPosition6 := position, tokenIndex, thunkPosition
							if !p.rules[7]() {
								goto l7
							}
							if !p.rules[19]() {
								goto l7
							}
							goto l6
						l7:
							position, tokenIndex, thunkPosition = position6, tokenIndex6, thunkPosition6
							{
								position8, tokenIndex8, thunkPosition8 := position, tokenIndex, thunkPosition
								if !matchDot() {
									goto l8
								}
								goto l0
							l8:
								position, tokenIndex, thunkPosition = position8, tokenIndex8, thunkPosition8
							}
						}
					l6:
						position, tokenIndex, thunkPosition = position5, tokenIndex5, thunkPosition5
					}
					if !(commit(0)) {
						goto l0
					}
					if begin4 != position {
						p.Add(RuleDefinition, begin4, position, tokenIndex)
						tokenIndex++
					}
				}
			l2:
				{
					position3, tokenIndex3, thunkPosition3 := position, tokenIndex, thunkPosition
					{
						begin9 := position
						if !p.rules[7]() {
							goto l3
						}
						do(22)
						if !p.rules[19]() {
							goto l3
						}
						if !p.rules[2]() {
							goto l3
						}
						do(23)
						{
							position10, tokenIndex10, thunkPosition10 := position, tokenIndex, thunkPosition
							{
								position11, tokenIndex11, thunkPosition11 := position, tokenIndex, thunkPosition
								if !p.rules[7]() {
									goto l12
								}
								if !p.rules[19]() {
									goto l12
								}
								goto l11
							l12:
								position, tokenIndex, thunkPosition = position11, tokenIndex11, thunkPosition11
								{
									position13, tokenIndex13, thunkPosition13 := position, tokenIndex, thunkPosition
									if !matchDot() {
										goto l13
									}
									goto l3
								l13:
									position, tokenIndex, thunkPosition = position13, tokenIndex13, thunkPosition13
								}
							}
						l11:
							position, tokenIndex, thunkPosition = position10, tokenIndex10, thunkPosition10
						}
						if !(commit(0)) {
							goto l3
						}
						if begin9 != position {
							p.Add(RuleDefinition, begin9, position, tokenIndex)
							tokenIndex++
						}
					}
					goto l2
				l3:
					position, tokenIndex, thunkPosition = position3, tokenIndex3, thunkPosition3
				}
				{
					begin14 := position
					{
						position15, tokenIndex15, thunkPosition15 := position, tokenIndex, thunkPosition
						if !matchDot() {
							goto l15
						}
						goto l0
					l15:
						position, tokenIndex, thunkPosition = position15, tokenIndex15, thunkPosition15
					}
					if begin14 != position {
						p.Add(RuleEndOfFile, begin14, position, tokenIndex)
						tokenIndex++
					}
				}
				if begin1 != position {
					p.Add(RuleGrammar, begin1, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l0:
			position, tokenIndex, thunkPosition = position0, tokenIndex0, thunkPosition0
			return false
		},
		/* 1 Definition <- <(Identifier { p.AddRule(buffer[begin:end]) } LeftArrow Expression { p.AddExpression() } &((Identifier LeftArrow) / !.) commit)> */
		nil,
		/* 2 Expression <- <((Sequence (Slash Sequence { p.AddAlternate() })* (Slash { p.AddNil(); p.AddAlternate() })?) / { p.AddNil() })> */
		func() bool {
			{
				begin18 := position
				{
					position19, tokenIndex19, thunkPosition19 := position, tokenIndex, thunkPosition
					if !p.rules[3]() {
						goto l20
					}
				l21:
					{
						position22, tokenIndex22, thunkPosition22 := position, tokenIndex, thunkPosition
						if !p.rules[20]() {
							goto l22
						}
						if !p.rules[3]() {
							goto l22
						}
						do(16)
						goto l21
					l22:
						position, tokenIndex, thunkPosition = position22, tokenIndex22, thunkPosition22
					}
					{
						position23, tokenIndex23, thunkPosition23 := position, tokenIndex, thunkPosition
						if !p.rules[20]() {
							goto l23
						}
						do(17)
						goto l24
					l23:
						position, tokenIndex, thunkPosition = position23, tokenIndex23, thunkPosition23
					}
				l24:
					goto l19
				l20:
					position, tokenIndex, thunkPosition = position19, tokenIndex19, thunkPosition19
					do(18)
				}
			l19:
				if begin18 != position {
					p.Add(RuleExpression, begin18, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		},
		/* 3 Sequence <- <(Prefix (Prefix { p.AddSequence() })*)> */
		func() bool {
			position25, tokenIndex25, thunkPosition25 := position, tokenIndex, thunkPosition
			{
				begin26 := position
				if !p.rules[4]() {
					goto l25
				}
			l27:
				{
					position28, tokenIndex28, thunkPosition28 := position, tokenIndex, thunkPosition
					if !p.rules[4]() {
						goto l28
					}
					do(46)
					goto l27
				l28:
					position, tokenIndex, thunkPosition = position28, tokenIndex28, thunkPosition28
				}
				if begin26 != position {
					p.Add(RuleSequence, begin26, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l25:
			position, tokenIndex, thunkPosition = position25, tokenIndex25, thunkPosition25
			return false
		},
		/* 4 Prefix <- <((And Action { p.AddPredicate(buffer[begin:end]) }) / ((&('!') (Not Suffix { p.AddPeekNot() })) | (&('&') (And Suffix { p.AddPeekFor() })) | (&('"' | '\'' | '(' | '.' | '<' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{') Suffix)))> */
		func() bool {
			position29, tokenIndex29, thunkPosition29 := position, tokenIndex, thunkPosition
			{
				begin30 := position
				{
					position31, tokenIndex31, thunkPosition31 := position, tokenIndex, thunkPosition
					if !p.rules[21]() {
						goto l32
					}
					if !p.rules[34]() {
						goto l32
					}
					do(8)
					goto l31
				l32:
					position, tokenIndex, thunkPosition = position31, tokenIndex31, thunkPosition31
					{
						if position == len(p.Buffer) {
							goto l29
						}
						switch p.Buffer[position] {
						case '!':
							{
								begin34 := position
								if !matchChar('!') {
									goto l29
								}
								if !p.rules[29]() {
									goto l29
								}
								if begin34 != position {
									p.Add(RuleNot, begin34, position, tokenIndex)
									tokenIndex++
								}
							}
							if !p.rules[5]() {
								goto l29
							}
							do(10)
							break
						case '&':
							if !p.rules[21]() {
								goto l29
							}
							if !p.rules[5]() {
								goto l29
							}
							do(9)
							break
						default:
							if !p.rules[5]() {
								goto l29
							}
							break
						}
					}

				}
			l31:
				if begin30 != position {
					p.Add(RulePrefix, begin30, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l29:
			position, tokenIndex, thunkPosition = position29, tokenIndex29, thunkPosition29
			return false
		},
		/* 5 Suffix <- <(Primary ((&('+') (Plus { p.AddPlus() })) | (&('*') (Star { p.AddStar() })) | (&('?') (Question { p.AddQuery() })))?)> */
		func() bool {
			position35, tokenIndex35, thunkPosition35 := position, tokenIndex, thunkPosition
			{
				begin36 := position
				{
					begin37 := position
					{
						position38, tokenIndex38, thunkPosition38 := position, tokenIndex, thunkPosition
						if !matchChar('c') {
							goto l39
						}
						if !matchChar('o') {
							goto l39
						}
						if !matchChar('m') {
							goto l39
						}
						if !matchChar('m') {
							goto l39
						}
						if !matchChar('i') {
							goto l39
						}
						if !matchChar('t') {
							goto l39
						}
						if !p.rules[29]() {
							goto l39
						}
						do(41)
						goto l38
					l39:
						position, tokenIndex, thunkPosition = position38, tokenIndex38, thunkPosition38
						{
							if position == len(p.Buffer) {
								goto l35
							}
							switch p.Buffer[position] {
							case '<':
								{
									begin41 := position
									if !matchChar('<') {
										goto l35
									}
									if !p.rules[29]() {
										goto l35
									}
									if begin41 != position {
										p.Add(RuleBegin, begin41, position, tokenIndex)
										tokenIndex++
									}
								}
								if !p.rules[2]() {
									goto l35
								}
								{
									begin42 := position
									if !matchChar('>') {
										goto l35
									}
									if !p.rules[29]() {
										goto l35
									}
									if begin42 != position {
										p.Add(RuleEnd, begin42, position, tokenIndex)
										tokenIndex++
									}
								}
								do(45)
								break
							case '{':
								if !p.rules[34]() {
									goto l35
								}
								do(44)
								break
							case '.':
								{
									begin43 := position
									if !matchChar('.') {
										goto l35
									}
									if !p.rules[29]() {
										goto l35
									}
									if begin43 != position {
										p.Add(RuleDot, begin43, position, tokenIndex)
										tokenIndex++
									}
								}
								do(43)
								break
							case '[':
								{
									begin44 := position
									{
										position45, tokenIndex45, thunkPosition45 := position, tokenIndex, thunkPosition
										if !matchChar('[') {
											goto l46
										}
										if !matchChar('[') {
											goto l46
										}
										{
											position47, tokenIndex47, thunkPosition47 := position, tokenIndex, thunkPosition
											{
												position49, tokenIndex49, thunkPosition49 := position, tokenIndex, thunkPosition
												if !matchChar('^') {
													goto l50
												}
												if !p.rules[13]() {
													goto l50
												}
												do(12)
												goto l49
											l50:
												position, tokenIndex, thunkPosition = position49, tokenIndex49, thunkPosition49
												if !p.rules[13]() {
													goto l47
												}
											}
										l49:
											goto l48
										l47:
											position, tokenIndex, thunkPosition = position47, tokenIndex47, thunkPosition47
										}
									l48:
										if !matchChar(']') {
											goto l46
										}
										if !matchChar(']') {
											goto l46
										}
										goto l45
									l46:
										position, tokenIndex, thunkPosition = position45, tokenIndex45, thunkPosition45
										if !matchChar('[') {
											goto l35
										}
										{
											position51, tokenIndex51, thunkPosition51 := position, tokenIndex, thunkPosition
											{
												position53, tokenIndex53, thunkPosition53 := position, tokenIndex, thunkPosition
												if !matchChar('^') {
													goto l54
												}
												if !p.rules[12]() {
													goto l54
												}
												do(13)
												goto l53
											l54:
												position, tokenIndex, thunkPosition = position53, tokenIndex53, thunkPosition53
												if !p.rules[12]() {
													goto l51
												}
											}
										l53:
											goto l52
										l51:
											position, tokenIndex, thunkPosition = position51, tokenIndex51, thunkPosition51
										}
									l52:
										if !matchChar(']') {
											goto l35
										}
									}
								l45:
									if !p.rules[29]() {
										goto l35
									}
									if begin44 != position {
										p.Add(RuleClass, begin44, position, tokenIndex)
										tokenIndex++
									}
								}
								break
							case '"', '\'':
								{
									begin55 := position
									{
										position56, tokenIndex56, thunkPosition56 := position, tokenIndex, thunkPosition
										if !matchChar('\'') {
											goto l57
										}
										{
											position58, tokenIndex58, thunkPosition58 := position, tokenIndex, thunkPosition
											{
												position60, tokenIndex60, thunkPosition60 := position, tokenIndex, thunkPosition
												if !matchChar('\'') {
													goto l60
												}
												goto l58
											l60:
												position, tokenIndex, thunkPosition = position60, tokenIndex60, thunkPosition60
											}
											if !p.rules[16]() {
												goto l58
											}
											goto l59
										l58:
											position, tokenIndex, thunkPosition = position58, tokenIndex58, thunkPosition58
										}
									l59:
									l61:
										{
											position62, tokenIndex62, thunkPosition62 := position, tokenIndex, thunkPosition
											{
												position63, tokenIndex63, thunkPosition63 := position, tokenIndex, thunkPosition
												if !matchChar('\'') {
													goto l63
												}
												goto l62
											l63:
												position, tokenIndex, thunkPosition = position63, tokenIndex63, thunkPosition63
											}
											if !p.rules[16]() {
												goto l62
											}
											do(14)
											goto l61
										l62:
											position, tokenIndex, thunkPosition = position62, tokenIndex62, thunkPosition62
										}
										if !matchChar('\'') {
											goto l57
										}
										if !p.rules[29]() {
											goto l57
										}
										goto l56
									l57:
										position, tokenIndex, thunkPosition = position56, tokenIndex56, thunkPosition56
										if !matchChar('"') {
											goto l35
										}
										{
											position64, tokenIndex64, thunkPosition64 := position, tokenIndex, thunkPosition
											{
												position66, tokenIndex66, thunkPosition66 := position, tokenIndex, thunkPosition
												if !matchChar('"') {
													goto l66
												}
												goto l64
											l66:
												position, tokenIndex, thunkPosition = position66, tokenIndex66, thunkPosition66
											}
											if !p.rules[17]() {
												goto l64
											}
											goto l65
										l64:
											position, tokenIndex, thunkPosition = position64, tokenIndex64, thunkPosition64
										}
									l65:
									l67:
										{
											position68, tokenIndex68, thunkPosition68 := position, tokenIndex, thunkPosition
											{
												position69, tokenIndex69, thunkPosition69 := position, tokenIndex, thunkPosition
												if !matchChar('"') {
													goto l69
												}
												goto l68
											l69:
												position, tokenIndex, thunkPosition = position69, tokenIndex69, thunkPosition69
											}
											if !p.rules[17]() {
												goto l68
											}
											do(15)
											goto l67
										l68:
											position, tokenIndex, thunkPosition = position68, tokenIndex68, thunkPosition68
										}
										if !matchChar('"') {
											goto l35
										}
										if !p.rules[29]() {
											goto l35
										}
									}
								l56:
									if begin55 != position {
										p.Add(RuleLiteral, begin55, position, tokenIndex)
										tokenIndex++
									}
								}
								break
							case '(':
								{
									begin70 := position
									if !matchChar('(') {
										goto l35
									}
									if !p.rules[29]() {
										goto l35
									}
									if begin70 != position {
										p.Add(RuleOpen, begin70, position, tokenIndex)
										tokenIndex++
									}
								}
								if !p.rules[2]() {
									goto l35
								}
								{
									begin71 := position
									if !matchChar(')') {
										goto l35
									}
									if !p.rules[29]() {
										goto l35
									}
									if begin71 != position {
										p.Add(RuleClose, begin71, position, tokenIndex)
										tokenIndex++
									}
								}
								break
							default:
								if !p.rules[7]() {
									goto l35
								}
								{
									position72, tokenIndex72, thunkPosition72 := position, tokenIndex, thunkPosition
									if !p.rules[19]() {
										goto l72
									}
									goto l35
								l72:
									position, tokenIndex, thunkPosition = position72, tokenIndex72, thunkPosition72
								}
								do(42)
								break
							}
						}

					}
				l38:
					if begin37 != position {
						p.Add(RulePrimary, begin37, position, tokenIndex)
						tokenIndex++
					}
				}
				{
					position73, tokenIndex73, thunkPosition73 := position, tokenIndex, thunkPosition
					{
						if position == len(p.Buffer) {
							goto l73
						}
						switch p.Buffer[position] {
						case '+':
							{
								begin76 := position
								if !matchChar('+') {
									goto l73
								}
								if !p.rules[29]() {
									goto l73
								}
								if begin76 != position {
									p.Add(RulePlus, begin76, position, tokenIndex)
									tokenIndex++
								}
							}
							do(21)
							break
						case '*':
							{
								begin77 := position
								if !matchChar('*') {
									goto l73
								}
								if !p.rules[29]() {
									goto l73
								}
								if begin77 != position {
									p.Add(RuleStar, begin77, position, tokenIndex)
									tokenIndex++
								}
							}
							do(20)
							break
						default:
							{
								begin78 := position
								if !matchChar('?') {
									goto l73
								}
								if !p.rules[29]() {
									goto l73
								}
								if begin78 != position {
									p.Add(RuleQuestion, begin78, position, tokenIndex)
									tokenIndex++
								}
							}
							do(19)
							break
						}
					}

					goto l74
				l73:
					position, tokenIndex, thunkPosition = position73, tokenIndex73, thunkPosition73
				}
			l74:
				if begin36 != position {
					p.Add(RuleSuffix, begin36, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l35:
			position, tokenIndex, thunkPosition = position35, tokenIndex35, thunkPosition35
			return false
		},
		/* 6 Primary <- <(('c' 'o' 'm' 'm' 'i' 't' Spacing { p.AddCommit() }) / ((&('<') (Begin Expression End { p.AddPush() })) | (&('{') (Action { p.AddAction(buffer[begin:end]) })) | (&('.') (Dot { p.AddDot() })) | (&('[') Class) | (&('"' | '\'') Literal) | (&('(') (Open Expression Close)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (Identifier !LeftArrow { p.AddName(buffer[begin:end]) }))))> */
		nil,
		/* 7 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position80, tokenIndex80, thunkPosition80 := position, tokenIndex, thunkPosition
			{
				begin81 := position
				{
					begin = position
					begin82 := position
					if !p.rules[8]() {
						goto l80
					}
				l83:
					{
						position84, tokenIndex84, thunkPosition84 := position, tokenIndex, thunkPosition
						{
							begin85 := position
							{
								position86, tokenIndex86, thunkPosition86 := position, tokenIndex, thunkPosition
								if !p.rules[8]() {
									goto l87
								}
								goto l86
							l87:
								position, tokenIndex, thunkPosition = position86, tokenIndex86, thunkPosition86
								if !matchRange('0', '9') {
									goto l84
								}
							}
						l86:
							if begin85 != position {
								p.Add(RuleIdentCont, begin85, position, tokenIndex)
								tokenIndex++
							}
						}
						goto l83
					l84:
						position, tokenIndex, thunkPosition = position84, tokenIndex84, thunkPosition84
					}
					end = position
					if begin82 != position {
						p.Add(RuleIdentifier, begin82, position, tokenIndex)
						tokenIndex++
					}
				}
				if !p.rules[29]() {
					goto l80
				}
				if begin81 != position {
					p.Add(RuleIdentifier, begin81, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l80:
			position, tokenIndex, thunkPosition = position80, tokenIndex80, thunkPosition80
			return false
		},
		/* 8 IdentStart <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position88, tokenIndex88, thunkPosition88 := position, tokenIndex, thunkPosition
			{
				begin89 := position
				{
					if position == len(p.Buffer) {
						goto l88
					}
					switch p.Buffer[position] {
					case '_':
						if !matchChar('_') {
							goto l88
						}
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if !matchRange('A', 'Z') {
							goto l88
						}
						break
					default:
						if !matchRange('a', 'z') {
							goto l88
						}
						break
					}
				}

				if begin89 != position {
					p.Add(RuleIdentStart, begin89, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l88:
			position, tokenIndex, thunkPosition = position88, tokenIndex88, thunkPosition88
			return false
		},
		/* 9 IdentCont <- <(IdentStart / [0-9])> */
		nil,
		/* 10 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char { p.AddSequence() })* '\'' Spacing) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar { p.AddSequence() })* '"' Spacing))> */
		nil,
		/* 11 Class <- <((('[' '[' (('^' DoubleRanges { p.AddPeekNot(); p.AddDot(); p.AddSequence() }) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges { p.AddPeekNot(); p.AddDot(); p.AddSequence() }) / Ranges)? ']')) Spacing)> */
		nil,
		/* 12 Ranges <- <(!']' Range (!']' Range { p.AddAlternate() })*)> */
		func() bool {
			position94, tokenIndex94, thunkPosition94 := position, tokenIndex, thunkPosition
			{
				begin95 := position
				{
					position96, tokenIndex96, thunkPosition96 := position, tokenIndex, thunkPosition
					if !matchChar(']') {
						goto l96
					}
					goto l94
				l96:
					position, tokenIndex, thunkPosition = position96, tokenIndex96, thunkPosition96
				}
				if !p.rules[14]() {
					goto l94
				}
			l97:
				{
					position98, tokenIndex98, thunkPosition98 := position, tokenIndex, thunkPosition
					{
						position99, tokenIndex99, thunkPosition99 := position, tokenIndex, thunkPosition
						if !matchChar(']') {
							goto l99
						}
						goto l98
					l99:
						position, tokenIndex, thunkPosition = position99, tokenIndex99, thunkPosition99
					}
					if !p.rules[14]() {
						goto l98
					}
					do(1)
					goto l97
				l98:
					position, tokenIndex, thunkPosition = position98, tokenIndex98, thunkPosition98
				}
				if begin95 != position {
					p.Add(RuleRanges, begin95, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l94:
			position, tokenIndex, thunkPosition = position94, tokenIndex94, thunkPosition94
			return false
		},
		/* 13 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange { p.AddAlternate() })*)> */
		func() bool {
			position100, tokenIndex100, thunkPosition100 := position, tokenIndex, thunkPosition
			{
				begin101 := position
				{
					position102, tokenIndex102, thunkPosition102 := position, tokenIndex, thunkPosition
					if !matchChar(']') {
						goto l102
					}
					if !matchChar(']') {
						goto l102
					}
					goto l100
				l102:
					position, tokenIndex, thunkPosition = position102, tokenIndex102, thunkPosition102
				}
				if !p.rules[15]() {
					goto l100
				}
			l103:
				{
					position104, tokenIndex104, thunkPosition104 := position, tokenIndex, thunkPosition
					{
						position105, tokenIndex105, thunkPosition105 := position, tokenIndex, thunkPosition
						if !matchChar(']') {
							goto l105
						}
						if !matchChar(']') {
							goto l105
						}
						goto l104
					l105:
						position, tokenIndex, thunkPosition = position105, tokenIndex105, thunkPosition105
					}
					if !p.rules[15]() {
						goto l104
					}
					do(2)
					goto l103
				l104:
					position, tokenIndex, thunkPosition = position104, tokenIndex104, thunkPosition104
				}
				if begin101 != position {
					p.Add(RuleDoubleRanges, begin101, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l100:
			position, tokenIndex, thunkPosition = position100, tokenIndex100, thunkPosition100
			return false
		},
		/* 14 Range <- <((Char '-' Char { p.AddRange() }) / Char)> */
		func() bool {
			position106, tokenIndex106, thunkPosition106 := position, tokenIndex, thunkPosition
			{
				begin107 := position
				{
					position108, tokenIndex108, thunkPosition108 := position, tokenIndex, thunkPosition
					if !p.rules[16]() {
						goto l109
					}
					if !matchChar('-') {
						goto l109
					}
					if !p.rules[16]() {
						goto l109
					}
					do(24)
					goto l108
				l109:
					position, tokenIndex, thunkPosition = position108, tokenIndex108, thunkPosition108
					if !p.rules[16]() {
						goto l106
					}
				}
			l108:
				if begin107 != position {
					p.Add(RuleRange, begin107, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l106:
			position, tokenIndex, thunkPosition = position106, tokenIndex106, thunkPosition106
			return false
		},
		/* 15 DoubleRange <- <((Char '-' Char { p.AddDoubleRange() }) / DoubleChar)> */
		func() bool {
			position110, tokenIndex110, thunkPosition110 := position, tokenIndex, thunkPosition
			{
				begin111 := position
				{
					position112, tokenIndex112, thunkPosition112 := position, tokenIndex, thunkPosition
					if !p.rules[16]() {
						goto l113
					}
					if !matchChar('-') {
						goto l113
					}
					if !p.rules[16]() {
						goto l113
					}
					do(0)
					goto l112
				l113:
					position, tokenIndex, thunkPosition = position112, tokenIndex112, thunkPosition112
					if !p.rules[17]() {
						goto l110
					}
				}
			l112:
				if begin111 != position {
					p.Add(RuleDoubleRange, begin111, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l110:
			position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
			return false
		},
		/* 16 Char <- <(Escape / (!'\\' <.> { p.AddCharacter(buffer[begin:end]) }))> */
		func() bool {
			position114, tokenIndex114, thunkPosition114 := position, tokenIndex, thunkPosition
			{
				begin115 := position
				{
					position116, tokenIndex116, thunkPosition116 := position, tokenIndex, thunkPosition
					if !p.rules[18]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, thunkPosition = position116, tokenIndex116, thunkPosition116
					{
						position118, tokenIndex118, thunkPosition118 := position, tokenIndex, thunkPosition
						if !matchChar('\\') {
							goto l118
						}
						goto l114
					l118:
						position, tokenIndex, thunkPosition = position118, tokenIndex118, thunkPosition118
					}
					{
						begin = position
						begin119 := position
						if !matchDot() {
							goto l114
						}
						end = position
						if begin119 != position {
							p.Add(RuleChar, begin119, position, tokenIndex)
							tokenIndex++
						}
					}
					do(11)
				}
			l116:
				if begin115 != position {
					p.Add(RuleChar, begin115, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l114:
			position, tokenIndex, thunkPosition = position114, tokenIndex114, thunkPosition114
			return false
		},
		/* 17 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> { p.AddDoubleCharacter(buffer[begin:end]) }) / (!'\\' <.> { p.AddCharacter(buffer[begin:end]) }))> */
		func() bool {
			position120, tokenIndex120, thunkPosition120 := position, tokenIndex, thunkPosition
			{
				begin121 := position
				{
					position122, tokenIndex122, thunkPosition122 := position, tokenIndex, thunkPosition
					if !p.rules[18]() {
						goto l123
					}
					goto l122
				l123:
					position, tokenIndex, thunkPosition = position122, tokenIndex122, thunkPosition122
					{
						begin = position
						begin125 := position
						{
							position126, tokenIndex126, thunkPosition126 := position, tokenIndex, thunkPosition
							if !matchRange('a', 'z') {
								goto l127
							}
							goto l126
						l127:
							position, tokenIndex, thunkPosition = position126, tokenIndex126, thunkPosition126
							if !matchRange('A', 'Z') {
								goto l124
							}
						}
					l126:
						end = position
						if begin125 != position {
							p.Add(RuleDoubleChar, begin125, position, tokenIndex)
							tokenIndex++
						}
					}
					do(3)
					goto l122
				l124:
					position, tokenIndex, thunkPosition = position122, tokenIndex122, thunkPosition122
					{
						position128, tokenIndex128, thunkPosition128 := position, tokenIndex, thunkPosition
						if !matchChar('\\') {
							goto l128
						}
						goto l120
					l128:
						position, tokenIndex, thunkPosition = position128, tokenIndex128, thunkPosition128
					}
					{
						begin = position
						begin129 := position
						if !matchDot() {
							goto l120
						}
						end = position
						if begin129 != position {
							p.Add(RuleDoubleChar, begin129, position, tokenIndex)
							tokenIndex++
						}
					}
					do(4)
				}
			l122:
				if begin121 != position {
					p.Add(RuleDoubleChar, begin121, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l120:
			position, tokenIndex, thunkPosition = position120, tokenIndex120, thunkPosition120
			return false
		},
		/* 18 Escape <- <(('\\' ('a' / 'A') { p.AddCharacter("\a") }) / ('\\' ('b' / 'B') { p.AddCharacter("\b") }) / ('\\' ('e' / 'E') { p.AddCharacter("\x1B") }) / ('\\' ('f' / 'F') { p.AddCharacter("\f") }) / ('\\' ('n' / 'N') { p.AddCharacter("\n") }) / ('\\' ('r' / 'R') { p.AddCharacter("\r") }) / ('\\' ('t' / 'T') { p.AddCharacter("\t") }) / ('\\' ('v' / 'V') { p.AddCharacter("\v") }) / ('\\' '\'' { p.AddCharacter("'") }) / ('\\' '"' { p.AddCharacter("\"") }) / ('\\' '[' { p.AddCharacter("[") }) / ('\\' ']' { p.AddCharacter("]") }) / ('\\' '-' { p.AddCharacter("-") }) / ('\\' <([0-3] [0-7] [0-7])> { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' <([0-7] [0-7]?)> { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' '\\' { p.AddCharacter("\\") }))> */
		func() bool {
			position130, tokenIndex130, thunkPosition130 := position, tokenIndex, thunkPosition
			{
				begin131 := position
				{
					position132, tokenIndex132, thunkPosition132 := position, tokenIndex, thunkPosition
					if !matchChar('\\') {
						goto l133
					}
					{
						position134, tokenIndex134, thunkPosition134 := position, tokenIndex, thunkPosition
						if !matchChar('a') {
							goto l135
						}
						goto l134
					l135:
						position, tokenIndex, thunkPosition = position134, tokenIndex134, thunkPosition134
						if !matchChar('A') {
							goto l133
						}
					}
				l134:
					do(25)
					goto l132
				l133:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l136
					}
					{
						position137, tokenIndex137, thunkPosition137 := position, tokenIndex, thunkPosition
						if !matchChar('b') {
							goto l138
						}
						goto l137
					l138:
						position, tokenIndex, thunkPosition = position137, tokenIndex137, thunkPosition137
						if !matchChar('B') {
							goto l136
						}
					}
				l137:
					do(26)
					goto l132
				l136:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l139
					}
					{
						position140, tokenIndex140, thunkPosition140 := position, tokenIndex, thunkPosition
						if !matchChar('e') {
							goto l141
						}
						goto l140
					l141:
						position, tokenIndex, thunkPosition = position140, tokenIndex140, thunkPosition140
						if !matchChar('E') {
							goto l139
						}
					}
				l140:
					do(27)
					goto l132
				l139:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l142
					}
					{
						position143, tokenIndex143, thunkPosition143 := position, tokenIndex, thunkPosition
						if !matchChar('f') {
							goto l144
						}
						goto l143
					l144:
						position, tokenIndex, thunkPosition = position143, tokenIndex143, thunkPosition143
						if !matchChar('F') {
							goto l142
						}
					}
				l143:
					do(28)
					goto l132
				l142:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l145
					}
					{
						position146, tokenIndex146, thunkPosition146 := position, tokenIndex, thunkPosition
						if !matchChar('n') {
							goto l147
						}
						goto l146
					l147:
						position, tokenIndex, thunkPosition = position146, tokenIndex146, thunkPosition146
						if !matchChar('N') {
							goto l145
						}
					}
				l146:
					do(29)
					goto l132
				l145:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l148
					}
					{
						position149, tokenIndex149, thunkPosition149 := position, tokenIndex, thunkPosition
						if !matchChar('r') {
							goto l150
						}
						goto l149
					l150:
						position, tokenIndex, thunkPosition = position149, tokenIndex149, thunkPosition149
						if !matchChar('R') {
							goto l148
						}
					}
				l149:
					do(30)
					goto l132
				l148:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l151
					}
					{
						position152, tokenIndex152, thunkPosition152 := position, tokenIndex, thunkPosition
						if !matchChar('t') {
							goto l153
						}
						goto l152
					l153:
						position, tokenIndex, thunkPosition = position152, tokenIndex152, thunkPosition152
						if !matchChar('T') {
							goto l151
						}
					}
				l152:
					do(31)
					goto l132
				l151:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l154
					}
					{
						position155, tokenIndex155, thunkPosition155 := position, tokenIndex, thunkPosition
						if !matchChar('v') {
							goto l156
						}
						goto l155
					l156:
						position, tokenIndex, thunkPosition = position155, tokenIndex155, thunkPosition155
						if !matchChar('V') {
							goto l154
						}
					}
				l155:
					do(32)
					goto l132
				l154:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l157
					}
					if !matchChar('\'') {
						goto l157
					}
					do(33)
					goto l132
				l157:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l158
					}
					if !matchChar('"') {
						goto l158
					}
					do(34)
					goto l132
				l158:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l159
					}
					if !matchChar('[') {
						goto l159
					}
					do(35)
					goto l132
				l159:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l160
					}
					if !matchChar(']') {
						goto l160
					}
					do(36)
					goto l132
				l160:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l161
					}
					if !matchChar('-') {
						goto l161
					}
					do(37)
					goto l132
				l161:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l162
					}
					{
						begin = position
						begin163 := position
						if !matchRange('0', '3') {
							goto l162
						}
						if !matchRange('0', '7') {
							goto l162
						}
						if !matchRange('0', '7') {
							goto l162
						}
						end = position
						if begin163 != position {
							p.Add(RuleEscape, begin163, position, tokenIndex)
							tokenIndex++
						}
					}
					do(38)
					goto l132
				l162:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l164
					}
					{
						begin = position
						begin165 := position
						if !matchRange('0', '7') {
							goto l164
						}
						{
							position166, tokenIndex166, thunkPosition166 := position, tokenIndex, thunkPosition
							if !matchRange('0', '7') {
								goto l166
							}
							goto l167
						l166:
							position, tokenIndex, thunkPosition = position166, tokenIndex166, thunkPosition166
						}
					l167:
						end = position
						if begin165 != position {
							p.Add(RuleEscape, begin165, position, tokenIndex)
							tokenIndex++
						}
					}
					do(39)
					goto l132
				l164:
					position, tokenIndex, thunkPosition = position132, tokenIndex132, thunkPosition132
					if !matchChar('\\') {
						goto l130
					}
					if !matchChar('\\') {
						goto l130
					}
					do(40)
				}
			l132:
				if begin131 != position {
					p.Add(RuleEscape, begin131, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l130:
			position, tokenIndex, thunkPosition = position130, tokenIndex130, thunkPosition130
			return false
		},
		/* 19 LeftArrow <- <('<' '-' Spacing)> */
		func() bool {
			position168, tokenIndex168, thunkPosition168 := position, tokenIndex, thunkPosition
			{
				begin169 := position
				if !matchChar('<') {
					goto l168
				}
				if !matchChar('-') {
					goto l168
				}
				if !p.rules[29]() {
					goto l168
				}
				if begin169 != position {
					p.Add(RuleLeftArrow, begin169, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l168:
			position, tokenIndex, thunkPosition = position168, tokenIndex168, thunkPosition168
			return false
		},
		/* 20 Slash <- <('/' Spacing)> */
		func() bool {
			position170, tokenIndex170, thunkPosition170 := position, tokenIndex, thunkPosition
			{
				begin171 := position
				if !matchChar('/') {
					goto l170
				}
				if !p.rules[29]() {
					goto l170
				}
				if begin171 != position {
					p.Add(RuleSlash, begin171, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l170:
			position, tokenIndex, thunkPosition = position170, tokenIndex170, thunkPosition170
			return false
		},
		/* 21 And <- <('&' Spacing)> */
		func() bool {
			position172, tokenIndex172, thunkPosition172 := position, tokenIndex, thunkPosition
			{
				begin173 := position
				if !matchChar('&') {
					goto l172
				}
				if !p.rules[29]() {
					goto l172
				}
				if begin173 != position {
					p.Add(RuleAnd, begin173, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l172:
			position, tokenIndex, thunkPosition = position172, tokenIndex172, thunkPosition172
			return false
		},
		/* 22 Not <- <('!' Spacing)> */
		nil,
		/* 23 Question <- <('?' Spacing)> */
		nil,
		/* 24 Star <- <('*' Spacing)> */
		nil,
		/* 25 Plus <- <('+' Spacing)> */
		nil,
		/* 26 Open <- <('(' Spacing)> */
		nil,
		/* 27 Close <- <(')' Spacing)> */
		nil,
		/* 28 Dot <- <('.' Spacing)> */
		nil,
		/* 29 Spacing <- <(Space / Comment)*> */
		func() bool {
			{
				begin182 := position
			l183:
				{
					position184, tokenIndex184, thunkPosition184 := position, tokenIndex, thunkPosition
					{
						position185, tokenIndex185, thunkPosition185 := position, tokenIndex, thunkPosition
						{
							begin187 := position
							{
								if position == len(p.Buffer) {
									goto l186
								}
								switch p.Buffer[position] {
								case '\t':
									if !matchChar('\t') {
										goto l186
									}
									break
								case ' ':
									if !matchChar(' ') {
										goto l186
									}
									break
								default:
									if !p.rules[32]() {
										goto l186
									}
									break
								}
							}

							if begin187 != position {
								p.Add(RuleSpace, begin187, position, tokenIndex)
								tokenIndex++
							}
						}
						goto l185
					l186:
						position, tokenIndex, thunkPosition = position185, tokenIndex185, thunkPosition185
						{
							begin189 := position
							if !matchChar('#') {
								goto l184
							}
						l190:
							{
								position191, tokenIndex191, thunkPosition191 := position, tokenIndex, thunkPosition
								{
									position192, tokenIndex192, thunkPosition192 := position, tokenIndex, thunkPosition
									if !p.rules[32]() {
										goto l192
									}
									goto l191
								l192:
									position, tokenIndex, thunkPosition = position192, tokenIndex192, thunkPosition192
								}
								if !matchDot() {
									goto l191
								}
								goto l190
							l191:
								position, tokenIndex, thunkPosition = position191, tokenIndex191, thunkPosition191
							}
							if !p.rules[32]() {
								goto l184
							}
							if begin189 != position {
								p.Add(RuleComment, begin189, position, tokenIndex)
								tokenIndex++
							}
						}
					}
				l185:
					goto l183
				l184:
					position, tokenIndex, thunkPosition = position184, tokenIndex184, thunkPosition184
				}
				if begin182 != position {
					p.Add(RuleSpacing, begin182, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		},
		/* 30 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		nil,
		/* 31 Space <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine))> */
		nil,
		/* 32 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position195, tokenIndex195, thunkPosition195 := position, tokenIndex, thunkPosition
			{
				begin196 := position
				{
					position197, tokenIndex197, thunkPosition197 := position, tokenIndex, thunkPosition
					if !matchChar('\r') {
						goto l198
					}
					if !matchChar('\n') {
						goto l198
					}
					goto l197
				l198:
					position, tokenIndex, thunkPosition = position197, tokenIndex197, thunkPosition197
					if !matchChar('\n') {
						goto l199
					}
					goto l197
				l199:
					position, tokenIndex, thunkPosition = position197, tokenIndex197, thunkPosition197
					if !matchChar('\r') {
						goto l195
					}
				}
			l197:
				if begin196 != position {
					p.Add(RuleEndOfLine, begin196, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l195:
			position, tokenIndex, thunkPosition = position195, tokenIndex195, thunkPosition195
			return false
		},
		/* 33 EndOfFile <- <!.> */
		nil,
		/* 34 Action <- <('{' <(!'}' .)*> '}' Spacing)> */
		func() bool {
			position201, tokenIndex201, thunkPosition201 := position, tokenIndex, thunkPosition
			{
				begin202 := position
				if !matchChar('{') {
					goto l201
				}
				{
					begin = position
					begin203 := position
				l204:
					{
						position205, tokenIndex205, thunkPosition205 := position, tokenIndex, thunkPosition
						{
							position206, tokenIndex206, thunkPosition206 := position, tokenIndex, thunkPosition
							if !matchChar('}') {
								goto l206
							}
							goto l205
						l206:
							position, tokenIndex, thunkPosition = position206, tokenIndex206, thunkPosition206
						}
						if !matchDot() {
							goto l205
						}
						goto l204
					l205:
						position, tokenIndex, thunkPosition = position205, tokenIndex205, thunkPosition205
					}
					end = position
					if begin203 != position {
						p.Add(RuleAction, begin203, position, tokenIndex)
						tokenIndex++
					}
				}
				if !matchChar('}') {
					goto l201
				}
				if !p.rules[29]() {
					goto l201
				}
				if begin202 != position {
					p.Add(RuleAction, begin202, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l201:
			position, tokenIndex, thunkPosition = position201, tokenIndex201, thunkPosition201
			return false
		},
		/* 35 Begin <- <('<' Spacing)> */
		nil,
		/* 36 End <- <('>' Spacing)> */
		nil,
	}
}
