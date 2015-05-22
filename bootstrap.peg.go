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
	rules  [91]func() bool
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
			rul3s[token.pegRule],
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
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)

		case ruleAction0:
			p.AddPackage(buffer[begin:end])
		case ruleAction1:
			p.AddPeg(buffer[begin:end])
		case ruleAction2:
			p.AddState(buffer[begin:end])
		case ruleAction3:
			p.AddImport(buffer[begin:end])
		case ruleAction4:
			p.AddRule(buffer[begin:end])
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
			p.AddPredicate(buffer[begin:end])
		case ruleAction11:
			p.AddStateChange(buffer[begin:end])
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
			p.AddName(buffer[begin:end])
		case ruleAction18:
			p.AddDot()
		case ruleAction19:
			p.AddAction(buffer[begin:end])
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
			p.AddCharacter(buffer[begin:end])
		case ruleAction30:
			p.AddDoubleCharacter(buffer[begin:end])
		case ruleAction31:
			p.AddCharacter(buffer[begin:end])
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
			p.AddHexaCharacter(buffer[begin:end])
		case ruleAction46:
			p.AddOctalCharacter(buffer[begin:end])
		case ruleAction47:
			p.AddOctalCharacter(buffer[begin:end])
		case ruleAction48:
			p.AddCharacter("\\")

		}
	}
	_, _, _ = buffer, begin, end
}

func (p *Peg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
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
		return &parseError{p}
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
				{
					add(ruleAction0, position)
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
						if !_rules[ruleSpacing]() {
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
							add(rulePegText, position6)
						}
						if buffer[position] != rune('"') {
							goto l4
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l4
						}
						{
							add(ruleAction3, position)
						}
						depth--
						add(ruleImport, position5)
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
				if !_rules[ruleMustSpacing]() {
					goto l0
				}
				if !_rules[ruleIdentifier]() {
					goto l0
				}
				{
					add(ruleAction1, position)
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
				{
					add(ruleAction2, position)
				}
				{
					position16 := position
					depth++
					if !_rules[ruleIdentifier]() {
						goto l0
					}
					{
						add(ruleAction4, position)
					}
					if !_rules[ruleLeftArrow]() {
						goto l0
					}
					if !_rules[ruleExpression]() {
						goto l0
					}
					{
						add(ruleAction5, position)
					}
					{
						position19, tokenIndex19, depth19 := position, tokenIndex, depth
						{
							position20, tokenIndex20, depth20 := position, tokenIndex, depth
							if !_rules[ruleIdentifier]() {
								goto l21
							}
							if !_rules[ruleLeftArrow]() {
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
					add(ruleDefinition, position16)
				}
			l14:
				{
					position15, tokenIndex15, depth15 := position, tokenIndex, depth
					{
						position23 := position
						depth++
						if !_rules[ruleIdentifier]() {
							goto l15
						}
						{
							add(ruleAction4, position)
						}
						if !_rules[ruleLeftArrow]() {
							goto l15
						}
						if !_rules[ruleExpression]() {
							goto l15
						}
						{
							add(ruleAction5, position)
						}
						{
							position26, tokenIndex26, depth26 := position, tokenIndex, depth
							{
								position27, tokenIndex27, depth27 := position, tokenIndex, depth
								if !_rules[ruleIdentifier]() {
									goto l28
								}
								if !_rules[ruleLeftArrow]() {
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
						add(ruleDefinition, position23)
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
					add(ruleEndOfFile, position30)
				}
				depth--
				add(ruleGrammar, position1)
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
					if !_rules[ruleSequence]() {
						goto l37
					}
				l38:
					{
						position39, tokenIndex39, depth39 := position, tokenIndex, depth
						if !_rules[ruleSlash]() {
							goto l39
						}
						if !_rules[ruleSequence]() {
							goto l39
						}
						{
							add(ruleAction6, position)
						}
						goto l38
					l39:
						position, tokenIndex, depth = position39, tokenIndex39, depth39
					}
					{
						position41, tokenIndex41, depth41 := position, tokenIndex, depth
						if !_rules[ruleSlash]() {
							goto l41
						}
						{
							add(ruleAction7, position)
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
						add(ruleAction8, position)
					}
				}
			l36:
				depth--
				add(ruleExpression, position35)
			}
			return true
		},
		/* 4 Sequence <- <(Prefix (Prefix Action9)*)> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if !_rules[rulePrefix]() {
					goto l45
				}
			l47:
				{
					position48, tokenIndex48, depth48 := position, tokenIndex, depth
					if !_rules[rulePrefix]() {
						goto l48
					}
					{
						add(ruleAction9, position)
					}
					goto l47
				l48:
					position, tokenIndex, depth = position48, tokenIndex48, depth48
				}
				depth--
				add(ruleSequence, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 5 Prefix <- <((And Action Action10) / (Not Action Action11) / ((&('!') (Not Suffix Action13)) | (&('&') (And Suffix Action12)) | (&('"' | '\'' | '(' | '.' | '<' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{') Suffix)))> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					if !_rules[ruleAnd]() {
						goto l53
					}
					if !_rules[ruleAction]() {
						goto l53
					}
					{
						add(ruleAction10, position)
					}
					goto l52
				l53:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
					if !_rules[ruleNot]() {
						goto l55
					}
					if !_rules[ruleAction]() {
						goto l55
					}
					{
						add(ruleAction11, position)
					}
					goto l52
				l55:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
					{
						switch buffer[position] {
						case '!':
							if !_rules[ruleNot]() {
								goto l50
							}
							if !_rules[ruleSuffix]() {
								goto l50
							}
							{
								add(ruleAction13, position)
							}
							break
						case '&':
							if !_rules[ruleAnd]() {
								goto l50
							}
							if !_rules[ruleSuffix]() {
								goto l50
							}
							{
								add(ruleAction12, position)
							}
							break
						default:
							if !_rules[ruleSuffix]() {
								goto l50
							}
							break
						}
					}

				}
			l52:
				depth--
				add(rulePrefix, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 6 Suffix <- <(Primary ((&('+') (Plus Action16)) | (&('*') (Star Action15)) | (&('?') (Question Action14)))?)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				{
					position62 := position
					depth++
					{
						switch buffer[position] {
						case '<':
							{
								position64 := position
								depth++
								if buffer[position] != rune('<') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								depth--
								add(ruleBegin, position64)
							}
							if !_rules[ruleExpression]() {
								goto l60
							}
							{
								position65 := position
								depth++
								if buffer[position] != rune('>') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								depth--
								add(ruleEnd, position65)
							}
							{
								add(ruleAction20, position)
							}
							break
						case '{':
							if !_rules[ruleAction]() {
								goto l60
							}
							{
								add(ruleAction19, position)
							}
							break
						case '.':
							{
								position68 := position
								depth++
								if buffer[position] != rune('.') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								depth--
								add(ruleDot, position68)
							}
							{
								add(ruleAction18, position)
							}
							break
						case '[':
							{
								position70 := position
								depth++
								{
									position71, tokenIndex71, depth71 := position, tokenIndex, depth
									if buffer[position] != rune('[') {
										goto l72
									}
									position++
									if buffer[position] != rune('[') {
										goto l72
									}
									position++
									{
										position73, tokenIndex73, depth73 := position, tokenIndex, depth
										{
											position75, tokenIndex75, depth75 := position, tokenIndex, depth
											if buffer[position] != rune('^') {
												goto l76
											}
											position++
											if !_rules[ruleDoubleRanges]() {
												goto l76
											}
											{
												add(ruleAction23, position)
											}
											goto l75
										l76:
											position, tokenIndex, depth = position75, tokenIndex75, depth75
											if !_rules[ruleDoubleRanges]() {
												goto l73
											}
										}
									l75:
										goto l74
									l73:
										position, tokenIndex, depth = position73, tokenIndex73, depth73
									}
								l74:
									if buffer[position] != rune(']') {
										goto l72
									}
									position++
									if buffer[position] != rune(']') {
										goto l72
									}
									position++
									goto l71
								l72:
									position, tokenIndex, depth = position71, tokenIndex71, depth71
									if buffer[position] != rune('[') {
										goto l60
									}
									position++
									{
										position78, tokenIndex78, depth78 := position, tokenIndex, depth
										{
											position80, tokenIndex80, depth80 := position, tokenIndex, depth
											if buffer[position] != rune('^') {
												goto l81
											}
											position++
											if !_rules[ruleRanges]() {
												goto l81
											}
											{
												add(ruleAction24, position)
											}
											goto l80
										l81:
											position, tokenIndex, depth = position80, tokenIndex80, depth80
											if !_rules[ruleRanges]() {
												goto l78
											}
										}
									l80:
										goto l79
									l78:
										position, tokenIndex, depth = position78, tokenIndex78, depth78
									}
								l79:
									if buffer[position] != rune(']') {
										goto l60
									}
									position++
								}
							l71:
								if !_rules[ruleSpacing]() {
									goto l60
								}
								depth--
								add(ruleClass, position70)
							}
							break
						case '"', '\'':
							{
								position83 := position
								depth++
								{
									position84, tokenIndex84, depth84 := position, tokenIndex, depth
									if buffer[position] != rune('\'') {
										goto l85
									}
									position++
									{
										position86, tokenIndex86, depth86 := position, tokenIndex, depth
										{
											position88, tokenIndex88, depth88 := position, tokenIndex, depth
											if buffer[position] != rune('\'') {
												goto l88
											}
											position++
											goto l86
										l88:
											position, tokenIndex, depth = position88, tokenIndex88, depth88
										}
										if !_rules[ruleChar]() {
											goto l86
										}
										goto l87
									l86:
										position, tokenIndex, depth = position86, tokenIndex86, depth86
									}
								l87:
								l89:
									{
										position90, tokenIndex90, depth90 := position, tokenIndex, depth
										{
											position91, tokenIndex91, depth91 := position, tokenIndex, depth
											if buffer[position] != rune('\'') {
												goto l91
											}
											position++
											goto l90
										l91:
											position, tokenIndex, depth = position91, tokenIndex91, depth91
										}
										if !_rules[ruleChar]() {
											goto l90
										}
										{
											add(ruleAction21, position)
										}
										goto l89
									l90:
										position, tokenIndex, depth = position90, tokenIndex90, depth90
									}
									if buffer[position] != rune('\'') {
										goto l85
									}
									position++
									if !_rules[ruleSpacing]() {
										goto l85
									}
									goto l84
								l85:
									position, tokenIndex, depth = position84, tokenIndex84, depth84
									if buffer[position] != rune('"') {
										goto l60
									}
									position++
									{
										position93, tokenIndex93, depth93 := position, tokenIndex, depth
										{
											position95, tokenIndex95, depth95 := position, tokenIndex, depth
											if buffer[position] != rune('"') {
												goto l95
											}
											position++
											goto l93
										l95:
											position, tokenIndex, depth = position95, tokenIndex95, depth95
										}
										if !_rules[ruleDoubleChar]() {
											goto l93
										}
										goto l94
									l93:
										position, tokenIndex, depth = position93, tokenIndex93, depth93
									}
								l94:
								l96:
									{
										position97, tokenIndex97, depth97 := position, tokenIndex, depth
										{
											position98, tokenIndex98, depth98 := position, tokenIndex, depth
											if buffer[position] != rune('"') {
												goto l98
											}
											position++
											goto l97
										l98:
											position, tokenIndex, depth = position98, tokenIndex98, depth98
										}
										if !_rules[ruleDoubleChar]() {
											goto l97
										}
										{
											add(ruleAction22, position)
										}
										goto l96
									l97:
										position, tokenIndex, depth = position97, tokenIndex97, depth97
									}
									if buffer[position] != rune('"') {
										goto l60
									}
									position++
									if !_rules[ruleSpacing]() {
										goto l60
									}
								}
							l84:
								depth--
								add(ruleLiteral, position83)
							}
							break
						case '(':
							{
								position100 := position
								depth++
								if buffer[position] != rune('(') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								depth--
								add(ruleOpen, position100)
							}
							if !_rules[ruleExpression]() {
								goto l60
							}
							{
								position101 := position
								depth++
								if buffer[position] != rune(')') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								depth--
								add(ruleClose, position101)
							}
							break
						default:
							if !_rules[ruleIdentifier]() {
								goto l60
							}
							{
								position102, tokenIndex102, depth102 := position, tokenIndex, depth
								if !_rules[ruleLeftArrow]() {
									goto l102
								}
								goto l60
							l102:
								position, tokenIndex, depth = position102, tokenIndex102, depth102
							}
							{
								add(ruleAction17, position)
							}
							break
						}
					}

					depth--
					add(rulePrimary, position62)
				}
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '+':
							{
								position107 := position
								depth++
								if buffer[position] != rune('+') {
									goto l104
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l104
								}
								depth--
								add(rulePlus, position107)
							}
							{
								add(ruleAction16, position)
							}
							break
						case '*':
							{
								position109 := position
								depth++
								if buffer[position] != rune('*') {
									goto l104
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l104
								}
								depth--
								add(ruleStar, position109)
							}
							{
								add(ruleAction15, position)
							}
							break
						default:
							{
								position111 := position
								depth++
								if buffer[position] != rune('?') {
									goto l104
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l104
								}
								depth--
								add(ruleQuestion, position111)
							}
							{
								add(ruleAction14, position)
							}
							break
						}
					}

					goto l105
				l104:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
				}
			l105:
				depth--
				add(ruleSuffix, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 7 Primary <- <((&('<') (Begin Expression End Action20)) | (&('{') (Action Action19)) | (&('.') (Dot Action18)) | (&('[') Class) | (&('"' | '\'') Literal) | (&('(') (Open Expression Close)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (Identifier !LeftArrow Action17)))> */
		nil,
		/* 8 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				{
					position116 := position
					depth++
					if !_rules[ruleIdentStart]() {
						goto l114
					}
				l117:
					{
						position118, tokenIndex118, depth118 := position, tokenIndex, depth
						{
							position119 := position
							depth++
							{
								position120, tokenIndex120, depth120 := position, tokenIndex, depth
								if !_rules[ruleIdentStart]() {
									goto l121
								}
								goto l120
							l121:
								position, tokenIndex, depth = position120, tokenIndex120, depth120
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l118
								}
								position++
							}
						l120:
							depth--
							add(ruleIdentCont, position119)
						}
						goto l117
					l118:
						position, tokenIndex, depth = position118, tokenIndex118, depth118
					}
					depth--
					add(rulePegText, position116)
				}
				if !_rules[ruleSpacing]() {
					goto l114
				}
				depth--
				add(ruleIdentifier, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 9 IdentStart <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l122
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l122
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l122
						}
						position++
						break
					}
				}

				depth--
				add(ruleIdentStart, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 10 IdentCont <- <(IdentStart / [0-9])> */
		nil,
		/* 11 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char Action21)* '\'' Spacing) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar Action22)* '"' Spacing))> */
		nil,
		/* 12 Class <- <((('[' '[' (('^' DoubleRanges Action23) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges Action24) / Ranges)? ']')) Spacing)> */
		nil,
		/* 13 Ranges <- <(!']' Range (!']' Range Action25)*)> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				{
					position130, tokenIndex130, depth130 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l130
					}
					position++
					goto l128
				l130:
					position, tokenIndex, depth = position130, tokenIndex130, depth130
				}
				if !_rules[ruleRange]() {
					goto l128
				}
			l131:
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					{
						position133, tokenIndex133, depth133 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l133
						}
						position++
						goto l132
					l133:
						position, tokenIndex, depth = position133, tokenIndex133, depth133
					}
					if !_rules[ruleRange]() {
						goto l132
					}
					{
						add(ruleAction25, position)
					}
					goto l131
				l132:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
				}
				depth--
				add(ruleRanges, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 14 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange Action26)*)> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l137
					}
					position++
					if buffer[position] != rune(']') {
						goto l137
					}
					position++
					goto l135
				l137:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
				}
				if !_rules[ruleDoubleRange]() {
					goto l135
				}
			l138:
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					{
						position140, tokenIndex140, depth140 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l140
						}
						position++
						if buffer[position] != rune(']') {
							goto l140
						}
						position++
						goto l139
					l140:
						position, tokenIndex, depth = position140, tokenIndex140, depth140
					}
					if !_rules[ruleDoubleRange]() {
						goto l139
					}
					{
						add(ruleAction26, position)
					}
					goto l138
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
				depth--
				add(ruleDoubleRanges, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 15 Range <- <((Char '-' Char Action27) / Char)> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if !_rules[ruleChar]() {
						goto l145
					}
					if buffer[position] != rune('-') {
						goto l145
					}
					position++
					if !_rules[ruleChar]() {
						goto l145
					}
					{
						add(ruleAction27, position)
					}
					goto l144
				l145:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if !_rules[ruleChar]() {
						goto l142
					}
				}
			l144:
				depth--
				add(ruleRange, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 16 DoubleRange <- <((Char '-' Char Action28) / DoubleChar)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if !_rules[ruleChar]() {
						goto l150
					}
					if buffer[position] != rune('-') {
						goto l150
					}
					position++
					if !_rules[ruleChar]() {
						goto l150
					}
					{
						add(ruleAction28, position)
					}
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if !_rules[ruleDoubleChar]() {
						goto l147
					}
				}
			l149:
				depth--
				add(ruleDoubleRange, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 17 Char <- <(Escape / (!'\\' <.> Action29))> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if !_rules[ruleEscape]() {
						goto l155
					}
					goto l154
				l155:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
					{
						position156, tokenIndex156, depth156 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l156
						}
						position++
						goto l152
					l156:
						position, tokenIndex, depth = position156, tokenIndex156, depth156
					}
					{
						position157 := position
						depth++
						if !matchDot() {
							goto l152
						}
						depth--
						add(rulePegText, position157)
					}
					{
						add(ruleAction29, position)
					}
				}
			l154:
				depth--
				add(ruleChar, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 18 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> Action30) / (!'\\' <.> Action31))> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				{
					position161, tokenIndex161, depth161 := position, tokenIndex, depth
					if !_rules[ruleEscape]() {
						goto l162
					}
					goto l161
				l162:
					position, tokenIndex, depth = position161, tokenIndex161, depth161
					{
						position164 := position
						depth++
						{
							position165, tokenIndex165, depth165 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l166
							}
							position++
							goto l165
						l166:
							position, tokenIndex, depth = position165, tokenIndex165, depth165
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l163
							}
							position++
						}
					l165:
						depth--
						add(rulePegText, position164)
					}
					{
						add(ruleAction30, position)
					}
					goto l161
				l163:
					position, tokenIndex, depth = position161, tokenIndex161, depth161
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l168
						}
						position++
						goto l159
					l168:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
					}
					{
						position169 := position
						depth++
						if !matchDot() {
							goto l159
						}
						depth--
						add(rulePegText, position169)
					}
					{
						add(ruleAction31, position)
					}
				}
			l161:
				depth--
				add(ruleDoubleChar, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 19 Escape <- <(('\\' ('a' / 'A') Action32) / ('\\' ('b' / 'B') Action33) / ('\\' ('e' / 'E') Action34) / ('\\' ('f' / 'F') Action35) / ('\\' ('n' / 'N') Action36) / ('\\' ('r' / 'R') Action37) / ('\\' ('t' / 'T') Action38) / ('\\' ('v' / 'V') Action39) / ('\\' '\'' Action40) / ('\\' '"' Action41) / ('\\' '[' Action42) / ('\\' ']' Action43) / ('\\' '-' Action44) / ('\\' ('0' ('x' / 'X')) <((&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))+> Action45) / ('\\' <([0-3] [0-7] [0-7])> Action46) / ('\\' <([0-7] [0-7]?)> Action47) / ('\\' '\\' Action48))> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l174
					}
					position++
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l176
						}
						position++
						goto l175
					l176:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if buffer[position] != rune('A') {
							goto l174
						}
						position++
					}
				l175:
					{
						add(ruleAction32, position)
					}
					goto l173
				l174:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l178
					}
					position++
					{
						position179, tokenIndex179, depth179 := position, tokenIndex, depth
						if buffer[position] != rune('b') {
							goto l180
						}
						position++
						goto l179
					l180:
						position, tokenIndex, depth = position179, tokenIndex179, depth179
						if buffer[position] != rune('B') {
							goto l178
						}
						position++
					}
				l179:
					{
						add(ruleAction33, position)
					}
					goto l173
				l178:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l182
					}
					position++
					{
						position183, tokenIndex183, depth183 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l184
						}
						position++
						goto l183
					l184:
						position, tokenIndex, depth = position183, tokenIndex183, depth183
						if buffer[position] != rune('E') {
							goto l182
						}
						position++
					}
				l183:
					{
						add(ruleAction34, position)
					}
					goto l173
				l182:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l186
					}
					position++
					{
						position187, tokenIndex187, depth187 := position, tokenIndex, depth
						if buffer[position] != rune('f') {
							goto l188
						}
						position++
						goto l187
					l188:
						position, tokenIndex, depth = position187, tokenIndex187, depth187
						if buffer[position] != rune('F') {
							goto l186
						}
						position++
					}
				l187:
					{
						add(ruleAction35, position)
					}
					goto l173
				l186:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l190
					}
					position++
					{
						position191, tokenIndex191, depth191 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l192
						}
						position++
						goto l191
					l192:
						position, tokenIndex, depth = position191, tokenIndex191, depth191
						if buffer[position] != rune('N') {
							goto l190
						}
						position++
					}
				l191:
					{
						add(ruleAction36, position)
					}
					goto l173
				l190:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l194
					}
					position++
					{
						position195, tokenIndex195, depth195 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l196
						}
						position++
						goto l195
					l196:
						position, tokenIndex, depth = position195, tokenIndex195, depth195
						if buffer[position] != rune('R') {
							goto l194
						}
						position++
					}
				l195:
					{
						add(ruleAction37, position)
					}
					goto l173
				l194:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l198
					}
					position++
					{
						position199, tokenIndex199, depth199 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l200
						}
						position++
						goto l199
					l200:
						position, tokenIndex, depth = position199, tokenIndex199, depth199
						if buffer[position] != rune('T') {
							goto l198
						}
						position++
					}
				l199:
					{
						add(ruleAction38, position)
					}
					goto l173
				l198:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l202
					}
					position++
					{
						position203, tokenIndex203, depth203 := position, tokenIndex, depth
						if buffer[position] != rune('v') {
							goto l204
						}
						position++
						goto l203
					l204:
						position, tokenIndex, depth = position203, tokenIndex203, depth203
						if buffer[position] != rune('V') {
							goto l202
						}
						position++
					}
				l203:
					{
						add(ruleAction39, position)
					}
					goto l173
				l202:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l206
					}
					position++
					if buffer[position] != rune('\'') {
						goto l206
					}
					position++
					{
						add(ruleAction40, position)
					}
					goto l173
				l206:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l208
					}
					position++
					if buffer[position] != rune('"') {
						goto l208
					}
					position++
					{
						add(ruleAction41, position)
					}
					goto l173
				l208:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l210
					}
					position++
					if buffer[position] != rune('[') {
						goto l210
					}
					position++
					{
						add(ruleAction42, position)
					}
					goto l173
				l210:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l212
					}
					position++
					if buffer[position] != rune(']') {
						goto l212
					}
					position++
					{
						add(ruleAction43, position)
					}
					goto l173
				l212:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l214
					}
					position++
					if buffer[position] != rune('-') {
						goto l214
					}
					position++
					{
						add(ruleAction44, position)
					}
					goto l173
				l214:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l216
					}
					position++
					if buffer[position] != rune('0') {
						goto l216
					}
					position++
					{
						position217, tokenIndex217, depth217 := position, tokenIndex, depth
						if buffer[position] != rune('x') {
							goto l218
						}
						position++
						goto l217
					l218:
						position, tokenIndex, depth = position217, tokenIndex217, depth217
						if buffer[position] != rune('X') {
							goto l216
						}
						position++
					}
				l217:
					{
						position219 := position
						depth++
						{
							switch buffer[position] {
							case 'A', 'B', 'C', 'D', 'E', 'F':
								if c := buffer[position]; c < rune('A') || c > rune('F') {
									goto l216
								}
								position++
								break
							case 'a', 'b', 'c', 'd', 'e', 'f':
								if c := buffer[position]; c < rune('a') || c > rune('f') {
									goto l216
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l216
								}
								position++
								break
							}
						}

					l220:
						{
							position221, tokenIndex221, depth221 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case 'A', 'B', 'C', 'D', 'E', 'F':
									if c := buffer[position]; c < rune('A') || c > rune('F') {
										goto l221
									}
									position++
									break
								case 'a', 'b', 'c', 'd', 'e', 'f':
									if c := buffer[position]; c < rune('a') || c > rune('f') {
										goto l221
									}
									position++
									break
								default:
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l221
									}
									position++
									break
								}
							}

							goto l220
						l221:
							position, tokenIndex, depth = position221, tokenIndex221, depth221
						}
						depth--
						add(rulePegText, position219)
					}
					{
						add(ruleAction45, position)
					}
					goto l173
				l216:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l225
					}
					position++
					{
						position226 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('3') {
							goto l225
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l225
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l225
						}
						position++
						depth--
						add(rulePegText, position226)
					}
					{
						add(ruleAction46, position)
					}
					goto l173
				l225:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l228
					}
					position++
					{
						position229 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l228
						}
						position++
						{
							position230, tokenIndex230, depth230 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('7') {
								goto l230
							}
							position++
							goto l231
						l230:
							position, tokenIndex, depth = position230, tokenIndex230, depth230
						}
					l231:
						depth--
						add(rulePegText, position229)
					}
					{
						add(ruleAction47, position)
					}
					goto l173
				l228:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if buffer[position] != rune('\\') {
						goto l171
					}
					position++
					if buffer[position] != rune('\\') {
						goto l171
					}
					position++
					{
						add(ruleAction48, position)
					}
				}
			l173:
				depth--
				add(ruleEscape, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 20 LeftArrow <- <((('<' '-') / '←') Spacing)> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					if buffer[position] != rune('<') {
						goto l237
					}
					position++
					if buffer[position] != rune('-') {
						goto l237
					}
					position++
					goto l236
				l237:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
					if buffer[position] != rune('←') {
						goto l234
					}
					position++
				}
			l236:
				if !_rules[ruleSpacing]() {
					goto l234
				}
				depth--
				add(ruleLeftArrow, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 21 Slash <- <('/' Spacing)> */
		func() bool {
			position238, tokenIndex238, depth238 := position, tokenIndex, depth
			{
				position239 := position
				depth++
				if buffer[position] != rune('/') {
					goto l238
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l238
				}
				depth--
				add(ruleSlash, position239)
			}
			return true
		l238:
			position, tokenIndex, depth = position238, tokenIndex238, depth238
			return false
		},
		/* 22 And <- <('&' Spacing)> */
		func() bool {
			position240, tokenIndex240, depth240 := position, tokenIndex, depth
			{
				position241 := position
				depth++
				if buffer[position] != rune('&') {
					goto l240
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l240
				}
				depth--
				add(ruleAnd, position241)
			}
			return true
		l240:
			position, tokenIndex, depth = position240, tokenIndex240, depth240
			return false
		},
		/* 23 Not <- <('!' Spacing)> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if buffer[position] != rune('!') {
					goto l242
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l242
				}
				depth--
				add(ruleNot, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
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
		/* 30 SpaceComment <- <(Space / Comment)> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					{
						position254 := position
						depth++
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l253
								}
								position++
								break
							case ' ':
								if buffer[position] != rune(' ') {
									goto l253
								}
								position++
								break
							default:
								if !_rules[ruleEndOfLine]() {
									goto l253
								}
								break
							}
						}

						depth--
						add(ruleSpace, position254)
					}
					goto l252
				l253:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
					{
						position256 := position
						depth++
						if buffer[position] != rune('#') {
							goto l250
						}
						position++
					l257:
						{
							position258, tokenIndex258, depth258 := position, tokenIndex, depth
							{
								position259, tokenIndex259, depth259 := position, tokenIndex, depth
								if !_rules[ruleEndOfLine]() {
									goto l259
								}
								goto l258
							l259:
								position, tokenIndex, depth = position259, tokenIndex259, depth259
							}
							if !matchDot() {
								goto l258
							}
							goto l257
						l258:
							position, tokenIndex, depth = position258, tokenIndex258, depth258
						}
						if !_rules[ruleEndOfLine]() {
							goto l250
						}
						depth--
						add(ruleComment, position256)
					}
				}
			l252:
				depth--
				add(ruleSpaceComment, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 31 Spacing <- <SpaceComment*> */
		func() bool {
			{
				position261 := position
				depth++
			l262:
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l263
					}
					goto l262
				l263:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
				}
				depth--
				add(ruleSpacing, position261)
			}
			return true
		},
		/* 32 MustSpacing <- <SpaceComment+> */
		func() bool {
			position264, tokenIndex264, depth264 := position, tokenIndex, depth
			{
				position265 := position
				depth++
				if !_rules[ruleSpaceComment]() {
					goto l264
				}
			l266:
				{
					position267, tokenIndex267, depth267 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l267
					}
					goto l266
				l267:
					position, tokenIndex, depth = position267, tokenIndex267, depth267
				}
				depth--
				add(ruleMustSpacing, position265)
			}
			return true
		l264:
			position, tokenIndex, depth = position264, tokenIndex264, depth264
			return false
		},
		/* 33 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		nil,
		/* 34 Space <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine))> */
		nil,
		/* 35 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position270, tokenIndex270, depth270 := position, tokenIndex, depth
			{
				position271 := position
				depth++
				{
					position272, tokenIndex272, depth272 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l273
					}
					position++
					if buffer[position] != rune('\n') {
						goto l273
					}
					position++
					goto l272
				l273:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
					if buffer[position] != rune('\n') {
						goto l274
					}
					position++
					goto l272
				l274:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
					if buffer[position] != rune('\r') {
						goto l270
					}
					position++
				}
			l272:
				depth--
				add(ruleEndOfLine, position271)
			}
			return true
		l270:
			position, tokenIndex, depth = position270, tokenIndex270, depth270
			return false
		},
		/* 36 EndOfFile <- <!.> */
		nil,
		/* 37 Action <- <('{' <(!'}' .)*> '}' Spacing)> */
		func() bool {
			position276, tokenIndex276, depth276 := position, tokenIndex, depth
			{
				position277 := position
				depth++
				if buffer[position] != rune('{') {
					goto l276
				}
				position++
				{
					position278 := position
					depth++
				l279:
					{
						position280, tokenIndex280, depth280 := position, tokenIndex, depth
						{
							position281, tokenIndex281, depth281 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l281
							}
							position++
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
					depth--
					add(rulePegText, position278)
				}
				if buffer[position] != rune('}') {
					goto l276
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l276
				}
				depth--
				add(ruleAction, position277)
			}
			return true
		l276:
			position, tokenIndex, depth = position276, tokenIndex276, depth276
			return false
		},
		/* 38 Begin <- <('<' Spacing)> */
		nil,
		/* 39 End <- <('>' Spacing)> */
		nil,
		/* 41 Action0 <- <{ p.AddPackage(buffer[begin:end]) }> */
		nil,
		/* 42 Action1 <- <{ p.AddPeg(buffer[begin:end]) }> */
		nil,
		/* 43 Action2 <- <{ p.AddState(buffer[begin:end]) }> */
		nil,
		nil,
		/* 45 Action3 <- <{ p.AddImport(buffer[begin:end]) }> */
		nil,
		/* 46 Action4 <- <{ p.AddRule(buffer[begin:end]) }> */
		nil,
		/* 47 Action5 <- <{ p.AddExpression() }> */
		nil,
		/* 48 Action6 <- <{ p.AddAlternate() }> */
		nil,
		/* 49 Action7 <- <{ p.AddNil(); p.AddAlternate() }> */
		nil,
		/* 50 Action8 <- <{ p.AddNil() }> */
		nil,
		/* 51 Action9 <- <{ p.AddSequence() }> */
		nil,
		/* 52 Action10 <- <{ p.AddPredicate(buffer[begin:end]) }> */
		nil,
		/* 53 Action11 <- <{ p.AddStateChange(buffer[begin:end]) }> */
		nil,
		/* 54 Action12 <- <{ p.AddPeekFor() }> */
		nil,
		/* 55 Action13 <- <{ p.AddPeekNot() }> */
		nil,
		/* 56 Action14 <- <{ p.AddQuery() }> */
		nil,
		/* 57 Action15 <- <{ p.AddStar() }> */
		nil,
		/* 58 Action16 <- <{ p.AddPlus() }> */
		nil,
		/* 59 Action17 <- <{ p.AddName(buffer[begin:end]) }> */
		nil,
		/* 60 Action18 <- <{ p.AddDot() }> */
		nil,
		/* 61 Action19 <- <{ p.AddAction(buffer[begin:end]) }> */
		nil,
		/* 62 Action20 <- <{ p.AddPush() }> */
		nil,
		/* 63 Action21 <- <{ p.AddSequence() }> */
		nil,
		/* 64 Action22 <- <{ p.AddSequence() }> */
		nil,
		/* 65 Action23 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 66 Action24 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 67 Action25 <- <{ p.AddAlternate() }> */
		nil,
		/* 68 Action26 <- <{ p.AddAlternate() }> */
		nil,
		/* 69 Action27 <- <{ p.AddRange() }> */
		nil,
		/* 70 Action28 <- <{ p.AddDoubleRange() }> */
		nil,
		/* 71 Action29 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 72 Action30 <- <{ p.AddDoubleCharacter(buffer[begin:end]) }> */
		nil,
		/* 73 Action31 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 74 Action32 <- <{ p.AddCharacter("\a") }> */
		nil,
		/* 75 Action33 <- <{ p.AddCharacter("\b") }> */
		nil,
		/* 76 Action34 <- <{ p.AddCharacter("\x1B") }> */
		nil,
		/* 77 Action35 <- <{ p.AddCharacter("\f") }> */
		nil,
		/* 78 Action36 <- <{ p.AddCharacter("\n") }> */
		nil,
		/* 79 Action37 <- <{ p.AddCharacter("\r") }> */
		nil,
		/* 80 Action38 <- <{ p.AddCharacter("\t") }> */
		nil,
		/* 81 Action39 <- <{ p.AddCharacter("\v") }> */
		nil,
		/* 82 Action40 <- <{ p.AddCharacter("'") }> */
		nil,
		/* 83 Action41 <- <{ p.AddCharacter("\"") }> */
		nil,
		/* 84 Action42 <- <{ p.AddCharacter("[") }> */
		nil,
		/* 85 Action43 <- <{ p.AddCharacter("]") }> */
		nil,
		/* 86 Action44 <- <{ p.AddCharacter("-") }> */
		nil,
		/* 87 Action45 <- <{ p.AddHexaCharacter(buffer[begin:end]) }> */
		nil,
		/* 88 Action46 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 89 Action47 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		nil,
		/* 90 Action48 <- <{ p.AddCharacter("\\") }> */
		nil,
	}
	p.rules = _rules
}
