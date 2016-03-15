package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleGrammar
	ruleImport
	ruleDefinition
	ruleExpression
	ruleSequence
	rulePrefix
	ruleSuffix
	rulePrimary
	ruleIdentifier
	ruleIdentStart
	ruleIdentCont
	ruleLiteral
	ruleClass
	ruleRanges
	ruleDoubleRanges
	ruleRange
	ruleDoubleRange
	ruleChar
	ruleDoubleChar
	ruleEscape
	ruleLeftArrow
	ruleSlash
	ruleAnd
	ruleNot
	ruleQuestion
	ruleStar
	rulePlus
	ruleOpen
	ruleClose
	ruleDot
	ruleSpaceComment
	ruleSpacing
	ruleMustSpacing
	ruleComment
	ruleSpace
	ruleEndOfLine
	ruleEndOfFile
	ruleAction
	ruleActionBody
	ruleBegin
	ruleEnd
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27
	ruleAction28
	ruleAction29
	ruleAction30
	ruleAction31
	ruleAction32
	ruleAction33
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38
	ruleAction39
	ruleAction40
	ruleAction41
	ruleAction42
	ruleAction43
	ruleAction44
	ruleAction45
	ruleAction46
	ruleAction47
	ruleAction48

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
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
	"SpaceComment",
	"Spacing",
	"MustSpacing",
	"Comment",
	"Space",
	"EndOfLine",
	"EndOfFile",
	"Action",
	"ActionBody",
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
	"Action47",
	"Action48",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next uint32, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
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
		if token.pegRule == ruleUnknown {
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
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
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
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
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
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
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
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
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
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
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
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

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
	rules  [92]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
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
	p   *Peg
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
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
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.AddPackage(text)
		case ruleAction1:
			p.AddPeg(text)
		case ruleAction2:
			p.AddState(text)
		case ruleAction3:
			p.AddImport(text)
		case ruleAction4:
			p.AddRule(text)
		case ruleAction5:
			p.AddExpression()
		case ruleAction6:
			p.AddAlternate()
		case ruleAction7:
			p.AddNil()
			p.AddAlternate()
		case ruleAction8:
			p.AddNil()
		case ruleAction9:
			p.AddSequence()
		case ruleAction10:
			p.AddPredicate(text)
		case ruleAction11:
			p.AddStateChange(text)
		case ruleAction12:
			p.AddPeekFor()
		case ruleAction13:
			p.AddPeekNot()
		case ruleAction14:
			p.AddQuery()
		case ruleAction15:
			p.AddStar()
		case ruleAction16:
			p.AddPlus()
		case ruleAction17:
			p.AddName(text)
		case ruleAction18:
			p.AddDot()
		case ruleAction19:
			p.AddAction(text)
		case ruleAction20:
			p.AddPush()
		case ruleAction21:
			p.AddSequence()
		case ruleAction22:
			p.AddSequence()
		case ruleAction23:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case ruleAction24:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case ruleAction25:
			p.AddAlternate()
		case ruleAction26:
			p.AddAlternate()
		case ruleAction27:
			p.AddRange()
		case ruleAction28:
			p.AddDoubleRange()
		case ruleAction29:
			p.AddCharacter(text)
		case ruleAction30:
			p.AddDoubleCharacter(text)
		case ruleAction31:
			p.AddCharacter(text)
		case ruleAction32:
			p.AddCharacter("\a")
		case ruleAction33:
			p.AddCharacter("\b")
		case ruleAction34:
			p.AddCharacter("\x1B")
		case ruleAction35:
			p.AddCharacter("\f")
		case ruleAction36:
			p.AddCharacter("\n")
		case ruleAction37:
			p.AddCharacter("\r")
		case ruleAction38:
			p.AddCharacter("\t")
		case ruleAction39:
			p.AddCharacter("\v")
		case ruleAction40:
			p.AddCharacter("'")
		case ruleAction41:
			p.AddCharacter("\"")
		case ruleAction42:
			p.AddCharacter("[")
		case ruleAction43:
			p.AddCharacter("]")
		case ruleAction44:
			p.AddCharacter("-")
		case ruleAction45:
			p.AddHexaCharacter(text)
		case ruleAction46:
			p.AddOctalCharacter(text)
		case ruleAction47:
			p.AddOctalCharacter(text)
		case ruleAction48:
			p.AddCharacter("\\")

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Peg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

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
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
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

	_rules = [...]func() bool{
		nil,
		/* 0 Grammar <- <(Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') MustSpacing Identifier Action0 Import* ('t' 'y' 'p' 'e') MustSpacing Identifier Action1 ('P' 'e' 'g') Spacing Action Action2 Definition+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[ruleSpacing]() {
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
				if !_rules[ruleMustSpacing]() {
					goto l0
				}
				if !_rules[ruleIdentifier]() {
					goto l0
				}
				if !_rules[ruleAction0]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					if !_rules[ruleImport]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
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
				if !_rules[ruleMustSpacing]() {
					goto l0
				}
				if !_rules[ruleIdentifier]() {
					goto l0
				}
				if !_rules[ruleAction1]() {
					goto l0
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
				if !_rules[ruleSpacing]() {
					goto l0
				}
				if !_rules[ruleAction]() {
					goto l0
				}
				if !_rules[ruleAction2]() {
					goto l0
				}
				if !_rules[ruleDefinition]() {
					goto l0
				}
			l4:
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !_rules[ruleDefinition]() {
						goto l5
					}
					goto l4
				l5:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
				}
				if !_rules[ruleEndOfFile]() {
					goto l0
				}
				depth--
				add(ruleGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Import <- <('i' 'm' 'p' 'o' 'r' 't' Spacing '"' <([a-z] / [A-Z] / '_' / '/' / '.' / '-')+> '"' Spacing Action3)> */
		func() bool {
			position6, tokenIndex6, depth6 := position, tokenIndex, depth
			{
				position7 := position
				depth++
				if buffer[position] != rune('i') {
					goto l6
				}
				position++
				if buffer[position] != rune('m') {
					goto l6
				}
				position++
				if buffer[position] != rune('p') {
					goto l6
				}
				position++
				if buffer[position] != rune('o') {
					goto l6
				}
				position++
				if buffer[position] != rune('r') {
					goto l6
				}
				position++
				if buffer[position] != rune('t') {
					goto l6
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l6
				}
				if buffer[position] != rune('"') {
					goto l6
				}
				position++
				{
					position8 := position
					depth++
					{
						position11, tokenIndex11, depth11 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l12
						}
						position++
						goto l11
					l12:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l13
						}
						position++
						goto l11
					l13:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('_') {
							goto l14
						}
						position++
						goto l11
					l14:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('/') {
							goto l15
						}
						position++
						goto l11
					l15:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('.') {
							goto l16
						}
						position++
						goto l11
					l16:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('-') {
							goto l6
						}
						position++
					}
				l11:
				l9:
					{
						position10, tokenIndex10, depth10 := position, tokenIndex, depth
						{
							position17, tokenIndex17, depth17 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l18
							}
							position++
							goto l17
						l18:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l19
							}
							position++
							goto l17
						l19:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('_') {
								goto l20
							}
							position++
							goto l17
						l20:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('/') {
								goto l21
							}
							position++
							goto l17
						l21:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('.') {
								goto l22
							}
							position++
							goto l17
						l22:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('-') {
								goto l10
							}
							position++
						}
					l17:
						goto l9
					l10:
						position, tokenIndex, depth = position10, tokenIndex10, depth10
					}
					depth--
					add(rulePegText, position8)
				}
				if buffer[position] != rune('"') {
					goto l6
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l6
				}
				if !_rules[ruleAction3]() {
					goto l6
				}
				depth--
				add(ruleImport, position7)
			}
			return true
		l6:
			position, tokenIndex, depth = position6, tokenIndex6, depth6
			return false
		},
		/* 2 Definition <- <(Identifier Action4 LeftArrow Expression Action5 &((Identifier LeftArrow) / !.))> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if !_rules[ruleIdentifier]() {
					goto l23
				}
				if !_rules[ruleAction4]() {
					goto l23
				}
				if !_rules[ruleLeftArrow]() {
					goto l23
				}
				if !_rules[ruleExpression]() {
					goto l23
				}
				if !_rules[ruleAction5]() {
					goto l23
				}
				{
					position25, tokenIndex25, depth25 := position, tokenIndex, depth
					{
						position26, tokenIndex26, depth26 := position, tokenIndex, depth
						if !_rules[ruleIdentifier]() {
							goto l27
						}
						if !_rules[ruleLeftArrow]() {
							goto l27
						}
						goto l26
					l27:
						position, tokenIndex, depth = position26, tokenIndex26, depth26
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							if !matchDot() {
								goto l28
							}
							goto l23
						l28:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
						}
					}
				l26:
					position, tokenIndex, depth = position25, tokenIndex25, depth25
				}
				depth--
				add(ruleDefinition, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 3 Expression <- <((Sequence (Slash Sequence Action6)* (Slash Action7)?) / Action8)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				{
					position31, tokenIndex31, depth31 := position, tokenIndex, depth
					if !_rules[ruleSequence]() {
						goto l32
					}
				l33:
					{
						position34, tokenIndex34, depth34 := position, tokenIndex, depth
						if !_rules[ruleSlash]() {
							goto l34
						}
						if !_rules[ruleSequence]() {
							goto l34
						}
						if !_rules[ruleAction6]() {
							goto l34
						}
						goto l33
					l34:
						position, tokenIndex, depth = position34, tokenIndex34, depth34
					}
					{
						position35, tokenIndex35, depth35 := position, tokenIndex, depth
						if !_rules[ruleSlash]() {
							goto l35
						}
						if !_rules[ruleAction7]() {
							goto l35
						}
						goto l36
					l35:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
					}
				l36:
					goto l31
				l32:
					position, tokenIndex, depth = position31, tokenIndex31, depth31
					if !_rules[ruleAction8]() {
						goto l29
					}
				}
			l31:
				depth--
				add(ruleExpression, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 4 Sequence <- <(Prefix (Prefix Action9)*)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if !_rules[rulePrefix]() {
					goto l37
				}
			l39:
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if !_rules[rulePrefix]() {
						goto l40
					}
					if !_rules[ruleAction9]() {
						goto l40
					}
					goto l39
				l40:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
				}
				depth--
				add(ruleSequence, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 5 Prefix <- <((And Action Action10) / (Not Action Action11) / (And Suffix Action12) / (Not Suffix Action13) / Suffix)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				{
					position43, tokenIndex43, depth43 := position, tokenIndex, depth
					if !_rules[ruleAnd]() {
						goto l44
					}
					if !_rules[ruleAction]() {
						goto l44
					}
					if !_rules[ruleAction10]() {
						goto l44
					}
					goto l43
				l44:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					if !_rules[ruleNot]() {
						goto l45
					}
					if !_rules[ruleAction]() {
						goto l45
					}
					if !_rules[ruleAction11]() {
						goto l45
					}
					goto l43
				l45:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					if !_rules[ruleAnd]() {
						goto l46
					}
					if !_rules[ruleSuffix]() {
						goto l46
					}
					if !_rules[ruleAction12]() {
						goto l46
					}
					goto l43
				l46:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					if !_rules[ruleNot]() {
						goto l47
					}
					if !_rules[ruleSuffix]() {
						goto l47
					}
					if !_rules[ruleAction13]() {
						goto l47
					}
					goto l43
				l47:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					if !_rules[ruleSuffix]() {
						goto l41
					}
				}
			l43:
				depth--
				add(rulePrefix, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 6 Suffix <- <(Primary ((Question Action14) / (Star Action15) / (Plus Action16))?)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !_rules[rulePrimary]() {
					goto l48
				}
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					{
						position52, tokenIndex52, depth52 := position, tokenIndex, depth
						if !_rules[ruleQuestion]() {
							goto l53
						}
						if !_rules[ruleAction14]() {
							goto l53
						}
						goto l52
					l53:
						position, tokenIndex, depth = position52, tokenIndex52, depth52
						if !_rules[ruleStar]() {
							goto l54
						}
						if !_rules[ruleAction15]() {
							goto l54
						}
						goto l52
					l54:
						position, tokenIndex, depth = position52, tokenIndex52, depth52
						if !_rules[rulePlus]() {
							goto l50
						}
						if !_rules[ruleAction16]() {
							goto l50
						}
					}
				l52:
					goto l51
				l50:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
				}
			l51:
				depth--
				add(ruleSuffix, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 7 Primary <- <((Identifier !LeftArrow Action17) / (Open Expression Close) / Literal / Class / (Dot Action18) / (Action Action19) / (Begin Expression End Action20))> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if !_rules[ruleIdentifier]() {
						goto l58
					}
					{
						position59, tokenIndex59, depth59 := position, tokenIndex, depth
						if !_rules[ruleLeftArrow]() {
							goto l59
						}
						goto l58
					l59:
						position, tokenIndex, depth = position59, tokenIndex59, depth59
					}
					if !_rules[ruleAction17]() {
						goto l58
					}
					goto l57
				l58:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if !_rules[ruleOpen]() {
						goto l60
					}
					if !_rules[ruleExpression]() {
						goto l60
					}
					if !_rules[ruleClose]() {
						goto l60
					}
					goto l57
				l60:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if !_rules[ruleLiteral]() {
						goto l61
					}
					goto l57
				l61:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if !_rules[ruleClass]() {
						goto l62
					}
					goto l57
				l62:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if !_rules[ruleDot]() {
						goto l63
					}
					if !_rules[ruleAction18]() {
						goto l63
					}
					goto l57
				l63:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if !_rules[ruleAction]() {
						goto l64
					}
					if !_rules[ruleAction19]() {
						goto l64
					}
					goto l57
				l64:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if !_rules[ruleBegin]() {
						goto l55
					}
					if !_rules[ruleExpression]() {
						goto l55
					}
					if !_rules[ruleEnd]() {
						goto l55
					}
					if !_rules[ruleAction20]() {
						goto l55
					}
				}
			l57:
				depth--
				add(rulePrimary, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 8 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				{
					position67 := position
					depth++
					if !_rules[ruleIdentStart]() {
						goto l65
					}
				l68:
					{
						position69, tokenIndex69, depth69 := position, tokenIndex, depth
						if !_rules[ruleIdentCont]() {
							goto l69
						}
						goto l68
					l69:
						position, tokenIndex, depth = position69, tokenIndex69, depth69
					}
					depth--
					add(rulePegText, position67)
				}
				if !_rules[ruleSpacing]() {
					goto l65
				}
				depth--
				add(ruleIdentifier, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 9 IdentStart <- <([a-z] / [A-Z] / '_')> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				{
					position72, tokenIndex72, depth72 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l73
					}
					position++
					goto l72
				l73:
					position, tokenIndex, depth = position72, tokenIndex72, depth72
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l74
					}
					position++
					goto l72
				l74:
					position, tokenIndex, depth = position72, tokenIndex72, depth72
					if buffer[position] != rune('_') {
						goto l70
					}
					position++
				}
			l72:
				depth--
				add(ruleIdentStart, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 10 IdentCont <- <(IdentStart / [0-9])> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					if !_rules[ruleIdentStart]() {
						goto l78
					}
					goto l77
				l78:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l75
					}
					position++
				}
			l77:
				depth--
				add(ruleIdentCont, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 11 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char Action21)* '\'' Spacing) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar Action22)* '"' Spacing))> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l82
					}
					position++
					{
						position83, tokenIndex83, depth83 := position, tokenIndex, depth
						{
							position85, tokenIndex85, depth85 := position, tokenIndex, depth
							if buffer[position] != rune('\'') {
								goto l85
							}
							position++
							goto l83
						l85:
							position, tokenIndex, depth = position85, tokenIndex85, depth85
						}
						if !_rules[ruleChar]() {
							goto l83
						}
						goto l84
					l83:
						position, tokenIndex, depth = position83, tokenIndex83, depth83
					}
				l84:
				l86:
					{
						position87, tokenIndex87, depth87 := position, tokenIndex, depth
						{
							position88, tokenIndex88, depth88 := position, tokenIndex, depth
							if buffer[position] != rune('\'') {
								goto l88
							}
							position++
							goto l87
						l88:
							position, tokenIndex, depth = position88, tokenIndex88, depth88
						}
						if !_rules[ruleChar]() {
							goto l87
						}
						if !_rules[ruleAction21]() {
							goto l87
						}
						goto l86
					l87:
						position, tokenIndex, depth = position87, tokenIndex87, depth87
					}
					if buffer[position] != rune('\'') {
						goto l82
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l82
					}
					goto l81
				l82:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
					if buffer[position] != rune('"') {
						goto l79
					}
					position++
					{
						position89, tokenIndex89, depth89 := position, tokenIndex, depth
						{
							position91, tokenIndex91, depth91 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l91
							}
							position++
							goto l89
						l91:
							position, tokenIndex, depth = position91, tokenIndex91, depth91
						}
						if !_rules[ruleDoubleChar]() {
							goto l89
						}
						goto l90
					l89:
						position, tokenIndex, depth = position89, tokenIndex89, depth89
					}
				l90:
				l92:
					{
						position93, tokenIndex93, depth93 := position, tokenIndex, depth
						{
							position94, tokenIndex94, depth94 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l94
							}
							position++
							goto l93
						l94:
							position, tokenIndex, depth = position94, tokenIndex94, depth94
						}
						if !_rules[ruleDoubleChar]() {
							goto l93
						}
						if !_rules[ruleAction22]() {
							goto l93
						}
						goto l92
					l93:
						position, tokenIndex, depth = position93, tokenIndex93, depth93
					}
					if buffer[position] != rune('"') {
						goto l79
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l79
					}
				}
			l81:
				depth--
				add(ruleLiteral, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 12 Class <- <((('[' '[' (('^' DoubleRanges Action23) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges Action24) / Ranges)? ']')) Spacing)> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l98
					}
					position++
					if buffer[position] != rune('[') {
						goto l98
					}
					position++
					{
						position99, tokenIndex99, depth99 := position, tokenIndex, depth
						{
							position101, tokenIndex101, depth101 := position, tokenIndex, depth
							if buffer[position] != rune('^') {
								goto l102
							}
							position++
							if !_rules[ruleDoubleRanges]() {
								goto l102
							}
							if !_rules[ruleAction23]() {
								goto l102
							}
							goto l101
						l102:
							position, tokenIndex, depth = position101, tokenIndex101, depth101
							if !_rules[ruleDoubleRanges]() {
								goto l99
							}
						}
					l101:
						goto l100
					l99:
						position, tokenIndex, depth = position99, tokenIndex99, depth99
					}
				l100:
					if buffer[position] != rune(']') {
						goto l98
					}
					position++
					if buffer[position] != rune(']') {
						goto l98
					}
					position++
					goto l97
				l98:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if buffer[position] != rune('[') {
						goto l95
					}
					position++
					{
						position103, tokenIndex103, depth103 := position, tokenIndex, depth
						{
							position105, tokenIndex105, depth105 := position, tokenIndex, depth
							if buffer[position] != rune('^') {
								goto l106
							}
							position++
							if !_rules[ruleRanges]() {
								goto l106
							}
							if !_rules[ruleAction24]() {
								goto l106
							}
							goto l105
						l106:
							position, tokenIndex, depth = position105, tokenIndex105, depth105
							if !_rules[ruleRanges]() {
								goto l103
							}
						}
					l105:
						goto l104
					l103:
						position, tokenIndex, depth = position103, tokenIndex103, depth103
					}
				l104:
					if buffer[position] != rune(']') {
						goto l95
					}
					position++
				}
			l97:
				if !_rules[ruleSpacing]() {
					goto l95
				}
				depth--
				add(ruleClass, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 13 Ranges <- <(!']' Range (!']' Range Action25)*)> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				{
					position109, tokenIndex109, depth109 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l109
					}
					position++
					goto l107
				l109:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
				}
				if !_rules[ruleRange]() {
					goto l107
				}
			l110:
				{
					position111, tokenIndex111, depth111 := position, tokenIndex, depth
					{
						position112, tokenIndex112, depth112 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l112
						}
						position++
						goto l111
					l112:
						position, tokenIndex, depth = position112, tokenIndex112, depth112
					}
					if !_rules[ruleRange]() {
						goto l111
					}
					if !_rules[ruleAction25]() {
						goto l111
					}
					goto l110
				l111:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
				}
				depth--
				add(ruleRanges, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 14 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange Action26)*)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l115
					}
					position++
					if buffer[position] != rune(']') {
						goto l115
					}
					position++
					goto l113
				l115:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
				}
				if !_rules[ruleDoubleRange]() {
					goto l113
				}
			l116:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					{
						position118, tokenIndex118, depth118 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l118
						}
						position++
						if buffer[position] != rune(']') {
							goto l118
						}
						position++
						goto l117
					l118:
						position, tokenIndex, depth = position118, tokenIndex118, depth118
					}
					if !_rules[ruleDoubleRange]() {
						goto l117
					}
					if !_rules[ruleAction26]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
				depth--
				add(ruleDoubleRanges, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 15 Range <- <((Char '-' Char Action27) / Char)> */
		func() bool {
			position119, tokenIndex119, depth119 := position, tokenIndex, depth
			{
				position120 := position
				depth++
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if !_rules[ruleChar]() {
						goto l122
					}
					if buffer[position] != rune('-') {
						goto l122
					}
					position++
					if !_rules[ruleChar]() {
						goto l122
					}
					if !_rules[ruleAction27]() {
						goto l122
					}
					goto l121
				l122:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleChar]() {
						goto l119
					}
				}
			l121:
				depth--
				add(ruleRange, position120)
			}
			return true
		l119:
			position, tokenIndex, depth = position119, tokenIndex119, depth119
			return false
		},
		/* 16 DoubleRange <- <((Char '-' Char Action28) / DoubleChar)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					if !_rules[ruleChar]() {
						goto l126
					}
					if buffer[position] != rune('-') {
						goto l126
					}
					position++
					if !_rules[ruleChar]() {
						goto l126
					}
					if !_rules[ruleAction28]() {
						goto l126
					}
					goto l125
				l126:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleDoubleChar]() {
						goto l123
					}
				}
			l125:
				depth--
				add(ruleDoubleRange, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 17 Char <- <(Escape / (!'\\' <.> Action29))> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if !_rules[ruleEscape]() {
						goto l130
					}
					goto l129
				l130:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					{
						position131, tokenIndex131, depth131 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l131
						}
						position++
						goto l127
					l131:
						position, tokenIndex, depth = position131, tokenIndex131, depth131
					}
					{
						position132 := position
						depth++
						if !matchDot() {
							goto l127
						}
						depth--
						add(rulePegText, position132)
					}
					if !_rules[ruleAction29]() {
						goto l127
					}
				}
			l129:
				depth--
				add(ruleChar, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 18 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> Action30) / (!'\\' <.> Action31))> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if !_rules[ruleEscape]() {
						goto l136
					}
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					{
						position138 := position
						depth++
						{
							position139, tokenIndex139, depth139 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l140
							}
							position++
							goto l139
						l140:
							position, tokenIndex, depth = position139, tokenIndex139, depth139
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l137
							}
							position++
						}
					l139:
						depth--
						add(rulePegText, position138)
					}
					if !_rules[ruleAction30]() {
						goto l137
					}
					goto l135
				l137:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					{
						position141, tokenIndex141, depth141 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l141
						}
						position++
						goto l133
					l141:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
					}
					{
						position142 := position
						depth++
						if !matchDot() {
							goto l133
						}
						depth--
						add(rulePegText, position142)
					}
					if !_rules[ruleAction31]() {
						goto l133
					}
				}
			l135:
				depth--
				add(ruleDoubleChar, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 19 Escape <- <(('\\' ('a' / 'A') Action32) / ('\\' ('b' / 'B') Action33) / ('\\' ('e' / 'E') Action34) / ('\\' ('f' / 'F') Action35) / ('\\' ('n' / 'N') Action36) / ('\\' ('r' / 'R') Action37) / ('\\' ('t' / 'T') Action38) / ('\\' ('v' / 'V') Action39) / ('\\' '\'' Action40) / ('\\' '"' Action41) / ('\\' '[' Action42) / ('\\' ']' Action43) / ('\\' '-' Action44) / ('\\' ('0' ('x' / 'X')) <([0-9] / [a-f] / [A-F])+> Action45) / ('\\' <([0-3] [0-7] [0-7])> Action46) / ('\\' <([0-7] [0-7]?)> Action47) / ('\\' '\\' Action48))> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l146
					}
					position++
					{
						position147, tokenIndex147, depth147 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l148
						}
						position++
						goto l147
					l148:
						position, tokenIndex, depth = position147, tokenIndex147, depth147
						if buffer[position] != rune('A') {
							goto l146
						}
						position++
					}
				l147:
					if !_rules[ruleAction32]() {
						goto l146
					}
					goto l145
				l146:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l149
					}
					position++
					{
						position150, tokenIndex150, depth150 := position, tokenIndex, depth
						if buffer[position] != rune('b') {
							goto l151
						}
						position++
						goto l150
					l151:
						position, tokenIndex, depth = position150, tokenIndex150, depth150
						if buffer[position] != rune('B') {
							goto l149
						}
						position++
					}
				l150:
					if !_rules[ruleAction33]() {
						goto l149
					}
					goto l145
				l149:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l152
					}
					position++
					{
						position153, tokenIndex153, depth153 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l154
						}
						position++
						goto l153
					l154:
						position, tokenIndex, depth = position153, tokenIndex153, depth153
						if buffer[position] != rune('E') {
							goto l152
						}
						position++
					}
				l153:
					if !_rules[ruleAction34]() {
						goto l152
					}
					goto l145
				l152:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l155
					}
					position++
					{
						position156, tokenIndex156, depth156 := position, tokenIndex, depth
						if buffer[position] != rune('f') {
							goto l157
						}
						position++
						goto l156
					l157:
						position, tokenIndex, depth = position156, tokenIndex156, depth156
						if buffer[position] != rune('F') {
							goto l155
						}
						position++
					}
				l156:
					if !_rules[ruleAction35]() {
						goto l155
					}
					goto l145
				l155:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l158
					}
					position++
					{
						position159, tokenIndex159, depth159 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l160
						}
						position++
						goto l159
					l160:
						position, tokenIndex, depth = position159, tokenIndex159, depth159
						if buffer[position] != rune('N') {
							goto l158
						}
						position++
					}
				l159:
					if !_rules[ruleAction36]() {
						goto l158
					}
					goto l145
				l158:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l161
					}
					position++
					{
						position162, tokenIndex162, depth162 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l163
						}
						position++
						goto l162
					l163:
						position, tokenIndex, depth = position162, tokenIndex162, depth162
						if buffer[position] != rune('R') {
							goto l161
						}
						position++
					}
				l162:
					if !_rules[ruleAction37]() {
						goto l161
					}
					goto l145
				l161:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l164
					}
					position++
					{
						position165, tokenIndex165, depth165 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l166
						}
						position++
						goto l165
					l166:
						position, tokenIndex, depth = position165, tokenIndex165, depth165
						if buffer[position] != rune('T') {
							goto l164
						}
						position++
					}
				l165:
					if !_rules[ruleAction38]() {
						goto l164
					}
					goto l145
				l164:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l167
					}
					position++
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						if buffer[position] != rune('v') {
							goto l169
						}
						position++
						goto l168
					l169:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
						if buffer[position] != rune('V') {
							goto l167
						}
						position++
					}
				l168:
					if !_rules[ruleAction39]() {
						goto l167
					}
					goto l145
				l167:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l170
					}
					position++
					if buffer[position] != rune('\'') {
						goto l170
					}
					position++
					if !_rules[ruleAction40]() {
						goto l170
					}
					goto l145
				l170:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l171
					}
					position++
					if buffer[position] != rune('"') {
						goto l171
					}
					position++
					if !_rules[ruleAction41]() {
						goto l171
					}
					goto l145
				l171:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l172
					}
					position++
					if buffer[position] != rune('[') {
						goto l172
					}
					position++
					if !_rules[ruleAction42]() {
						goto l172
					}
					goto l145
				l172:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l173
					}
					position++
					if buffer[position] != rune(']') {
						goto l173
					}
					position++
					if !_rules[ruleAction43]() {
						goto l173
					}
					goto l145
				l173:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l174
					}
					position++
					if buffer[position] != rune('-') {
						goto l174
					}
					position++
					if !_rules[ruleAction44]() {
						goto l174
					}
					goto l145
				l174:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l175
					}
					position++
					if buffer[position] != rune('0') {
						goto l175
					}
					position++
					{
						position176, tokenIndex176, depth176 := position, tokenIndex, depth
						if buffer[position] != rune('x') {
							goto l177
						}
						position++
						goto l176
					l177:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
						if buffer[position] != rune('X') {
							goto l175
						}
						position++
					}
				l176:
					{
						position178 := position
						depth++
						{
							position181, tokenIndex181, depth181 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l182
							}
							position++
							goto l181
						l182:
							position, tokenIndex, depth = position181, tokenIndex181, depth181
							if c := buffer[position]; c < rune('a') || c > rune('f') {
								goto l183
							}
							position++
							goto l181
						l183:
							position, tokenIndex, depth = position181, tokenIndex181, depth181
							if c := buffer[position]; c < rune('A') || c > rune('F') {
								goto l175
							}
							position++
						}
					l181:
					l179:
						{
							position180, tokenIndex180, depth180 := position, tokenIndex, depth
							{
								position184, tokenIndex184, depth184 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l185
								}
								position++
								goto l184
							l185:
								position, tokenIndex, depth = position184, tokenIndex184, depth184
								if c := buffer[position]; c < rune('a') || c > rune('f') {
									goto l186
								}
								position++
								goto l184
							l186:
								position, tokenIndex, depth = position184, tokenIndex184, depth184
								if c := buffer[position]; c < rune('A') || c > rune('F') {
									goto l180
								}
								position++
							}
						l184:
							goto l179
						l180:
							position, tokenIndex, depth = position180, tokenIndex180, depth180
						}
						depth--
						add(rulePegText, position178)
					}
					if !_rules[ruleAction45]() {
						goto l175
					}
					goto l145
				l175:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l187
					}
					position++
					{
						position188 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('3') {
							goto l187
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l187
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l187
						}
						position++
						depth--
						add(rulePegText, position188)
					}
					if !_rules[ruleAction46]() {
						goto l187
					}
					goto l145
				l187:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l189
					}
					position++
					{
						position190 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l189
						}
						position++
						{
							position191, tokenIndex191, depth191 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('7') {
								goto l191
							}
							position++
							goto l192
						l191:
							position, tokenIndex, depth = position191, tokenIndex191, depth191
						}
					l192:
						depth--
						add(rulePegText, position190)
					}
					if !_rules[ruleAction47]() {
						goto l189
					}
					goto l145
				l189:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('\\') {
						goto l143
					}
					position++
					if buffer[position] != rune('\\') {
						goto l143
					}
					position++
					if !_rules[ruleAction48]() {
						goto l143
					}
				}
			l145:
				depth--
				add(ruleEscape, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 20 LeftArrow <- <((('<' '-') / '←') Spacing)> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if buffer[position] != rune('<') {
						goto l196
					}
					position++
					if buffer[position] != rune('-') {
						goto l196
					}
					position++
					goto l195
				l196:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
					if buffer[position] != rune('←') {
						goto l193
					}
					position++
				}
			l195:
				if !_rules[ruleSpacing]() {
					goto l193
				}
				depth--
				add(ruleLeftArrow, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 21 Slash <- <('/' Spacing)> */
		func() bool {
			position197, tokenIndex197, depth197 := position, tokenIndex, depth
			{
				position198 := position
				depth++
				if buffer[position] != rune('/') {
					goto l197
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l197
				}
				depth--
				add(ruleSlash, position198)
			}
			return true
		l197:
			position, tokenIndex, depth = position197, tokenIndex197, depth197
			return false
		},
		/* 22 And <- <('&' Spacing)> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if buffer[position] != rune('&') {
					goto l199
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l199
				}
				depth--
				add(ruleAnd, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 23 Not <- <('!' Spacing)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				if buffer[position] != rune('!') {
					goto l201
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l201
				}
				depth--
				add(ruleNot, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 24 Question <- <('?' Spacing)> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				if buffer[position] != rune('?') {
					goto l203
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l203
				}
				depth--
				add(ruleQuestion, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 25 Star <- <('*' Spacing)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				if buffer[position] != rune('*') {
					goto l205
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l205
				}
				depth--
				add(ruleStar, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 26 Plus <- <('+' Spacing)> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				if buffer[position] != rune('+') {
					goto l207
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l207
				}
				depth--
				add(rulePlus, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 27 Open <- <('(' Spacing)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				if buffer[position] != rune('(') {
					goto l209
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l209
				}
				depth--
				add(ruleOpen, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 28 Close <- <(')' Spacing)> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				if buffer[position] != rune(')') {
					goto l211
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l211
				}
				depth--
				add(ruleClose, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 29 Dot <- <('.' Spacing)> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				if buffer[position] != rune('.') {
					goto l213
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l213
				}
				depth--
				add(ruleDot, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 30 SpaceComment <- <(Space / Comment)> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l218
					}
					goto l217
				l218:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
					if !_rules[ruleComment]() {
						goto l215
					}
				}
			l217:
				depth--
				add(ruleSpaceComment, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 31 Spacing <- <SpaceComment*> */
		func() bool {
			{
				position220 := position
				depth++
			l221:
				{
					position222, tokenIndex222, depth222 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l222
					}
					goto l221
				l222:
					position, tokenIndex, depth = position222, tokenIndex222, depth222
				}
				depth--
				add(ruleSpacing, position220)
			}
			return true
		},
		/* 32 MustSpacing <- <SpaceComment+> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				if !_rules[ruleSpaceComment]() {
					goto l223
				}
			l225:
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l226
					}
					goto l225
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
				depth--
				add(ruleMustSpacing, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 33 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if buffer[position] != rune('#') {
					goto l227
				}
				position++
			l229:
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						if !_rules[ruleEndOfLine]() {
							goto l231
						}
						goto l230
					l231:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
					}
					if !matchDot() {
						goto l230
					}
					goto l229
				l230:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
				}
				if !_rules[ruleEndOfLine]() {
					goto l227
				}
				depth--
				add(ruleComment, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 34 Space <- <(' ' / '\t' / EndOfLine)> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l235
					}
					position++
					goto l234
				l235:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if buffer[position] != rune('\t') {
						goto l236
					}
					position++
					goto l234
				l236:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if !_rules[ruleEndOfLine]() {
						goto l232
					}
				}
			l234:
				depth--
				add(ruleSpace, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 35 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				{
					position239, tokenIndex239, depth239 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l240
					}
					position++
					if buffer[position] != rune('\n') {
						goto l240
					}
					position++
					goto l239
				l240:
					position, tokenIndex, depth = position239, tokenIndex239, depth239
					if buffer[position] != rune('\n') {
						goto l241
					}
					position++
					goto l239
				l241:
					position, tokenIndex, depth = position239, tokenIndex239, depth239
					if buffer[position] != rune('\r') {
						goto l237
					}
					position++
				}
			l239:
				depth--
				add(ruleEndOfLine, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 36 EndOfFile <- <!.> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				{
					position244, tokenIndex244, depth244 := position, tokenIndex, depth
					if !matchDot() {
						goto l244
					}
					goto l242
				l244:
					position, tokenIndex, depth = position244, tokenIndex244, depth244
				}
				depth--
				add(ruleEndOfFile, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 37 Action <- <('{' <ActionBody*> '}' Spacing)> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				if buffer[position] != rune('{') {
					goto l245
				}
				position++
				{
					position247 := position
					depth++
				l248:
					{
						position249, tokenIndex249, depth249 := position, tokenIndex, depth
						if !_rules[ruleActionBody]() {
							goto l249
						}
						goto l248
					l249:
						position, tokenIndex, depth = position249, tokenIndex249, depth249
					}
					depth--
					add(rulePegText, position247)
				}
				if buffer[position] != rune('}') {
					goto l245
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l245
				}
				depth--
				add(ruleAction, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 38 ActionBody <- <((!('{' / '}') .) / ('{' ActionBody* '}'))> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					{
						position254, tokenIndex254, depth254 := position, tokenIndex, depth
						{
							position255, tokenIndex255, depth255 := position, tokenIndex, depth
							if buffer[position] != rune('{') {
								goto l256
							}
							position++
							goto l255
						l256:
							position, tokenIndex, depth = position255, tokenIndex255, depth255
							if buffer[position] != rune('}') {
								goto l254
							}
							position++
						}
					l255:
						goto l253
					l254:
						position, tokenIndex, depth = position254, tokenIndex254, depth254
					}
					if !matchDot() {
						goto l253
					}
					goto l252
				l253:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
					if buffer[position] != rune('{') {
						goto l250
					}
					position++
				l257:
					{
						position258, tokenIndex258, depth258 := position, tokenIndex, depth
						if !_rules[ruleActionBody]() {
							goto l258
						}
						goto l257
					l258:
						position, tokenIndex, depth = position258, tokenIndex258, depth258
					}
					if buffer[position] != rune('}') {
						goto l250
					}
					position++
				}
			l252:
				depth--
				add(ruleActionBody, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 39 Begin <- <('<' Spacing)> */
		func() bool {
			position259, tokenIndex259, depth259 := position, tokenIndex, depth
			{
				position260 := position
				depth++
				if buffer[position] != rune('<') {
					goto l259
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l259
				}
				depth--
				add(ruleBegin, position260)
			}
			return true
		l259:
			position, tokenIndex, depth = position259, tokenIndex259, depth259
			return false
		},
		/* 40 End <- <('>' Spacing)> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				if buffer[position] != rune('>') {
					goto l261
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l261
				}
				depth--
				add(ruleEnd, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 42 Action0 <- <{ p.AddPackage(text) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 43 Action1 <- <{ p.AddPeg(text) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 44 Action2 <- <{ p.AddState(text) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 46 Action3 <- <{ p.AddImport(text) }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 47 Action4 <- <{ p.AddRule(text) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 48 Action5 <- <{ p.AddExpression() }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 49 Action6 <- <{ p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 50 Action7 <- <{ p.AddNil(); p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 51 Action8 <- <{ p.AddNil() }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 52 Action9 <- <{ p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 53 Action10 <- <{ p.AddPredicate(text) }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 54 Action11 <- <{ p.AddStateChange(text) }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 55 Action12 <- <{ p.AddPeekFor() }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 56 Action13 <- <{ p.AddPeekNot() }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 57 Action14 <- <{ p.AddQuery() }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 58 Action15 <- <{ p.AddStar() }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 59 Action16 <- <{ p.AddPlus() }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 60 Action17 <- <{ p.AddName(text) }> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 61 Action18 <- <{ p.AddDot() }> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 62 Action19 <- <{ p.AddAction(text) }> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 63 Action20 <- <{ p.AddPush() }> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 64 Action21 <- <{ p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 65 Action22 <- <{ p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
		/* 66 Action23 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 67 Action24 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
		/* 68 Action25 <- <{ p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction25, position)
			}
			return true
		},
		/* 69 Action26 <- <{ p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction26, position)
			}
			return true
		},
		/* 70 Action27 <- <{ p.AddRange() }> */
		func() bool {
			{
				add(ruleAction27, position)
			}
			return true
		},
		/* 71 Action28 <- <{ p.AddDoubleRange() }> */
		func() bool {
			{
				add(ruleAction28, position)
			}
			return true
		},
		/* 72 Action29 <- <{ p.AddCharacter(text) }> */
		func() bool {
			{
				add(ruleAction29, position)
			}
			return true
		},
		/* 73 Action30 <- <{ p.AddDoubleCharacter(text) }> */
		func() bool {
			{
				add(ruleAction30, position)
			}
			return true
		},
		/* 74 Action31 <- <{ p.AddCharacter(text) }> */
		func() bool {
			{
				add(ruleAction31, position)
			}
			return true
		},
		/* 75 Action32 <- <{ p.AddCharacter("\a") }> */
		func() bool {
			{
				add(ruleAction32, position)
			}
			return true
		},
		/* 76 Action33 <- <{ p.AddCharacter("\b") }> */
		func() bool {
			{
				add(ruleAction33, position)
			}
			return true
		},
		/* 77 Action34 <- <{ p.AddCharacter("\x1B") }> */
		func() bool {
			{
				add(ruleAction34, position)
			}
			return true
		},
		/* 78 Action35 <- <{ p.AddCharacter("\f") }> */
		func() bool {
			{
				add(ruleAction35, position)
			}
			return true
		},
		/* 79 Action36 <- <{ p.AddCharacter("\n") }> */
		func() bool {
			{
				add(ruleAction36, position)
			}
			return true
		},
		/* 80 Action37 <- <{ p.AddCharacter("\r") }> */
		func() bool {
			{
				add(ruleAction37, position)
			}
			return true
		},
		/* 81 Action38 <- <{ p.AddCharacter("\t") }> */
		func() bool {
			{
				add(ruleAction38, position)
			}
			return true
		},
		/* 82 Action39 <- <{ p.AddCharacter("\v") }> */
		func() bool {
			{
				add(ruleAction39, position)
			}
			return true
		},
		/* 83 Action40 <- <{ p.AddCharacter("'") }> */
		func() bool {
			{
				add(ruleAction40, position)
			}
			return true
		},
		/* 84 Action41 <- <{ p.AddCharacter("\"") }> */
		func() bool {
			{
				add(ruleAction41, position)
			}
			return true
		},
		/* 85 Action42 <- <{ p.AddCharacter("[") }> */
		func() bool {
			{
				add(ruleAction42, position)
			}
			return true
		},
		/* 86 Action43 <- <{ p.AddCharacter("]") }> */
		func() bool {
			{
				add(ruleAction43, position)
			}
			return true
		},
		/* 87 Action44 <- <{ p.AddCharacter("-") }> */
		func() bool {
			{
				add(ruleAction44, position)
			}
			return true
		},
		/* 88 Action45 <- <{ p.AddHexaCharacter(text) }> */
		func() bool {
			{
				add(ruleAction45, position)
			}
			return true
		},
		/* 89 Action46 <- <{ p.AddOctalCharacter(text) }> */
		func() bool {
			{
				add(ruleAction46, position)
			}
			return true
		},
		/* 90 Action47 <- <{ p.AddOctalCharacter(text) }> */
		func() bool {
			{
				add(ruleAction47, position)
			}
			return true
		},
		/* 91 Action48 <- <{ p.AddCharacter("\\") }> */
		func() bool {
			{
				add(ruleAction48, position)
			}
			return true
		},
	}
	p.rules = _rules
}
