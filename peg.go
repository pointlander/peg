// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"github.com/pointlander/jetset"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const PEG_HEADER_TEMPLATE = `package {{.PackageName}}

import (
	{{range .Imports}}"{{.}}"
	{{end}}
)

const end_symbol rune = {{.EndSymbol}}

/* The rule types inferred from the grammar are below. */
type pegRule {{.PegRuleType}}

const (
	ruleUnknown pegRule = iota
	{{range .RuleNames}}rule{{.String}}
	{{end}}
	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string {
	"Unknown",
	{{range .RuleNames}}"{{.String}}",
	{{end}}
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
			node.up.print(depth + 1, buffer)
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

{{range .Sizes}}

/* ${@} bit structure for abstract syntax tree */
type token{{.}} struct {
	pegRule
	begin, end, next uint{{.}}
}

func (t *token{{.}}) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token{{.}}) isParentOf(u token{{.}}) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token{{.}}) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token{{.}}) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens{{.}} struct {
	tree		[]token{{.}}
	ordered		[][]token{{.}}
}

func (t *tokens{{.}}) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens{{.}}) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens{{.}}) Order() [][]token{{.}} {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int{{.}}, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth + 1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token{{.}}, len(depths)), make([]token{{.}}, len(t.tree) + len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint{{.}}(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state{{.}} struct {
	token{{.}}
	depths []int{{.}}
	leaf bool
}

func (t *tokens{{.}}) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32:<-tokens}}
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

func (t *tokens{{.}}) PreOrder() (<-chan state{{.}}, [][]token{{.}}) {
	s, ordered := make(chan state{{.}}, 6), t.Order()
	go func() {
		var states [8]state{{.}}
		for i, _ := range states {
			states[i].depths = make([]int{{.}}, len(ordered))
		}
		depths, state, depth := make([]int{{.}}, len(ordered)), 0, 1
		write := func(t token{{.}}, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state + 1) % 8, t.pegRule, t.begin, t.end, uint{{.}}(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token{{.}} = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth - 1][depths[depth - 1] - 1], ordered[depth][depths[depth]]
		depthFirstSearch: for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i - 1], depths[depth - 1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth - 1][j - 2].isParentOf(c)) {
						if c.end != b.begin {
							write(token{{.}} {pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token{{.}} {pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
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
					write(token{{.}} {pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth - 1][depths[depth - 1] - 1], a, ordered[depth][depths[depth]]
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

func (t *tokens{{.}}) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i] - 1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i] - 1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max + 1 < i {
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
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i] - 1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens{{.}}) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens{{.}}) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token{{.}}{pegRule: rule, begin: uint{{.}}(begin), end: uint{{.}}(end), next: uint{{.}}(depth)}
}

func (t *tokens{{.}}) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens{{.}}) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length - 1
	for i, _ := range tokens {
		o := ordered[length - i]
		if len(o) > 1 {
			tokens[i] = o[len(o) - 2].getToken32()
		}
	}
	return tokens
}
{{end}}

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
		expanded := make([]token32, 2 * len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type {{.StructName}} struct {
	{{.StructVariables}}
	Buffer		string
	buffer		[]rune
	rules		[{{.RulesCount}}]func() bool
	Parse		func(rule ...int) error
	Reset		func()
	Pretty 	bool
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int] textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

	search: for i, c := range buffer {
		if c == '\n' {line, symbol = line + 1, 0} else {symbol++}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {if i != positions[j] {continue search}}
			break search
		}
 	}

	return translations
}

type parseError struct {
	p *{{.StructName}}
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2 * len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p + 1
		positions[p], p = int(token.end), p + 1
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

func (p *{{.StructName}}) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *{{.StructName}}) Highlighter() {
	p.tokenTree.PrintSyntax()
}

{{if .HasActions}}
func (p *{{.StructName}}) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.tokenTree.Tokens() {
		switch (token.pegRule) {
		{{if .HasPush}}
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])
		{{end}}
		{{range .Actions}}case ruleAction{{.GetId}}:
			{{.String}}
		{{end}}
		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}
{{end}}

func (p *{{.StructName}}) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer) - 1] != end_symbol {
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

	{{if .HasDot}}
	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}
	{{end}}

	{{if .HasCharacter}}
	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/
	{{end}}

	{{if .HasString}}
	matchString := func(s string) bool {
		i := position
		for _, c := range s {
			if buffer[i] != c {
				return false
			}
			i++
		}
		position = i
		return true
	}
	{{end}}

	{{if .HasRange}}
	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/
	{{end}}

	_rules = [...]func() bool {
		nil,`

type Type uint8

const (
	TypeUnknown Type = iota
	TypeRule
	TypeName
	TypeDot
	TypeCharacter
	TypeRange
	TypeString
	TypePredicate
	TypeStateChange
	TypeCommit
	TypeAction
	TypePackage
	TypeImport
	TypeState
	TypeAlternate
	TypeUnorderedAlternate
	TypeSequence
	TypePeekFor
	TypePeekNot
	TypeQuery
	TypeStar
	TypePlus
	TypePeg
	TypePush
	TypeImplicitPush
	TypeNil
	TypeLast
)

var TypeMap = [...]string{
	"TypeUnknown",
	"TypeRule",
	"TypeName",
	"TypeDot",
	"TypeCharacter",
	"TypeRange",
	"TypeString",
	"TypePredicate",
	"TypeCommit",
	"TypeAction",
	"TypePackage",
	"TypeImport",
	"TypeState",
	"TypeAlternate",
	"TypeUnorderedAlternate",
	"TypeSequence",
	"TypePeekFor",
	"TypePeekNot",
	"TypeQuery",
	"TypeStar",
	"TypePlus",
	"TypePeg",
	"TypePush",
	"TypeImplicitPush",
	"TypeNil",
	"TypeLast"}

func (t Type) GetType() Type {
	return t
}

type Node interface {
	fmt.Stringer
	debug()

	Escaped() string
	SetString(s string)

	GetType() Type
	SetType(t Type)

	GetId() int
	SetId(id int)

	Init()
	Front() *node
	Next() *node
	PushFront(value *node)
	PopFront() *node
	PushBack(value *node)
	Len() int
	Copy() *node
	Slice() []*node
}

type node struct {
	Type
	string
	id int

	front  *node
	back   *node
	length int

	/* use hash table here instead of Copy? */
	next *node
}

func (n *node) String() string {
	return n.string
}

func (n *node) debug() {
	if len(n.string) == 1 {
		fmt.Printf("%v %v '%v' %d\n", n.id, TypeMap[n.Type], n.string, n.string[0])
	} else {
		fmt.Printf("%v %v '%v'\n", n.id, TypeMap[n.Type], n.string)
	}
}

func (n *node) Escaped() string {
	return escape(n.string)
}

func (n *node) SetString(s string) {
	n.string = s
}

func (n *node) SetType(t Type) {
	n.Type = t
}

func (n *node) GetId() int {
	return n.id
}

func (n *node) SetId(id int) {
	n.id = id
}

func (n *node) Init() {
	n.front = nil
	n.back = nil
	n.length = 0
}

func (n *node) Front() *node {
	return n.front
}

func (n *node) Next() *node {
	return n.next
}

func (n *node) PushFront(value *node) {
	if n.back == nil {
		n.back = value
	} else {
		value.next = n.front
	}
	n.front = value
	n.length++
}

func (n *node) PopFront() *node {
	front := n.front

	switch true {
	case front == nil:
		panic("tree is empty")
	case front == n.back:
		n.front, n.back = nil, nil
	default:
		n.front, front.next = front.next, nil
	}

	n.length--
	return front
}

func (n *node) PushBack(value *node) {
	if n.front == nil {
		n.front = value
	} else {
		n.back.next = value
	}
	n.back = value
	n.length++
}

func (n *node) Len() (c int) {
	return n.length
}

func (n *node) Copy() *node {
	return &node{Type: n.Type, string: n.string, id: n.id, front: n.front, back: n.back, length: n.length}
}

func (n *node) Slice() []*node {
	s := make([]*node, n.length)
	for element, i := n.Front(), 0; element != nil; element, i = element.Next(), i+1 {
		s[i] = element
	}
	return s
}

/* A tree data structure into which a PEG can be parsed. */
type Tree struct {
	Rules      map[string]Node
	rulesCount map[string]uint
	node
	inline, _switch bool

	RuleNames       []Node
	Sizes           [1]int
	PackageName     string
	Imports         []string
	EndSymbol       rune
	PegRuleType     string
	StructName      string
	StructVariables string
	RulesCount      int
	Bits            int
	HasActions      bool
	Actions         []Node
	HasPush         bool
	HasCommit       bool
	HasDot          bool
	HasCharacter    bool
	HasString       bool
	HasRange        bool
}

func New(inline, _switch bool) *Tree {
	return &Tree{Rules: make(map[string]Node),
		Sizes:      [1]int{32},
		rulesCount: make(map[string]uint),
		inline:     inline,
		_switch:    _switch}
}

func (t *Tree) AddRule(name string) {
	t.PushFront(&node{Type: TypeRule, string: name, id: t.RulesCount})
	t.RulesCount++
}

func (t *Tree) AddExpression() {
	expression := t.PopFront()
	rule := t.PopFront()
	rule.PushBack(expression)
	t.PushBack(rule)
}

func (t *Tree) AddName(text string) {
	t.PushFront(&node{Type: TypeName, string: text})
}

func (t *Tree) AddDot() { t.PushFront(&node{Type: TypeDot, string: "."}) }
func (t *Tree) AddCharacter(text string) {
	t.PushFront(&node{Type: TypeCharacter, string: text})
}
func (t *Tree) AddDoubleCharacter(text string) {
	t.PushFront(&node{Type: TypeCharacter, string: strings.ToLower(text)})
	t.PushFront(&node{Type: TypeCharacter, string: strings.ToUpper(text)})
	t.AddAlternate()
}
func (t *Tree) AddHexaCharacter(text string) {
	hexa, _ := strconv.ParseInt(text, 16, 32)
	t.PushFront(&node{Type: TypeCharacter, string: string(hexa)})
}
func (t *Tree) AddOctalCharacter(text string) {
	octal, _ := strconv.ParseInt(text, 8, 8)
	t.PushFront(&node{Type: TypeCharacter, string: string(octal)})
}
func (t *Tree) AddPredicate(text string)   { t.PushFront(&node{Type: TypePredicate, string: text}) }
func (t *Tree) AddStateChange(text string) { t.PushFront(&node{Type: TypeStateChange, string: text}) }
func (t *Tree) AddNil()                    { t.PushFront(&node{Type: TypeNil, string: "<nil>"}) }
func (t *Tree) AddAction(text string)      { t.PushFront(&node{Type: TypeAction, string: text}) }
func (t *Tree) AddPackage(text string)     { t.PushBack(&node{Type: TypePackage, string: text}) }
func (t *Tree) AddImport(text string)      { t.PushBack(&node{Type: TypeImport, string: text}) }
func (t *Tree) AddState(text string) {
	peg := t.PopFront()
	peg.PushBack(&node{Type: TypeState, string: text})
	t.PushBack(peg)
}

func (t *Tree) addList(listType Type) {
	a := t.PopFront()
	b := t.PopFront()
	var l *node
	if b.GetType() == listType {
		l = b
	} else {
		l = &node{Type: listType}
		l.PushBack(b)
	}
	l.PushBack(a)
	t.PushFront(l)
}
func (t *Tree) AddAlternate() { t.addList(TypeAlternate) }
func (t *Tree) AddSequence()  { t.addList(TypeSequence) }
func (t *Tree) AddRange()     { t.addList(TypeRange) }
func (t *Tree) AddDoubleRange() {
	a := t.PopFront()
	b := t.PopFront()

	t.AddCharacter(strings.ToLower(b.String()))
	t.AddCharacter(strings.ToLower(a.String()))
	t.addList(TypeRange)

	t.AddCharacter(strings.ToUpper(b.String()))
	t.AddCharacter(strings.ToUpper(a.String()))
	t.addList(TypeRange)

	t.AddAlternate()
}

func (t *Tree) addFix(fixType Type) {
	n := &node{Type: fixType}
	n.PushBack(t.PopFront())
	t.PushFront(n)
}
func (t *Tree) AddPeekFor() { t.addFix(TypePeekFor) }
func (t *Tree) AddPeekNot() { t.addFix(TypePeekNot) }
func (t *Tree) AddQuery()   { t.addFix(TypeQuery) }
func (t *Tree) AddStar()    { t.addFix(TypeStar) }
func (t *Tree) AddPlus()    { t.addFix(TypePlus) }
func (t *Tree) AddPush()    { t.addFix(TypePush) }

func (t *Tree) AddPeg(text string) { t.PushFront(&node{Type: TypePeg, string: text}) }

func join(tasks []func()) {
	length := len(tasks)
	done := make(chan int, length)
	for _, task := range tasks {
		go func(task func()) { task(); done <- 1 }(task)
	}
	for d := <-done; d < length; d += <-done {
	}
}

func escape(c string) string {
	switch c {
	case "'":
		return "\\'"
	case "\"":
		return "\""
	default:
		c = strconv.Quote(c)
		return c[1 : len(c)-1]
	}
}

func (t *Tree) Compile(file string, out io.Writer) {
	t.AddImport("fmt")
	t.AddImport("math")
	t.AddImport("sort")
	t.AddImport("strconv")
	t.EndSymbol = 0x110000
	t.RulesCount++

	counts := [TypeLast]uint{}
	{
		var rule *node
		var link func(node Node)
		link = func(n Node) {
			nodeType := n.GetType()
			id := counts[nodeType]
			counts[nodeType]++
			switch nodeType {
			case TypeAction:
				n.SetId(int(id))
				copy, name := n.Copy(), fmt.Sprintf("Action%v", id)
				t.Actions = append(t.Actions, copy)
				n.Init()
				n.SetType(TypeName)
				n.SetString(name)
				n.SetId(t.RulesCount)

				emptyRule := &node{Type: TypeRule, string: name, id: t.RulesCount}
				implicitPush := &node{Type: TypeImplicitPush}
				emptyRule.PushBack(implicitPush)
				implicitPush.PushBack(copy)
				implicitPush.PushBack(emptyRule.Copy())
				t.PushBack(emptyRule)
				t.RulesCount++

				t.Rules[name] = emptyRule
				t.RuleNames = append(t.RuleNames, emptyRule)
			case TypeName:
				name := n.String()
				if _, ok := t.Rules[name]; !ok {
					emptyRule := &node{Type: TypeRule, string: name, id: t.RulesCount}
					implicitPush := &node{Type: TypeImplicitPush}
					emptyRule.PushBack(implicitPush)
					implicitPush.PushBack(&node{Type: TypeNil, string: "<nil>"})
					implicitPush.PushBack(emptyRule.Copy())
					t.PushBack(emptyRule)
					t.RulesCount++

					t.Rules[name] = emptyRule
					t.RuleNames = append(t.RuleNames, emptyRule)
				}
			case TypePush:
				copy, name := rule.Copy(), "PegText"
				copy.SetString(name)
				if _, ok := t.Rules[name]; !ok {
					emptyRule := &node{Type: TypeRule, string: name, id: t.RulesCount}
					emptyRule.PushBack(&node{Type: TypeNil, string: "<nil>"})
					t.PushBack(emptyRule)
					t.RulesCount++

					t.Rules[name] = emptyRule
					t.RuleNames = append(t.RuleNames, emptyRule)
				}
				n.PushBack(copy)
				fallthrough
			case TypeImplicitPush:
				link(n.Front())
			case TypeRule, TypeAlternate, TypeUnorderedAlternate, TypeSequence,
				TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
				for _, node := range n.Slice() {
					link(node)
				}
			}
		}
		/* first pass */
		for _, node := range t.Slice() {
			switch node.GetType() {
			case TypePackage:
				t.PackageName = node.String()
			case TypeImport:
				t.Imports = append(t.Imports, node.String())
			case TypePeg:
				t.StructName = node.String()
				t.StructVariables = node.Front().String()
			case TypeRule:
				if _, ok := t.Rules[node.String()]; !ok {
					expression := node.Front()
					copy := expression.Copy()
					expression.Init()
					expression.SetType(TypeImplicitPush)
					expression.PushBack(copy)
					expression.PushBack(node.Copy())

					t.Rules[node.String()] = node
					t.RuleNames = append(t.RuleNames, node)
				}
			}
		}
		/* second pass */
		for _, node := range t.Slice() {
			if node.GetType() == TypeRule {
				rule = node
				link(node)
			}
		}
	}

	join([]func(){
		func() {
			var countRules func(node Node)
			ruleReached := make([]bool, t.RulesCount)
			countRules = func(node Node) {
				switch node.GetType() {
				case TypeRule:
					name, id := node.String(), node.GetId()
					if count, ok := t.rulesCount[name]; ok {
						t.rulesCount[name] = count + 1
					} else {
						t.rulesCount[name] = 1
					}
					if ruleReached[id] {
						return
					}
					ruleReached[id] = true
					countRules(node.Front())
				case TypeName:
					countRules(t.Rules[node.String()])
				case TypeImplicitPush, TypePush:
					countRules(node.Front())
				case TypeAlternate, TypeUnorderedAlternate, TypeSequence,
					TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
					for _, element := range node.Slice() {
						countRules(element)
					}
				}
			}
			for _, node := range t.Slice() {
				if node.GetType() == TypeRule {
					countRules(node)
					break
				}
			}
		},
		func() {
			var checkRecursion func(node Node) bool
			ruleReached := make([]bool, t.RulesCount)
			checkRecursion = func(node Node) bool {
				switch node.GetType() {
				case TypeRule:
					id := node.GetId()
					if ruleReached[id] {
						fmt.Fprintf(os.Stderr, "possible infinite left recursion in rule '%v'\n", node)
						return false
					}
					ruleReached[id] = true
					consumes := checkRecursion(node.Front())
					ruleReached[id] = false
					return consumes
				case TypeAlternate:
					for _, element := range node.Slice() {
						if !checkRecursion(element) {
							return false
						}
					}
					return true
				case TypeSequence:
					for _, element := range node.Slice() {
						if checkRecursion(element) {
							return true
						}
					}
				case TypeName:
					return checkRecursion(t.Rules[node.String()])
				case TypePlus, TypePush, TypeImplicitPush:
					return checkRecursion(node.Front())
				case TypeCharacter, TypeString:
					return len(node.String()) > 0
				case TypeDot, TypeRange:
					return true
				}
				return false
			}
			for _, node := range t.Slice() {
				if node.GetType() == TypeRule {
					checkRecursion(node)
				}
			}
		}})

	if t._switch {
		var optimizeAlternates func(node Node) (consumes bool, s jetset.Set)
		cache, firstPass := make([]struct {
			reached, consumes bool
			s                 jetset.Set
		}, t.RulesCount), true
		optimizeAlternates = func(n Node) (consumes bool, s jetset.Set) {
			/*n.debug()*/
			switch n.GetType() {
			case TypeRule:
				cache := &cache[n.GetId()]
				if cache.reached {
					consumes, s = cache.consumes, cache.s
					return
				}

				cache.reached = true
				consumes, s = optimizeAlternates(n.Front())
				cache.consumes, cache.s = consumes, s
			case TypeName:
				consumes, s = optimizeAlternates(t.Rules[n.String()])
			case TypeDot:
				consumes = true
				/* TypeDot set doesn't include the EndSymbol */
				s = s.Add(uint64(t.EndSymbol))
				s = s.Complement(uint64(t.EndSymbol))
			case TypeString, TypeCharacter:
				consumes = true
				s = s.Add(uint64([]rune(n.String())[0]))
			case TypeRange:
				consumes = true
				element := n.Front()
				lower := []rune(element.String())[0]
				element = element.Next()
				upper := []rune(element.String())[0]
				s = s.AddRange(uint64(lower), uint64(upper))
			case TypeAlternate:
				consumes = true
				mconsumes, properties, c :=
					consumes, make([]struct {
						intersects bool
						s          jetset.Set
					}, n.Len()), 0
				for _, element := range n.Slice() {
					mconsumes, properties[c].s = optimizeAlternates(element)
					consumes = consumes && mconsumes
					s = s.Union(properties[c].s)
					c++
				}

				if firstPass {
					break
				}

				intersections := 2
			compare:
				for ai, a := range properties[0 : len(properties)-1] {
					for _, b := range properties[ai+1:] {
						if a.s.Intersects(b.s) {
							intersections++
							properties[ai].intersects = true
							continue compare
						}
					}
				}
				if intersections >= len(properties) {
					break
				}

				c, unordered, ordered, max :=
					0, &node{Type: TypeUnorderedAlternate}, &node{Type: TypeAlternate}, 0
				for _, element := range n.Slice() {
					if properties[c].intersects {
						ordered.PushBack(element.Copy())
					} else {
						class := &node{Type: TypeUnorderedAlternate}
						for d := 0; d < 256; d++ {
							if properties[c].s.Has(uint64(d)) {
								class.PushBack(&node{Type: TypeCharacter, string: string(d)})
							}
						}

						sequence, predicate, length :=
							&node{Type: TypeSequence}, &node{Type: TypePeekFor}, properties[c].s.Len()
						if length == 0 {
							class.PushBack(&node{Type: TypeNil, string: "<nil>"})
						}
						predicate.PushBack(class)
						sequence.PushBack(predicate)
						sequence.PushBack(element.Copy())

						if element.GetType() == TypeNil {
							unordered.PushBack(sequence)
						} else if length > max {
							unordered.PushBack(sequence)
							max = length
						} else {
							unordered.PushFront(sequence)
						}
					}
					c++
				}
				n.Init()
				if ordered.Front() == nil {
					n.SetType(TypeUnorderedAlternate)
					for _, element := range unordered.Slice() {
						n.PushBack(element.Copy())
					}
				} else {
					for _, element := range ordered.Slice() {
						n.PushBack(element.Copy())
					}
					n.PushBack(unordered)
				}
			case TypeSequence:
				classes, elements :=
					make([]struct {
						s jetset.Set
					}, n.Len()), n.Slice()

				for c, element := range elements {
					consumes, classes[c].s = optimizeAlternates(element)
					if consumes {
						elements, classes = elements[c+1:], classes[:c+1]
						break
					}
				}

				for c := len(classes) - 1; c >= 0; c-- {
					s = s.Union(classes[c].s)
				}

				for _, element := range elements {
					optimizeAlternates(element)
				}
			case TypePeekNot, TypePeekFor:
				optimizeAlternates(n.Front())
			case TypeQuery, TypeStar:
				_, s = optimizeAlternates(n.Front())
			case TypePlus, TypePush, TypeImplicitPush:
				consumes, s = optimizeAlternates(n.Front())
			case TypeAction, TypeNil:
				//empty
			}
			return
		}
		for _, element := range t.Slice() {
			if element.GetType() == TypeRule {
				optimizeAlternates(element)
				break
			}
		}

		for i, _ := range cache {
			cache[i].reached = false
		}
		firstPass = false
		for _, element := range t.Slice() {
			if element.GetType() == TypeRule {
				optimizeAlternates(element)
				break
			}
		}
	}

	var buffer bytes.Buffer
	defer func() {
		fileSet := token.NewFileSet()
		code, error := parser.ParseFile(fileSet, file, &buffer, parser.ParseComments)
		if error != nil {
			buffer.WriteTo(out)
			fmt.Printf("%v: %v\n", file, error)
			return
		}
		formatter := printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
		error = formatter.Fprint(out, fileSet, code)
		if error != nil {
			buffer.WriteTo(out)
			fmt.Printf("%v: %v\n", file, error)
			return
		}

	}()

	_print := func(format string, a ...interface{}) { fmt.Fprintf(&buffer, format, a...) }
	printSave := func(n uint) { _print("\n   position%d, tokenIndex%d, depth%d := position, tokenIndex, depth", n, n, n) }
	printRestore := func(n uint) { _print("\n   position, tokenIndex, depth = position%d, tokenIndex%d, depth%d", n, n, n) }
	printTemplate := func(s string) {
		if error := template.Must(template.New("peg").Parse(s)).Execute(&buffer, t); error != nil {
			panic(error)
		}
	}

	t.HasActions = counts[TypeAction] > 0
	t.HasPush = counts[TypePush] > 0
	t.HasCommit = counts[TypeCommit] > 0
	t.HasDot = counts[TypeDot] > 0
	t.HasCharacter = counts[TypeCharacter] > 0
	t.HasString = counts[TypeString] > 0
	t.HasRange = counts[TypeRange] > 0

	var printRule func(n Node)
	var compile func(expression Node, ko uint)
	var label uint
	labels := make(map[uint]bool)
	printBegin := func() { _print("\n   {") }
	printEnd := func() { _print("\n   }") }
	printLabel := func(n uint) {
		_print("\n")
		if labels[n] {
			_print("   l%d:\t", n)
		}
	}
	printJump := func(n uint) {
		_print("\n   goto l%d", n)
		labels[n] = true
	}
	printRule = func(n Node) {
		switch n.GetType() {
		case TypeRule:
			_print("%v <- ", n)
			printRule(n.Front())
		case TypeDot:
			_print(".")
		case TypeName:
			_print("%v", n)
		case TypeCharacter:
			_print("'%v'", escape(n.String()))
		case TypeString:
			s := escape(n.String())
			_print("'%v'", s[1:len(s)-1])
		case TypeRange:
			element := n.Front()
			lower := element
			element = element.Next()
			upper := element
			_print("[%v-%v]", escape(lower.String()), escape(upper.String()))
		case TypePredicate:
			_print("&{%v}", n)
		case TypeStateChange:
			_print("!{%v}", n)
		case TypeAction:
			_print("{%v}", n)
		case TypeCommit:
			_print("commit")
		case TypeAlternate:
			_print("(")
			elements := n.Slice()
			printRule(elements[0])
			for _, element := range elements[1:] {
				_print(" / ")
				printRule(element)
			}
			_print(")")
		case TypeUnorderedAlternate:
			_print("(")
			elements := n.Slice()
			printRule(elements[0])
			for _, element := range elements[1:] {
				_print(" | ")
				printRule(element)
			}
			_print(")")
		case TypeSequence:
			_print("(")
			elements := n.Slice()
			printRule(elements[0])
			for _, element := range elements[1:] {
				_print(" ")
				printRule(element)
			}
			_print(")")
		case TypePeekFor:
			_print("&")
			printRule(n.Front())
		case TypePeekNot:
			_print("!")
			printRule(n.Front())
		case TypeQuery:
			printRule(n.Front())
			_print("?")
		case TypeStar:
			printRule(n.Front())
			_print("*")
		case TypePlus:
			printRule(n.Front())
			_print("+")
		case TypePush, TypeImplicitPush:
			_print("<")
			printRule(n.Front())
			_print(">")
		case TypeNil:
		default:
			fmt.Fprintf(os.Stderr, "illegal node type: %v\n", n.GetType())
		}
	}
	compile = func(n Node, ko uint) {
		switch n.GetType() {
		case TypeRule:
			fmt.Fprintf(os.Stderr, "internal error #1 (%v)\n", n)
		case TypeDot:
			_print("\n   if !matchDot() {")
			/*print("\n   if buffer[position] == end_symbol {")*/
			printJump(ko)
			/*print("}\nposition++")*/
			_print("}")
		case TypeName:
			name := n.String()
			rule := t.Rules[name]
			if t.inline && t.rulesCount[name] == 1 {
				compile(rule.Front(), ko)
				return
			}
			_print("\n   if !_rules[rule%v]() {", name /*rule.GetId()*/)
			printJump(ko)
			_print("}")
		case TypeRange:
			element := n.Front()
			lower := element
			element = element.Next()
			upper := element
			/*print("\n   if !matchRange('%v', '%v') {", escape(lower.String()), escape(upper.String()))*/
			_print("\n   if c := buffer[position]; c < rune('%v') || c > rune('%v') {", escape(lower.String()), escape(upper.String()))
			printJump(ko)
			_print("}\nposition++")
		case TypeCharacter:
			/*print("\n   if !matchChar('%v') {", escape(n.String()))*/
			_print("\n   if buffer[position] != rune('%v') {", escape(n.String()))
			printJump(ko)
			_print("}\nposition++")
		case TypeString:
			_print("\n   if !matchString(%v) {", strconv.Quote(n.String()))
			printJump(ko)
			_print("}")
		case TypePredicate:
			_print("\n   if !(%v) {", n)
			printJump(ko)
			_print("}")
		case TypeStateChange:
			_print("\n   %v", n)
		case TypeAction:
		case TypeCommit:
		case TypePush:
			fallthrough
		case TypeImplicitPush:
			ok, element := label, n.Front()
			label++
			nodeType, rule := element.GetType(), element.Next()
			printBegin()
			if nodeType == TypeAction {
				_print("\nadd(rule%v, position)", rule)
			} else {
				_print("\nposition%d := position", ok)
				_print("\ndepth++")
				compile(element, ko)
				_print("\ndepth--")
				_print("\nadd(rule%v, position%d)", rule, ok)
			}
			printEnd()
		case TypeAlternate:
			ok := label
			label++
			printBegin()
			elements := n.Slice()
			printSave(ok)
			for _, element := range elements[:len(elements)-1] {
				next := label
				label++
				compile(element, next)
				printJump(ok)
				printLabel(next)
				printRestore(ok)
			}
			compile(elements[len(elements)-1], ko)
			printEnd()
			printLabel(ok)
		case TypeUnorderedAlternate:
			done, ok := ko, label
			label++
			printBegin()
			_print("\n   switch buffer[position] {")
			elements := n.Slice()
			elements, last := elements[:len(elements)-1], elements[len(elements)-1].Front().Next()
			for _, element := range elements {
				sequence := element.Front()
				class := sequence.Front()
				sequence = sequence.Next()
				_print("\n   case")
				comma := false
				for _, character := range class.Slice() {
					if comma {
						_print(",")
					} else {
						comma = true
					}
					_print(" '%s'", escape(character.String()))
				}
				_print(":")
				compile(sequence, done)
				_print("\nbreak")
			}
			_print("\n   default:")
			compile(last, done)
			_print("\nbreak")
			_print("\n   }")
			printEnd()
			printLabel(ok)
		case TypeSequence:
			for _, element := range n.Slice() {
				compile(element, ko)
			}
		case TypePeekFor:
			ok := label
			label++
			printBegin()
			printSave(ok)
			compile(n.Front(), ko)
			printRestore(ok)
			printEnd()
		case TypePeekNot:
			ok := label
			label++
			printBegin()
			printSave(ok)
			compile(n.Front(), ok)
			printJump(ko)
			printLabel(ok)
			printRestore(ok)
			printEnd()
		case TypeQuery:
			qko := label
			label++
			qok := label
			label++
			printBegin()
			printSave(qko)
			compile(n.Front(), qko)
			printJump(qok)
			printLabel(qko)
			printRestore(qko)
			printEnd()
			printLabel(qok)
		case TypeStar:
			again := label
			label++
			out := label
			label++
			printLabel(again)
			printBegin()
			printSave(out)
			compile(n.Front(), out)
			printJump(again)
			printLabel(out)
			printRestore(out)
			printEnd()
		case TypePlus:
			again := label
			label++
			out := label
			label++
			compile(n.Front(), ko)
			printLabel(again)
			printBegin()
			printSave(out)
			compile(n.Front(), out)
			printJump(again)
			printLabel(out)
			printRestore(out)
			printEnd()
		case TypeNil:
		default:
			fmt.Fprintf(os.Stderr, "illegal node type: %v\n", n.GetType())
		}
	}

	/* lets figure out which jump labels are going to be used with this dry compile */
	printTemp, _print := _print, func(format string, a ...interface{}) {}
	for _, element := range t.Slice() {
		if element.GetType() != TypeRule {
			continue
		}
		expression := element.Front()
		if expression.GetType() == TypeNil {
			continue
		}
		ko := label
		label++
		if count, ok := t.rulesCount[element.String()]; !ok {
			continue
		} else if t.inline && count == 1 && ko != 0 {
			continue
		}
		compile(expression, ko)
	}
	_print, label = printTemp, 0

	/* now for the real compile pass */
	t.PegRuleType = "uint8"
	if length := int64(t.Len()); length > math.MaxUint32 {
		t.PegRuleType = "uint64"
	} else if length > math.MaxUint16 {
		t.PegRuleType = "uint32"
	} else if length > math.MaxUint8 {
		t.PegRuleType = "uint16"
	}
	printTemplate(PEG_HEADER_TEMPLATE)
	for _, element := range t.Slice() {
		if element.GetType() != TypeRule {
			continue
		}
		expression := element.Front()
		if implicit := expression.Front(); expression.GetType() == TypeNil || implicit.GetType() == TypeNil {
			if element.String() != "PegText" {
				fmt.Fprintf(os.Stderr, "rule '%v' used but not defined\n", element)
			}
			_print("\n  nil,")
			continue
		}
		ko := label
		label++
		_print("\n  /* %v ", element.GetId())
		printRule(element)
		_print(" */")
		if count, ok := t.rulesCount[element.String()]; !ok {
			fmt.Fprintf(os.Stderr, "rule '%v' defined but not used\n", element)
			_print("\n  nil,")
			continue
		} else if t.inline && count == 1 && ko != 0 {
			_print("\n  nil,")
			continue
		}
		_print("\n  func() bool {")
		if labels[ko] {
			printSave(ko)
		}
		compile(expression, ko)
		//print("\n  fmt.Printf(\"%v\\n\")", element.String())
		_print("\n   return true")
		if labels[ko] {
			printLabel(ko)
			printRestore(ko)
			_print("\n   return false")
		}
		_print("\n  },")
	}
	_print("\n }\n p.rules = _rules")
	_print("\n}\n")
}
