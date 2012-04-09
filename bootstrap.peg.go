package main

import (
	/*"bytes"*/
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

const END_SYMBOL byte = 0

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
	RuleAction0
	RuleAction1
	RuleAction2
	RuleAction3
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
	RulePegText
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

	RulePre_
	Rule_In_
	Rule_Suf
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
	"Action0",
	"Action1",
	"Action2",
	"Action3",
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
	"PegText",
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

type Peg struct {
	*Tree

	Buffer string
	rules  [85]func() bool
	Parse  func(rule ...int) os.Error
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
	p *Peg
}

func (e *parseError) String() string {
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
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/ )
	}

	return error
}

func (p *Peg) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Peg) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *Peg) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.TokenTree.Tokens() {
		switch token.Rule {
		case RulePegText:
			begin, end = int(token.begin), int(token.end)
		case RuleAction0:
			p.AddPackage(buffer[begin:end])
		case RuleAction1:
			p.AddPeg(buffer[begin:end])
		case RuleAction2:
			p.AddState(buffer[begin:end])
		case RuleAction3:
			p.AddRule(buffer[begin:end])
		case RuleAction4:
			p.AddExpression()
		case RuleAction5:
			p.AddAlternate()
		case RuleAction6:
			p.AddNil()
			p.AddAlternate()
		case RuleAction7:
			p.AddNil()
		case RuleAction8:
			p.AddSequence()
		case RuleAction9:
			p.AddPredicate(buffer[begin:end])
		case RuleAction10:
			p.AddPeekFor()
		case RuleAction11:
			p.AddPeekNot()
		case RuleAction12:
			p.AddQuery()
		case RuleAction13:
			p.AddStar()
		case RuleAction14:
			p.AddPlus()
		case RuleAction15:
			p.AddName(buffer[begin:end])
		case RuleAction16:
			p.AddDot()
		case RuleAction17:
			p.AddAction(buffer[begin:end])
		case RuleAction18:
			p.AddPush()
		case RuleAction19:
			p.AddSequence()
		case RuleAction20:
			p.AddSequence()
		case RuleAction21:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case RuleAction22:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case RuleAction23:
			p.AddAlternate()
		case RuleAction24:
			p.AddAlternate()
		case RuleAction25:
			p.AddRange()
		case RuleAction26:
			p.AddDoubleRange()
		case RuleAction27:
			p.AddCharacter(buffer[begin:end])
		case RuleAction28:
			p.AddDoubleCharacter(buffer[begin:end])
		case RuleAction29:
			p.AddCharacter(buffer[begin:end])
		case RuleAction30:
			p.AddCharacter("\a")
		case RuleAction31:
			p.AddCharacter("\b")
		case RuleAction32:
			p.AddCharacter("\x1B")
		case RuleAction33:
			p.AddCharacter("\f")
		case RuleAction34:
			p.AddCharacter("\n")
		case RuleAction35:
			p.AddCharacter("\r")
		case RuleAction36:
			p.AddCharacter("\t")
		case RuleAction37:
			p.AddCharacter("\v")
		case RuleAction38:
			p.AddCharacter("'")
		case RuleAction39:
			p.AddCharacter("\"")
		case RuleAction40:
			p.AddCharacter("[")
		case RuleAction41:
			p.AddCharacter("]")
		case RuleAction42:
			p.AddCharacter("-")
		case RuleAction43:
			p.AddOctalCharacter(buffer[begin:end])
		case RuleAction44:
			p.AddOctalCharacter(buffer[begin:end])
		case RuleAction45:
			p.AddCharacter("\\")

		}
	}
}

func (p *Peg) Init() {
	if p.Buffer[len(p.Buffer)-1] != END_SYMBOL {
		p.Buffer = p.Buffer + string(END_SYMBOL)
	}

	var tree TokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.Buffer, p.rules

	p.Parse = func(rule ...int) os.Error {
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
		/* 0 Grammar <- <(Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') Spacing Identifier Action0 ('t' 'y' 'p' 'e') Spacing Identifier Action1 ('P' 'e' 'g') Spacing Action Action2 Definition+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !rules[RuleSpacing]() {
					goto l0
				}
				if buffer[position] != 'p' {
					goto l0
				}
				position++
				if buffer[position] != 'a' {
					goto l0
				}
				position++
				if buffer[position] != 'c' {
					goto l0
				}
				position++
				if buffer[position] != 'k' {
					goto l0
				}
				position++
				if buffer[position] != 'a' {
					goto l0
				}
				position++
				if buffer[position] != 'g' {
					goto l0
				}
				position++
				if buffer[position] != 'e' {
					goto l0
				}
				position++
				if !rules[RuleSpacing]() {
					goto l0
				}
				if !rules[RuleIdentifier]() {
					goto l0
				}
				{
					add(RuleAction0, position)
				}
				if buffer[position] != 't' {
					goto l0
				}
				position++
				if buffer[position] != 'y' {
					goto l0
				}
				position++
				if buffer[position] != 'p' {
					goto l0
				}
				position++
				if buffer[position] != 'e' {
					goto l0
				}
				position++
				if !rules[RuleSpacing]() {
					goto l0
				}
				if !rules[RuleIdentifier]() {
					goto l0
				}
				{
					add(RuleAction1, position)
				}
				if buffer[position] != 'P' {
					goto l0
				}
				position++
				if buffer[position] != 'e' {
					goto l0
				}
				position++
				if buffer[position] != 'g' {
					goto l0
				}
				position++
				if !rules[RuleSpacing]() {
					goto l0
				}
				if !rules[RuleAction]() {
					goto l0
				}
				{
					add(RuleAction2, position)
				}
				{
					position7 := position
					depth++
					if !rules[RuleIdentifier]() {
						goto l0
					}
					{
						add(RuleAction3, position)
					}
					if !rules[RuleLeftArrow]() {
						goto l0
					}
					if !rules[RuleExpression]() {
						goto l0
					}
					{
						add(RuleAction4, position)
					}
					{
						position10, tokenIndex10, depth10 := position, tokenIndex, depth
						{
							position11, tokenIndex11, depth11 := position, tokenIndex, depth
							if !rules[RuleIdentifier]() {
								goto l12
							}
							if !rules[RuleLeftArrow]() {
								goto l12
							}
							goto l11
						l12:
							position, tokenIndex, depth = position11, tokenIndex11, depth11
							{
								position13, tokenIndex13, depth13 := position, tokenIndex, depth
								if !matchDot() {
									goto l13
								}
								goto l0
							l13:
								position, tokenIndex, depth = position13, tokenIndex13, depth13
							}
						}
					l11:
						position, tokenIndex, depth = position10, tokenIndex10, depth10
					}
					depth--
					add(RuleDefinition, position7)
				}
			l5:
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					{
						position14 := position
						depth++
						if !rules[RuleIdentifier]() {
							goto l6
						}
						{
							add(RuleAction3, position)
						}
						if !rules[RuleLeftArrow]() {
							goto l6
						}
						if !rules[RuleExpression]() {
							goto l6
						}
						{
							add(RuleAction4, position)
						}
						{
							position17, tokenIndex17, depth17 := position, tokenIndex, depth
							{
								position18, tokenIndex18, depth18 := position, tokenIndex, depth
								if !rules[RuleIdentifier]() {
									goto l19
								}
								if !rules[RuleLeftArrow]() {
									goto l19
								}
								goto l18
							l19:
								position, tokenIndex, depth = position18, tokenIndex18, depth18
								{
									position20, tokenIndex20, depth20 := position, tokenIndex, depth
									if !matchDot() {
										goto l20
									}
									goto l6
								l20:
									position, tokenIndex, depth = position20, tokenIndex20, depth20
								}
							}
						l18:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
						}
						depth--
						add(RuleDefinition, position14)
					}
					goto l5
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
				{
					position21 := position
					depth++
					{
						position22, tokenIndex22, depth22 := position, tokenIndex, depth
						if !matchDot() {
							goto l22
						}
						goto l0
					l22:
						position, tokenIndex, depth = position22, tokenIndex22, depth22
					}
					depth--
					add(RuleEndOfFile, position21)
				}
				depth--
				add(RuleGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Definition <- <(Identifier Action3 LeftArrow Expression Action4 &((Identifier LeftArrow) / !.))> */
		nil,
		/* 2 Expression <- <((Sequence (Slash Sequence Action5)* (Slash Action6)?) / Action7)> */
		func() bool {
			{
				position25 := position
				depth++
				{
					position26, tokenIndex26, depth26 := position, tokenIndex, depth
					if !rules[RuleSequence]() {
						goto l27
					}
				l28:
					{
						position29, tokenIndex29, depth29 := position, tokenIndex, depth
						if !rules[RuleSlash]() {
							goto l29
						}
						if !rules[RuleSequence]() {
							goto l29
						}
						{
							add(RuleAction5, position)
						}
						goto l28
					l29:
						position, tokenIndex, depth = position29, tokenIndex29, depth29
					}
					{
						position31, tokenIndex31, depth31 := position, tokenIndex, depth
						if !rules[RuleSlash]() {
							goto l31
						}
						{
							add(RuleAction6, position)
						}
						goto l32
					l31:
						position, tokenIndex, depth = position31, tokenIndex31, depth31
					}
				l32:
					goto l26
				l27:
					position, tokenIndex, depth = position26, tokenIndex26, depth26
					{
						add(RuleAction7, position)
					}
				}
			l26:
				depth--
				add(RuleExpression, position25)
			}
			return true
		},
		/* 3 Sequence <- <(Prefix (Prefix Action8)*)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				if !rules[RulePrefix]() {
					goto l35
				}
			l37:
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if !rules[RulePrefix]() {
						goto l38
					}
					{
						add(RuleAction8, position)
					}
					goto l37
				l38:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
				}
				depth--
				add(RuleSequence, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 4 Prefix <- <((And Action Action9) / ((&('!') (Not Suffix Action11)) | (&('&') (And Suffix Action10)) | (&('"' | '\'' | '(' | '.' | '<' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{') Suffix)))> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !rules[RuleAnd]() {
						goto l43
					}
					if !rules[RuleAction]() {
						goto l43
					}
					{
						add(RuleAction9, position)
					}
					goto l42
				l43:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					{
						switch buffer[position] {
						case '!':
							{
								position46 := position
								depth++
								if buffer[position] != '!' {
									goto l40
								}
								position++
								if !rules[RuleSpacing]() {
									goto l40
								}
								depth--
								add(RuleNot, position46)
							}
							if !rules[RuleSuffix]() {
								goto l40
							}
							{
								add(RuleAction11, position)
							}
							break
						case '&':
							if !rules[RuleAnd]() {
								goto l40
							}
							if !rules[RuleSuffix]() {
								goto l40
							}
							{
								add(RuleAction10, position)
							}
							break
						default:
							if !rules[RuleSuffix]() {
								goto l40
							}
							break
						}
					}

				}
			l42:
				depth--
				add(RulePrefix, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 5 Suffix <- <(Primary ((&('+') (Plus Action14)) | (&('*') (Star Action13)) | (&('?') (Question Action12)))?)> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				{
					position51 := position
					depth++
					{
						switch buffer[position] {
						case '<':
							{
								position53 := position
								depth++
								if buffer[position] != '<' {
									goto l49
								}
								position++
								if !rules[RuleSpacing]() {
									goto l49
								}
								depth--
								add(RuleBegin, position53)
							}
							if !rules[RuleExpression]() {
								goto l49
							}
							{
								position54 := position
								depth++
								if buffer[position] != '>' {
									goto l49
								}
								position++
								if !rules[RuleSpacing]() {
									goto l49
								}
								depth--
								add(RuleEnd, position54)
							}
							{
								add(RuleAction18, position)
							}
							break
						case '{':
							if !rules[RuleAction]() {
								goto l49
							}
							{
								add(RuleAction17, position)
							}
							break
						case '.':
							{
								position57 := position
								depth++
								if buffer[position] != '.' {
									goto l49
								}
								position++
								if !rules[RuleSpacing]() {
									goto l49
								}
								depth--
								add(RuleDot, position57)
							}
							{
								add(RuleAction16, position)
							}
							break
						case '[':
							{
								position59 := position
								depth++
								{
									position60, tokenIndex60, depth60 := position, tokenIndex, depth
									if buffer[position] != '[' {
										goto l61
									}
									position++
									if buffer[position] != '[' {
										goto l61
									}
									position++
									{
										position62, tokenIndex62, depth62 := position, tokenIndex, depth
										{
											position64, tokenIndex64, depth64 := position, tokenIndex, depth
											if buffer[position] != '^' {
												goto l65
											}
											position++
											if !rules[RuleDoubleRanges]() {
												goto l65
											}
											{
												add(RuleAction21, position)
											}
											goto l64
										l65:
											position, tokenIndex, depth = position64, tokenIndex64, depth64
											if !rules[RuleDoubleRanges]() {
												goto l62
											}
										}
									l64:
										goto l63
									l62:
										position, tokenIndex, depth = position62, tokenIndex62, depth62
									}
								l63:
									if buffer[position] != ']' {
										goto l61
									}
									position++
									if buffer[position] != ']' {
										goto l61
									}
									position++
									goto l60
								l61:
									position, tokenIndex, depth = position60, tokenIndex60, depth60
									if buffer[position] != '[' {
										goto l49
									}
									position++
									{
										position67, tokenIndex67, depth67 := position, tokenIndex, depth
										{
											position69, tokenIndex69, depth69 := position, tokenIndex, depth
											if buffer[position] != '^' {
												goto l70
											}
											position++
											if !rules[RuleRanges]() {
												goto l70
											}
											{
												add(RuleAction22, position)
											}
											goto l69
										l70:
											position, tokenIndex, depth = position69, tokenIndex69, depth69
											if !rules[RuleRanges]() {
												goto l67
											}
										}
									l69:
										goto l68
									l67:
										position, tokenIndex, depth = position67, tokenIndex67, depth67
									}
								l68:
									if buffer[position] != ']' {
										goto l49
									}
									position++
								}
							l60:
								if !rules[RuleSpacing]() {
									goto l49
								}
								depth--
								add(RuleClass, position59)
							}
							break
						case '"', '\'':
							{
								position72 := position
								depth++
								{
									position73, tokenIndex73, depth73 := position, tokenIndex, depth
									if buffer[position] != '\'' {
										goto l74
									}
									position++
									{
										position75, tokenIndex75, depth75 := position, tokenIndex, depth
										{
											position77, tokenIndex77, depth77 := position, tokenIndex, depth
											if buffer[position] != '\'' {
												goto l77
											}
											position++
											goto l75
										l77:
											position, tokenIndex, depth = position77, tokenIndex77, depth77
										}
										if !rules[RuleChar]() {
											goto l75
										}
										goto l76
									l75:
										position, tokenIndex, depth = position75, tokenIndex75, depth75
									}
								l76:
								l78:
									{
										position79, tokenIndex79, depth79 := position, tokenIndex, depth
										{
											position80, tokenIndex80, depth80 := position, tokenIndex, depth
											if buffer[position] != '\'' {
												goto l80
											}
											position++
											goto l79
										l80:
											position, tokenIndex, depth = position80, tokenIndex80, depth80
										}
										if !rules[RuleChar]() {
											goto l79
										}
										{
											add(RuleAction19, position)
										}
										goto l78
									l79:
										position, tokenIndex, depth = position79, tokenIndex79, depth79
									}
									if buffer[position] != '\'' {
										goto l74
									}
									position++
									if !rules[RuleSpacing]() {
										goto l74
									}
									goto l73
								l74:
									position, tokenIndex, depth = position73, tokenIndex73, depth73
									if buffer[position] != '"' {
										goto l49
									}
									position++
									{
										position82, tokenIndex82, depth82 := position, tokenIndex, depth
										{
											position84, tokenIndex84, depth84 := position, tokenIndex, depth
											if buffer[position] != '"' {
												goto l84
											}
											position++
											goto l82
										l84:
											position, tokenIndex, depth = position84, tokenIndex84, depth84
										}
										if !rules[RuleDoubleChar]() {
											goto l82
										}
										goto l83
									l82:
										position, tokenIndex, depth = position82, tokenIndex82, depth82
									}
								l83:
								l85:
									{
										position86, tokenIndex86, depth86 := position, tokenIndex, depth
										{
											position87, tokenIndex87, depth87 := position, tokenIndex, depth
											if buffer[position] != '"' {
												goto l87
											}
											position++
											goto l86
										l87:
											position, tokenIndex, depth = position87, tokenIndex87, depth87
										}
										if !rules[RuleDoubleChar]() {
											goto l86
										}
										{
											add(RuleAction20, position)
										}
										goto l85
									l86:
										position, tokenIndex, depth = position86, tokenIndex86, depth86
									}
									if buffer[position] != '"' {
										goto l49
									}
									position++
									if !rules[RuleSpacing]() {
										goto l49
									}
								}
							l73:
								depth--
								add(RuleLiteral, position72)
							}
							break
						case '(':
							{
								position89 := position
								depth++
								if buffer[position] != '(' {
									goto l49
								}
								position++
								if !rules[RuleSpacing]() {
									goto l49
								}
								depth--
								add(RuleOpen, position89)
							}
							if !rules[RuleExpression]() {
								goto l49
							}
							{
								position90 := position
								depth++
								if buffer[position] != ')' {
									goto l49
								}
								position++
								if !rules[RuleSpacing]() {
									goto l49
								}
								depth--
								add(RuleClose, position90)
							}
							break
						default:
							if !rules[RuleIdentifier]() {
								goto l49
							}
							{
								position91, tokenIndex91, depth91 := position, tokenIndex, depth
								if !rules[RuleLeftArrow]() {
									goto l91
								}
								goto l49
							l91:
								position, tokenIndex, depth = position91, tokenIndex91, depth91
							}
							{
								add(RuleAction15, position)
							}
							break
						}
					}

					depth--
					add(RulePrimary, position51)
				}
				{
					position93, tokenIndex93, depth93 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '+':
							{
								position96 := position
								depth++
								if buffer[position] != '+' {
									goto l93
								}
								position++
								if !rules[RuleSpacing]() {
									goto l93
								}
								depth--
								add(RulePlus, position96)
							}
							{
								add(RuleAction14, position)
							}
							break
						case '*':
							{
								position98 := position
								depth++
								if buffer[position] != '*' {
									goto l93
								}
								position++
								if !rules[RuleSpacing]() {
									goto l93
								}
								depth--
								add(RuleStar, position98)
							}
							{
								add(RuleAction13, position)
							}
							break
						default:
							{
								position100 := position
								depth++
								if buffer[position] != '?' {
									goto l93
								}
								position++
								if !rules[RuleSpacing]() {
									goto l93
								}
								depth--
								add(RuleQuestion, position100)
							}
							{
								add(RuleAction12, position)
							}
							break
						}
					}

					goto l94
				l93:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
				}
			l94:
				depth--
				add(RuleSuffix, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 6 Primary <- <((&('<') (Begin Expression End Action18)) | (&('{') (Action Action17)) | (&('.') (Dot Action16)) | (&('[') Class) | (&('"' | '\'') Literal) | (&('(') (Open Expression Close)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (Identifier !LeftArrow Action15)))> */
		nil,
		/* 7 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				{
					position105 := position
					depth++
					if !rules[RuleIdentStart]() {
						goto l103
					}
				l106:
					{
						position107, tokenIndex107, depth107 := position, tokenIndex, depth
						{
							position108 := position
							depth++
							{
								position109, tokenIndex109, depth109 := position, tokenIndex, depth
								if !rules[RuleIdentStart]() {
									goto l110
								}
								goto l109
							l110:
								position, tokenIndex, depth = position109, tokenIndex109, depth109
								if c := buffer[position]; c < '0' || c > '9' {
									goto l107
								}
								position++
							}
						l109:
							depth--
							add(RuleIdentCont, position108)
						}
						goto l106
					l107:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
					}
					depth--
					add(RulePegText, position105)
				}
				if !rules[RuleSpacing]() {
					goto l103
				}
				depth--
				add(RuleIdentifier, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 8 IdentStart <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != '_' {
							goto l111
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < 'A' || c > 'Z' {
							goto l111
						}
						position++
						break
					default:
						if c := buffer[position]; c < 'a' || c > 'z' {
							goto l111
						}
						position++
						break
					}
				}

				depth--
				add(RuleIdentStart, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 9 IdentCont <- <(IdentStart / [0-9])> */
		nil,
		/* 10 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char Action19)* '\'' Spacing) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar Action20)* '"' Spacing))> */
		nil,
		/* 11 Class <- <((('[' '[' (('^' DoubleRanges Action21) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges Action22) / Ranges)? ']')) Spacing)> */
		nil,
		/* 12 Ranges <- <(!']' Range (!']' Range Action23)*)> */
		func() bool {
			position117, tokenIndex117, depth117 := position, tokenIndex, depth
			{
				position118 := position
				depth++
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if buffer[position] != ']' {
						goto l119
					}
					position++
					goto l117
				l119:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
				}
				if !rules[RuleRange]() {
					goto l117
				}
			l120:
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					{
						position122, tokenIndex122, depth122 := position, tokenIndex, depth
						if buffer[position] != ']' {
							goto l122
						}
						position++
						goto l121
					l122:
						position, tokenIndex, depth = position122, tokenIndex122, depth122
					}
					if !rules[RuleRange]() {
						goto l121
					}
					{
						add(RuleAction23, position)
					}
					goto l120
				l121:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
				}
				depth--
				add(RuleRanges, position118)
			}
			return true
		l117:
			position, tokenIndex, depth = position117, tokenIndex117, depth117
			return false
		},
		/* 13 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange Action24)*)> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				{
					position126, tokenIndex126, depth126 := position, tokenIndex, depth
					if buffer[position] != ']' {
						goto l126
					}
					position++
					if buffer[position] != ']' {
						goto l126
					}
					position++
					goto l124
				l126:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
				}
				if !rules[RuleDoubleRange]() {
					goto l124
				}
			l127:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					{
						position129, tokenIndex129, depth129 := position, tokenIndex, depth
						if buffer[position] != ']' {
							goto l129
						}
						position++
						if buffer[position] != ']' {
							goto l129
						}
						position++
						goto l128
					l129:
						position, tokenIndex, depth = position129, tokenIndex129, depth129
					}
					if !rules[RuleDoubleRange]() {
						goto l128
					}
					{
						add(RuleAction24, position)
					}
					goto l127
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
				depth--
				add(RuleDoubleRanges, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 14 Range <- <((Char '-' Char Action25) / Char)> */
		func() bool {
			position131, tokenIndex131, depth131 := position, tokenIndex, depth
			{
				position132 := position
				depth++
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					if !rules[RuleChar]() {
						goto l134
					}
					if buffer[position] != '-' {
						goto l134
					}
					position++
					if !rules[RuleChar]() {
						goto l134
					}
					{
						add(RuleAction25, position)
					}
					goto l133
				l134:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !rules[RuleChar]() {
						goto l131
					}
				}
			l133:
				depth--
				add(RuleRange, position132)
			}
			return true
		l131:
			position, tokenIndex, depth = position131, tokenIndex131, depth131
			return false
		},
		/* 15 DoubleRange <- <((Char '-' Char Action26) / DoubleChar)> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if !rules[RuleChar]() {
						goto l139
					}
					if buffer[position] != '-' {
						goto l139
					}
					position++
					if !rules[RuleChar]() {
						goto l139
					}
					{
						add(RuleAction26, position)
					}
					goto l138
				l139:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
					if !rules[RuleDoubleChar]() {
						goto l136
					}
				}
			l138:
				depth--
				add(RuleDoubleRange, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 16 Char <- <(Escape / (!'\\' <.> Action27))> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if !rules[RuleEscape]() {
						goto l144
					}
					goto l143
				l144:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					{
						position145, tokenIndex145, depth145 := position, tokenIndex, depth
						if buffer[position] != '\\' {
							goto l145
						}
						position++
						goto l141
					l145:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
					}
					{
						position146 := position
						depth++
						if !matchDot() {
							goto l141
						}
						depth--
						add(RulePegText, position146)
					}
					{
						add(RuleAction27, position)
					}
				}
			l143:
				depth--
				add(RuleChar, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 17 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> Action28) / (!'\\' <.> Action29))> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !rules[RuleEscape]() {
						goto l151
					}
					goto l150
				l151:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					{
						position153 := position
						depth++
						{
							position154, tokenIndex154, depth154 := position, tokenIndex, depth
							if c := buffer[position]; c < 'a' || c > 'z' {
								goto l155
							}
							position++
							goto l154
						l155:
							position, tokenIndex, depth = position154, tokenIndex154, depth154
							if c := buffer[position]; c < 'A' || c > 'Z' {
								goto l152
							}
							position++
						}
					l154:
						depth--
						add(RulePegText, position153)
					}
					{
						add(RuleAction28, position)
					}
					goto l150
				l152:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					{
						position157, tokenIndex157, depth157 := position, tokenIndex, depth
						if buffer[position] != '\\' {
							goto l157
						}
						position++
						goto l148
					l157:
						position, tokenIndex, depth = position157, tokenIndex157, depth157
					}
					{
						position158 := position
						depth++
						if !matchDot() {
							goto l148
						}
						depth--
						add(RulePegText, position158)
					}
					{
						add(RuleAction29, position)
					}
				}
			l150:
				depth--
				add(RuleDoubleChar, position149)
			}
			return true
		l148:
			position, tokenIndex, depth = position148, tokenIndex148, depth148
			return false
		},
		/* 18 Escape <- <(('\\' ('a' / 'A') Action30) / ('\\' ('b' / 'B') Action31) / ('\\' ('e' / 'E') Action32) / ('\\' ('f' / 'F') Action33) / ('\\' ('n' / 'N') Action34) / ('\\' ('r' / 'R') Action35) / ('\\' ('t' / 'T') Action36) / ('\\' ('v' / 'V') Action37) / ('\\' '\'' Action38) / ('\\' '"' Action39) / ('\\' '[' Action40) / ('\\' ']' Action41) / ('\\' '-' Action42) / ('\\' <([0-3] [0-7] [0-7])> Action43) / ('\\' <([0-7] [0-7]?)> Action44) / ('\\' '\\' Action45))> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if buffer[position] != '\\' {
						goto l163
					}
					position++
					{
						position164, tokenIndex164, depth164 := position, tokenIndex, depth
						if buffer[position] != 'a' {
							goto l165
						}
						position++
						goto l164
					l165:
						position, tokenIndex, depth = position164, tokenIndex164, depth164
						if buffer[position] != 'A' {
							goto l163
						}
						position++
					}
				l164:
					{
						add(RuleAction30, position)
					}
					goto l162
				l163:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l167
					}
					position++
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						if buffer[position] != 'b' {
							goto l169
						}
						position++
						goto l168
					l169:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
						if buffer[position] != 'B' {
							goto l167
						}
						position++
					}
				l168:
					{
						add(RuleAction31, position)
					}
					goto l162
				l167:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l171
					}
					position++
					{
						position172, tokenIndex172, depth172 := position, tokenIndex, depth
						if buffer[position] != 'e' {
							goto l173
						}
						position++
						goto l172
					l173:
						position, tokenIndex, depth = position172, tokenIndex172, depth172
						if buffer[position] != 'E' {
							goto l171
						}
						position++
					}
				l172:
					{
						add(RuleAction32, position)
					}
					goto l162
				l171:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l175
					}
					position++
					{
						position176, tokenIndex176, depth176 := position, tokenIndex, depth
						if buffer[position] != 'f' {
							goto l177
						}
						position++
						goto l176
					l177:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
						if buffer[position] != 'F' {
							goto l175
						}
						position++
					}
				l176:
					{
						add(RuleAction33, position)
					}
					goto l162
				l175:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l179
					}
					position++
					{
						position180, tokenIndex180, depth180 := position, tokenIndex, depth
						if buffer[position] != 'n' {
							goto l181
						}
						position++
						goto l180
					l181:
						position, tokenIndex, depth = position180, tokenIndex180, depth180
						if buffer[position] != 'N' {
							goto l179
						}
						position++
					}
				l180:
					{
						add(RuleAction34, position)
					}
					goto l162
				l179:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l183
					}
					position++
					{
						position184, tokenIndex184, depth184 := position, tokenIndex, depth
						if buffer[position] != 'r' {
							goto l185
						}
						position++
						goto l184
					l185:
						position, tokenIndex, depth = position184, tokenIndex184, depth184
						if buffer[position] != 'R' {
							goto l183
						}
						position++
					}
				l184:
					{
						add(RuleAction35, position)
					}
					goto l162
				l183:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l187
					}
					position++
					{
						position188, tokenIndex188, depth188 := position, tokenIndex, depth
						if buffer[position] != 't' {
							goto l189
						}
						position++
						goto l188
					l189:
						position, tokenIndex, depth = position188, tokenIndex188, depth188
						if buffer[position] != 'T' {
							goto l187
						}
						position++
					}
				l188:
					{
						add(RuleAction36, position)
					}
					goto l162
				l187:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l191
					}
					position++
					{
						position192, tokenIndex192, depth192 := position, tokenIndex, depth
						if buffer[position] != 'v' {
							goto l193
						}
						position++
						goto l192
					l193:
						position, tokenIndex, depth = position192, tokenIndex192, depth192
						if buffer[position] != 'V' {
							goto l191
						}
						position++
					}
				l192:
					{
						add(RuleAction37, position)
					}
					goto l162
				l191:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l195
					}
					position++
					if buffer[position] != '\'' {
						goto l195
					}
					position++
					{
						add(RuleAction38, position)
					}
					goto l162
				l195:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l197
					}
					position++
					if buffer[position] != '"' {
						goto l197
					}
					position++
					{
						add(RuleAction39, position)
					}
					goto l162
				l197:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l199
					}
					position++
					if buffer[position] != '[' {
						goto l199
					}
					position++
					{
						add(RuleAction40, position)
					}
					goto l162
				l199:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l201
					}
					position++
					if buffer[position] != ']' {
						goto l201
					}
					position++
					{
						add(RuleAction41, position)
					}
					goto l162
				l201:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l203
					}
					position++
					if buffer[position] != '-' {
						goto l203
					}
					position++
					{
						add(RuleAction42, position)
					}
					goto l162
				l203:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l205
					}
					position++
					{
						position206 := position
						depth++
						if c := buffer[position]; c < '0' || c > '3' {
							goto l205
						}
						position++
						if c := buffer[position]; c < '0' || c > '7' {
							goto l205
						}
						position++
						if c := buffer[position]; c < '0' || c > '7' {
							goto l205
						}
						position++
						depth--
						add(RulePegText, position206)
					}
					{
						add(RuleAction43, position)
					}
					goto l162
				l205:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l208
					}
					position++
					{
						position209 := position
						depth++
						if c := buffer[position]; c < '0' || c > '7' {
							goto l208
						}
						position++
						{
							position210, tokenIndex210, depth210 := position, tokenIndex, depth
							if c := buffer[position]; c < '0' || c > '7' {
								goto l210
							}
							position++
							goto l211
						l210:
							position, tokenIndex, depth = position210, tokenIndex210, depth210
						}
					l211:
						depth--
						add(RulePegText, position209)
					}
					{
						add(RuleAction44, position)
					}
					goto l162
				l208:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if buffer[position] != '\\' {
						goto l160
					}
					position++
					if buffer[position] != '\\' {
						goto l160
					}
					position++
					{
						add(RuleAction45, position)
					}
				}
			l162:
				depth--
				add(RuleEscape, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 19 LeftArrow <- <('<' '-' Spacing)> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				if buffer[position] != '<' {
					goto l214
				}
				position++
				if buffer[position] != '-' {
					goto l214
				}
				position++
				if !rules[RuleSpacing]() {
					goto l214
				}
				depth--
				add(RuleLeftArrow, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 20 Slash <- <('/' Spacing)> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				if buffer[position] != '/' {
					goto l216
				}
				position++
				if !rules[RuleSpacing]() {
					goto l216
				}
				depth--
				add(RuleSlash, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 21 And <- <('&' Spacing)> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				if buffer[position] != '&' {
					goto l218
				}
				position++
				if !rules[RuleSpacing]() {
					goto l218
				}
				depth--
				add(RuleAnd, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
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
				position228 := position
				depth++
			l229:
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						{
							position233 := position
							depth++
							{
								switch buffer[position] {
								case '\t':
									if buffer[position] != '\t' {
										goto l232
									}
									position++
									break
								case ' ':
									if buffer[position] != ' ' {
										goto l232
									}
									position++
									break
								default:
									if !rules[RuleEndOfLine]() {
										goto l232
									}
									break
								}
							}

							depth--
							add(RuleSpace, position233)
						}
						goto l231
					l232:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
						{
							position235 := position
							depth++
							if buffer[position] != '#' {
								goto l230
							}
							position++
						l236:
							{
								position237, tokenIndex237, depth237 := position, tokenIndex, depth
								{
									position238, tokenIndex238, depth238 := position, tokenIndex, depth
									if !rules[RuleEndOfLine]() {
										goto l238
									}
									goto l237
								l238:
									position, tokenIndex, depth = position238, tokenIndex238, depth238
								}
								if !matchDot() {
									goto l237
								}
								goto l236
							l237:
								position, tokenIndex, depth = position237, tokenIndex237, depth237
							}
							if !rules[RuleEndOfLine]() {
								goto l230
							}
							depth--
							add(RuleComment, position235)
						}
					}
				l231:
					goto l229
				l230:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
				}
				depth--
				add(RuleSpacing, position228)
			}
			return true
		},
		/* 30 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		nil,
		/* 31 Space <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine))> */
		nil,
		/* 32 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position241, tokenIndex241, depth241 := position, tokenIndex, depth
			{
				position242 := position
				depth++
				{
					position243, tokenIndex243, depth243 := position, tokenIndex, depth
					if buffer[position] != '\r' {
						goto l244
					}
					position++
					if buffer[position] != '\n' {
						goto l244
					}
					position++
					goto l243
				l244:
					position, tokenIndex, depth = position243, tokenIndex243, depth243
					if buffer[position] != '\n' {
						goto l245
					}
					position++
					goto l243
				l245:
					position, tokenIndex, depth = position243, tokenIndex243, depth243
					if buffer[position] != '\r' {
						goto l241
					}
					position++
				}
			l243:
				depth--
				add(RuleEndOfLine, position242)
			}
			return true
		l241:
			position, tokenIndex, depth = position241, tokenIndex241, depth241
			return false
		},
		/* 33 EndOfFile <- <!.> */
		nil,
		/* 34 Action <- <('{' <(!'}' .)*> '}' Spacing)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				if buffer[position] != '{' {
					goto l247
				}
				position++
				{
					position249 := position
					depth++
				l250:
					{
						position251, tokenIndex251, depth251 := position, tokenIndex, depth
						{
							position252, tokenIndex252, depth252 := position, tokenIndex, depth
							if buffer[position] != '}' {
								goto l252
							}
							position++
							goto l251
						l252:
							position, tokenIndex, depth = position252, tokenIndex252, depth252
						}
						if !matchDot() {
							goto l251
						}
						goto l250
					l251:
						position, tokenIndex, depth = position251, tokenIndex251, depth251
					}
					depth--
					add(RulePegText, position249)
				}
				if buffer[position] != '}' {
					goto l247
				}
				position++
				if !rules[RuleSpacing]() {
					goto l247
				}
				depth--
				add(RuleAction, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 35 Begin <- <('<' Spacing)> */
		nil,
		/* 36 End <- <('>' Spacing)> */
		nil,
		/* 38 Action0 <- <{ p.AddPackage(buffer[begin:end]) }> */
		nil,
		/* 39 Action1 <- <{ p.AddPeg(buffer[begin:end]) }> */
		nil,
		/* 40 Action2 <- <{ p.AddState(buffer[begin:end]) }> */
		nil,
		/* 41 Action3 <- <{ p.AddRule(buffer[begin:end]) }> */
		nil,
		/* 42 Action4 <- <{ p.AddExpression() }> */
		nil,
		/* 43 Action5 <- <{ p.AddAlternate() }> */
		nil,
		/* 44 Action6 <- <{ p.AddNil(); p.AddAlternate() }> */
		nil,
		/* 45 Action7 <- <{ p.AddNil() }> */
		nil,
		/* 46 Action8 <- <{ p.AddSequence() }> */
		nil,
		/* 47 Action9 <- <{ p.AddPredicate(buffer[begin:end]) }> */
		nil,
		/* 48 Action10 <- <{ p.AddPeekFor() }> */
		nil,
		/* 49 Action11 <- <{ p.AddPeekNot() }> */
		nil,
		/* 50 Action12 <- <{ p.AddQuery() }> */
		nil,
		/* 51 Action13 <- <{ p.AddStar() }> */
		nil,
		/* 52 Action14 <- <{ p.AddPlus() }> */
		nil,
		/* 53 Action15 <- <{ p.AddName(buffer[begin:end]) }> */
		nil,
		/* 54 Action16 <- <{ p.AddDot() }> */
		nil,
		/* 55 Action17 <- <{ p.AddAction(buffer[begin:end]) }> */
		nil,
		/* 56 Action18 <- <{ p.AddPush() }> */
		nil,
		nil,
		/* 58 Action19 <- <{ p.AddSequence() }> */
		nil,
		/* 59 Action20 <- <{ p.AddSequence() }> */
		nil,
		/* 60 Action21 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 61 Action22 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 62 Action23 <- <{ p.AddAlternate() }> */
		nil,
		/* 63 Action24 <- <{ p.AddAlternate() }> */
		nil,
		/* 64 Action25 <- <{ p.AddRange() }> */
		nil,
		/* 65 Action26 <- <{ p.AddDoubleRange() }> */
		nil,
		/* 66 Action27 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 67 Action28 <- <{ p.AddDoubleCharacter(buffer[begin:end]) }> */
		nil,
		/* 68 Action29 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 69 Action30 <- <{ p.AddCharacter("\a") }> */
		nil,
		/* 70 Action31 <- <{ p.AddCharacter("\b") }> */
		nil,
		/* 71 Action32 <- <{ p.AddCharacter("\x1B") }> */
		nil,
		/* 72 Action33 <- <{ p.AddCharacter("\f") }> */
		nil,
		/* 73 Action34 <- <{ p.AddCharacter("\n") }> */
		nil,
		/* 74 Action35 <- <{ p.AddCharacter("\r") }> */
		nil,
		/* 75 Action36 <- <{ p.AddCharacter("\t") }> */
		nil,
		/* 76 Action37 <- <{ p.AddCharacter("\v") }> */
		nil,
		/* 77 Action38 <- <{ p.AddCharacter("'") }> */
		nil,
		/* 78 Action39 <- <{ p.AddCharacter("\"") }> */
		nil,
		/* 79 Action40 <- <{ p.AddCharacter("[") }> */
		nil,
		/* 80 Action41 <- <{ p.AddCharacter("]") }> */
		nil,
		/* 81 Action42 <- <{ p.AddCharacter("-") }> */
		nil,
		/* 82 Action43 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 83 Action44 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 84 Action45 <- <{ p.AddCharacter("\\") }> */
		nil,
	}
	p.rules = rules
}
