// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"container/list"
	"fmt"
	"go/parser"
        "go/printer"
        "go/token"
//	"math"
	"os"
	"strconv"
	"template"
)

const PEG_HEADER_TEMPLATE =
`package {{.PackageName}}

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
	{{range .RuleNames}}Rule{{.String}}
	{{end}}
)

var Rul3s = [...]string {
	"Unknown",
	{{range .RuleNames}}"{{.String}}",
	{{end}}
}

type TokenTree interface {
	sort.Interface
	Prepare()
	Add(rule Rule, begin, end, next int)
	Expand(index int) TokenTree
	Stack() []token32
	Tokens() <-chan token32
}

{{range .Sizes}}

/* ${@} bit structure for abstract syntax tree */
type token{{.}} struct {
	Rule
	begin, end, next int{{.}}
}

func (t *token{{.}}) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

type tokens{{.}} struct {
	tree		[]token{{.}}
	stackSize	int32
}

func (t *tokens{{.}}) Len() int {
	return len(t.tree)
}

func (t *tokens{{.}}) Less(i, j int) bool {
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

func (t *tokens{{.}}) Swap(i, j int) {
	t.tree[i], t.tree[j] = t.tree[j], t.tree[i]
}

func (t *tokens{{.}}) Prepare() {
	sort.Sort(t)
	size := int(t.tree[0].next) + 1

	tree, stack, top := t.tree[0:size], make([]token{{.}}, size), -1
	for i, token := range tree {
		token.next = int{{.}}(i)
		for top >= 0 && token.begin >= stack[top].end {
			tree[stack[top].next].next, top = token.next, top - 1
		}
		stack[top + 1], top = token, top + 1
	}

	for i, token := range stack {
		if token.isZero() {
			t.stackSize = int32(i)
			break
		}
	}

	t.tree = tree
}

func (t *tokens{{.}}) Add(rule Rule, begin, end, next int) {
	t.tree[next] = token{{.}}{Rule: rule, begin: int{{.}}(begin), end: int{{.}}(end), next: int{{.}}(next)}
}

func (t *tokens{{.}}) Stack() []token32 {
	if t.stackSize == 0 {
		t.Prepare()
	}
	return make([]token32, t.stackSize)
}

func (t *tokens{{.}}) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- token32{Rule: v.Rule, begin: int32(v.begin), end: int32(v.end), next: int32(v.next)}
		}
		close(s)
	}()
	return s
}
{{end}}

func (t *tokens16) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
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
		expanded := make([]token32, 2 * len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type {{.StructName}} struct {
	{{.StructVariables}}
	Buffer		string
	Min, Max	int
	rules		[{{.RulesCount}}]func() bool

	TokenTree
}

func (p *{{.StructName}}) Add(rule Rule, begin, end, next int) {
	if tree := p.TokenTree.Expand(next); tree != nil {
		p.TokenTree = tree
	}
	p.TokenTree.Add(rule, begin, end, next)
}

func (p *{{.StructName}}) Parse() os.Error {
	if p.rules[0]() {
		return nil
	}
	return &parseError{ p }
}

type parseError struct {
	p *{{.StructName}}
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

func (p *{{.StructName}}) PrintSyntaxTree() {
	tokenTree := p.TokenTree
	stack, top, i := tokenTree.Stack(), -1, 0
	for token := range tokenTree.Tokens() {
		if top >= 0 && int(stack[top].next) == i {
			for top >= 0 && int(stack[top].next) == i {
				top--
			}
		}
		stack[top + 1], top, i = token, top + 1, i + 1

		for c := 0; c < top; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(p.Buffer[token.begin:token.end]))
	}
}

func (p *{{.StructName}}) Highlighter() {
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
		stack[top + 1], top, i = token, top + 1, i + 1
	}
}

func (p *{{.StructName}}) Init() {
	position, tokenIndex := 0, 0
	p.TokenTree = &tokens16{tree: make([]token16, 65536)}

	{{if .HasActions}}
 	actions := [...]func(buffer string, begin, end int) {
		{{range .Actions}}/* {{.GetId}} */
		func(buffer string, begin, end int) {
			{{.String}}
		},
		{{end}}
	}

	var thunkPosition, begin, end int
	thunks := make([]struct {action uint{{.Bits}}; begin, end int}, 32)
	do := func(action uint{{.Bits}}) {
		if thunkPosition == len(thunks) {
			newThunks := make([]struct {action uint{{.Bits}}; begin, end int}, 2 * len(thunks))
			copy(newThunks, thunks)
			thunks = newThunks
		}
		thunks[thunkPosition].action = action
		thunks[thunkPosition].begin = begin
		thunks[thunkPosition].end = end
		thunkPosition++
	}

	{{if .HasCommit}}
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
	{{end}}
	{{end}}

	{{if .HasDot}}
	matchDot := func() bool {
		if position < len(p.Buffer) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	{{end}}

	{{if .HasCharacter}}
	matchChar := func(c byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] == c) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	{{end}}

	{{if .HasString}}
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
	{{end}}

	{{if .HasRange}}
	matchRange := func(lower byte, upper byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] >= lower) && (p.Buffer[position] <= upper) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	{{end}}

	p.rules = [...]func() bool {`

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
	TypeCommit
	TypeAction
	TypePackage
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

func (t Type) GetType() Type {
	return t
}

type Node interface {
	fmt.Stringer

	GetType() Type
	SetType(t Type)

	GetId() int
	SetId(id int)

	Init()
	Front() *node
	Next() *node
	PushFront(value *node)
	PushBack(value *node)
	Len() int
	Copy() *node
}

type node struct {
	Type
	string
	id int

	front *node
	back *node
	length int

	next *node
}

func (n *node) String() string {
	return n.string
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

/* Used to represent character classes. */
type characterClass [32]uint8

func (c *characterClass) copy() (class *characterClass) {
	class = new(characterClass)
	copy(class[0:], c[0:])
	return
}
func (c *characterClass) add(character uint8)      { c[character>>3] |= (1 << (character & 7)) }
func (c *characterClass) has(character uint8) bool { return c[character>>3]&(1<<(character&7)) != 0 }
func (c *characterClass) complement() {
	for i := range *c {
		c[i] = ^c[i]
	}
}
func (c *characterClass) union(class *characterClass) {
	for index, value := range *class {
		c[index] |= value
	}
}
func (c *characterClass) intersection(class *characterClass) {
	for index, value := range *class {
		c[index] &= value
	}
}
func (c *characterClass) len() (length int) {
	for character := 0; character < 256; character++ {
		if c.has(uint8(character)) {
			length++
		}
	}
	return
}

/* A tree data structure into which a PEG can be parsed. */
type Tree struct {
	Rules		map[string]Node
	rulesCount	map[string]uint
	ruleId     	int
	list.List
	classes         map[string]*characterClass
	stack           [1024]*node
	top             int
	inline, _switch bool

	RuleNames	[]Node
	Sizes		[2]int
	PackageName     string
	StructName      string
	StructVariables string
	RulesCount      int
	Bits            int
	HasActions	bool
	Actions         []Node
	HasCommit       bool
	HasDot          bool
	HasCharacter    bool
	HasString       bool
	HasRange        bool
}

func New(inline, _switch bool) *Tree {
	return &Tree{Rules: make(map[string]Node),
		Sizes: [2]int{16, 32},
		rulesCount: make(map[string]uint),
		classes:    make(map[string]*characterClass),
		inline:     inline,
		_switch:    _switch}
}

func (t *Tree) push(n *node) {
	t.top++
	t.stack[t.top] = n
}

func (t *Tree) pop() *node {
	n := t.stack[t.top]
	t.top--
	return n
}

func (t *Tree) currentRule() *node {
	return t.stack[1]
}

func (t *Tree) AddRule(name string) {
	t.push(&node{Type: TypeRule, string: name, id: t.ruleId})
	t.ruleId++
}

func (t *Tree) AddExpression() {
	expression := t.pop()
	rule := t.pop()
	rule.PushBack(expression)
	t.PushBack(rule)
}

func (t *Tree) AddName(text string) {
	t.Rules[text] = &node{Type: TypeRule}
	t.push(&node{Type: TypeName, string: text})
}

func (t *Tree) AddDot() { t.push(&node{Type: TypeDot, string: "."}) }
func (t *Tree) AddCharacter(text string) {
	t.push(&node{Type: TypeCharacter, string: text})
}
func (t *Tree) AddOctalCharacter(text string) {
	octal, _ := strconv.Btoui64(text, 8)
	t.push(&node{Type: TypeCharacter, string: string(octal)})
}
func (t *Tree) AddPredicate(text string) { t.push(&node{Type: TypePredicate, string: text}) }
func (t *Tree) AddCommit() { t.push(&node{Type: TypeCommit, string: "commit"}) }
func (t *Tree) AddNil() { t.push(&node{Type: TypeNil, string: "<nil>"}) }
func (t *Tree) AddAction(text string) { t.push(&node{Type: TypeAction, string: text}) }
func (t *Tree) AddPackage(text string) { t.PushBack(&node{Type: TypePackage, string: text}) }
func (t *Tree) AddState(text string) {
	peg := t.pop()
	peg.PushBack(&node{Type: TypeState, string: text})
	t.PushBack(peg)
}

func (t *Tree) addList(listType Type) {
	a := t.pop()
	b := t.pop()
	var l *node
	if b.GetType() == listType {
		l = b
	} else {
		l = &node{Type: listType}
		l.PushBack(b)
	}
	l.PushBack(a)
	t.push(l)
}
func (t *Tree) AddAlternate() { t.addList(TypeAlternate) }
func (t *Tree) AddSequence() { t.addList(TypeSequence) }
func (t *Tree) AddRange()    { t.addList(TypeRange) }

func (t *Tree) addFix(fixType Type) {
	n := &node{Type: fixType}
	n.PushBack(t.pop())
	t.push(n)
}
func (t *Tree) AddPeekFor()        { t.addFix(TypePeekFor) }
func (t *Tree) AddPeekNot()        { t.addFix(TypePeekNot) }
func (t *Tree) AddQuery()          { t.addFix(TypeQuery) }
func (t *Tree) AddStar()           { t.addFix(TypeStar) }
func (t *Tree) AddPlus()           { t.addFix(TypePlus) }
func (t *Tree) AddPush()	   { t.addFix(TypePush) }

func (t *Tree) AddPeg(text string) { t.push(&node{Type: TypePeg, string: text}) }

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
		return c[1:len(c) - 1]
	}
	return ""
}

type estimate struct {
	consumed, depth int
	consumes bool
	name string
	estimates []estimate
}

func (t *Tree) Compile(file string) {
	counts := [TypeLast]uint{}

	for element := t.Front(); element != nil; element = element.Next() {
		node := element.Value.(Node)
		switch node.GetType() {
		case TypePackage:
			t.PackageName = node.String()
		case TypePeg:
			t.StructName = node.String()
			t.StructVariables = node.Front().String()
		case TypeRule:
			t.Rules[node.String()] = node
			t.RuleNames = append(t.RuleNames, node)
		}
	}

	/*Needed for undefined rules!*/
	for name, r := range t.Rules {
		if r.String() == "" {
			r := &node{Type: TypeRule, string: name, id: t.ruleId}
			r.PushBack(&node{Type:TypeNil, string: "<nil>"})
			t.ruleId++
			t.Rules[name] = r
			t.PushBack(r)
		}
	}
	t.RulesCount = len(t.Rules)

	join([]func(){
		func() {
			var countTypes func(node Node)
			countTypes = func(node Node) {
				nodeType := node.GetType()
				id := counts[nodeType]
				counts[nodeType]++
				switch nodeType {
				case TypeAction:
					node.SetId(int(id))
					t.Actions = append(t.Actions, node)
				case TypeRule, TypeAlternate, TypeUnorderedAlternate, TypeSequence,
					TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus, TypePush:
					for element := node.Front(); element != nil; element = element.Next() {
						countTypes(element)
					}
				}
			}
			for _, rule := range t.Rules {
				countTypes(rule)
			}
		},
		func() {
			var countRules func(node Node)
			ruleReached := make([]bool, len(t.Rules))
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
				case TypeAlternate, TypeUnorderedAlternate, TypeSequence,
					TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus, TypePush:
					for element := node.Front(); element != nil; element = element.Next() {
						countRules(element)
					}
				}
			}
			for element := t.Front(); element != nil; element = element.Next() {
				node := element.Value.(Node)
				if node.GetType() == TypeRule {
					countRules(node)
					break
				}
			}
		},
		func() {
			var checkRecursion func(node Node) bool
			ruleReached := make([]bool, len(t.Rules))
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
					for element := node.Front(); element != nil; element = element.Next() {
						if !checkRecursion(element) {
							return false
						}
					}
					return true
				case TypeSequence:
					for element := node.Front(); element != nil; element = element.Next() {
						if checkRecursion(element) {
							return true
						}
					}
				case TypeName:
					return checkRecursion(t.Rules[node.String()])
				case TypePlus, TypePush:
					return checkRecursion(node.Front())
				case TypeCharacter, TypeString:
					return len(node.String()) > 0
				case TypeDot, TypeRange:
					return true
				}
				return false
			}
			for _, rule := range t.Rules {
				checkRecursion(rule)
			}
		}})

	var ast func(node, rule Node)
	ast = func(node, rule Node) {
		if node.GetType() == TypePush {
			node.PushBack(rule.Copy())
			if node.Front() != nil {
				ast(node.Front(), rule)
			}
			return
		}
		for element := node.Front(); element != nil; element = element.Next() {
			ast(element, rule)
		}
	}
	for _, rule := range t.Rules {
		ast(rule, rule)
		expression := rule.Front()
		copy := expression.Copy()
		expression.Init()
		expression.SetType(TypeImplicitPush)
		expression.PushBack(copy)
		expression.PushBack(rule.Copy())
	}

	/*var estimateMemory func(node Node, depth int) (estimates []estimate, consumed int)
	ruleReached := make([]bool, len(t.Rules))
	estimateMemory = func(node Node, depth int) (estimates []estimate, consumed int) {
		switch node.GetType() {
		case TypeRule:
			id := node.GetId()
			if ruleReached[id] {
				return
			}
			ruleReached[id] = true
			estimates, consumed = estimateMemory(node.Front(), depth)
			ruleReached[id] = false
			return
		case TypeAlternate:
			consumes := math.MaxInt32
			//need to account for emtpy alternate!!!
			for element := node.Front(); element != nil; element = element.Next() {
				e, c := estimateMemory(element, depth)
				if c == 0 {
					consumed = 0
					for i, _ := range estimates {
						estimates[i].consumes = false
					}
					return
				} else if c < consumes {
					consumed, consumes, estimates = c, c, e
				}
			}
			return
		case TypeSequence:
			for element := node.Front(); element != nil; element = element.Next() {
				e, c := estimateMemory(element, depth)
				if c != 0 {
					consumed += c
					for _, est := range e {
						estimates = append(estimates, est)
					}
				}
			}
			return
		case TypePlus:
			estimates, consumed = estimateMemory(node.Front(), depth)
			for _, e := range estimates {
				e.consumes = false
				estimates = append(estimates, e)
			}
			return
		case TypeName:
			estimates, consumed = estimateMemory(t.Rules[node.String()], depth)
			return
		case TypePush, TypeImplicitPush:
			depth++
			token := node.Front()
			estimates, consumed = estimateMemory(token, depth)
			e := estimate{consumed: consumed, depth: depth, consumes: true, name: token.Next().String(), estimates: estimates}
			estimates = make([]estimate, 1)
			estimates[0] = e
			return
		case TypeCharacter, TypeString:
			consumed = len(node.String())
			return
		case TypeDot, TypeRange:
			consumed = 1
			return
		}
		return
	}
	var printEstimates func(estimates []estimate)
	printEstimates = func(estimates []estimate) {
 		for _, e := range estimates {
			for c := 0; c < e.depth; c++ {
				fmt.Printf(" ")
			}
			fmt.Printf("%v %v %v %v\n", e.consumed, e.consumes, e.name, float32(e.depth)/float32(e.consumed))
			printEstimates(e.estimates)
		}
	}
	for ruleName, rule := range t.Rules {
		fmt.Printf("%s\n", ruleName)
		estimates, _ := estimateMemory(rule, 0)
		printEstimates(estimates)
	}*/

	if t._switch {
		var optimizeAlternates func(node Node) (consumes, eof, peek bool, class *characterClass)
		cache := make([]struct {
			reached, consumes, eof, peek bool
			class                        *characterClass
		}, len(t.Rules))
		optimizeAlternates = func(n Node) (consumes, eof, peek bool, class *characterClass) {
			switch n.GetType() {
			case TypeRule:
				cache := &cache[n.GetId()]
				if cache.reached {
					consumes, eof, peek, class = cache.consumes, cache.eof, cache.peek, cache.class
					return
				}
				cache.reached = true
				consumes, eof, peek, class = optimizeAlternates(n.Front())
				cache.consumes, cache.eof, cache.peek, cache.class = consumes, eof, peek, class
			case TypeName:
				consumes, eof, peek, class = optimizeAlternates(t.Rules[n.String()])
			case TypeDot:
				consumes, class = true, new(characterClass)
				for index, _ := range *class {
					class[index] = 0xff
				}
			case TypeString, TypeCharacter:
				consumes, class = true, new(characterClass)
				class.add(n.String()[0])
			case TypeRange:
				consumes, class = true, new(characterClass)
				element := n.Front()
				lower := element.String()[0]
				element = element.Next()
				upper := element.String()[0]
				for c := lower; c <= upper; c++ {
					class.add(c)
				}
			case TypeAlternate:
				consumes, peek, class = true, true, new(characterClass)
				mconsumes, meof, mpeek, properties, c :=
					consumes, eof, peek, make([]struct {
						intersects bool
						class      *characterClass
					}, n.Len()), 0
				for element := n.Front(); element != nil; element = element.Next() {
					mconsumes, meof, mpeek, properties[c].class = optimizeAlternates(element)
					consumes, eof, peek = consumes && mconsumes, eof || meof, peek && mpeek
					if properties[c].class != nil {
						class.union(properties[c].class)
					}
					c++
				}
				if eof {
					break
				}
				intersections := 2
			compare:
				for ai, a := range properties[0 : len(properties)-1] {
					for _, b := range properties[ai+1:] {
						for i, v := range *a.class {
							if (b.class[i] & v) != 0 {
								intersections++
								properties[ai].intersects = true
								continue compare
							}
						}
					}
				}
				if intersections < len(properties) {
					c, unordered, ordered, max :=
						0, &node{Type: TypeUnorderedAlternate}, &node{Type: TypeAlternate}, 0
					for element := n.Front(); element != nil; element = element.Next() {
						if properties[c].intersects {
							ordered.PushBack(element.Copy())
						} else {
							class := &node{Type: TypeUnorderedAlternate}
							for d := 0; d < 256; d++ {
								if properties[c].class.has(uint8(d)) {
									class.PushBack(&node{Type: TypeCharacter, string: string(d)})
								}
							}

							sequence, predicate, length :=
								&node{Type: TypeSequence}, &node{Type: TypePeekFor}, properties[c].class.len()
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
						for element := unordered.Front(); element != nil; element = element.Next() {
							n.PushBack(element.Copy())
						}
					} else {
						for element := ordered.Front(); element != nil; element = element.Next() {
							n.PushBack(element.Copy())
						}
						n.PushBack(unordered)
					}
				}
			case TypeSequence:
				sequence := n
				meof, classes, c, element :=
					eof, make([]struct {
						peek  bool
						class *characterClass
					}, sequence.Len()), 0, sequence.Front()
				for ; !consumes && element != nil; element, c = element.Next(), c + 1 {
					consumes, meof, classes[c].peek, classes[c].class = optimizeAlternates(element)
					eof, peek = eof || meof, peek || classes[c].peek
				}
				eof, peek, class = !consumes && eof, !consumes && peek, new(characterClass)
				for c--; c >= 0; c-- {
					if classes[c].class != nil {
						if classes[c].peek {
							class.intersection(classes[c].class)
						} else {
							class.union(classes[c].class)
						}
					}
				}
				for ; element != nil; element = element.Next() {
					optimizeAlternates(element)
				}
			case TypePeekNot:
				peek = true
				_, eof, _, class = optimizeAlternates(n.Front())
				eof = !eof
				class = class.copy()
				class.complement()
			case TypePeekFor:
				peek = true
				fallthrough
			case TypeQuery, TypeStar:
				_, eof, _, class = optimizeAlternates(n.Front())
			case TypePlus, TypePush:
				consumes, eof, peek, class = optimizeAlternates(n.Front())
			case TypeAction, TypeNil:
				class = new(characterClass)
			}
			return
		}
		for element := t.Front(); element != nil; element = element.Next() {
			n := element.Value.(Node)
			if n.GetType() == TypeRule {
				optimizeAlternates(n.(*node))
				break
			}
		}
	}

	out, error := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if error != nil {
		fmt.Printf("%v: %v\n", file, error)
		return
	}
	defer out.Close()

	var buffer bytes.Buffer
	defer func() {
		fileSet := token.NewFileSet()
		code, error := parser.ParseFile(fileSet, file, &buffer, parser.ParseComments)
		if error != nil {
			buffer.WriteTo(out)
			fmt.Printf("%v: %v\n", file, error)
			return
		}
		formatter := printer.Config{printer.TabIndent | printer.UseSpaces, 8}
		_, error = formatter.Fprint(out, fileSet, code)
		if error != nil {
			buffer.WriteTo(out)
			fmt.Printf("%v: %v\n", file, error)
			return
		}

	}()

	print := func(format string, a ...interface{}) { fmt.Fprintf(&buffer, format, a...) }
	printSave := func(n uint) { print("\n   position%d, tokenIndex%d := position, tokenIndex", n, n) }
	printRestore := func(n uint) { print("   position, tokenIndex = position%d, tokenIndex%d", n, n) }
	printTemplate := func(s string) { if error := template.Must(template.New("peg").Parse(s)).Execute(&buffer, t); error != nil { panic(error) } }

	if t.HasActions = counts[TypeAction] > 0; t.HasActions {
		bits := 0
		for length := len(t.Actions); length != 0; length >>= 1 {
			bits++
		}
		switch {
		case bits < 8:
			bits = 8
		case bits < 16:
			bits = 16
		case bits < 32:
			bits = 32
		case bits < 64:
			bits = 64
		}
		t.Bits = bits

		printSave = func(n uint) { print("\n   position%d, tokenIndex%d, thunkPosition%d := position, tokenIndex, thunkPosition", n, n, n) }
		printRestore = func(n uint) { print("   position, tokenIndex, thunkPosition = position%d, tokenIndex%d, thunkPosition%d", n, n, n) }
	}

	t.HasCommit = counts[TypeCommit] > 0
	t.HasDot = counts[TypeDot] > 0
	t.HasCharacter = counts[TypeCharacter] > 0
	t.HasString = counts[TypeString] > 0
	t.HasRange = counts[TypeRange] > 0

	var printRule func(n Node)
	var compile func(expression Node, ko uint)
	var label uint
	labels := make(map[uint]bool)
	printBegin := func() { print("\n   {") }
	printEnd := func() { print("\n   }") }
	printLabel := func(n uint) {
		print("\n")
		if labels[n] {
			print("   l%d:\t", n)
		}
	}
	printJump := func(n uint) {
		print("\n   goto l%d", n)
		labels[n] = true
	}
	printRule = func(n Node) {
		switch n.GetType() {
		case TypeRule:
			print("%v <- ", n)
			printRule(n.Front())
		case TypeDot:
			print(".")
		case TypeName:
			print("%v", n)
		case TypeCharacter:
			print("'%v'", escape(n.String()))
		case TypeString:
			s := escape(n.String())
			print("'%v'", s[1:len(s) - 1])
		case TypeRange:
			element := n.Front()
			lower := element
			element = element.Next()
			upper := element
			print("[%v-%v]", lower, upper)
		case TypePredicate:
			print("&{%v}", n)
		case TypeAction:
			print("{%v}", n)
		case TypeCommit:
			print("commit")
		case TypeAlternate:
			print("(")
			element := n.Front()
			printRule(element)
			for element = element.Next(); element != nil; element = element.Next() {
				print(" / ")
				printRule(element)
			}
			print(")")
		case TypeUnorderedAlternate:
			print("(")
			element := n.Front()
			printRule(element)
			for element = element.Next(); element != nil; element = element.Next() {
				print(" | ")
				printRule(element)
			}
			print(")")
		case TypeSequence:
			print("(")
			element := n.Front()
			printRule(element)
			for element = element.Next(); element != nil; element = element.Next() {
				print(" ")
				printRule(element)
			}
			print(")")
		case TypePeekFor:
			print("&")
			printRule(n.Front())
		case TypePeekNot:
			print("!")
			printRule(n.Front())
		case TypeQuery:
			printRule(n.Front())
			print("?")
		case TypeStar:
			printRule(n.Front())
			print("*")
		case TypePlus:
			printRule(n.Front())
			print("+")
		case TypePush, TypeImplicitPush:
			print("<")
			printRule(n.Front())
			print(">")
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
			print("\n   if !matchDot() {")
			printJump(ko)
			print("}")
		case TypeName:
			name := n.String()
			rule := t.Rules[name]
			if t.inline && t.rulesCount[name] == 1 {
				compile(rule.Front(), ko)
				return
			}
			print("\n   if !p.rules[%d]() {", rule.GetId())
			printJump(ko)
			print("}")
		case TypeRange:
			element := n.Front()
			lower := element
			element = element.Next()
			upper := element
			print("\n   if !matchRange('%v', '%v') {", escape(lower.String()), escape(upper.String()))
			printJump(ko)
			print("}")
		case TypeCharacter:
			print("\n   if !matchChar('%v') {", escape(n.String()))
			printJump(ko)
			print("}")
		case TypeString:
			print("\n   if !matchString(%v) {", strconv.Quote(n.String()))
			printJump(ko)
			print("}")
		case TypePredicate:
			print("\n   if !(%v) {", n)
			printJump(ko)
			print("}")
		case TypeAction:
			print("\n   do(%d)", n.GetId())
		case TypeCommit:
			/*print("\n   if !(commit(thunkPosition0)) {")*/
			print("\n   if !(commit(0)) {")
			printJump(ko)
			print("}")
		case TypePush:
			begin := label
			label++
			element := n.Front()
			printBegin()
			if t.HasActions {
				print("\n   begin = position")
			}
			print("\nbegin%d := position", begin)
			compile(element, ko)
			if t.HasActions {
				print("\n   end = position")
			}
			print("\nif begin%d != position {p.Add(Rule%v, begin%d, position, tokenIndex)", begin, element.Next(), begin)
			print("\ntokenIndex++}")
			printEnd()
		case TypeImplicitPush:
			begin := label
			label++
			element := n.Front()
			printBegin()
			print("\nbegin%d := position", begin)
			compile(element, ko)
			print("\nif begin%d != position {p.Add(Rule%v, begin%d, position, tokenIndex)", begin, element.Next(), begin)
			print("\ntokenIndex++}")
			printEnd()
		case TypeAlternate:
			ok := label
			label++
			printBegin()
			element := n.Front()
			if element.Next() != nil {
				printSave(ok)
			}
			for element.Next() != nil {
				next := label
				label++
				compile(element, next)
				printJump(ok)
				printLabel(next)
				printRestore(ok)
				element = element.Next()
			}
			compile(element, ko)
			printEnd()
			printLabel(ok)
		case TypeUnorderedAlternate:
			done, ok := ko, label
			label++
			printBegin()
			print("\n   if position == len(p.Buffer) {")
			printJump(done)
			print("}")
			print("\n   switch p.Buffer[position] {")
			element := n.Front()
			for ; element.Next() != nil; element = element.Next() {
				sequence := element.Front()
				class := sequence.Front()
				sequence = sequence.Next()
				print("\n   case")
				comma := false
				for character := class.Front(); character != nil; character = character.Next() {
					if comma {
						print(",")
					} else {
						comma = true
					}
					print(" '%s'", escape(character.String()))
				}
				print(":")
				compile(sequence, done)
				print("\nbreak")
			}
			print("\n   default:")
			compile(element.Front().Next(), done)
			print("\nbreak")
			print("\n   }")
			printEnd()
			printLabel(ok)
		case TypeSequence:
			for element := n.Front(); element != nil; element = element.Next() {
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
	printTemp, print := print, func(format string, a ...interface{}) {}
	for element := t.Front(); element != nil; element = element.Next() {
		rule := element.Value.(Node)
		if rule.GetType() != TypeRule {
			continue
		}
		expression := rule.Front()
		if expression.GetType() == TypeNil {
			continue
		}
		ko := label
		label++
		if count, ok := t.rulesCount[rule.String()]; !ok {
			continue
		} else if t.inline && count == 1 && ko != 0 {
			continue
		}
		compile(expression, ko)
	}
	print, label = printTemp, 0

	/* now for the real compile pass */
	printTemplate(PEG_HEADER_TEMPLATE)
	for element := t.Front(); element != nil; element = element.Next() {
		rule := element.Value.(Node)
		if rule.GetType() != TypeRule {
			continue
		}
		expression := rule.Front()
		if expression.GetType() == TypeNil {
			fmt.Fprintf(os.Stderr, "rule '%v' used but not defined\n", rule)
			print("\n  nil,")
			continue
		}
		ko := label
		label++
		print("\n  /* %v ", rule.GetId())
		printRule(rule)
		print(" */")
		if count, ok := t.rulesCount[rule.String()]; !ok {
			fmt.Fprintf(os.Stderr, "rule '%v' defined but not used\n", rule)
			print("\n  nil,")
			continue
		} else if t.inline && count == 1 && ko != 0 {
			print("\n  nil,")
			continue
		}
		print("\n  func() bool {")
		if labels[ko] {
			printSave(ko)
		}
		compile(expression, ko)
		print("\n   return true")
		if labels[ko] {
			printLabel(ko)
			printRestore(ko)
			print("\n   return false")
		}
		print("\n  },")
	}
	print("\n }")
	print("\n}\n")
}
