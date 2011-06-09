// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"container/list"
	"container/vector"
	"fmt"
	"go/parser"
        "go/printer"
        tok "go/token"
	"os"
	"strconv"
	"template"
)

const PEG_HEADER_TEMPLATE =
`package ${PackageName}

import (
	"bytes"
	"fmt"
 	"os"
)

type ${StructName} struct {
	${StructVariables}
	Buffer		string
	Min, Max	int
	rules		[${RulesCount}]func() bool
}

type parseError struct {
	p *${StructName}
}

func (p *${StructName}) Parse() os.Error {
	if p.rules[0]() {
		return nil
	}
	return &parseError{ p }
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

func (p *${StructName}) Init() {
	var position int

	${.section HasActions}
 	actions := [...]func(buffer string, begin, end int) {
		${.repeated section Actions}
		/* ${GetId} */
		func(buffer string, begin, end int) {
			${String}
		},
		${.end}
	}

	var thunkPosition, begin, end int
	thunks := make([]struct {action uint${Bits}; begin, end int}, 32)
	do := func(action uint${Bits}) {
		if thunkPosition == len(thunks) {
			newThunks := make([]struct {action uint${Bits}; begin, end int}, 2 * len(thunks))
			copy(newThunks, thunks)
			thunks = newThunks
		}
		thunks[thunkPosition].action = action
		thunks[thunkPosition].begin = begin
		thunks[thunkPosition].end = end
		thunkPosition++
	}

	${.section HasCommit}
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
	${.end}
	${.end}

	${.section HasDot}
	matchDot := func() bool {
		if position < len(p.Buffer) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	${.end}

	${.section HasCharacter}
	matchChar := func(c byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] == c) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	${.end}

	${.section HasString}
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
	${.end}

	${.section HasRange}
	matchRange := func(lower byte, upper byte) bool {
		if (position < len(p.Buffer)) && (p.Buffer[position] >= lower) && (p.Buffer[position] <= upper) {
			position++
			return true
		} else if position >= p.Max {
			p.Max = position
		}
		return false
	}
	${.end}`

type Type uint8
const (
	TypeUnknown Type = iota
	TypeRule
	TypeVariable
	TypeName
	TypeDot
	TypeCharacter
	TypeRange
	TypeString
	TypePredicate
	TypeCommit
	TypeBegin
	TypeEnd
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
	TypeNil
	TypeLast
)

func (t Type) GetType() Type {
	return t
}

type Node interface {
	fmt.Stringer
	GetType() Type
	GetId() int
	SetId(id int)
}

/* Used to represent TypeRule*/
type Rule interface {
	Node
	GetExpression() Node
	SetExpression(e Node)
}

type rule struct {
	name       string
	id         int
	expression Node
}

func (r *rule) GetType() Type {
	return TypeRule
}

func (r *rule) GetId() int {
	return r.id
}

func (r *rule) SetId(id int) {
	r.id = id
}

func (r *rule) GetExpression() Node {
	return r.expression
}

func (r *rule) SetExpression(e Node) {
	r.expression = e
}

func (r *rule) String() string {
	return r.name
}

/* Used to represent TypeAction, TypeName, TypeDot, TypeCharacter, TypeString, TypePredicate, and TypeNil. */
type Token interface {
	Node
}

type token struct {
	Type
	string
	id 	int
}

func (t *token) String() string {
	return t.string
}

func (t *token) GetId() int {
	return t.id
}

func (t *token) SetId(id int) {
	t.id = id
}

/* Used to represent a TypeAlternate, TypeSequence, TypeUnorderedAlternate, TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypeRange, or TypePlus */
type List interface {
	Node
	SetType(t Type)

	Init() *list.List
	Front() *list.Element
	PushBack(value interface{}) *list.Element
	Len() int
}

type nodeList struct {
	Type
	string
	id int

	list.List
}

func (l *nodeList) SetType(t Type) {
	l.Type = t
}

func (l *nodeList) String() string {
	return l.string
}

func (l *nodeList) GetId() int {
	return l.id
}

func (l *nodeList) SetId(id int) {
	l.id = id
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
	rules      map[string]*rule
	rulesCount map[string]uint
	ruleId     int
	list.List
	classes         map[string]*characterClass
	stack           [1024]Node
	top             int
	inline, _switch bool

	PackageName     string
	StructName      string
	StructVariables string
	RulesCount      int
	Bits            int
	HasActions	bool
	Actions         vector.Vector
	HasCommit       bool
	HasDot          bool
	HasCharacter    bool
	HasString       bool
	HasRange        bool
}

func New(inline, _switch bool) *Tree {
	return &Tree{rules: make(map[string]*rule),
		rulesCount: make(map[string]uint),
		classes:    make(map[string]*characterClass),
		inline:     inline,
		_switch:    _switch}
}

func (t *Tree) push(n Node) {
	t.top++
	t.stack[t.top] = n
}

func (t *Tree) pop() Node {
	n := t.stack[t.top]
	t.top--
	return n
}

func (t *Tree) currentRule() Rule {
	return t.stack[1].(Rule)
}

func (t *Tree) AddRule(name string) {
	t.push(&rule{name: name, id: t.ruleId})
	t.ruleId++
}

func (t *Tree) AddExpression() {
	expression := t.pop()
	rule := t.pop().(Rule)
	rule.SetExpression(expression)
	t.PushBack(rule)
}

func (t *Tree) AddName(text string) {
	t.rules[text] = &rule{}
	t.push(&token{Type: TypeName, string: text})
}

var (
	dot = &token{Type: TypeDot, string: "."}
	commit = &token{Type: TypeCommit, string: "commit"}
	begin = &token{Type: TypeBegin, string: "<"}
	end = &token{Type: TypeEnd, string: ">"}
	nilNode = &token{Type: TypeNil, string: "<nil>"}
)

func (t *Tree) AddDot() { t.push(dot) }
func (t *Tree) AddCharacter(text string) {
	t.push(&token{Type: TypeCharacter, string: text})
}
func (t *Tree) AddOctalCharacter(text string) {
	octal, _ := strconv.Btoui64(text, 8)
	t.push(&token{Type: TypeCharacter, string: string(octal)})
}
func (t *Tree) AddPredicate(text string) { t.push(&token{Type: TypePredicate, string: text}) }
func (t *Tree) AddCommit() { t.push(commit) }
func (t *Tree) AddBegin() { t.push(begin) }
func (t *Tree) AddEnd() { t.push(end) }
func (t *Tree) AddNil() { t.push(nilNode) }
func (t *Tree) AddAction(text string) { t.push(&token{Type: TypeAction, string: text}) }
func (t *Tree) AddPackage(text string) { t.PushBack(&token{Type: TypePackage, string: text}) }
func (t *Tree) AddState(text string) {
	peg := t.pop().(List)
	peg.PushBack(&token{Type: TypeState, string: text})
	t.PushBack(peg)
}

func (t *Tree) addList(listType Type) {
	a := t.pop()
	b := t.pop()
	var l List
	if b.GetType() == listType {
		l = b.(List)
	} else {
		l = &nodeList{Type: listType}
		l.PushBack(b)
	}
	l.PushBack(a)
	t.push(l)
}
func (t *Tree) AddAlternate() { t.addList(TypeAlternate) }
func (t *Tree) AddSequence() { t.addList(TypeSequence) }
func (t *Tree) AddRange()    { t.addList(TypeRange) }

func (t *Tree) addFix(fixType Type) {
	n := &nodeList{Type: fixType}
	n.PushBack(t.pop())
	t.push(n)
}
func (t *Tree) AddPeekFor()        { t.addFix(TypePeekFor) }
func (t *Tree) AddPeekNot()        { t.addFix(TypePeekNot) }
func (t *Tree) AddQuery()          { t.addFix(TypeQuery) }
func (t *Tree) AddStar()           { t.addFix(TypeStar) }
func (t *Tree) AddPlus()           { t.addFix(TypePlus) }
func (t *Tree) AddPeg(text string) { t.push(&nodeList{Type: TypePeg, string: text}) }

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

func (t *Tree) Compile(file string) {
	counts := [TypeLast]uint{}

	for element := t.Front(); element != nil; element = element.Next() {
		node := element.Value.(Node)
		switch node.GetType() {
		case TypePackage:
			t.PackageName = node.(Token).String()
		case TypePeg:
			peg := node.(List)
			t.StructName = peg.String()
			t.StructVariables = peg.Front().Value.(Token).String()
		case TypeRule:
			rule := node.(*rule)
			t.rules[rule.String()] = rule
		}
	}

	for name, r := range t.rules {
		if r.name == "" {
			r := &rule{name: name, id: t.ruleId}
			t.ruleId++
			t.rules[name] = r
			t.PushBack(r)
		}
	}
	t.RulesCount = len(t.rules)

	join([]func(){
		func() {
			var countTypes func(node Node)
			countTypes = func(node Node) {
				nodeType := node.GetType()
				id := counts[nodeType]
				counts[nodeType]++
				switch nodeType {
				case TypeRule:
					countTypes(node.(Rule).GetExpression())
				case TypeAction:
					node.SetId(int(id))
					t.Actions.Push(node)
				case TypeAlternate, TypeUnorderedAlternate, TypeSequence,
					TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
					for element := node.(List).Front(); element != nil; element = element.Next() {
						countTypes(element.Value.(Node))
					}
				}
			}
			for _, rule := range t.rules {
				countTypes(rule)
			}
		},
		func() {
			var countRules func(node Node)
			ruleReached := make([]bool, len(t.rules))
			countRules = func(node Node) {
				switch node.GetType() {
				case TypeRule:
					rule := node.(Rule)
					name, id := rule.String(), rule.GetId()
					if count, ok := t.rulesCount[name]; ok {
						t.rulesCount[name] = count + 1
					} else {
						t.rulesCount[name] = 1
					}
					if ruleReached[id] {
						return
					}
					ruleReached[id] = true
					countRules(rule.GetExpression())
				case TypeName:
					countRules(t.rules[node.String()])
				case TypeAlternate, TypeUnorderedAlternate, TypeSequence,
					TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
					for element := node.(List).Front(); element != nil; element = element.Next() {
						countRules(element.Value.(Node))
					}
				}
			}
			for element := t.Front(); element != nil; element = element.Next() {
				node := element.Value.(Node)
				if node.GetType() == TypeRule {
					countRules(node.(*rule))
					break
				}
			}
		},
		func() {
			var checkRecursion func(node Node) bool
			ruleReached := make([]bool, len(t.rules))
			checkRecursion = func(node Node) bool {
				switch node.GetType() {
				case TypeRule:
					rule := node.(Rule)
					id := rule.GetId()
					if ruleReached[id] {
						fmt.Fprintf(os.Stderr, "possible infinite left recursion in rule '%v'\n", node)
						return false
					}
					ruleReached[id] = true
					consumes := checkRecursion(rule.GetExpression())
					ruleReached[id] = false
					return consumes
				case TypeAlternate:
					for element := node.(List).Front(); element != nil; element = element.Next() {
						if !checkRecursion(element.Value.(Node)) {
							return false
						}
					}
					return true
				case TypeSequence:
					for element := node.(List).Front(); element != nil; element = element.Next() {
						if checkRecursion(element.Value.(Node)) {
							return true
						}
					}
				case TypeName:
					return checkRecursion(t.rules[node.String()])
				case TypePlus:
					return checkRecursion(node.(List).Front().Value.(Node))
				case TypeCharacter, TypeString:
					return len(node.String()) > 0
				case TypeDot:
					return true
				}
				return false
			}
			for _, rule := range t.rules {
				checkRecursion(rule)
			}
		}})

	if t._switch {
		var optimizeAlternates func(node Node) (consumes, eof, peek bool, class *characterClass)
		cache := make([]struct {
			reached, consumes, eof, peek bool
			class                        *characterClass
		}, len(t.rules))
		optimizeAlternates = func(node Node) (consumes, eof, peek bool, class *characterClass) {
			switch node.GetType() {
			case TypeRule:
				rule := node.(Rule)
				cache := &cache[rule.GetId()]
				if cache.reached {
					consumes, eof, peek, class = cache.consumes, cache.eof, cache.peek, cache.class
					return
				}
				cache.reached = true
				consumes, eof, peek, class = optimizeAlternates(rule.GetExpression())
				cache.consumes, cache.eof, cache.peek, cache.class = consumes, eof, peek, class
			case TypeName:
				consumes, eof, peek, class = optimizeAlternates(t.rules[node.String()])
			case TypeDot:
				consumes, class = true, new(characterClass)
				for index, _ := range *class {
					class[index] = 0xff
				}
			case TypeString, TypeCharacter:
				consumes, class = true, new(characterClass)
				class.add(node.String()[0])
			case TypeRange:
				consumes, class = true, new(characterClass)
				list := node.(List)
				element := list.Front()
				lower := element.Value.(Node).String()[0]
				element = element.Next()
				upper := element.Value.(Node).String()[0]
				for c := lower; c <= upper; c++ {
					class.add(c)
				}
			case TypeAlternate:
				consumes, peek, class = true, true, new(characterClass)
				alternate := node.(List)
				mconsumes, meof, mpeek, properties, c :=
					consumes, eof, peek, make([]struct {
						intersects bool
						class      *characterClass
					}, alternate.Len()), 0
				for element := alternate.Front(); element != nil; element = element.Next() {
					mconsumes, meof, mpeek, properties[c].class = optimizeAlternates(element.Value.(Node))
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
						0, &nodeList{Type: TypeUnorderedAlternate}, &nodeList{Type: TypeAlternate}, 0
					for element := alternate.Front(); element != nil; element = element.Next() {
						if properties[c].intersects {
							ordered.PushBack(element.Value)
						} else {
							class := &nodeList{Type: TypeUnorderedAlternate}
							for d := 0; d < 256; d++ {
								if properties[c].class.has(uint8(d)) {
									class.PushBack(&token{Type: TypeCharacter, string: string(d)})
								}
							}

							sequence, predicate, length :=
								&nodeList{Type: TypeSequence}, &nodeList{Type: TypePeekFor}, properties[c].class.len()
							if length == 0 {
								class.PushBack(nilNode)
							}
							predicate.PushBack(class)
							sequence.PushBack(predicate)
							sequence.PushBack(element.Value)

							if element.Value.(Node).GetType() == TypeNil {
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
					alternate.Init()
					if ordered.Len() == 0 {
						alternate.SetType(TypeUnorderedAlternate)
						for element := unordered.Front(); element != nil; element = element.Next() {
							alternate.PushBack(element.Value)
						}
					} else {
						for element := ordered.Front(); element != nil; element = element.Next() {
							alternate.PushBack(element.Value)
						}
						alternate.PushBack(unordered)
					}
				}
			case TypeSequence:
				sequence := node.(List)
				meof, classes, c, element :=
					eof, make([]struct {
						peek  bool
						class *characterClass
					}, sequence.Len()), 0, sequence.Front()
				for ; !consumes && element != nil; element, c = element.Next(), c + 1 {
					consumes, meof, classes[c].peek, classes[c].class = optimizeAlternates(element.Value.(Node))
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
					optimizeAlternates(element.Value.(Node))
				}
			case TypePeekNot:
				peek = true
				_, eof, _, class = optimizeAlternates(node.(List).Front().Value.(Node))
				eof = !eof
				class = class.copy()
				class.complement()
			case TypePeekFor:
				peek = true
				fallthrough
			case TypeQuery, TypeStar:
				_, eof, _, class = optimizeAlternates(node.(List).Front().Value.(Node))
			case TypePlus:
				consumes, eof, peek, class = optimizeAlternates(node.(List).Front().Value.(Node))
			case TypeAction, TypeNil:
				class = new(characterClass)
			}
			return
		}
		for element := t.Front(); element != nil; element = element.Next() {
			node := element.Value.(Node)
			if node.GetType() == TypeRule {
				optimizeAlternates(node.(*rule))
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
		fileSet := tok.NewFileSet()
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
	printSave := func(n uint) { print("\n   position%d := position", n) }
	printRestore := func(n uint) { print("   position = position%d", n) }
	printTemplate := func(s string) {
		templateEngine := template.New(nil)
		templateEngine.SetDelims("${", "}")
		if error := templateEngine.Parse(s); error != nil { panic(error) }
		if error := templateEngine.Execute(&buffer, t); error != nil { panic(error) }
	}

	if t.HasActions = counts[TypeAction] > 0; t.HasActions {
		bits := 0
		for length := t.Actions.Len(); length != 0; length >>= 1 {
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

		printSave = func(n uint) { print("\n   position%d,  thunkPosition%d := position, thunkPosition", n, n) }
		printRestore = func(n uint) { print("   position, thunkPosition = position%d, thunkPosition%d", n, n) }
	}

	t.HasCommit = counts[TypeCommit] > 0
	t.HasDot = counts[TypeDot] > 0
	t.HasCharacter = counts[TypeCharacter] > 0
	t.HasString = counts[TypeString] > 0
	t.HasRange = counts[TypeRange] > 0
	printTemplate(PEG_HEADER_TEMPLATE)

	var printRule func(node Node)
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
	printRule = func(node Node) {
		switch node.GetType() {
		case TypeRule:
			print("%v <- ", node)
			printRule(node.(Rule).GetExpression())
		case TypeDot:
			print(".")
		case TypeName:
			print("%v", node)
		case TypeCharacter:
			print("'%v'", escape(node.String()))
		case TypeString:
			s := escape(node.String())
			print("'%v'", s[1:len(s) - 1])
		case TypeRange:
			list := node.(List)
			element := list.Front()
			lower := element.Value
			element = element.Next()
			upper := element.Value
			print("[%v-%v]", lower, upper)
		case TypePredicate:
			print("&{%v}", node)
		case TypeAction:
			print("{%v}", node)
		case TypeCommit:
			print("commit")
		case TypeBegin:
			print("<")
		case TypeEnd:
			print(">")
		case TypeAlternate:
			print("(")
			list := node.(List)
			element := list.Front()
			printRule(element.Value.(Node))
			for element = element.Next(); element != nil; element = element.Next() {
				print(" / ")
				printRule(element.Value.(Node))
			}
			print(")")
		case TypeUnorderedAlternate:
			print("(")
			element := node.(List).Front()
			printRule(element.Value.(Node))
			for element = element.Next(); element != nil; element = element.Next() {
				print(" | ")
				printRule(element.Value.(Node))
			}
			print(")")
		case TypeSequence:
			print("(")
			element := node.(List).Front()
			printRule(element.Value.(Node))
			for element = element.Next(); element != nil; element = element.Next() {
				print(" ")
				printRule(element.Value.(Node))
			}
			print(")")
		case TypePeekFor:
			print("&")
			printRule(node.(List).Front().Value.(Node))
		case TypePeekNot:
			print("!")
			printRule(node.(List).Front().Value.(Node))
		case TypeQuery:
			printRule(node.(List).Front().Value.(Node))
			print("?")
		case TypeStar:
			printRule(node.(List).Front().Value.(Node))
			print("*")
		case TypePlus:
			printRule(node.(List).Front().Value.(Node))
			print("+")
		case TypeNil:
		default:
			fmt.Fprintf(os.Stderr, "illegal node type: %v\n", node.GetType())
		}
	}
	compile = func(node Node, ko uint) {
		switch node.GetType() {
		case TypeRule:
			fmt.Fprintf(os.Stderr, "internal error #1 (%v)\n", node)
		case TypeDot:
			print("\n   if !matchDot() {")
			printJump(ko)
			print("}")
		case TypeName:
			name := node.String()
			rule := t.rules[name]
			if t.inline && t.rulesCount[name] == 1 {
				compile(rule.GetExpression(), ko)
				return
			}
			print("\n   if !p.rules[%d]() {", rule.GetId())
			printJump(ko)
			print("}")
		case TypeRange:
			list := node.(List)
			element := list.Front()
			lower := element.Value.(Node)
			element = element.Next()
			upper := element.Value.(Node)
			print("\n   if !matchRange('%v', '%v') {", escape(lower.String()), escape(upper.String()))
			printJump(ko)
			print("}")
		case TypeCharacter:
			print("\n   if !matchChar('%v') {", escape(node.String()))
			printJump(ko)
			print("}")
		case TypeString:
			print("\n   if !matchString(%v) {", strconv.Quote(node.String()))
			printJump(ko)
			print("}")
		case TypePredicate:
			print("\n   if !(%v) {", node)
			printJump(ko)
			print("}")
		case TypeAction:
			print("\n   do(%d)", node.GetId())
		case TypeCommit:
			print("\n   if !(commit(thunkPosition0)) {")
			printJump(ko)
			print("}")
		case TypeBegin:
			if t.HasActions {
				print("\n   begin = position")
			}
		case TypeEnd:
			if t.HasActions {
				print("\n   end = position")
			}
		case TypeAlternate:
			list := node.(List)
			ok := label
			label++
			printBegin()
			element := list.Front()
			if element.Next() != nil {
				printSave(ok)
			}
			for element.Next() != nil {
				next := label
				label++
				compile(element.Value.(Node), next)
				printJump(ok)
				printLabel(next)
				printRestore(ok)
				element = element.Next()
			}
			compile(element.Value.(Node), ko)
			printEnd()
			printLabel(ok)
		case TypeUnorderedAlternate:
			list := node.(List)
			done, ok := ko, label
			label++
			printBegin()
			print("\n   if position == len(p.Buffer) {")
			printJump(done)
			print("}")
			print("\n   switch p.Buffer[position] {")
			element := list.Front()
			for ; element.Next() != nil; element = element.Next() {
				sequence := element.Value.(List).Front()
				class := sequence.Value.(List).Front().Value.(List)
				sequence = sequence.Next()
				print("\n   case")
				comma := false
				for character := class.Front(); character != nil; character = character.Next() {
					if comma {
						print(",")
					} else {
						comma = true
					}
					print(" '%s'", escape(character.Value.(Token).String()))
				}
				print(":")
				compile(sequence.Value.(Node), done)
				print("\nbreak")
			}
			print("\n   default:")
			compile(element.Value.(List).Front().Next().Value.(Node), done)
			print("\nbreak")
			print("\n   }")
			printEnd()
			printLabel(ok)
		case TypeSequence:
			for element := node.(List).Front(); element != nil; element = element.Next() {
				compile(element.Value.(Node), ko)
			}
		case TypePeekFor:
			ok := label
			label++
			printBegin()
			printSave(ok)
			compile(node.(List).Front().Value.(Node), ko)
			printRestore(ok)
			printEnd()
		case TypePeekNot:
			ok := label
			label++
			printBegin()
			printSave(ok)
			compile(node.(List).Front().Value.(Node), ok)
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
			compile(node.(List).Front().Value.(Node), qko)
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
			compile(node.(List).Front().Value.(Node), out)
			printJump(again)
			printLabel(out)
			printRestore(out)
			printEnd()
		case TypePlus:
			again := label
			label++
			out := label
			label++
			compile(node.(List).Front().Value.(Node), ko)
			printLabel(again)
			printBegin()
			printSave(out)
			compile(node.(List).Front().Value.(Node), out)
			printJump(again)
			printLabel(out)
			printRestore(out)
			printEnd()
		case TypeNil:
		default:
			fmt.Fprintf(os.Stderr, "illegal node type: %v\n", node.GetType())
		}
	}

	/* lets figure out which jump labels are going to be used with this dry compile */
	printTemp, print := print, func(format string, a ...interface{}) {}
	for element := t.Front(); element != nil; element = element.Next() {
		node := element.Value.(Node)
		if node.GetType() != TypeRule {
			continue
		}
		rule := node.(*rule)
		expression := rule.GetExpression()
		if expression == nilNode {
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
	print("\n p.rules = [...]func() bool {")
	for element := t.Front(); element != nil; element = element.Next() {
		node := element.Value.(Node)
		if node.GetType() != TypeRule {
			continue
		}
		rule := node.(*rule)
		expression := rule.GetExpression()
		if expression == nilNode {
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
			printSave(0)
		}
		compile(expression, ko)
		print("\n   return true")
		if labels[ko] {
			printLabel(ko)
			printRestore(0)
			print("\n   return false")
		}
		print("\n  },")
	}
	print("\n }")
	print("\n}\n")
}
