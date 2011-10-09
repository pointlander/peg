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
	RuleRange
	RuleChar
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
	"Range",
	"Char",
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

type Peg struct {
	*Tree

	Buffer   string
	Min, Max int
	rules    [33]func() bool

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
			p.AddPush()
		},
		/* 41 */
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
		/* 0 Grammar <- <(Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') Spacing Identifier { p.AddPackage(buffer[begin:end]) } ('t' 'y' 'p' 'e') Spacing Identifier { p.AddPeg(buffer[begin:end]) } ('P' 'e' 'g') Spacing Action { p.AddState(buffer[begin:end]) } commit Definition+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0, thunkPosition0 := position, tokenIndex, thunkPosition
			{
				begin1 := position
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
				if !(commit(0)) {
					goto l0
				}
				{
					begin4 := position
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
						position5, tokenIndex5, thunkPosition5 := position, tokenIndex, thunkPosition
						{
							position6, tokenIndex6, thunkPosition6 := position, tokenIndex, thunkPosition
							if !p.rules[7]() {
								goto l7
							}
							if !p.rules[15]() {
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
						do(33)
						if !p.rules[15]() {
							goto l3
						}
						if !p.rules[2]() {
							goto l3
						}
						do(34)
						{
							position10, tokenIndex10, thunkPosition10 := position, tokenIndex, thunkPosition
							{
								position11, tokenIndex11, thunkPosition11 := position, tokenIndex, thunkPosition
								if !p.rules[7]() {
									goto l12
								}
								if !p.rules[15]() {
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
						if !p.rules[16]() {
							goto l22
						}
						if !p.rules[3]() {
							goto l22
						}
						do(27)
						goto l21
					l22:
						position, tokenIndex, thunkPosition = position22, tokenIndex22, thunkPosition22
					}
					{
						position23, tokenIndex23, thunkPosition23 := position, tokenIndex, thunkPosition
						if !p.rules[16]() {
							goto l23
						}
						do(28)
						goto l24
					l23:
						position, tokenIndex, thunkPosition = position23, tokenIndex23, thunkPosition23
					}
				l24:
					goto l19
				l20:
					position, tokenIndex, thunkPosition = position19, tokenIndex19, thunkPosition19
					do(29)
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
					do(41)
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
		/* 4 Prefix <- <((And Action { p.AddPredicate(buffer[begin:end]) }) / (And Suffix { p.AddPeekFor() }) / (Not Suffix { p.AddPeekNot() }) / Suffix)> */
		func() bool {
			position29, tokenIndex29, thunkPosition29 := position, tokenIndex, thunkPosition
			{
				begin30 := position
				{
					position31, tokenIndex31, thunkPosition31 := position, tokenIndex, thunkPosition
					if !p.rules[17]() {
						goto l32
					}
					if !p.rules[30]() {
						goto l32
					}
					do(4)
					goto l31
				l32:
					position, tokenIndex, thunkPosition = position31, tokenIndex31, thunkPosition31
					if !p.rules[17]() {
						goto l33
					}
					if !p.rules[5]() {
						goto l33
					}
					do(5)
					goto l31
				l33:
					position, tokenIndex, thunkPosition = position31, tokenIndex31, thunkPosition31
					{
						begin35 := position
						if !matchChar('!') {
							goto l34
						}
						if !p.rules[25]() {
							goto l34
						}
						if begin35 != position {
							p.Add(RuleNot, begin35, position, tokenIndex)
							tokenIndex++
						}
					}
					if !p.rules[5]() {
						goto l34
					}
					do(6)
					goto l31
				l34:
					position, tokenIndex, thunkPosition = position31, tokenIndex31, thunkPosition31
					if !p.rules[5]() {
						goto l29
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
		/* 5 Suffix <- <(Primary ((Question { p.AddQuery() }) / (Star { p.AddStar() }) / (Plus { p.AddPlus() }))?)> */
		func() bool {
			position36, tokenIndex36, thunkPosition36 := position, tokenIndex, thunkPosition
			{
				begin37 := position
				{
					begin38 := position
					{
						position39, tokenIndex39, thunkPosition39 := position, tokenIndex, thunkPosition
						if !matchChar('c') {
							goto l40
						}
						if !matchChar('o') {
							goto l40
						}
						if !matchChar('m') {
							goto l40
						}
						if !matchChar('m') {
							goto l40
						}
						if !matchChar('i') {
							goto l40
						}
						if !matchChar('t') {
							goto l40
						}
						if !p.rules[25]() {
							goto l40
						}
						do(36)
						goto l39
					l40:
						position, tokenIndex, thunkPosition = position39, tokenIndex39, thunkPosition39
						if !p.rules[7]() {
							goto l41
						}
						{
							position42, tokenIndex42, thunkPosition42 := position, tokenIndex, thunkPosition
							if !p.rules[15]() {
								goto l42
							}
							goto l41
						l42:
							position, tokenIndex, thunkPosition = position42, tokenIndex42, thunkPosition42
						}
						do(37)
						goto l39
					l41:
						position, tokenIndex, thunkPosition = position39, tokenIndex39, thunkPosition39
						{
							begin44 := position
							if !matchChar('(') {
								goto l43
							}
							if !p.rules[25]() {
								goto l43
							}
							if begin44 != position {
								p.Add(RuleOpen, begin44, position, tokenIndex)
								tokenIndex++
							}
						}
						if !p.rules[2]() {
							goto l43
						}
						{
							begin45 := position
							if !matchChar(')') {
								goto l43
							}
							if !p.rules[25]() {
								goto l43
							}
							if begin45 != position {
								p.Add(RuleClose, begin45, position, tokenIndex)
								tokenIndex++
							}
						}
						goto l39
					l43:
						position, tokenIndex, thunkPosition = position39, tokenIndex39, thunkPosition39
						{
							begin47 := position
							{
								position48, tokenIndex48, thunkPosition48 := position, tokenIndex, thunkPosition
								if !matchChar('\'') {
									goto l49
								}
								{
									position50, tokenIndex50, thunkPosition50 := position, tokenIndex, thunkPosition
									{
										position52, tokenIndex52, thunkPosition52 := position, tokenIndex, thunkPosition
										if !matchChar('\'') {
											goto l52
										}
										goto l50
									l52:
										position, tokenIndex, thunkPosition = position52, tokenIndex52, thunkPosition52
									}
									if !p.rules[14]() {
										goto l50
									}
									goto l51
								l50:
									position, tokenIndex, thunkPosition = position50, tokenIndex50, thunkPosition50
								}
							l51:
							l53:
								{
									position54, tokenIndex54, thunkPosition54 := position, tokenIndex, thunkPosition
									{
										position55, tokenIndex55, thunkPosition55 := position, tokenIndex, thunkPosition
										if !matchChar('\'') {
											goto l55
										}
										goto l54
									l55:
										position, tokenIndex, thunkPosition = position55, tokenIndex55, thunkPosition55
									}
									if !p.rules[14]() {
										goto l54
									}
									do(25)
									goto l53
								l54:
									position, tokenIndex, thunkPosition = position54, tokenIndex54, thunkPosition54
								}
								if !matchChar('\'') {
									goto l49
								}
								if !p.rules[25]() {
									goto l49
								}
								goto l48
							l49:
								position, tokenIndex, thunkPosition = position48, tokenIndex48, thunkPosition48
								if !matchChar('"') {
									goto l46
								}
								{
									position56, tokenIndex56, thunkPosition56 := position, tokenIndex, thunkPosition
									{
										position58, tokenIndex58, thunkPosition58 := position, tokenIndex, thunkPosition
										if !matchChar('"') {
											goto l58
										}
										goto l56
									l58:
										position, tokenIndex, thunkPosition = position58, tokenIndex58, thunkPosition58
									}
									if !p.rules[14]() {
										goto l56
									}
									goto l57
								l56:
									position, tokenIndex, thunkPosition = position56, tokenIndex56, thunkPosition56
								}
							l57:
							l59:
								{
									position60, tokenIndex60, thunkPosition60 := position, tokenIndex, thunkPosition
									{
										position61, tokenIndex61, thunkPosition61 := position, tokenIndex, thunkPosition
										if !matchChar('"') {
											goto l61
										}
										goto l60
									l61:
										position, tokenIndex, thunkPosition = position61, tokenIndex61, thunkPosition61
									}
									if !p.rules[14]() {
										goto l60
									}
									do(26)
									goto l59
								l60:
									position, tokenIndex, thunkPosition = position60, tokenIndex60, thunkPosition60
								}
								if !matchChar('"') {
									goto l46
								}
								if !p.rules[25]() {
									goto l46
								}
							}
						l48:
							if begin47 != position {
								p.Add(RuleLiteral, begin47, position, tokenIndex)
								tokenIndex++
							}
						}
						goto l39
					l46:
						position, tokenIndex, thunkPosition = position39, tokenIndex39, thunkPosition39
						{
							begin63 := position
							if !matchChar('[') {
								goto l62
							}
							{
								position64, tokenIndex64, thunkPosition64 := position, tokenIndex, thunkPosition
								{
									position66, tokenIndex66, thunkPosition66 := position, tokenIndex, thunkPosition
									if !matchChar('^') {
										goto l67
									}
									if !p.rules[12]() {
										goto l67
									}
									do(24)
									goto l66
								l67:
									position, tokenIndex, thunkPosition = position66, tokenIndex66, thunkPosition66
									if !p.rules[12]() {
										goto l64
									}
								}
							l66:
								goto l65
							l64:
								position, tokenIndex, thunkPosition = position64, tokenIndex64, thunkPosition64
							}
						l65:
							if !matchChar(']') {
								goto l62
							}
							if !p.rules[25]() {
								goto l62
							}
							if begin63 != position {
								p.Add(RuleClass, begin63, position, tokenIndex)
								tokenIndex++
							}
						}
						goto l39
					l62:
						position, tokenIndex, thunkPosition = position39, tokenIndex39, thunkPosition39
						{
							begin69 := position
							if !matchChar('.') {
								goto l68
							}
							if !p.rules[25]() {
								goto l68
							}
							if begin69 != position {
								p.Add(RuleDot, begin69, position, tokenIndex)
								tokenIndex++
							}
						}
						do(38)
						goto l39
					l68:
						position, tokenIndex, thunkPosition = position39, tokenIndex39, thunkPosition39
						if !p.rules[30]() {
							goto l70
						}
						do(39)
						goto l39
					l70:
						position, tokenIndex, thunkPosition = position39, tokenIndex39, thunkPosition39
						{
							begin71 := position
							if !matchChar('<') {
								goto l36
							}
							if !p.rules[25]() {
								goto l36
							}
							if begin71 != position {
								p.Add(RuleBegin, begin71, position, tokenIndex)
								tokenIndex++
							}
						}
						if !p.rules[2]() {
							goto l36
						}
						{
							begin72 := position
							if !matchChar('>') {
								goto l36
							}
							if !p.rules[25]() {
								goto l36
							}
							if begin72 != position {
								p.Add(RuleEnd, begin72, position, tokenIndex)
								tokenIndex++
							}
						}
						do(40)
					}
				l39:
					if begin38 != position {
						p.Add(RulePrimary, begin38, position, tokenIndex)
						tokenIndex++
					}
				}
				{
					position73, tokenIndex73, thunkPosition73 := position, tokenIndex, thunkPosition
					{
						position75, tokenIndex75, thunkPosition75 := position, tokenIndex, thunkPosition
						{
							begin77 := position
							if !matchChar('?') {
								goto l76
							}
							if !p.rules[25]() {
								goto l76
							}
							if begin77 != position {
								p.Add(RuleQuestion, begin77, position, tokenIndex)
								tokenIndex++
							}
						}
						do(30)
						goto l75
					l76:
						position, tokenIndex, thunkPosition = position75, tokenIndex75, thunkPosition75
						{
							begin79 := position
							if !matchChar('*') {
								goto l78
							}
							if !p.rules[25]() {
								goto l78
							}
							if begin79 != position {
								p.Add(RuleStar, begin79, position, tokenIndex)
								tokenIndex++
							}
						}
						do(31)
						goto l75
					l78:
						position, tokenIndex, thunkPosition = position75, tokenIndex75, thunkPosition75
						{
							begin80 := position
							if !matchChar('+') {
								goto l73
							}
							if !p.rules[25]() {
								goto l73
							}
							if begin80 != position {
								p.Add(RulePlus, begin80, position, tokenIndex)
								tokenIndex++
							}
						}
						do(32)
					}
				l75:
					goto l74
				l73:
					position, tokenIndex, thunkPosition = position73, tokenIndex73, thunkPosition73
				}
			l74:
				if begin37 != position {
					p.Add(RuleSuffix, begin37, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l36:
			position, tokenIndex, thunkPosition = position36, tokenIndex36, thunkPosition36
			return false
		},
		/* 6 Primary <- <(('c' 'o' 'm' 'm' 'i' 't' Spacing { p.AddCommit() }) / (Identifier !LeftArrow { p.AddName(buffer[begin:end]) }) / (Open Expression Close) / Literal / Class / (Dot { p.AddDot() }) / (Action { p.AddAction(buffer[begin:end]) }) / (Begin Expression End { p.AddPush() }))> */
		nil,
		/* 7 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position82, tokenIndex82, thunkPosition82 := position, tokenIndex, thunkPosition
			{
				begin83 := position
				{
					begin = position
					begin84 := position
					if !p.rules[8]() {
						goto l82
					}
				l85:
					{
						position86, tokenIndex86, thunkPosition86 := position, tokenIndex, thunkPosition
						{
							begin87 := position
							{
								position88, tokenIndex88, thunkPosition88 := position, tokenIndex, thunkPosition
								if !p.rules[8]() {
									goto l89
								}
								goto l88
							l89:
								position, tokenIndex, thunkPosition = position88, tokenIndex88, thunkPosition88
								if !matchRange('0', '9') {
									goto l86
								}
							}
						l88:
							if begin87 != position {
								p.Add(RuleIdentCont, begin87, position, tokenIndex)
								tokenIndex++
							}
						}
						goto l85
					l86:
						position, tokenIndex, thunkPosition = position86, tokenIndex86, thunkPosition86
					}
					end = position
					if begin84 != position {
						p.Add(RuleIdentifier, begin84, position, tokenIndex)
						tokenIndex++
					}
				}
				if !p.rules[25]() {
					goto l82
				}
				if begin83 != position {
					p.Add(RuleIdentifier, begin83, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l82:
			position, tokenIndex, thunkPosition = position82, tokenIndex82, thunkPosition82
			return false
		},
		/* 8 IdentStart <- <([a-z] / [A-Z] / '_')> */
		func() bool {
			position90, tokenIndex90, thunkPosition90 := position, tokenIndex, thunkPosition
			{
				begin91 := position
				{
					position92, tokenIndex92, thunkPosition92 := position, tokenIndex, thunkPosition
					if !matchRange('a', 'z') {
						goto l93
					}
					goto l92
				l93:
					position, tokenIndex, thunkPosition = position92, tokenIndex92, thunkPosition92
					if !matchRange('A', 'Z') {
						goto l94
					}
					goto l92
				l94:
					position, tokenIndex, thunkPosition = position92, tokenIndex92, thunkPosition92
					if !matchChar('_') {
						goto l90
					}
				}
			l92:
				if begin91 != position {
					p.Add(RuleIdentStart, begin91, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l90:
			position, tokenIndex, thunkPosition = position90, tokenIndex90, thunkPosition90
			return false
		},
		/* 9 IdentCont <- <(IdentStart / [0-9])> */
		nil,
		/* 10 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char { p.AddSequence() })* '\'' Spacing) / ('"' (!'"' Char)? (!'"' Char { p.AddSequence() })* '"' Spacing))> */
		nil,
		/* 11 Class <- <('[' (('^' Ranges { p.AddPeekNot(); p.AddDot(); p.AddSequence() }) / Ranges)? ']' Spacing)> */
		nil,
		/* 12 Ranges <- <(!']' Range (!']' Range { p.AddAlternate() })*)> */
		func() bool {
			position98, tokenIndex98, thunkPosition98 := position, tokenIndex, thunkPosition
			{
				begin99 := position
				{
					position100, tokenIndex100, thunkPosition100 := position, tokenIndex, thunkPosition
					if !matchChar(']') {
						goto l100
					}
					goto l98
				l100:
					position, tokenIndex, thunkPosition = position100, tokenIndex100, thunkPosition100
				}
				if !p.rules[13]() {
					goto l98
				}
			l101:
				{
					position102, tokenIndex102, thunkPosition102 := position, tokenIndex, thunkPosition
					{
						position103, tokenIndex103, thunkPosition103 := position, tokenIndex, thunkPosition
						if !matchChar(']') {
							goto l103
						}
						goto l102
					l103:
						position, tokenIndex, thunkPosition = position103, tokenIndex103, thunkPosition103
					}
					if !p.rules[13]() {
						goto l102
					}
					do(0)
					goto l101
				l102:
					position, tokenIndex, thunkPosition = position102, tokenIndex102, thunkPosition102
				}
				if begin99 != position {
					p.Add(RuleRanges, begin99, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l98:
			position, tokenIndex, thunkPosition = position98, tokenIndex98, thunkPosition98
			return false
		},
		/* 13 Range <- <((Char '-' Char { p.AddRange() }) / Char)> */
		func() bool {
			position104, tokenIndex104, thunkPosition104 := position, tokenIndex, thunkPosition
			{
				begin105 := position
				{
					position106, tokenIndex106, thunkPosition106 := position, tokenIndex, thunkPosition
					if !p.rules[14]() {
						goto l107
					}
					if !matchChar('-') {
						goto l107
					}
					if !p.rules[14]() {
						goto l107
					}
					do(35)
					goto l106
				l107:
					position, tokenIndex, thunkPosition = position106, tokenIndex106, thunkPosition106
					if !p.rules[14]() {
						goto l104
					}
				}
			l106:
				if begin105 != position {
					p.Add(RuleRange, begin105, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l104:
			position, tokenIndex, thunkPosition = position104, tokenIndex104, thunkPosition104
			return false
		},
		/* 14 Char <- <(('\\' 'a' { p.AddCharacter("\a") }) / ('\\' 'b' { p.AddCharacter("\b") }) / ('\\' 'e' { p.AddCharacter("\x1B") }) / ('\\' 'f' { p.AddCharacter("\f") }) / ('\\' 'n' { p.AddCharacter("\n") }) / ('\\' 'r' { p.AddCharacter("\r") }) / ('\\' 't' { p.AddCharacter("\t") }) / ('\\' 'v' { p.AddCharacter("\v") }) / ('\\' '\'' { p.AddCharacter("'") }) / ('\\' '"' { p.AddCharacter("\"") }) / ('\\' '[' { p.AddCharacter("[") }) / ('\\' ']' { p.AddCharacter("]") }) / ('\\' '-' { p.AddCharacter("-") }) / ('\\' <([0-3] [0-7] [0-7])> { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' <([0-7] [0-7]?)> { p.AddOctalCharacter(buffer[begin:end]) }) / ('\\' '\\' { p.AddCharacter("\\") }) / (!'\\' <.> { p.AddCharacter(buffer[begin:end]) }))> */
		func() bool {
			position108, tokenIndex108, thunkPosition108 := position, tokenIndex, thunkPosition
			{
				begin109 := position
				{
					position110, tokenIndex110, thunkPosition110 := position, tokenIndex, thunkPosition
					if !matchChar('\\') {
						goto l111
					}
					if !matchChar('a') {
						goto l111
					}
					do(7)
					goto l110
				l111:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l112
					}
					if !matchChar('b') {
						goto l112
					}
					do(8)
					goto l110
				l112:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l113
					}
					if !matchChar('e') {
						goto l113
					}
					do(9)
					goto l110
				l113:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l114
					}
					if !matchChar('f') {
						goto l114
					}
					do(10)
					goto l110
				l114:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l115
					}
					if !matchChar('n') {
						goto l115
					}
					do(11)
					goto l110
				l115:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l116
					}
					if !matchChar('r') {
						goto l116
					}
					do(12)
					goto l110
				l116:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l117
					}
					if !matchChar('t') {
						goto l117
					}
					do(13)
					goto l110
				l117:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l118
					}
					if !matchChar('v') {
						goto l118
					}
					do(14)
					goto l110
				l118:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l119
					}
					if !matchChar('\'') {
						goto l119
					}
					do(15)
					goto l110
				l119:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l120
					}
					if !matchChar('"') {
						goto l120
					}
					do(16)
					goto l110
				l120:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l121
					}
					if !matchChar('[') {
						goto l121
					}
					do(17)
					goto l110
				l121:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l122
					}
					if !matchChar(']') {
						goto l122
					}
					do(18)
					goto l110
				l122:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l123
					}
					if !matchChar('-') {
						goto l123
					}
					do(19)
					goto l110
				l123:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l124
					}
					{
						begin = position
						begin125 := position
						if !matchRange('0', '3') {
							goto l124
						}
						if !matchRange('0', '7') {
							goto l124
						}
						if !matchRange('0', '7') {
							goto l124
						}
						end = position
						if begin125 != position {
							p.Add(RuleChar, begin125, position, tokenIndex)
							tokenIndex++
						}
					}
					do(20)
					goto l110
				l124:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l126
					}
					{
						begin = position
						begin127 := position
						if !matchRange('0', '7') {
							goto l126
						}
						{
							position128, tokenIndex128, thunkPosition128 := position, tokenIndex, thunkPosition
							if !matchRange('0', '7') {
								goto l128
							}
							goto l129
						l128:
							position, tokenIndex, thunkPosition = position128, tokenIndex128, thunkPosition128
						}
					l129:
						end = position
						if begin127 != position {
							p.Add(RuleChar, begin127, position, tokenIndex)
							tokenIndex++
						}
					}
					do(21)
					goto l110
				l126:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					if !matchChar('\\') {
						goto l130
					}
					if !matchChar('\\') {
						goto l130
					}
					do(22)
					goto l110
				l130:
					position, tokenIndex, thunkPosition = position110, tokenIndex110, thunkPosition110
					{
						position131, tokenIndex131, thunkPosition131 := position, tokenIndex, thunkPosition
						if !matchChar('\\') {
							goto l131
						}
						goto l108
					l131:
						position, tokenIndex, thunkPosition = position131, tokenIndex131, thunkPosition131
					}
					{
						begin = position
						begin132 := position
						if !matchDot() {
							goto l108
						}
						end = position
						if begin132 != position {
							p.Add(RuleChar, begin132, position, tokenIndex)
							tokenIndex++
						}
					}
					do(23)
				}
			l110:
				if begin109 != position {
					p.Add(RuleChar, begin109, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l108:
			position, tokenIndex, thunkPosition = position108, tokenIndex108, thunkPosition108
			return false
		},
		/* 15 LeftArrow <- <('<' '-' Spacing)> */
		func() bool {
			position133, tokenIndex133, thunkPosition133 := position, tokenIndex, thunkPosition
			{
				begin134 := position
				if !matchChar('<') {
					goto l133
				}
				if !matchChar('-') {
					goto l133
				}
				if !p.rules[25]() {
					goto l133
				}
				if begin134 != position {
					p.Add(RuleLeftArrow, begin134, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l133:
			position, tokenIndex, thunkPosition = position133, tokenIndex133, thunkPosition133
			return false
		},
		/* 16 Slash <- <('/' Spacing)> */
		func() bool {
			position135, tokenIndex135, thunkPosition135 := position, tokenIndex, thunkPosition
			{
				begin136 := position
				if !matchChar('/') {
					goto l135
				}
				if !p.rules[25]() {
					goto l135
				}
				if begin136 != position {
					p.Add(RuleSlash, begin136, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l135:
			position, tokenIndex, thunkPosition = position135, tokenIndex135, thunkPosition135
			return false
		},
		/* 17 And <- <('&' Spacing)> */
		func() bool {
			position137, tokenIndex137, thunkPosition137 := position, tokenIndex, thunkPosition
			{
				begin138 := position
				if !matchChar('&') {
					goto l137
				}
				if !p.rules[25]() {
					goto l137
				}
				if begin138 != position {
					p.Add(RuleAnd, begin138, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l137:
			position, tokenIndex, thunkPosition = position137, tokenIndex137, thunkPosition137
			return false
		},
		/* 18 Not <- <('!' Spacing)> */
		nil,
		/* 19 Question <- <('?' Spacing)> */
		nil,
		/* 20 Star <- <('*' Spacing)> */
		nil,
		/* 21 Plus <- <('+' Spacing)> */
		nil,
		/* 22 Open <- <('(' Spacing)> */
		nil,
		/* 23 Close <- <(')' Spacing)> */
		nil,
		/* 24 Dot <- <('.' Spacing)> */
		nil,
		/* 25 Spacing <- <(Space / Comment)*> */
		func() bool {
			{
				begin147 := position
			l148:
				{
					position149, tokenIndex149, thunkPosition149 := position, tokenIndex, thunkPosition
					{
						position150, tokenIndex150, thunkPosition150 := position, tokenIndex, thunkPosition
						{
							begin152 := position
							{
								position153, tokenIndex153, thunkPosition153 := position, tokenIndex, thunkPosition
								if !matchChar(' ') {
									goto l154
								}
								goto l153
							l154:
								position, tokenIndex, thunkPosition = position153, tokenIndex153, thunkPosition153
								if !matchChar('\t') {
									goto l155
								}
								goto l153
							l155:
								position, tokenIndex, thunkPosition = position153, tokenIndex153, thunkPosition153
								if !p.rules[28]() {
									goto l151
								}
							}
						l153:
							if begin152 != position {
								p.Add(RuleSpace, begin152, position, tokenIndex)
								tokenIndex++
							}
						}
						goto l150
					l151:
						position, tokenIndex, thunkPosition = position150, tokenIndex150, thunkPosition150
						{
							begin156 := position
							if !matchChar('#') {
								goto l149
							}
						l157:
							{
								position158, tokenIndex158, thunkPosition158 := position, tokenIndex, thunkPosition
								{
									position159, tokenIndex159, thunkPosition159 := position, tokenIndex, thunkPosition
									if !p.rules[28]() {
										goto l159
									}
									goto l158
								l159:
									position, tokenIndex, thunkPosition = position159, tokenIndex159, thunkPosition159
								}
								if !matchDot() {
									goto l158
								}
								goto l157
							l158:
								position, tokenIndex, thunkPosition = position158, tokenIndex158, thunkPosition158
							}
							if !p.rules[28]() {
								goto l149
							}
							if begin156 != position {
								p.Add(RuleComment, begin156, position, tokenIndex)
								tokenIndex++
							}
						}
					}
				l150:
					goto l148
				l149:
					position, tokenIndex, thunkPosition = position149, tokenIndex149, thunkPosition149
				}
				if begin147 != position {
					p.Add(RuleSpacing, begin147, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		},
		/* 26 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		nil,
		/* 27 Space <- <(' ' / '\t' / EndOfLine)> */
		nil,
		/* 28 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position162, tokenIndex162, thunkPosition162 := position, tokenIndex, thunkPosition
			{
				begin163 := position
				{
					position164, tokenIndex164, thunkPosition164 := position, tokenIndex, thunkPosition
					if !matchChar('\r') {
						goto l165
					}
					if !matchChar('\n') {
						goto l165
					}
					goto l164
				l165:
					position, tokenIndex, thunkPosition = position164, tokenIndex164, thunkPosition164
					if !matchChar('\n') {
						goto l166
					}
					goto l164
				l166:
					position, tokenIndex, thunkPosition = position164, tokenIndex164, thunkPosition164
					if !matchChar('\r') {
						goto l162
					}
				}
			l164:
				if begin163 != position {
					p.Add(RuleEndOfLine, begin163, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l162:
			position, tokenIndex, thunkPosition = position162, tokenIndex162, thunkPosition162
			return false
		},
		/* 29 EndOfFile <- <!.> */
		nil,
		/* 30 Action <- <('{' <(!'}' .)*> '}' Spacing)> */
		func() bool {
			position168, tokenIndex168, thunkPosition168 := position, tokenIndex, thunkPosition
			{
				begin169 := position
				if !matchChar('{') {
					goto l168
				}
				{
					begin = position
					begin170 := position
				l171:
					{
						position172, tokenIndex172, thunkPosition172 := position, tokenIndex, thunkPosition
						{
							position173, tokenIndex173, thunkPosition173 := position, tokenIndex, thunkPosition
							if !matchChar('}') {
								goto l173
							}
							goto l172
						l173:
							position, tokenIndex, thunkPosition = position173, tokenIndex173, thunkPosition173
						}
						if !matchDot() {
							goto l172
						}
						goto l171
					l172:
						position, tokenIndex, thunkPosition = position172, tokenIndex172, thunkPosition172
					}
					end = position
					if begin170 != position {
						p.Add(RuleAction, begin170, position, tokenIndex)
						tokenIndex++
					}
				}
				if !matchChar('}') {
					goto l168
				}
				if !p.rules[25]() {
					goto l168
				}
				if begin169 != position {
					p.Add(RuleAction, begin169, position, tokenIndex)
					tokenIndex++
				}
			}
			return true
		l168:
			position, tokenIndex, thunkPosition = position168, tokenIndex168, thunkPosition168
			return false
		},
		/* 31 Begin <- <('<' Spacing)> */
		nil,
		/* 32 End <- <('>' Spacing)> */
		nil,
	}
}
