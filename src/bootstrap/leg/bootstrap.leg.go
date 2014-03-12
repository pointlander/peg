package main

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
)

const END_SYMBOL rune = 4

type YYSTYPE int

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	RuleGrammar
	RuleDeclaration
	RuleTrailer
	RuleDefinition
	RuleExpression
	RuleSequence
	RulePrefix
	RuleSuffix
	RulePrimary
	RuleIdentifier
	RuleLiteral
	RuleClass
	RuleRanges
	RuleDoubleRanges
	RuleRange
	RuleDoubleRange
	RuleChar
	RuleDoubleChar
	RuleEscape
	RuleAction
	RuleBraces
	RuleEqual
	RuleColon
	RuleBar
	RuleAnd
	RuleNot
	RuleQuestion
	RuleStar
	RulePlus
	RuleOpen
	RuleClose
	RuleDot
	RuleRPERCENT
	Rule_
	RuleComment
	RuleSpace
	RuleEndOfLine
	RuleEndOfFile
	RuleBegin
	RuleEnd
	RuleAction0
	RuleAction1
	RuleAction2
	RuleAction3
	RulePegText
	RuleAction4
	RuleAction5
	RuleAction6
	RuleAction7
	RuleAction8
	RuleAction9
	RuleAction10
	RuleAction11
	RuleAction12
	RuleAction13
	RuleAction14
	RuleAction15
	RuleAction16
	RuleAction17
	RuleAction18
	RuleAction19
	RuleAction20
	RuleAction21
	RuleAction22
	RuleAction23
	RuleAction24
	RuleAction25
	RuleAction26
	RuleAction27
	RuleAction28
	RuleAction29
	RuleAction30
	RuleAction31
	RuleAction32
	RuleAction33
	RuleAction34
	RuleAction35
	RuleAction36
	RuleAction37
	RuleAction38
	RuleAction39
	RuleAction40
	RuleAction41
	RuleAction42
	RuleAction43
	RuleAction44
	RuleAction45
	RuleAction46
	RuleAction47
	RuleAction48
	RuleAction49
	RuleAction50

	RuleActionPush
	RuleActionPop
	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"Grammar",
	"Declaration",
	"Trailer",
	"Definition",
	"Expression",
	"Sequence",
	"Prefix",
	"Suffix",
	"Primary",
	"Identifier",
	"Literal",
	"Class",
	"Ranges",
	"DoubleRanges",
	"Range",
	"DoubleRange",
	"Char",
	"DoubleChar",
	"Escape",
	"Action",
	"Braces",
	"Equal",
	"Colon",
	"Bar",
	"And",
	"Not",
	"Question",
	"Star",
	"Plus",
	"Open",
	"Close",
	"Dot",
	"RPERCENT",
	"_",
	"Comment",
	"Space",
	"EndOfLine",
	"EndOfFile",
	"Begin",
	"End",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"PegText",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",
	"Action28",
	"Action29",
	"Action30",
	"Action31",
	"Action32",
	"Action33",
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",
	"Action39",
	"Action40",
	"Action41",
	"Action42",
	"Action43",
	"Action44",
	"Action45",
	"Action46",
	"Action47",
	"Action48",
	"Action49",
	"Action50",

	"RuleActionPush",
	"RuleActionPop",
	"Pre_",
	"_In_",
	"_Suf",
}

type TokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule Rule, begin, end, next, depth int)
	Expand(index int) TokenTree
	Tokens() <-chan token32
	Error() []token32
	trim(length int)
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	Rule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) PreOrder() (<-chan State16, [][]token16) {
	s, ordered := make(chan State16, 6), t.Order()
	go func() {
		var states [8]State16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token16{Rule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	Rule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) PreOrder() (<-chan State32, [][]token32) {
	s, ordered := make(chan State32, 6), t.Order()
	go func() {
		var states [8]State32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token32{Rule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.GetToken32()
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

type Leg struct {
	*Tree

	Buffer string
	buffer []rune
	rules  [93]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	TokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *Leg
}

func (e *parseError) Error() string {
	tokens, error := e.p.TokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			Rul3s[token.Rule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *Leg) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Leg) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *Leg) Execute() {
	buffer, begin, end := p.Buffer, 0, 0

	for token := range p.TokenTree.Tokens() {
		switch token.Rule {
		case RulePegText:
			begin, end = int(token.begin), int(token.end)
		case RuleAction0:
			p.AddPackage(buffer[begin:end])
		case RuleAction1:
			p.AddYYSType(buffer[begin:end])
		case RuleAction2:
			p.AddLeg(buffer[begin:end])
		case RuleAction3:
			p.AddState(buffer[begin:end])
		case RuleAction4:
			p.AddDeclaration(buffer[begin:end])
		case RuleAction5:
			p.AddTrailer(buffer[begin:end])
		case RuleAction6:
			p.AddRule(buffer[begin:end])
		case RuleAction7:
			p.AddExpression()
		case RuleAction8:
			p.AddAlternate()
		case RuleAction9:
			p.AddNil()
			p.AddAlternate()
		case RuleAction10:
			p.AddNil()
		case RuleAction11:
			p.AddSequence()
		case RuleAction12:
			p.AddPredicate(buffer[begin:end])
		case RuleAction13:
			p.AddPeekFor()
		case RuleAction14:
			p.AddPeekNot()
		case RuleAction15:
			p.AddQuery()
		case RuleAction16:
			p.AddStar()
		case RuleAction17:
			p.AddPlus()
		case RuleAction18:
			p.AddVariable(buffer[begin:end])
		case RuleAction19:
			p.AddName(buffer[begin:end])
		case RuleAction20:
			p.AddName(buffer[begin:end])
		case RuleAction21:
			p.AddDot()
		case RuleAction22:
			p.AddAction(buffer[begin:end])
		case RuleAction23:
			p.AddPush()
		case RuleAction24:
			p.AddSequence()
		case RuleAction25:
			p.AddSequence()
		case RuleAction26:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case RuleAction27:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case RuleAction28:
			p.AddAlternate()
		case RuleAction29:
			p.AddAlternate()
		case RuleAction30:
			p.AddRange()
		case RuleAction31:
			p.AddDoubleRange()
		case RuleAction32:
			p.AddCharacter(buffer[begin:end])
		case RuleAction33:
			p.AddDoubleCharacter(buffer[begin:end])
		case RuleAction34:
			p.AddCharacter(buffer[begin:end])
		case RuleAction35:
			p.AddCharacter("\a")
		case RuleAction36:
			p.AddCharacter("\b")
		case RuleAction37:
			p.AddCharacter("\x1B")
		case RuleAction38:
			p.AddCharacter("\f")
		case RuleAction39:
			p.AddCharacter("\n")
		case RuleAction40:
			p.AddCharacter("\r")
		case RuleAction41:
			p.AddCharacter("\t")
		case RuleAction42:
			p.AddCharacter("\v")
		case RuleAction43:
			p.AddCharacter("'")
		case RuleAction44:
			p.AddCharacter("\"")
		case RuleAction45:
			p.AddCharacter("[")
		case RuleAction46:
			p.AddCharacter("]")
		case RuleAction47:
			p.AddCharacter("-")
		case RuleAction48:
			p.AddOctalCharacter(buffer[begin:end])
		case RuleAction49:
			p.AddOctalCharacter(buffer[begin:end])
		case RuleAction50:
			p.AddCharacter("\\")

		}
	}
}

func (p *Leg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != END_SYMBOL {
		p.buffer = append(p.buffer, END_SYMBOL)
	}

	var tree TokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.TokenTree = tree
		if matches {
			p.TokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule Rule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != END_SYMBOL {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
	    if buffer[position] == c {
	        position++
	        return true
	    }
	    return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
	    if c := buffer[position]; c >= lower && c <= upper {
	        position++
	        return true
	    }
	    return false
	}*/

	rules = [...]func() bool{
		nil,
		/* 0 Grammar <- <(_ ('p' 'a' 'c' 'k' 'a' 'g' 'e') _ Identifier Action0 ('Y' 'Y' 'S' 'T' 'Y' 'P' 'E') _ Identifier Action1 ('t' 'y' 'p' 'e') _ Identifier Action2 ('P' 'e' 'g') _ Action Action3 (Declaration / Definition)+ Trailer? EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{

				position1 := position
				depth++
				if !rules[Rule_]() {
					goto l0
				}
				if buffer[position] != rune('p') {
					goto l0
				}
				position++
				if buffer[position] != rune('a') {
					goto l0
				}
				position++
				if buffer[position] != rune('c') {
					goto l0
				}
				position++
				if buffer[position] != rune('k') {
					goto l0
				}
				position++
				if buffer[position] != rune('a') {
					goto l0
				}
				position++
				if buffer[position] != rune('g') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !rules[Rule_]() {
					goto l0
				}
				if !rules[RuleIdentifier]() {
					goto l0
				}
				{

					add(RuleAction0, position)
				}
				if buffer[position] != rune('Y') {
					goto l0
				}
				position++
				if buffer[position] != rune('Y') {
					goto l0
				}
				position++
				if buffer[position] != rune('S') {
					goto l0
				}
				position++
				if buffer[position] != rune('T') {
					goto l0
				}
				position++
				if buffer[position] != rune('Y') {
					goto l0
				}
				position++
				if buffer[position] != rune('P') {
					goto l0
				}
				position++
				if buffer[position] != rune('E') {
					goto l0
				}
				position++
				if !rules[Rule_]() {
					goto l0
				}
				if !rules[RuleIdentifier]() {
					goto l0
				}
				{

					add(RuleAction1, position)
				}
				if buffer[position] != rune('t') {
					goto l0
				}
				position++
				if buffer[position] != rune('y') {
					goto l0
				}
				position++
				if buffer[position] != rune('p') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !rules[Rule_]() {
					goto l0
				}
				if !rules[RuleIdentifier]() {
					goto l0
				}
				{

					add(RuleAction2, position)
				}
				if buffer[position] != rune('P') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if buffer[position] != rune('g') {
					goto l0
				}
				position++
				if !rules[Rule_]() {
					goto l0
				}
				if !rules[RuleAction]() {
					goto l0
				}
				{

					add(RuleAction3, position)
				}
				{

					position8, tokenIndex8, depth8 := position, tokenIndex, depth
					{

						position10 := position
						depth++
						{

							position11 := position
							depth++
							if buffer[position] != rune('%') {
								goto l9
							}
							position++
							if buffer[position] != rune('{') {
								goto l9
							}
							position++
							depth--
							add(RulePegText, position11)
						}
						{

							position12 := position
							depth++
						l13:
							{

								position14, tokenIndex14, depth14 := position, tokenIndex, depth
								{

									position15, tokenIndex15, depth15 := position, tokenIndex, depth
									{

										position16 := position
										depth++
										if buffer[position] != rune('%') {
											goto l15
										}
										position++
										if buffer[position] != rune('}') {
											goto l15
										}
										position++
										depth--
										add(RulePegText, position16)
									}
									goto l14
								l15:
									position, tokenIndex, depth = position15, tokenIndex15, depth15
								}
								if !matchDot() {
									goto l14
								}
								goto l13
							l14:
								position, tokenIndex, depth = position14, tokenIndex14, depth14
							}
							depth--
							add(RulePegText, position12)
						}
						{

							position17 := position
							depth++
							if buffer[position] != rune('%') {
								goto l9
							}
							position++
							if buffer[position] != rune('}') {
								goto l9
							}
							position++
							if !rules[Rule_]() {
								goto l9
							}
							depth--
							add(RuleRPERCENT, position17)
						}
						{

							add(RuleAction4, position)
						}
						depth--
						add(RuleDeclaration, position10)
					}
					goto l8
				l9:
					position, tokenIndex, depth = position8, tokenIndex8, depth8
					{

						position19 := position
						depth++
						if !rules[RuleIdentifier]() {
							goto l0
						}
						{

							add(RuleAction6, position)
						}
						if !rules[RuleEqual]() {
							goto l0
						}
						if !rules[RuleExpression]() {
							goto l0
						}
						{

							add(RuleAction7, position)
						}
						depth--
						add(RuleDefinition, position19)
					}
				}
			l8:
			l6:
				{

					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					{

						position22, tokenIndex22, depth22 := position, tokenIndex, depth
						{

							position24 := position
							depth++
							{

								position25 := position
								depth++
								if buffer[position] != rune('%') {
									goto l23
								}
								position++
								if buffer[position] != rune('{') {
									goto l23
								}
								position++
								depth--
								add(RulePegText, position25)
							}
							{

								position26 := position
								depth++
							l27:
								{

									position28, tokenIndex28, depth28 := position, tokenIndex, depth
									{

										position29, tokenIndex29, depth29 := position, tokenIndex, depth
										{

											position30 := position
											depth++
											if buffer[position] != rune('%') {
												goto l29
											}
											position++
											if buffer[position] != rune('}') {
												goto l29
											}
											position++
											depth--
											add(RulePegText, position30)
										}
										goto l28
									l29:
										position, tokenIndex, depth = position29, tokenIndex29, depth29
									}
									if !matchDot() {
										goto l28
									}
									goto l27
								l28:
									position, tokenIndex, depth = position28, tokenIndex28, depth28
								}
								depth--
								add(RulePegText, position26)
							}
							{

								position31 := position
								depth++
								if buffer[position] != rune('%') {
									goto l23
								}
								position++
								if buffer[position] != rune('}') {
									goto l23
								}
								position++
								if !rules[Rule_]() {
									goto l23
								}
								depth--
								add(RuleRPERCENT, position31)
							}
							{

								add(RuleAction4, position)
							}
							depth--
							add(RuleDeclaration, position24)
						}
						goto l22
					l23:
						position, tokenIndex, depth = position22, tokenIndex22, depth22
						{

							position33 := position
							depth++
							if !rules[RuleIdentifier]() {
								goto l7
							}
							{

								add(RuleAction6, position)
							}
							if !rules[RuleEqual]() {
								goto l7
							}
							if !rules[RuleExpression]() {
								goto l7
							}
							{

								add(RuleAction7, position)
							}
							depth--
							add(RuleDefinition, position33)
						}
					}
				l22:
					goto l6
				l7:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
				}
				{

					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					{

						position38 := position
						depth++
						if buffer[position] != rune('%') {
							goto l36
						}
						position++
						if buffer[position] != rune('%') {
							goto l36
						}
						position++
						{

							position39 := position
							depth++
						l40:
							{

								position41, tokenIndex41, depth41 := position, tokenIndex, depth
								if !matchDot() {
									goto l41
								}
								goto l40
							l41:
								position, tokenIndex, depth = position41, tokenIndex41, depth41
							}
							depth--
							add(RulePegText, position39)
						}
						{

							add(RuleAction5, position)
						}
						depth--
						add(RuleTrailer, position38)
					}
					goto l37
				l36:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
				}
			l37:
				{

					position43 := position
					depth++
					{

						position44, tokenIndex44, depth44 := position, tokenIndex, depth
						if !matchDot() {
							goto l44
						}
						goto l0
					l44:
						position, tokenIndex, depth = position44, tokenIndex44, depth44
					}
					depth--
					add(RuleEndOfFile, position43)
				}
				depth--
				add(RuleGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Declaration <- <(<('%' '{')> <(!<('%' '}')> .)*> RPERCENT Action4)> */
		nil,
		/* 2 Trailer <- <('%' '%' (<.*> Action5))> */
		nil,
		/* 3 Definition <- <(Identifier Action6 Equal Expression Action7)> */
		nil,
		/* 4 Expression <- <((Sequence (Bar Sequence Action8)* (Bar Action9)?) / Action10)> */
		func() bool {
			{

				position49 := position
				depth++
				{

					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if !rules[RuleSequence]() {
						goto l51
					}
				l52:
					{

						position53, tokenIndex53, depth53 := position, tokenIndex, depth
						if !rules[RuleBar]() {
							goto l53
						}
						if !rules[RuleSequence]() {
							goto l53
						}
						{

							add(RuleAction8, position)
						}
						goto l52
					l53:
						position, tokenIndex, depth = position53, tokenIndex53, depth53
					}
					{

						position55, tokenIndex55, depth55 := position, tokenIndex, depth
						if !rules[RuleBar]() {
							goto l55
						}
						{

							add(RuleAction9, position)
						}
						goto l56
					l55:
						position, tokenIndex, depth = position55, tokenIndex55, depth55
					}
				l56:
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					{

						add(RuleAction10, position)
					}
				}
			l50:
				depth--
				add(RuleExpression, position49)
			}
			return true
		},
		/* 5 Sequence <- <(Prefix (Prefix Action11)*)> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{

				position60 := position
				depth++
				if !rules[RulePrefix]() {
					goto l59
				}
			l61:
				{

					position62, tokenIndex62, depth62 := position, tokenIndex, depth
					if !rules[RulePrefix]() {
						goto l62
					}
					{

						add(RuleAction11, position)
					}
					goto l61
				l62:
					position, tokenIndex, depth = position62, tokenIndex62, depth62
				}
				depth--
				add(RuleSequence, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 6 Prefix <- <((And Action Action12) / ((&('!') (Not Suffix Action14)) | (&('&') (And Suffix Action13)) | (&('"' | '\'' | '(' | '-' | '.' | '<' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{') Suffix)))> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{

				position65 := position
				depth++
				{

					position66, tokenIndex66, depth66 := position, tokenIndex, depth
					if !rules[RuleAnd]() {
						goto l67
					}
					if !rules[RuleAction]() {
						goto l67
					}
					{

						add(RuleAction12, position)
					}
					goto l66
				l67:
					position, tokenIndex, depth = position66, tokenIndex66, depth66
					{

						switch buffer[position] {
						case '!':
							{

								position70 := position
								depth++
								if buffer[position] != rune('!') {
									goto l64
								}
								position++
								if !rules[Rule_]() {
									goto l64
								}
								depth--
								add(RuleNot, position70)
							}
							if !rules[RuleSuffix]() {
								goto l64
							}
							{

								add(RuleAction14, position)
							}
							break
						case '&':
							if !rules[RuleAnd]() {
								goto l64
							}
							if !rules[RuleSuffix]() {
								goto l64
							}
							{

								add(RuleAction13, position)
							}
							break
						default:
							if !rules[RuleSuffix]() {
								goto l64
							}
							break
						}
					}

				}
			l66:
				depth--
				add(RulePrefix, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 7 Suffix <- <(Primary ((&('+') (Plus Action17)) | (&('*') (Star Action16)) | (&('?') (Question Action15)))?)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{

				position74 := position
				depth++
				{

					position75 := position
					depth++
					{

						position76, tokenIndex76, depth76 := position, tokenIndex, depth
						if !rules[RuleIdentifier]() {
							goto l77
						}
						{

							add(RuleAction18, position)
						}
						{

							position79 := position
							depth++
							if buffer[position] != rune(':') {
								goto l77
							}
							position++
							if !rules[Rule_]() {
								goto l77
							}
							depth--
							add(RuleColon, position79)
						}
						if !rules[RuleIdentifier]() {
							goto l77
						}
						{

							position80, tokenIndex80, depth80 := position, tokenIndex, depth
							if !rules[RuleEqual]() {
								goto l80
							}
							goto l77
						l80:
							position, tokenIndex, depth = position80, tokenIndex80, depth80
						}
						{

							add(RuleAction19, position)
						}
						goto l76
					l77:
						position, tokenIndex, depth = position76, tokenIndex76, depth76
						{

							switch buffer[position] {
							case '<':
								{

									position83 := position
									depth++
									if buffer[position] != rune('<') {
										goto l73
									}
									position++
									if !rules[Rule_]() {
										goto l73
									}
									depth--
									add(RuleBegin, position83)
								}
								if !rules[RuleExpression]() {
									goto l73
								}
								{

									position84 := position
									depth++
									if buffer[position] != rune('>') {
										goto l73
									}
									position++
									if !rules[Rule_]() {
										goto l73
									}
									depth--
									add(RuleEnd, position84)
								}
								{

									add(RuleAction23, position)
								}
								break
							case '{':
								if !rules[RuleAction]() {
									goto l73
								}
								{

									add(RuleAction22, position)
								}
								break
							case '.':
								{

									position87 := position
									depth++
									if buffer[position] != rune('.') {
										goto l73
									}
									position++
									if !rules[Rule_]() {
										goto l73
									}
									depth--
									add(RuleDot, position87)
								}
								{

									add(RuleAction21, position)
								}
								break
							case '[':
								{

									position89 := position
									depth++
									{

										position90, tokenIndex90, depth90 := position, tokenIndex, depth
										if buffer[position] != rune('[') {
											goto l91
										}
										position++
										if buffer[position] != rune('[') {
											goto l91
										}
										position++
										{

											position92, tokenIndex92, depth92 := position, tokenIndex, depth
											{

												position94, tokenIndex94, depth94 := position, tokenIndex, depth
												if buffer[position] != rune('^') {
													goto l95
												}
												position++
												if !rules[RuleDoubleRanges]() {
													goto l95
												}
												{

													add(RuleAction26, position)
												}
												goto l94
											l95:
												position, tokenIndex, depth = position94, tokenIndex94, depth94
												if !rules[RuleDoubleRanges]() {
													goto l92
												}
											}
										l94:
											goto l93
										l92:
											position, tokenIndex, depth = position92, tokenIndex92, depth92
										}
									l93:
										if buffer[position] != rune(']') {
											goto l91
										}
										position++
										if buffer[position] != rune(']') {
											goto l91
										}
										position++
										goto l90
									l91:
										position, tokenIndex, depth = position90, tokenIndex90, depth90
										if buffer[position] != rune('[') {
											goto l73
										}
										position++
										{

											position97, tokenIndex97, depth97 := position, tokenIndex, depth
											{

												position99, tokenIndex99, depth99 := position, tokenIndex, depth
												if buffer[position] != rune('^') {
													goto l100
												}
												position++
												if !rules[RuleRanges]() {
													goto l100
												}
												{

													add(RuleAction27, position)
												}
												goto l99
											l100:
												position, tokenIndex, depth = position99, tokenIndex99, depth99
												if !rules[RuleRanges]() {
													goto l97
												}
											}
										l99:
											goto l98
										l97:
											position, tokenIndex, depth = position97, tokenIndex97, depth97
										}
									l98:
										if buffer[position] != rune(']') {
											goto l73
										}
										position++
									}
								l90:
									if !rules[Rule_]() {
										goto l73
									}
									depth--
									add(RuleClass, position89)
								}
								break
							case '"', '\'':
								{

									position102 := position
									depth++
									{

										position103, tokenIndex103, depth103 := position, tokenIndex, depth
										if buffer[position] != rune('\'') {
											goto l104
										}
										position++
										{

											position105, tokenIndex105, depth105 := position, tokenIndex, depth
											{

												position107, tokenIndex107, depth107 := position, tokenIndex, depth
												if buffer[position] != rune('\'') {
													goto l107
												}
												position++
												goto l105
											l107:
												position, tokenIndex, depth = position107, tokenIndex107, depth107
											}
											if !rules[RuleChar]() {
												goto l105
											}
											goto l106
										l105:
											position, tokenIndex, depth = position105, tokenIndex105, depth105
										}
									l106:
									l108:
										{

											position109, tokenIndex109, depth109 := position, tokenIndex, depth
											{

												position110, tokenIndex110, depth110 := position, tokenIndex, depth
												if buffer[position] != rune('\'') {
													goto l110
												}
												position++
												goto l109
											l110:
												position, tokenIndex, depth = position110, tokenIndex110, depth110
											}
											if !rules[RuleChar]() {
												goto l109
											}
											{

												add(RuleAction24, position)
											}
											goto l108
										l109:
											position, tokenIndex, depth = position109, tokenIndex109, depth109
										}
										if buffer[position] != rune('\'') {
											goto l104
										}
										position++
										if !rules[Rule_]() {
											goto l104
										}
										goto l103
									l104:
										position, tokenIndex, depth = position103, tokenIndex103, depth103
										if buffer[position] != rune('"') {
											goto l73
										}
										position++
										{

											position112, tokenIndex112, depth112 := position, tokenIndex, depth
											{

												position114, tokenIndex114, depth114 := position, tokenIndex, depth
												if buffer[position] != rune('"') {
													goto l114
												}
												position++
												goto l112
											l114:
												position, tokenIndex, depth = position114, tokenIndex114, depth114
											}
											if !rules[RuleDoubleChar]() {
												goto l112
											}
											goto l113
										l112:
											position, tokenIndex, depth = position112, tokenIndex112, depth112
										}
									l113:
									l115:
										{

											position116, tokenIndex116, depth116 := position, tokenIndex, depth
											{

												position117, tokenIndex117, depth117 := position, tokenIndex, depth
												if buffer[position] != rune('"') {
													goto l117
												}
												position++
												goto l116
											l117:
												position, tokenIndex, depth = position117, tokenIndex117, depth117
											}
											if !rules[RuleDoubleChar]() {
												goto l116
											}
											{

												add(RuleAction25, position)
											}
											goto l115
										l116:
											position, tokenIndex, depth = position116, tokenIndex116, depth116
										}
										if buffer[position] != rune('"') {
											goto l73
										}
										position++
										if !rules[Rule_]() {
											goto l73
										}
									}
								l103:
									depth--
									add(RuleLiteral, position102)
								}
								break
							case '(':
								{

									position119 := position
									depth++
									if buffer[position] != rune('(') {
										goto l73
									}
									position++
									if !rules[Rule_]() {
										goto l73
									}
									depth--
									add(RuleOpen, position119)
								}
								if !rules[RuleExpression]() {
									goto l73
								}
								{

									position120 := position
									depth++
									if buffer[position] != rune(')') {
										goto l73
									}
									position++
									if !rules[Rule_]() {
										goto l73
									}
									depth--
									add(RuleClose, position120)
								}
								break
							default:
								if !rules[RuleIdentifier]() {
									goto l73
								}
								{

									position121, tokenIndex121, depth121 := position, tokenIndex, depth
									if !rules[RuleEqual]() {
										goto l121
									}
									goto l73
								l121:
									position, tokenIndex, depth = position121, tokenIndex121, depth121
								}
								{

									add(RuleAction20, position)
								}
								break
							}
						}

					}
				l76:
					depth--
					add(RulePrimary, position75)
				}
				{

					position123, tokenIndex123, depth123 := position, tokenIndex, depth
					{

						switch buffer[position] {
						case '+':
							{

								position126 := position
								depth++
								if buffer[position] != rune('+') {
									goto l123
								}
								position++
								if !rules[Rule_]() {
									goto l123
								}
								depth--
								add(RulePlus, position126)
							}
							{

								add(RuleAction17, position)
							}
							break
						case '*':
							{

								position128 := position
								depth++
								if buffer[position] != rune('*') {
									goto l123
								}
								position++
								if !rules[Rule_]() {
									goto l123
								}
								depth--
								add(RuleStar, position128)
							}
							{

								add(RuleAction16, position)
							}
							break
						default:
							{

								position130 := position
								depth++
								if buffer[position] != rune('?') {
									goto l123
								}
								position++
								if !rules[Rule_]() {
									goto l123
								}
								depth--
								add(RuleQuestion, position130)
							}
							{

								add(RuleAction15, position)
							}
							break
						}
					}

					goto l124
				l123:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
				}
			l124:
				depth--
				add(RuleSuffix, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 8 Primary <- <((Identifier Action18 Colon Identifier !Equal Action19) / ((&('<') (Begin Expression End Action23)) | (&('{') (Action Action22)) | (&('.') (Dot Action21)) | (&('[') Class) | (&('"' | '\'') Literal) | (&('(') (Open Expression Close)) | (&('-' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (Identifier !Equal Action20))))> */
		nil,
		/* 9 Identifier <- <(<(((&('_') '_') | (&('-') '-') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') ([a-z] / [A-Z]))) ((&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('-') '-') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') ([a-z] / [A-Z])))*)> _)> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{

				position134 := position
				depth++
				{

					position135 := position
					depth++
					{

						switch buffer[position] {
						case '_':
							if buffer[position] != rune('_') {
								goto l133
							}
							position++
							break
						case '-':
							if buffer[position] != rune('-') {
								goto l133
							}
							position++
							break
						default:
							{

								position137, tokenIndex137, depth137 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l138
								}
								position++
								goto l137
							l138:
								position, tokenIndex, depth = position137, tokenIndex137, depth137
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l133
								}
								position++
							}
						l137:
							break
						}
					}

				l139:
					{

						position140, tokenIndex140, depth140 := position, tokenIndex, depth
						{

							switch buffer[position] {
							case '_':
								if buffer[position] != rune('_') {
									goto l140
								}
								position++
								break
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l140
								}
								position++
								break
							case '-':
								if buffer[position] != rune('-') {
									goto l140
								}
								position++
								break
							default:
								{

									position142, tokenIndex142, depth142 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l143
									}
									position++
									goto l142
								l143:
									position, tokenIndex, depth = position142, tokenIndex142, depth142
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l140
									}
									position++
								}
							l142:
								break
							}
						}

						goto l139
					l140:
						position, tokenIndex, depth = position140, tokenIndex140, depth140
					}
					depth--
					add(RulePegText, position135)
				}
				if !rules[Rule_]() {
					goto l133
				}
				depth--
				add(RuleIdentifier, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 10 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char Action24)* '\'' _) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar Action25)* '"' _))> */
		nil,
		/* 11 Class <- <((('[' '[' (('^' DoubleRanges Action26) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges Action27) / Ranges)? ']')) _)> */
		nil,
		/* 12 Ranges <- <(!']' Range (!']' Range Action28)*)> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{

				position147 := position
				depth++
				{

					position148, tokenIndex148, depth148 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l148
					}
					position++
					goto l146
				l148:
					position, tokenIndex, depth = position148, tokenIndex148, depth148
				}
				if !rules[RuleRange]() {
					goto l146
				}
			l149:
				{

					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					{

						position151, tokenIndex151, depth151 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l151
						}
						position++
						goto l150
					l151:
						position, tokenIndex, depth = position151, tokenIndex151, depth151
					}
					if !rules[RuleRange]() {
						goto l150
					}
					{

						add(RuleAction28, position)
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(RuleRanges, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 13 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange Action29)*)> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{

				position154 := position
				depth++
				{

					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l155
					}
					position++
					if buffer[position] != rune(']') {
						goto l155
					}
					position++
					goto l153
				l155:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
				}
				if !rules[RuleDoubleRange]() {
					goto l153
				}
			l156:
				{

					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					{

						position158, tokenIndex158, depth158 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l158
						}
						position++
						if buffer[position] != rune(']') {
							goto l158
						}
						position++
						goto l157
					l158:
						position, tokenIndex, depth = position158, tokenIndex158, depth158
					}
					if !rules[RuleDoubleRange]() {
						goto l157
					}
					{

						add(RuleAction29, position)
					}
					goto l156
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
				depth--
				add(RuleDoubleRanges, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 14 Range <- <((Char '-' Char Action30) / Char)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{

				position161 := position
				depth++
				{

					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if !rules[RuleChar]() {
						goto l163
					}
					if buffer[position] != rune('-') {
						goto l163
					}
					position++
					if !rules[RuleChar]() {
						goto l163
					}
					{

						add(RuleAction30, position)
					}
					goto l162
				l163:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if !rules[RuleChar]() {
						goto l160
					}
				}
			l162:
				depth--
				add(RuleRange, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 15 DoubleRange <- <((Char '-' Char Action31) / DoubleChar)> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{

				position166 := position
				depth++
				{

					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					if !rules[RuleChar]() {
						goto l168
					}
					if buffer[position] != rune('-') {
						goto l168
					}
					position++
					if !rules[RuleChar]() {
						goto l168
					}
					{

						add(RuleAction31, position)
					}
					goto l167
				l168:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
					if !rules[RuleDoubleChar]() {
						goto l165
					}
				}
			l167:
				depth--
				add(RuleDoubleRange, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 16 Char <- <(Escape / (!'\\' <.> Action32))> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{

				position171 := position
				depth++
				{

					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if !rules[RuleEscape]() {
						goto l173
					}
					goto l172
				l173:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					{

						position174, tokenIndex174, depth174 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l174
						}
						position++
						goto l170
					l174:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
					}
					{

						position175 := position
						depth++
						if !matchDot() {
							goto l170
						}
						depth--
						add(RulePegText, position175)
					}
					{

						add(RuleAction32, position)
					}
				}
			l172:
				depth--
				add(RuleChar, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 17 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> Action33) / (!'\\' <.> Action34))> */
		func() bool {
			position177, tokenIndex177, depth177 := position, tokenIndex, depth
			{

				position178 := position
				depth++
				{

					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					if !rules[RuleEscape]() {
						goto l180
					}
					goto l179
				l180:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
					{

						position182 := position
						depth++
						{

							position183, tokenIndex183, depth183 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l184
							}
							position++
							goto l183
						l184:
							position, tokenIndex, depth = position183, tokenIndex183, depth183
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l181
							}
							position++
						}
					l183:
						depth--
						add(RulePegText, position182)
					}
					{

						add(RuleAction33, position)
					}
					goto l179
				l181:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
					{

						position186, tokenIndex186, depth186 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l186
						}
						position++
						goto l177
					l186:
						position, tokenIndex, depth = position186, tokenIndex186, depth186
					}
					{

						position187 := position
						depth++
						if !matchDot() {
							goto l177
						}
						depth--
						add(RulePegText, position187)
					}
					{

						add(RuleAction34, position)
					}
				}
			l179:
				depth--
				add(RuleDoubleChar, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 18 Escape <- <(('\\' ('a' / 'A') Action35) / ('\\' ('b' / 'B') Action36) / ('\\' ('e' / 'E') Action37) / ('\\' ('f' / 'F') Action38) / ('\\' ('n' / 'N') Action39) / ('\\' ('r' / 'R') Action40) / ('\\' ('t' / 'T') Action41) / ('\\' ('v' / 'V') Action42) / ('\\' '\'' Action43) / ('\\' '"' Action44) / ('\\' '[' Action45) / ('\\' ']' Action46) / ('\\' '-' Action47) / ('\\' <([0-3] [0-7] [0-7])> Action48) / ('\\' <([0-7] [0-7]?)> Action49) / ('\\' '\\' Action50))> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{

				position190 := position
				depth++
				{

					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l192
					}
					position++
					{

						position193, tokenIndex193, depth193 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l194
						}
						position++
						goto l193
					l194:
						position, tokenIndex, depth = position193, tokenIndex193, depth193
						if buffer[position] != rune('A') {
							goto l192
						}
						position++
					}
				l193:
					{

						add(RuleAction35, position)
					}
					goto l191
				l192:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l196
					}
					position++
					{

						position197, tokenIndex197, depth197 := position, tokenIndex, depth
						if buffer[position] != rune('b') {
							goto l198
						}
						position++
						goto l197
					l198:
						position, tokenIndex, depth = position197, tokenIndex197, depth197
						if buffer[position] != rune('B') {
							goto l196
						}
						position++
					}
				l197:
					{

						add(RuleAction36, position)
					}
					goto l191
				l196:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l200
					}
					position++
					{

						position201, tokenIndex201, depth201 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l202
						}
						position++
						goto l201
					l202:
						position, tokenIndex, depth = position201, tokenIndex201, depth201
						if buffer[position] != rune('E') {
							goto l200
						}
						position++
					}
				l201:
					{

						add(RuleAction37, position)
					}
					goto l191
				l200:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l204
					}
					position++
					{

						position205, tokenIndex205, depth205 := position, tokenIndex, depth
						if buffer[position] != rune('f') {
							goto l206
						}
						position++
						goto l205
					l206:
						position, tokenIndex, depth = position205, tokenIndex205, depth205
						if buffer[position] != rune('F') {
							goto l204
						}
						position++
					}
				l205:
					{

						add(RuleAction38, position)
					}
					goto l191
				l204:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l208
					}
					position++
					{

						position209, tokenIndex209, depth209 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l210
						}
						position++
						goto l209
					l210:
						position, tokenIndex, depth = position209, tokenIndex209, depth209
						if buffer[position] != rune('N') {
							goto l208
						}
						position++
					}
				l209:
					{

						add(RuleAction39, position)
					}
					goto l191
				l208:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l212
					}
					position++
					{

						position213, tokenIndex213, depth213 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l214
						}
						position++
						goto l213
					l214:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if buffer[position] != rune('R') {
							goto l212
						}
						position++
					}
				l213:
					{

						add(RuleAction40, position)
					}
					goto l191
				l212:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l216
					}
					position++
					{

						position217, tokenIndex217, depth217 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l218
						}
						position++
						goto l217
					l218:
						position, tokenIndex, depth = position217, tokenIndex217, depth217
						if buffer[position] != rune('T') {
							goto l216
						}
						position++
					}
				l217:
					{

						add(RuleAction41, position)
					}
					goto l191
				l216:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l220
					}
					position++
					{

						position221, tokenIndex221, depth221 := position, tokenIndex, depth
						if buffer[position] != rune('v') {
							goto l222
						}
						position++
						goto l221
					l222:
						position, tokenIndex, depth = position221, tokenIndex221, depth221
						if buffer[position] != rune('V') {
							goto l220
						}
						position++
					}
				l221:
					{

						add(RuleAction42, position)
					}
					goto l191
				l220:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l224
					}
					position++
					if buffer[position] != rune('\'') {
						goto l224
					}
					position++
					{

						add(RuleAction43, position)
					}
					goto l191
				l224:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l226
					}
					position++
					if buffer[position] != rune('"') {
						goto l226
					}
					position++
					{

						add(RuleAction44, position)
					}
					goto l191
				l226:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l228
					}
					position++
					if buffer[position] != rune('[') {
						goto l228
					}
					position++
					{

						add(RuleAction45, position)
					}
					goto l191
				l228:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l230
					}
					position++
					if buffer[position] != rune(']') {
						goto l230
					}
					position++
					{

						add(RuleAction46, position)
					}
					goto l191
				l230:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l232
					}
					position++
					if buffer[position] != rune('-') {
						goto l232
					}
					position++
					{

						add(RuleAction47, position)
					}
					goto l191
				l232:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l234
					}
					position++
					{

						position235 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('3') {
							goto l234
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l234
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l234
						}
						position++
						depth--
						add(RulePegText, position235)
					}
					{

						add(RuleAction48, position)
					}
					goto l191
				l234:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l237
					}
					position++
					{

						position238 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l237
						}
						position++
						{

							position239, tokenIndex239, depth239 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('7') {
								goto l239
							}
							position++
							goto l240
						l239:
							position, tokenIndex, depth = position239, tokenIndex239, depth239
						}
					l240:
						depth--
						add(RulePegText, position238)
					}
					{

						add(RuleAction49, position)
					}
					goto l191
				l237:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
					if buffer[position] != rune('\\') {
						goto l189
					}
					position++
					if buffer[position] != rune('\\') {
						goto l189
					}
					position++
					{

						add(RuleAction50, position)
					}
				}
			l191:
				depth--
				add(RuleEscape, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 19 Action <- <('{' <Braces*> '}' _)> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{

				position244 := position
				depth++
				if buffer[position] != rune('{') {
					goto l243
				}
				position++
				{

					position245 := position
					depth++
				l246:
					{

						position247, tokenIndex247, depth247 := position, tokenIndex, depth
						if !rules[RuleBraces]() {
							goto l247
						}
						goto l246
					l247:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
					}
					depth--
					add(RulePegText, position245)
				}
				if buffer[position] != rune('}') {
					goto l243
				}
				position++
				if !rules[Rule_]() {
					goto l243
				}
				depth--
				add(RuleAction, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 20 Braces <- <(('{' Braces* '}') / (!'}' .))> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{

				position249 := position
				depth++
				{

					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if buffer[position] != rune('{') {
						goto l251
					}
					position++
				l252:
					{

						position253, tokenIndex253, depth253 := position, tokenIndex, depth
						if !rules[RuleBraces]() {
							goto l253
						}
						goto l252
					l253:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
					}
					if buffer[position] != rune('}') {
						goto l251
					}
					position++
					goto l250
				l251:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
					{

						position254, tokenIndex254, depth254 := position, tokenIndex, depth
						if buffer[position] != rune('}') {
							goto l254
						}
						position++
						goto l248
					l254:
						position, tokenIndex, depth = position254, tokenIndex254, depth254
					}
					if !matchDot() {
						goto l248
					}
				}
			l250:
				depth--
				add(RuleBraces, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 21 Equal <- <('=' _)> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{

				position256 := position
				depth++
				if buffer[position] != rune('=') {
					goto l255
				}
				position++
				if !rules[Rule_]() {
					goto l255
				}
				depth--
				add(RuleEqual, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 22 Colon <- <(':' _)> */
		nil,
		/* 23 Bar <- <('|' _)> */
		func() bool {
			position258, tokenIndex258, depth258 := position, tokenIndex, depth
			{

				position259 := position
				depth++
				if buffer[position] != rune('|') {
					goto l258
				}
				position++
				if !rules[Rule_]() {
					goto l258
				}
				depth--
				add(RuleBar, position259)
			}
			return true
		l258:
			position, tokenIndex, depth = position258, tokenIndex258, depth258
			return false
		},
		/* 24 And <- <('&' _)> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{

				position261 := position
				depth++
				if buffer[position] != rune('&') {
					goto l260
				}
				position++
				if !rules[Rule_]() {
					goto l260
				}
				depth--
				add(RuleAnd, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 25 Not <- <('!' _)> */
		nil,
		/* 26 Question <- <('?' _)> */
		nil,
		/* 27 Star <- <('*' _)> */
		nil,
		/* 28 Plus <- <('+' _)> */
		nil,
		/* 29 Open <- <('(' _)> */
		nil,
		/* 30 Close <- <(')' _)> */
		nil,
		/* 31 Dot <- <('.' _)> */
		nil,
		/* 32 RPERCENT <- <('%' '}' _)> */
		nil,
		/* 33 _ <- <(Space / Comment)*> */
		func() bool {
			{

				position271 := position
				depth++
			l272:
				{

					position273, tokenIndex273, depth273 := position, tokenIndex, depth
					{

						position274, tokenIndex274, depth274 := position, tokenIndex, depth
						{

							position276 := position
							depth++
							{

								switch buffer[position] {
								case '\t':
									if buffer[position] != rune('\t') {
										goto l275
									}
									position++
									break
								case ' ':
									if buffer[position] != rune(' ') {
										goto l275
									}
									position++
									break
								default:
									if !rules[RuleEndOfLine]() {
										goto l275
									}
									break
								}
							}

							depth--
							add(RuleSpace, position276)
						}
						goto l274
					l275:
						position, tokenIndex, depth = position274, tokenIndex274, depth274
						{

							position278 := position
							depth++
							if buffer[position] != rune('#') {
								goto l273
							}
							position++
						l279:
							{

								position280, tokenIndex280, depth280 := position, tokenIndex, depth
								{

									position281, tokenIndex281, depth281 := position, tokenIndex, depth
									if !rules[RuleEndOfLine]() {
										goto l281
									}
									goto l280
								l281:
									position, tokenIndex, depth = position281, tokenIndex281, depth281
								}
								if !matchDot() {
									goto l280
								}
								goto l279
							l280:
								position, tokenIndex, depth = position280, tokenIndex280, depth280
							}
							if !rules[RuleEndOfLine]() {
								goto l273
							}
							depth--
							add(RuleComment, position278)
						}
					}
				l274:
					goto l272
				l273:
					position, tokenIndex, depth = position273, tokenIndex273, depth273
				}
				depth--
				add(Rule_, position271)
			}
			return true
		},
		/* 34 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		nil,
		/* 35 Space <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine))> */
		nil,
		/* 36 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{

				position285 := position
				depth++
				{

					position286, tokenIndex286, depth286 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l287
					}
					position++
					if buffer[position] != rune('\n') {
						goto l287
					}
					position++
					goto l286
				l287:
					position, tokenIndex, depth = position286, tokenIndex286, depth286
					if buffer[position] != rune('\n') {
						goto l288
					}
					position++
					goto l286
				l288:
					position, tokenIndex, depth = position286, tokenIndex286, depth286
					if buffer[position] != rune('\r') {
						goto l284
					}
					position++
				}
			l286:
				depth--
				add(RuleEndOfLine, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 37 EndOfFile <- <!.> */
		nil,
		/* 38 Begin <- <('<' _)> */
		nil,
		/* 39 End <- <('>' _)> */
		nil,
		/* 41 Action0 <- <{ p.AddPackage(buffer[begin:end]) }> */
		nil,
		/* 42 Action1 <- <{ p.AddYYSType(buffer[begin:end]) }> */
		nil,
		/* 43 Action2 <- <{ p.AddLeg(buffer[begin:end]) }> */
		nil,
		/* 44 Action3 <- <{ p.AddState(buffer[begin:end]) }> */
		nil,
		nil,
		/* 46 Action4 <- <{  p.AddDeclaration(buffer[begin:end])  }> */
		nil,
		/* 47 Action5 <- <{ p.AddTrailer(buffer[begin:end]) }> */
		nil,
		/* 48 Action6 <- <{ p.AddRule(buffer[begin:end]) }> */
		nil,
		/* 49 Action7 <- <{ p.AddExpression() }> */
		nil,
		/* 50 Action8 <- <{ p.AddAlternate() }> */
		nil,
		/* 51 Action9 <- <{ p.AddNil(); p.AddAlternate() }> */
		nil,
		/* 52 Action10 <- <{ p.AddNil() }> */
		nil,
		/* 53 Action11 <- <{ p.AddSequence() }> */
		nil,
		/* 54 Action12 <- <{ p.AddPredicate(buffer[begin:end]) }> */
		nil,
		/* 55 Action13 <- <{ p.AddPeekFor() }> */
		nil,
		/* 56 Action14 <- <{ p.AddPeekNot() }> */
		nil,
		/* 57 Action15 <- <{ p.AddQuery() }> */
		nil,
		/* 58 Action16 <- <{ p.AddStar() }> */
		nil,
		/* 59 Action17 <- <{ p.AddPlus() }> */
		nil,
		/* 60 Action18 <- <{ p.AddVariable(buffer[begin:end]) }> */
		nil,
		/* 61 Action19 <- <{ p.AddName(buffer[begin:end]) }> */
		nil,
		/* 62 Action20 <- <{ p.AddName(buffer[begin:end]) }> */
		nil,
		/* 63 Action21 <- <{ p.AddDot() }> */
		nil,
		/* 64 Action22 <- <{ p.AddAction(buffer[begin:end]) }> */
		nil,
		/* 65 Action23 <- <{ p.AddPush() }> */
		nil,
		/* 66 Action24 <- <{ p.AddSequence() }> */
		nil,
		/* 67 Action25 <- <{ p.AddSequence() }> */
		nil,
		/* 68 Action26 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 69 Action27 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 70 Action28 <- <{ p.AddAlternate() }> */
		nil,
		/* 71 Action29 <- <{ p.AddAlternate() }> */
		nil,
		/* 72 Action30 <- <{ p.AddRange() }> */
		nil,
		/* 73 Action31 <- <{ p.AddDoubleRange() }> */
		nil,
		/* 74 Action32 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 75 Action33 <- <{ p.AddDoubleCharacter(buffer[begin:end]) }> */
		nil,
		/* 76 Action34 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 77 Action35 <- <{ p.AddCharacter("\a") }> */
		nil,
		/* 78 Action36 <- <{ p.AddCharacter("\b") }> */
		nil,
		/* 79 Action37 <- <{ p.AddCharacter("\x1B") }> */
		nil,
		/* 80 Action38 <- <{ p.AddCharacter("\f") }> */
		nil,
		/* 81 Action39 <- <{ p.AddCharacter("\n") }> */
		nil,
		/* 82 Action40 <- <{ p.AddCharacter("\r") }> */
		nil,
		/* 83 Action41 <- <{ p.AddCharacter("\t") }> */
		nil,
		/* 84 Action42 <- <{ p.AddCharacter("\v") }> */
		nil,
		/* 85 Action43 <- <{ p.AddCharacter("'") }> */
		nil,
		/* 86 Action44 <- <{ p.AddCharacter("\"") }> */
		nil,
		/* 87 Action45 <- <{ p.AddCharacter("[") }> */
		nil,
		/* 88 Action46 <- <{ p.AddCharacter("]") }> */
		nil,
		/* 89 Action47 <- <{ p.AddCharacter("-") }> */
		nil,
		/* 90 Action48 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 91 Action49 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 92 Action50 <- <{ p.AddCharacter("\\") }> */
		nil,
	}
	p.rules = rules
}
