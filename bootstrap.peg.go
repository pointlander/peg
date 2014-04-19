package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const END_SYMBOL rune = 4

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	RuleGrammar
	RuleImport
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
	RulePegText
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

	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"Grammar",
	"Import",
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
	"PegText",
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

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule Rule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *Node32
	Error() []token32
	trim(length int)
}

type Node32 struct {
	token32
	up, next *Node32
}

func (node *Node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[node.Rule], strconv.Quote(buffer[node.begin:node.end]))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *Node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *Node32
	down *element
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

func (t *tokens16) AST() *Node32 {
	tokens := t.Tokens()
	stack := &element{node: &Node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &Node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
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

func (t *tokens32) AST() *Node32 {
	tokens := t.Tokens()
	stack := &element{node: &Node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &Node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
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

func (t *tokens16) Expand(index int) tokenTree {
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

func (t *tokens32) Expand(index int) tokenTree {
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
	buffer []rune
	rules  [87]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
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

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
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

func (p *Peg) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Peg) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *Peg) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
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
			p.AddImport(buffer[begin:end])
		case RuleAction4:
			p.AddRule(buffer[begin:end])
		case RuleAction5:
			p.AddExpression()
		case RuleAction6:
			p.AddAlternate()
		case RuleAction7:
			p.AddNil()
			p.AddAlternate()
		case RuleAction8:
			p.AddNil()
		case RuleAction9:
			p.AddSequence()
		case RuleAction10:
			p.AddPredicate(buffer[begin:end])
		case RuleAction11:
			p.AddPeekFor()
		case RuleAction12:
			p.AddPeekNot()
		case RuleAction13:
			p.AddQuery()
		case RuleAction14:
			p.AddStar()
		case RuleAction15:
			p.AddPlus()
		case RuleAction16:
			p.AddName(buffer[begin:end])
		case RuleAction17:
			p.AddDot()
		case RuleAction18:
			p.AddAction(buffer[begin:end])
		case RuleAction19:
			p.AddPush()
		case RuleAction20:
			p.AddSequence()
		case RuleAction21:
			p.AddSequence()
		case RuleAction22:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case RuleAction23:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case RuleAction24:
			p.AddAlternate()
		case RuleAction25:
			p.AddAlternate()
		case RuleAction26:
			p.AddRange()
		case RuleAction27:
			p.AddDoubleRange()
		case RuleAction28:
			p.AddCharacter(buffer[begin:end])
		case RuleAction29:
			p.AddDoubleCharacter(buffer[begin:end])
		case RuleAction30:
			p.AddCharacter(buffer[begin:end])
		case RuleAction31:
			p.AddCharacter("\a")
		case RuleAction32:
			p.AddCharacter("\b")
		case RuleAction33:
			p.AddCharacter("\x1B")
		case RuleAction34:
			p.AddCharacter("\f")
		case RuleAction35:
			p.AddCharacter("\n")
		case RuleAction36:
			p.AddCharacter("\r")
		case RuleAction37:
			p.AddCharacter("\t")
		case RuleAction38:
			p.AddCharacter("\v")
		case RuleAction39:
			p.AddCharacter("'")
		case RuleAction40:
			p.AddCharacter("\"")
		case RuleAction41:
			p.AddCharacter("[")
		case RuleAction42:
			p.AddCharacter("]")
		case RuleAction43:
			p.AddCharacter("-")
		case RuleAction44:
			p.AddOctalCharacter(buffer[begin:end])
		case RuleAction45:
			p.AddOctalCharacter(buffer[begin:end])
		case RuleAction46:
			p.AddCharacter("\\")

		}
	}
}

func (p *Peg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != END_SYMBOL {
		p.buffer = append(p.buffer, END_SYMBOL)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
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
		/* 0 Grammar <- <(Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') Spacing Identifier Action0 Import* ('t' 'y' 'p' 'e') Spacing Identifier Action1 ('P' 'e' 'g') Spacing Action Action2 Definition+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !rules[RuleSpacing]() {
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
				if !rules[RuleSpacing]() {
					goto l0
				}
				if !rules[RuleIdentifier]() {
					goto l0
				}
				{
					add(RuleAction0, position)
				}
			l3:
				{
					position4, tokenIndex4, depth4 := position, tokenIndex, depth
					{
						position5 := position
						depth++
						if buffer[position] != rune('i') {
							goto l4
						}
						position++
						if buffer[position] != rune('m') {
							goto l4
						}
						position++
						if buffer[position] != rune('p') {
							goto l4
						}
						position++
						if buffer[position] != rune('o') {
							goto l4
						}
						position++
						if buffer[position] != rune('r') {
							goto l4
						}
						position++
						if buffer[position] != rune('t') {
							goto l4
						}
						position++
						if !rules[RuleSpacing]() {
							goto l4
						}
						if buffer[position] != rune('"') {
							goto l4
						}
						position++
						{
							position6 := position
							depth++
							{
								switch buffer[position] {
								case '-':
									if buffer[position] != rune('-') {
										goto l4
									}
									position++
									break
								case '.':
									if buffer[position] != rune('.') {
										goto l4
									}
									position++
									break
								case '/':
									if buffer[position] != rune('/') {
										goto l4
									}
									position++
									break
								case '_':
									if buffer[position] != rune('_') {
										goto l4
									}
									position++
									break
								case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l4
									}
									position++
									break
								default:
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l4
									}
									position++
									break
								}
							}

						l7:
							{
								position8, tokenIndex8, depth8 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '-':
										if buffer[position] != rune('-') {
											goto l8
										}
										position++
										break
									case '.':
										if buffer[position] != rune('.') {
											goto l8
										}
										position++
										break
									case '/':
										if buffer[position] != rune('/') {
											goto l8
										}
										position++
										break
									case '_':
										if buffer[position] != rune('_') {
											goto l8
										}
										position++
										break
									case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
										if c := buffer[position]; c < rune('A') || c > rune('Z') {
											goto l8
										}
										position++
										break
									default:
										if c := buffer[position]; c < rune('a') || c > rune('z') {
											goto l8
										}
										position++
										break
									}
								}

								goto l7
							l8:
								position, tokenIndex, depth = position8, tokenIndex8, depth8
							}
							depth--
							add(RulePegText, position6)
						}
						if buffer[position] != rune('"') {
							goto l4
						}
						position++
						if !rules[RuleSpacing]() {
							goto l4
						}
						{
							add(RuleAction3, position)
						}
						depth--
						add(RuleImport, position5)
					}
					goto l3
				l4:
					position, tokenIndex, depth = position4, tokenIndex4, depth4
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
				if !rules[RuleSpacing]() {
					goto l0
				}
				if !rules[RuleIdentifier]() {
					goto l0
				}
				{
					add(RuleAction1, position)
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
					position16 := position
					depth++
					if !rules[RuleIdentifier]() {
						goto l0
					}
					{
						add(RuleAction4, position)
					}
					if !rules[RuleLeftArrow]() {
						goto l0
					}
					if !rules[RuleExpression]() {
						goto l0
					}
					{
						add(RuleAction5, position)
					}
					{
						position19, tokenIndex19, depth19 := position, tokenIndex, depth
						{
							position20, tokenIndex20, depth20 := position, tokenIndex, depth
							if !rules[RuleIdentifier]() {
								goto l21
							}
							if !rules[RuleLeftArrow]() {
								goto l21
							}
							goto l20
						l21:
							position, tokenIndex, depth = position20, tokenIndex20, depth20
							{
								position22, tokenIndex22, depth22 := position, tokenIndex, depth
								if !matchDot() {
									goto l22
								}
								goto l0
							l22:
								position, tokenIndex, depth = position22, tokenIndex22, depth22
							}
						}
					l20:
						position, tokenIndex, depth = position19, tokenIndex19, depth19
					}
					depth--
					add(RuleDefinition, position16)
				}
			l14:
				{
					position15, tokenIndex15, depth15 := position, tokenIndex, depth
					{
						position23 := position
						depth++
						if !rules[RuleIdentifier]() {
							goto l15
						}
						{
							add(RuleAction4, position)
						}
						if !rules[RuleLeftArrow]() {
							goto l15
						}
						if !rules[RuleExpression]() {
							goto l15
						}
						{
							add(RuleAction5, position)
						}
						{
							position26, tokenIndex26, depth26 := position, tokenIndex, depth
							{
								position27, tokenIndex27, depth27 := position, tokenIndex, depth
								if !rules[RuleIdentifier]() {
									goto l28
								}
								if !rules[RuleLeftArrow]() {
									goto l28
								}
								goto l27
							l28:
								position, tokenIndex, depth = position27, tokenIndex27, depth27
								{
									position29, tokenIndex29, depth29 := position, tokenIndex, depth
									if !matchDot() {
										goto l29
									}
									goto l15
								l29:
									position, tokenIndex, depth = position29, tokenIndex29, depth29
								}
							}
						l27:
							position, tokenIndex, depth = position26, tokenIndex26, depth26
						}
						depth--
						add(RuleDefinition, position23)
					}
					goto l14
				l15:
					position, tokenIndex, depth = position15, tokenIndex15, depth15
				}
				{
					position30 := position
					depth++
					{
						position31, tokenIndex31, depth31 := position, tokenIndex, depth
						if !matchDot() {
							goto l31
						}
						goto l0
					l31:
						position, tokenIndex, depth = position31, tokenIndex31, depth31
					}
					depth--
					add(RuleEndOfFile, position30)
				}
				depth--
				add(RuleGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Import <- <('i' 'm' 'p' 'o' 'r' 't' Spacing '"' <((&('-') '-') | (&('.') '.') | (&('/') '/') | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> '"' Spacing Action3)> */
		nil,
		/* 2 Definition <- <(Identifier Action4 LeftArrow Expression Action5 &((Identifier LeftArrow) / !.))> */
		nil,
		/* 3 Expression <- <((Sequence (Slash Sequence Action6)* (Slash Action7)?) / Action8)> */
		func() bool {
			{
				position35 := position
				depth++
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					if !rules[RuleSequence]() {
						goto l37
					}
				l38:
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if !rules[RuleSlash]() {
							goto l39
						}
						if !rules[RuleSequence]() {
							goto l39
						}
						{
							add(RuleAction6, position)
						}
						goto l38
					l39:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
					}
					{
						position41, tokenIndex41, depth41 := position, tokenIndex, depth
						if !rules[RuleSlash]() {
							goto l41
						}
						{
							add(RuleAction7, position)
						}
						goto l42
					l41:
						position, tokenIndex, depth = position41, tokenIndex41, depth41
					}
				l42:
					goto l36
				l37:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
					{
						add(RuleAction8, position)
					}
				}
			l36:
				depth--
				add(RuleExpression, position35)
			}
			return true
		},
		/* 4 Sequence <- <(Prefix (Prefix Action9)*)> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if !rules[RulePrefix]() {
					goto l45
				}
			l47:
				{
					position48, tokenIndex48, depth48 := position, tokenIndex, depth
					if !rules[RulePrefix]() {
						goto l48
					}
					{
						add(RuleAction9, position)
					}
					goto l47
				l48:
					position, tokenIndex, depth = position48, tokenIndex48, depth48
				}
				depth--
				add(RuleSequence, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 5 Prefix <- <((And Action Action10) / ((&('!') (Not Suffix Action12)) | (&('&') (And Suffix Action11)) | (&('"' | '\'' | '(' | '.' | '<' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{') Suffix)))> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					if !rules[RuleAnd]() {
						goto l53
					}
					if !rules[RuleAction]() {
						goto l53
					}
					{
						add(RuleAction10, position)
					}
					goto l52
				l53:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
					{
						switch buffer[position] {
						case '!':
							{
								position56 := position
								depth++
								if buffer[position] != rune('!') {
									goto l50
								}
								position++
								if !rules[RuleSpacing]() {
									goto l50
								}
								depth--
								add(RuleNot, position56)
							}
							if !rules[RuleSuffix]() {
								goto l50
							}
							{
								add(RuleAction12, position)
							}
							break
						case '&':
							if !rules[RuleAnd]() {
								goto l50
							}
							if !rules[RuleSuffix]() {
								goto l50
							}
							{
								add(RuleAction11, position)
							}
							break
						default:
							if !rules[RuleSuffix]() {
								goto l50
							}
							break
						}
					}

				}
			l52:
				depth--
				add(RulePrefix, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 6 Suffix <- <(Primary ((&('+') (Plus Action15)) | (&('*') (Star Action14)) | (&('?') (Question Action13)))?)> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
				{
					position61 := position
					depth++
					{
						switch buffer[position] {
						case '<':
							{
								position63 := position
								depth++
								if buffer[position] != rune('<') {
									goto l59
								}
								position++
								if !rules[RuleSpacing]() {
									goto l59
								}
								depth--
								add(RuleBegin, position63)
							}
							if !rules[RuleExpression]() {
								goto l59
							}
							{
								position64 := position
								depth++
								if buffer[position] != rune('>') {
									goto l59
								}
								position++
								if !rules[RuleSpacing]() {
									goto l59
								}
								depth--
								add(RuleEnd, position64)
							}
							{
								add(RuleAction19, position)
							}
							break
						case '{':
							if !rules[RuleAction]() {
								goto l59
							}
							{
								add(RuleAction18, position)
							}
							break
						case '.':
							{
								position67 := position
								depth++
								if buffer[position] != rune('.') {
									goto l59
								}
								position++
								if !rules[RuleSpacing]() {
									goto l59
								}
								depth--
								add(RuleDot, position67)
							}
							{
								add(RuleAction17, position)
							}
							break
						case '[':
							{
								position69 := position
								depth++
								{
									position70, tokenIndex70, depth70 := position, tokenIndex, depth
									if buffer[position] != rune('[') {
										goto l71
									}
									position++
									if buffer[position] != rune('[') {
										goto l71
									}
									position++
									{
										position72, tokenIndex72, depth72 := position, tokenIndex, depth
										{
											position74, tokenIndex74, depth74 := position, tokenIndex, depth
											if buffer[position] != rune('^') {
												goto l75
											}
											position++
											if !rules[RuleDoubleRanges]() {
												goto l75
											}
											{
												add(RuleAction22, position)
											}
											goto l74
										l75:
											position, tokenIndex, depth = position74, tokenIndex74, depth74
											if !rules[RuleDoubleRanges]() {
												goto l72
											}
										}
									l74:
										goto l73
									l72:
										position, tokenIndex, depth = position72, tokenIndex72, depth72
									}
								l73:
									if buffer[position] != rune(']') {
										goto l71
									}
									position++
									if buffer[position] != rune(']') {
										goto l71
									}
									position++
									goto l70
								l71:
									position, tokenIndex, depth = position70, tokenIndex70, depth70
									if buffer[position] != rune('[') {
										goto l59
									}
									position++
									{
										position77, tokenIndex77, depth77 := position, tokenIndex, depth
										{
											position79, tokenIndex79, depth79 := position, tokenIndex, depth
											if buffer[position] != rune('^') {
												goto l80
											}
											position++
											if !rules[RuleRanges]() {
												goto l80
											}
											{
												add(RuleAction23, position)
											}
											goto l79
										l80:
											position, tokenIndex, depth = position79, tokenIndex79, depth79
											if !rules[RuleRanges]() {
												goto l77
											}
										}
									l79:
										goto l78
									l77:
										position, tokenIndex, depth = position77, tokenIndex77, depth77
									}
								l78:
									if buffer[position] != rune(']') {
										goto l59
									}
									position++
								}
							l70:
								if !rules[RuleSpacing]() {
									goto l59
								}
								depth--
								add(RuleClass, position69)
							}
							break
						case '"', '\'':
							{
								position82 := position
								depth++
								{
									position83, tokenIndex83, depth83 := position, tokenIndex, depth
									if buffer[position] != rune('\'') {
										goto l84
									}
									position++
									{
										position85, tokenIndex85, depth85 := position, tokenIndex, depth
										{
											position87, tokenIndex87, depth87 := position, tokenIndex, depth
											if buffer[position] != rune('\'') {
												goto l87
											}
											position++
											goto l85
										l87:
											position, tokenIndex, depth = position87, tokenIndex87, depth87
										}
										if !rules[RuleChar]() {
											goto l85
										}
										goto l86
									l85:
										position, tokenIndex, depth = position85, tokenIndex85, depth85
									}
								l86:
								l88:
									{
										position89, tokenIndex89, depth89 := position, tokenIndex, depth
										{
											position90, tokenIndex90, depth90 := position, tokenIndex, depth
											if buffer[position] != rune('\'') {
												goto l90
											}
											position++
											goto l89
										l90:
											position, tokenIndex, depth = position90, tokenIndex90, depth90
										}
										if !rules[RuleChar]() {
											goto l89
										}
										{
											add(RuleAction20, position)
										}
										goto l88
									l89:
										position, tokenIndex, depth = position89, tokenIndex89, depth89
									}
									if buffer[position] != rune('\'') {
										goto l84
									}
									position++
									if !rules[RuleSpacing]() {
										goto l84
									}
									goto l83
								l84:
									position, tokenIndex, depth = position83, tokenIndex83, depth83
									if buffer[position] != rune('"') {
										goto l59
									}
									position++
									{
										position92, tokenIndex92, depth92 := position, tokenIndex, depth
										{
											position94, tokenIndex94, depth94 := position, tokenIndex, depth
											if buffer[position] != rune('"') {
												goto l94
											}
											position++
											goto l92
										l94:
											position, tokenIndex, depth = position94, tokenIndex94, depth94
										}
										if !rules[RuleDoubleChar]() {
											goto l92
										}
										goto l93
									l92:
										position, tokenIndex, depth = position92, tokenIndex92, depth92
									}
								l93:
								l95:
									{
										position96, tokenIndex96, depth96 := position, tokenIndex, depth
										{
											position97, tokenIndex97, depth97 := position, tokenIndex, depth
											if buffer[position] != rune('"') {
												goto l97
											}
											position++
											goto l96
										l97:
											position, tokenIndex, depth = position97, tokenIndex97, depth97
										}
										if !rules[RuleDoubleChar]() {
											goto l96
										}
										{
											add(RuleAction21, position)
										}
										goto l95
									l96:
										position, tokenIndex, depth = position96, tokenIndex96, depth96
									}
									if buffer[position] != rune('"') {
										goto l59
									}
									position++
									if !rules[RuleSpacing]() {
										goto l59
									}
								}
							l83:
								depth--
								add(RuleLiteral, position82)
							}
							break
						case '(':
							{
								position99 := position
								depth++
								if buffer[position] != rune('(') {
									goto l59
								}
								position++
								if !rules[RuleSpacing]() {
									goto l59
								}
								depth--
								add(RuleOpen, position99)
							}
							if !rules[RuleExpression]() {
								goto l59
							}
							{
								position100 := position
								depth++
								if buffer[position] != rune(')') {
									goto l59
								}
								position++
								if !rules[RuleSpacing]() {
									goto l59
								}
								depth--
								add(RuleClose, position100)
							}
							break
						default:
							if !rules[RuleIdentifier]() {
								goto l59
							}
							{
								position101, tokenIndex101, depth101 := position, tokenIndex, depth
								if !rules[RuleLeftArrow]() {
									goto l101
								}
								goto l59
							l101:
								position, tokenIndex, depth = position101, tokenIndex101, depth101
							}
							{
								add(RuleAction16, position)
							}
							break
						}
					}

					depth--
					add(RulePrimary, position61)
				}
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '+':
							{
								position106 := position
								depth++
								if buffer[position] != rune('+') {
									goto l103
								}
								position++
								if !rules[RuleSpacing]() {
									goto l103
								}
								depth--
								add(RulePlus, position106)
							}
							{
								add(RuleAction15, position)
							}
							break
						case '*':
							{
								position108 := position
								depth++
								if buffer[position] != rune('*') {
									goto l103
								}
								position++
								if !rules[RuleSpacing]() {
									goto l103
								}
								depth--
								add(RuleStar, position108)
							}
							{
								add(RuleAction14, position)
							}
							break
						default:
							{
								position110 := position
								depth++
								if buffer[position] != rune('?') {
									goto l103
								}
								position++
								if !rules[RuleSpacing]() {
									goto l103
								}
								depth--
								add(RuleQuestion, position110)
							}
							{
								add(RuleAction13, position)
							}
							break
						}
					}

					goto l104
				l103:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
				}
			l104:
				depth--
				add(RuleSuffix, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 7 Primary <- <((&('<') (Begin Expression End Action19)) | (&('{') (Action Action18)) | (&('.') (Dot Action17)) | (&('[') Class) | (&('"' | '\'') Literal) | (&('(') (Open Expression Close)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (Identifier !LeftArrow Action16)))> */
		nil,
		/* 8 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				{
					position115 := position
					depth++
					if !rules[RuleIdentStart]() {
						goto l113
					}
				l116:
					{
						position117, tokenIndex117, depth117 := position, tokenIndex, depth
						{
							position118 := position
							depth++
							{
								position119, tokenIndex119, depth119 := position, tokenIndex, depth
								if !rules[RuleIdentStart]() {
									goto l120
								}
								goto l119
							l120:
								position, tokenIndex, depth = position119, tokenIndex119, depth119
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l117
								}
								position++
							}
						l119:
							depth--
							add(RuleIdentCont, position118)
						}
						goto l116
					l117:
						position, tokenIndex, depth = position117, tokenIndex117, depth117
					}
					depth--
					add(RulePegText, position115)
				}
				if !rules[RuleSpacing]() {
					goto l113
				}
				depth--
				add(RuleIdentifier, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 9 IdentStart <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l121
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l121
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l121
						}
						position++
						break
					}
				}

				depth--
				add(RuleIdentStart, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 10 IdentCont <- <(IdentStart / [0-9])> */
		nil,
		/* 11 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char Action20)* '\'' Spacing) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar Action21)* '"' Spacing))> */
		nil,
		/* 12 Class <- <((('[' '[' (('^' DoubleRanges Action22) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges Action23) / Ranges)? ']')) Spacing)> */
		nil,
		/* 13 Ranges <- <(!']' Range (!']' Range Action24)*)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l129
					}
					position++
					goto l127
				l129:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
				}
				if !rules[RuleRange]() {
					goto l127
				}
			l130:
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					{
						position132, tokenIndex132, depth132 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l132
						}
						position++
						goto l131
					l132:
						position, tokenIndex, depth = position132, tokenIndex132, depth132
					}
					if !rules[RuleRange]() {
						goto l131
					}
					{
						add(RuleAction24, position)
					}
					goto l130
				l131:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
				}
				depth--
				add(RuleRanges, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 14 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange Action25)*)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				{
					position136, tokenIndex136, depth136 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l136
					}
					position++
					if buffer[position] != rune(']') {
						goto l136
					}
					position++
					goto l134
				l136:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
				}
				if !rules[RuleDoubleRange]() {
					goto l134
				}
			l137:
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					{
						position139, tokenIndex139, depth139 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l139
						}
						position++
						if buffer[position] != rune(']') {
							goto l139
						}
						position++
						goto l138
					l139:
						position, tokenIndex, depth = position139, tokenIndex139, depth139
					}
					if !rules[RuleDoubleRange]() {
						goto l138
					}
					{
						add(RuleAction25, position)
					}
					goto l137
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
				}
				depth--
				add(RuleDoubleRanges, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 15 Range <- <((Char '-' Char Action26) / Char)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if !rules[RuleChar]() {
						goto l144
					}
					if buffer[position] != rune('-') {
						goto l144
					}
					position++
					if !rules[RuleChar]() {
						goto l144
					}
					{
						add(RuleAction26, position)
					}
					goto l143
				l144:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					if !rules[RuleChar]() {
						goto l141
					}
				}
			l143:
				depth--
				add(RuleRange, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 16 DoubleRange <- <((Char '-' Char Action27) / DoubleChar)> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				{
					position148, tokenIndex148, depth148 := position, tokenIndex, depth
					if !rules[RuleChar]() {
						goto l149
					}
					if buffer[position] != rune('-') {
						goto l149
					}
					position++
					if !rules[RuleChar]() {
						goto l149
					}
					{
						add(RuleAction27, position)
					}
					goto l148
				l149:
					position, tokenIndex, depth = position148, tokenIndex148, depth148
					if !rules[RuleDoubleChar]() {
						goto l146
					}
				}
			l148:
				depth--
				add(RuleDoubleRange, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 17 Char <- <(Escape / (!'\\' <.> Action28))> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !rules[RuleEscape]() {
						goto l154
					}
					goto l153
				l154:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
					{
						position155, tokenIndex155, depth155 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l155
						}
						position++
						goto l151
					l155:
						position, tokenIndex, depth = position155, tokenIndex155, depth155
					}
					{
						position156 := position
						depth++
						if !matchDot() {
							goto l151
						}
						depth--
						add(RulePegText, position156)
					}
					{
						add(RuleAction28, position)
					}
				}
			l153:
				depth--
				add(RuleChar, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 18 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> Action29) / (!'\\' <.> Action30))> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if !rules[RuleEscape]() {
						goto l161
					}
					goto l160
				l161:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					{
						position163 := position
						depth++
						{
							position164, tokenIndex164, depth164 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l165
							}
							position++
							goto l164
						l165:
							position, tokenIndex, depth = position164, tokenIndex164, depth164
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l162
							}
							position++
						}
					l164:
						depth--
						add(RulePegText, position163)
					}
					{
						add(RuleAction29, position)
					}
					goto l160
				l162:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					{
						position167, tokenIndex167, depth167 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l167
						}
						position++
						goto l158
					l167:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
					}
					{
						position168 := position
						depth++
						if !matchDot() {
							goto l158
						}
						depth--
						add(RulePegText, position168)
					}
					{
						add(RuleAction30, position)
					}
				}
			l160:
				depth--
				add(RuleDoubleChar, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 19 Escape <- <(('\\' ('a' / 'A') Action31) / ('\\' ('b' / 'B') Action32) / ('\\' ('e' / 'E') Action33) / ('\\' ('f' / 'F') Action34) / ('\\' ('n' / 'N') Action35) / ('\\' ('r' / 'R') Action36) / ('\\' ('t' / 'T') Action37) / ('\\' ('v' / 'V') Action38) / ('\\' '\'' Action39) / ('\\' '"' Action40) / ('\\' '[' Action41) / ('\\' ']' Action42) / ('\\' '-' Action43) / ('\\' <([0-3] [0-7] [0-7])> Action44) / ('\\' <([0-7] [0-7]?)> Action45) / ('\\' '\\' Action46))> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l173
					}
					position++
					{
						position174, tokenIndex174, depth174 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l175
						}
						position++
						goto l174
					l175:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
						if buffer[position] != rune('A') {
							goto l173
						}
						position++
					}
				l174:
					{
						add(RuleAction31, position)
					}
					goto l172
				l173:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l177
					}
					position++
					{
						position178, tokenIndex178, depth178 := position, tokenIndex, depth
						if buffer[position] != rune('b') {
							goto l179
						}
						position++
						goto l178
					l179:
						position, tokenIndex, depth = position178, tokenIndex178, depth178
						if buffer[position] != rune('B') {
							goto l177
						}
						position++
					}
				l178:
					{
						add(RuleAction32, position)
					}
					goto l172
				l177:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l181
					}
					position++
					{
						position182, tokenIndex182, depth182 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l183
						}
						position++
						goto l182
					l183:
						position, tokenIndex, depth = position182, tokenIndex182, depth182
						if buffer[position] != rune('E') {
							goto l181
						}
						position++
					}
				l182:
					{
						add(RuleAction33, position)
					}
					goto l172
				l181:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l185
					}
					position++
					{
						position186, tokenIndex186, depth186 := position, tokenIndex, depth
						if buffer[position] != rune('f') {
							goto l187
						}
						position++
						goto l186
					l187:
						position, tokenIndex, depth = position186, tokenIndex186, depth186
						if buffer[position] != rune('F') {
							goto l185
						}
						position++
					}
				l186:
					{
						add(RuleAction34, position)
					}
					goto l172
				l185:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l189
					}
					position++
					{
						position190, tokenIndex190, depth190 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l191
						}
						position++
						goto l190
					l191:
						position, tokenIndex, depth = position190, tokenIndex190, depth190
						if buffer[position] != rune('N') {
							goto l189
						}
						position++
					}
				l190:
					{
						add(RuleAction35, position)
					}
					goto l172
				l189:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l193
					}
					position++
					{
						position194, tokenIndex194, depth194 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l195
						}
						position++
						goto l194
					l195:
						position, tokenIndex, depth = position194, tokenIndex194, depth194
						if buffer[position] != rune('R') {
							goto l193
						}
						position++
					}
				l194:
					{
						add(RuleAction36, position)
					}
					goto l172
				l193:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l197
					}
					position++
					{
						position198, tokenIndex198, depth198 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l199
						}
						position++
						goto l198
					l199:
						position, tokenIndex, depth = position198, tokenIndex198, depth198
						if buffer[position] != rune('T') {
							goto l197
						}
						position++
					}
				l198:
					{
						add(RuleAction37, position)
					}
					goto l172
				l197:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l201
					}
					position++
					{
						position202, tokenIndex202, depth202 := position, tokenIndex, depth
						if buffer[position] != rune('v') {
							goto l203
						}
						position++
						goto l202
					l203:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						if buffer[position] != rune('V') {
							goto l201
						}
						position++
					}
				l202:
					{
						add(RuleAction38, position)
					}
					goto l172
				l201:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l205
					}
					position++
					if buffer[position] != rune('\'') {
						goto l205
					}
					position++
					{
						add(RuleAction39, position)
					}
					goto l172
				l205:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l207
					}
					position++
					if buffer[position] != rune('"') {
						goto l207
					}
					position++
					{
						add(RuleAction40, position)
					}
					goto l172
				l207:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l209
					}
					position++
					if buffer[position] != rune('[') {
						goto l209
					}
					position++
					{
						add(RuleAction41, position)
					}
					goto l172
				l209:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l211
					}
					position++
					if buffer[position] != rune(']') {
						goto l211
					}
					position++
					{
						add(RuleAction42, position)
					}
					goto l172
				l211:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l213
					}
					position++
					if buffer[position] != rune('-') {
						goto l213
					}
					position++
					{
						add(RuleAction43, position)
					}
					goto l172
				l213:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l215
					}
					position++
					{
						position216 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('3') {
							goto l215
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l215
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l215
						}
						position++
						depth--
						add(RulePegText, position216)
					}
					{
						add(RuleAction44, position)
					}
					goto l172
				l215:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l218
					}
					position++
					{
						position219 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l218
						}
						position++
						{
							position220, tokenIndex220, depth220 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('7') {
								goto l220
							}
							position++
							goto l221
						l220:
							position, tokenIndex, depth = position220, tokenIndex220, depth220
						}
					l221:
						depth--
						add(RulePegText, position219)
					}
					{
						add(RuleAction45, position)
					}
					goto l172
				l218:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if buffer[position] != rune('\\') {
						goto l170
					}
					position++
					if buffer[position] != rune('\\') {
						goto l170
					}
					position++
					{
						add(RuleAction46, position)
					}
				}
			l172:
				depth--
				add(RuleEscape, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 20 LeftArrow <- <('<' '-' Spacing)> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				if buffer[position] != rune('<') {
					goto l224
				}
				position++
				if buffer[position] != rune('-') {
					goto l224
				}
				position++
				if !rules[RuleSpacing]() {
					goto l224
				}
				depth--
				add(RuleLeftArrow, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 21 Slash <- <('/' Spacing)> */
		func() bool {
			position226, tokenIndex226, depth226 := position, tokenIndex, depth
			{
				position227 := position
				depth++
				if buffer[position] != rune('/') {
					goto l226
				}
				position++
				if !rules[RuleSpacing]() {
					goto l226
				}
				depth--
				add(RuleSlash, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 22 And <- <('&' Spacing)> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				if buffer[position] != rune('&') {
					goto l228
				}
				position++
				if !rules[RuleSpacing]() {
					goto l228
				}
				depth--
				add(RuleAnd, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 23 Not <- <('!' Spacing)> */
		nil,
		/* 24 Question <- <('?' Spacing)> */
		nil,
		/* 25 Star <- <('*' Spacing)> */
		nil,
		/* 26 Plus <- <('+' Spacing)> */
		nil,
		/* 27 Open <- <('(' Spacing)> */
		nil,
		/* 28 Close <- <(')' Spacing)> */
		nil,
		/* 29 Dot <- <('.' Spacing)> */
		nil,
		/* 30 Spacing <- <(Space / Comment)*> */
		func() bool {
			{
				position238 := position
				depth++
			l239:
				{
					position240, tokenIndex240, depth240 := position, tokenIndex, depth
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						{
							position243 := position
							depth++
							{
								switch buffer[position] {
								case '\t':
									if buffer[position] != rune('\t') {
										goto l242
									}
									position++
									break
								case ' ':
									if buffer[position] != rune(' ') {
										goto l242
									}
									position++
									break
								default:
									if !rules[RuleEndOfLine]() {
										goto l242
									}
									break
								}
							}

							depth--
							add(RuleSpace, position243)
						}
						goto l241
					l242:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
						{
							position245 := position
							depth++
							if buffer[position] != rune('#') {
								goto l240
							}
							position++
						l246:
							{
								position247, tokenIndex247, depth247 := position, tokenIndex, depth
								{
									position248, tokenIndex248, depth248 := position, tokenIndex, depth
									if !rules[RuleEndOfLine]() {
										goto l248
									}
									goto l247
								l248:
									position, tokenIndex, depth = position248, tokenIndex248, depth248
								}
								if !matchDot() {
									goto l247
								}
								goto l246
							l247:
								position, tokenIndex, depth = position247, tokenIndex247, depth247
							}
							if !rules[RuleEndOfLine]() {
								goto l240
							}
							depth--
							add(RuleComment, position245)
						}
					}
				l241:
					goto l239
				l240:
					position, tokenIndex, depth = position240, tokenIndex240, depth240
				}
				depth--
				add(RuleSpacing, position238)
			}
			return true
		},
		/* 31 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		nil,
		/* 32 Space <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine))> */
		nil,
		/* 33 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position251, tokenIndex251, depth251 := position, tokenIndex, depth
			{
				position252 := position
				depth++
				{
					position253, tokenIndex253, depth253 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l254
					}
					position++
					if buffer[position] != rune('\n') {
						goto l254
					}
					position++
					goto l253
				l254:
					position, tokenIndex, depth = position253, tokenIndex253, depth253
					if buffer[position] != rune('\n') {
						goto l255
					}
					position++
					goto l253
				l255:
					position, tokenIndex, depth = position253, tokenIndex253, depth253
					if buffer[position] != rune('\r') {
						goto l251
					}
					position++
				}
			l253:
				depth--
				add(RuleEndOfLine, position252)
			}
			return true
		l251:
			position, tokenIndex, depth = position251, tokenIndex251, depth251
			return false
		},
		/* 34 EndOfFile <- <!.> */
		nil,
		/* 35 Action <- <('{' <(!'}' .)*> '}' Spacing)> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				if buffer[position] != rune('{') {
					goto l257
				}
				position++
				{
					position259 := position
					depth++
				l260:
					{
						position261, tokenIndex261, depth261 := position, tokenIndex, depth
						{
							position262, tokenIndex262, depth262 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l262
							}
							position++
							goto l261
						l262:
							position, tokenIndex, depth = position262, tokenIndex262, depth262
						}
						if !matchDot() {
							goto l261
						}
						goto l260
					l261:
						position, tokenIndex, depth = position261, tokenIndex261, depth261
					}
					depth--
					add(RulePegText, position259)
				}
				if buffer[position] != rune('}') {
					goto l257
				}
				position++
				if !rules[RuleSpacing]() {
					goto l257
				}
				depth--
				add(RuleAction, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
		/* 36 Begin <- <('<' Spacing)> */
		nil,
		/* 37 End <- <('>' Spacing)> */
		nil,
		/* 39 Action0 <- <{ p.AddPackage(buffer[begin:end]) }> */
		nil,
		/* 40 Action1 <- <{ p.AddPeg(buffer[begin:end]) }> */
		nil,
		/* 41 Action2 <- <{ p.AddState(buffer[begin:end]) }> */
		nil,
		nil,
		/* 43 Action3 <- <{ p.AddImport(buffer[begin:end]) }> */
		nil,
		/* 44 Action4 <- <{ p.AddRule(buffer[begin:end]) }> */
		nil,
		/* 45 Action5 <- <{ p.AddExpression() }> */
		nil,
		/* 46 Action6 <- <{ p.AddAlternate() }> */
		nil,
		/* 47 Action7 <- <{ p.AddNil(); p.AddAlternate() }> */
		nil,
		/* 48 Action8 <- <{ p.AddNil() }> */
		nil,
		/* 49 Action9 <- <{ p.AddSequence() }> */
		nil,
		/* 50 Action10 <- <{ p.AddPredicate(buffer[begin:end]) }> */
		nil,
		/* 51 Action11 <- <{ p.AddPeekFor() }> */
		nil,
		/* 52 Action12 <- <{ p.AddPeekNot() }> */
		nil,
		/* 53 Action13 <- <{ p.AddQuery() }> */
		nil,
		/* 54 Action14 <- <{ p.AddStar() }> */
		nil,
		/* 55 Action15 <- <{ p.AddPlus() }> */
		nil,
		/* 56 Action16 <- <{ p.AddName(buffer[begin:end]) }> */
		nil,
		/* 57 Action17 <- <{ p.AddDot() }> */
		nil,
		/* 58 Action18 <- <{ p.AddAction(buffer[begin:end]) }> */
		nil,
		/* 59 Action19 <- <{ p.AddPush() }> */
		nil,
		/* 60 Action20 <- <{ p.AddSequence() }> */
		nil,
		/* 61 Action21 <- <{ p.AddSequence() }> */
		nil,
		/* 62 Action22 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 63 Action23 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 64 Action24 <- <{ p.AddAlternate() }> */
		nil,
		/* 65 Action25 <- <{ p.AddAlternate() }> */
		nil,
		/* 66 Action26 <- <{ p.AddRange() }> */
		nil,
		/* 67 Action27 <- <{ p.AddDoubleRange() }> */
		nil,
		/* 68 Action28 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 69 Action29 <- <{ p.AddDoubleCharacter(buffer[begin:end]) }> */
		nil,
		/* 70 Action30 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 71 Action31 <- <{ p.AddCharacter("\a") }> */
		nil,
		/* 72 Action32 <- <{ p.AddCharacter("\b") }> */
		nil,
		/* 73 Action33 <- <{ p.AddCharacter("\x1B") }> */
		nil,
		/* 74 Action34 <- <{ p.AddCharacter("\f") }> */
		nil,
		/* 75 Action35 <- <{ p.AddCharacter("\n") }> */
		nil,
		/* 76 Action36 <- <{ p.AddCharacter("\r") }> */
		nil,
		/* 77 Action37 <- <{ p.AddCharacter("\t") }> */
		nil,
		/* 78 Action38 <- <{ p.AddCharacter("\v") }> */
		nil,
		/* 79 Action39 <- <{ p.AddCharacter("'") }> */
		nil,
		/* 80 Action40 <- <{ p.AddCharacter("\"") }> */
		nil,
		/* 81 Action41 <- <{ p.AddCharacter("[") }> */
		nil,
		/* 82 Action42 <- <{ p.AddCharacter("]") }> */
		nil,
		/* 83 Action43 <- <{ p.AddCharacter("-") }> */
		nil,
		/* 84 Action44 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 85 Action45 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 86 Action46 <- <{ p.AddCharacter("\\") }> */
		nil,
	}
	p.rules = rules
}
