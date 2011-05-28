// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package peg

import (
	"fmt"
	"container/list"
	"os"
)

type Type uint8

const (
	TypeUnknown Type = iota
	TypeRule
	TypeVariable
	TypeName
	TypeDot
	TypeCharacter
	TypeString
	TypeClass
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

func (t Type) IsSafe() bool {
	return (t == TypeQuery) || (t == TypeStar)
}

type Node interface {
	fmt.Stringer
	GetType() Type
}

/* Used to represent TypeRule*/
type Rule interface {
	Node
	GetId() int
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

func (r *rule) GetExpression() Node {
	if r.expression == nil {
		return nilNode
	}
	return r.expression
}

func (r *rule) SetExpression(e Node) {
	r.expression = e
}

func (r *rule) String() string {
	return r.name
}

/* Used to represent TypeName, TypeDot, TypeCharacter, TypeString, TypeClass, TypePredicate, and TypeNil. */
type Token interface {
	Node
	GetClass() *characterClass
}

type token struct {
	Type
	string
	class *characterClass
}

func (t *token) GetClass() *characterClass {
	return t.class
}

func (t *token) String() string {
	return t.string
}

var nilNode = &token{Type: TypeNil, string: "<nil>"}

/* Used to represent TypeAction. */
type Action interface {
	Node
	GetId() int
	GetRule() string
}

type action struct {
	text string
	id   int
	rule string
}

func (a *action) GetType() Type {
	return TypeAction
}

func (a *action) String() string {
	return a.text
}

func (a *action) GetId() int {
	return a.id
}

func (a *action) GetRule() string {
	return a.rule
}

/* Used to represent a TypeAlternate or TypeSequence. */
type List interface {
	Node
	SetType(t Type)
	SetLastIsEmpty(lastIsEmpty bool)
	GetLastIsEmpty() bool
	Init() *list.List
	Front() *list.Element
	PushBack(value interface{}) *list.Element
	Len() int
}

type nodeList struct {
	Type
	list.List
	lastIsEmpty bool
}

func (l *nodeList) SetType(t Type) {
	l.Type = t
}

func (l *nodeList) SetLastIsEmpty(lastIsEmpty bool) {
	l.lastIsEmpty = lastIsEmpty
}

func (l *nodeList) GetLastIsEmpty() bool {
	return l.lastIsEmpty
}

func (l *nodeList) String() string {
	i := l.List.Front()
	s := "(" + i.Value.(fmt.Stringer).String()
	for i = i.Next(); i != nil; i = i.Next() {
		s += " / " + i.Value.(fmt.Stringer).String()
	}
	return s + ")"
}

/* Used to represent a TypePeekFor, TypePeekNot, TypeQuery, TypeStar, or TypePlus */
type Fix interface {
	Node
	GetNode() Node
	SetNode(node Node)
}

type fix struct {
	Type
	string
	node Node
}

func (f *fix) String() string {
	return f.string
}

func (f *fix) GetNode() Node {
	return f.node
}

func (f *fix) SetNode(node Node) {
	f.node = node
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
func (c *characterClass) String() (class string) {
	escape := func(c uint8) string {
		s := ""
		switch uint8(c) {
		case '\a':
			s = `\a` /* bel */
		case '\b':
			s = `\b` /* bs */
		case '\f':
			s = `\f` /* ff */
		case '\n':
			s = `\n` /* nl */
		case '\r':
			s = `\r` /* cr */
		case '\t':
			s = `\t` /* ht */
		case '\v':
			s = `\v` /* vt */
		case '\'':
			s = `\'` /* ' */
		case '"':
			s = `\"` /* " */
		case '[':
			s = `\[` /* [ */
		case ']':
			s = `\]` /* ] */
		case '\\':
			s = `\\` /* \ */
		case '-':
			s = `\-` /* - */
		default:
			s = fmt.Sprintf("%c", c)
		}
		return s
	}
	class = ""
	l := 0
	for character := 0; character < 256; character++ {
		if c.has(uint8(character)) {
			if l == 0 {
				class += escape(uint8(character))
			}
			l++
		} else {
			if l == 2 {
				class += escape(uint8(character - 1))
			} else if l > 2 {
				class += "-" + escape(uint8(character-1))
			}
			l = 0
		}
	}
	if l >= 2 {
		class += "-" + escape(255)
	}
	return
}

/* A tree data structure into which a PEG can be parsed. */
type Tree struct {
	rules      map[string]*rule
	rulesCount map[string]uint
	ruleId     int
	list.List
	actions         list.List
	classes         map[string]*characterClass
	stack           [1024]Node
	top             int
	inline, _switch bool
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

var dot *token = &token{Type: TypeDot, string: "."}

func (t *Tree) AddDot() { t.push(dot) }
func (t *Tree) AddString(text string) {
	length := len(text)
	if (length == 1) || ((length == 2) && (text[0] == '\\')) {
		t.push(&token{Type: TypeCharacter, string: text})
	} else {
		t.push(&token{Type: TypeString, string: text})
	}
}
func (t *Tree) AddClass(text string) {
	t.push(&token{Type: TypeClass, string: text})
	if c, ok := t.classes[text]; !ok {
		c = new(characterClass)
		t.classes[text] = c
		inverse := false
		if text[0] == '^' {
			inverse = true
			text = text[1:]
		}
		var last uint8
		hasLast := false
		for i := 0; i < (len(text) - 1); i++ {
			switch {
			case (text[i] == '-') && hasLast:
				i++
				for j := last; j <= text[i]; j++ {
					c.add(j)
				}
				hasLast = false
			case (text[i] == '\\'):
				i++
				last, hasLast = text[i], true
				switch last {
				case 'a':
					last = '\a' /* bel */
				case 'b':
					last = '\b' /* bs */
				case 'f':
					last = '\f' /* ff */
				case 'n':
					last = '\n' /* nl */
				case 'r':
					last = '\r' /* cr */
				case 't':
					last = '\t' /* ht */
				case 'v':
					last = '\v' /* vt */
				}
				c.add(last)
			default:
				last, hasLast = text[i], true
				c.add(last)
			}
		}
		c.add(text[len(text)-1])
		if inverse {
			c.complement()
		}
	}
}
func (t *Tree) AddPredicate(text string) { t.push(&token{Type: TypePredicate, string: text}) }

var commit *token = &token{Type: TypeCommit, string: "commit"}

func (t *Tree) AddCommit() { t.push(commit) }

var begin *token = &token{Type: TypeBegin, string: "<"}

func (t *Tree) AddBegin() { t.push(begin) }

var end *token = &token{Type: TypeEnd, string: ">"}

func (t *Tree) AddEnd() { t.push(end) }
func (t *Tree) AddAction(text string) {
	a := &action{text: text, id: t.actions.Len(), rule: t.currentRule().String()}
	t.actions.PushBack(a)
	t.push(a)
}
func (t *Tree) AddPackage(text string) { t.PushBack(&token{Type: TypePackage, string: text}) }
func (t *Tree) AddState(text string) {
	peg := t.pop().(Fix)
	peg.SetNode(&token{Type: TypeState, string: text})
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
func (t *Tree) AddEmptyAlternate() {
	a := t.pop()
	var l List
	if a.GetType() == TypeAlternate {
		l = a.(List)
		l.SetLastIsEmpty(true)
	} else {
		l = &nodeList{Type: TypeAlternate, lastIsEmpty: true}
		l.PushBack(a)
	}
	t.push(l)
}
func (t *Tree) AddSequence() { t.addList(TypeSequence) }

func (t *Tree) addFix(fixType Type) { t.push(&fix{Type: fixType, node: t.pop()}) }
func (t *Tree) AddPeekFor()         { t.addFix(TypePeekFor) }
func (t *Tree) AddPeekNot()         { t.addFix(TypePeekNot) }
func (t *Tree) AddQuery()           { t.addFix(TypeQuery) }
func (t *Tree) AddStar()            { t.addFix(TypeStar) }
func (t *Tree) AddPlus()            { t.addFix(TypePlus) }
func (t *Tree) AddPeg(text string)  { t.push(&fix{Type: TypePeg, string: text}) }

func join(tasks []func()) {
	length := len(tasks)
	done := make(chan int, length)
	for _, task := range tasks {
		go func(task func()) { task(); done <- 1 }(task)
	}
	for d := <-done; d < length; d += <-done {
	}
}

func (t *Tree) Compile(file string) {
	_package, state, name, counts :=
		"", "", "", [TypeLast]uint{}
	for element := t.Front(); element != nil; element = element.Next() {
		node := element.Value.(Node)
		switch node.GetType() {
		case TypePackage:
			_package = node.(Token).String()
		case TypePeg:
			peg := node.(Fix)
			name = peg.String()
			state = peg.GetNode().(Token).String()
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

	join([]func(){
		func() {
			var countTypes func(node Node)
			countTypes = func(node Node) {
				t := node.GetType()
				counts[t]++
				switch t {
				case TypeRule:
					countTypes(node.(Rule).GetExpression())
				case TypeAlternate, TypeUnorderedAlternate, TypeSequence:
					for element := node.(List).Front(); element != nil; element = element.Next() {
						countTypes(element.Value.(Node))
					}
				case TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
					countTypes(node.(Fix).GetNode())
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
				case TypeAlternate, TypeUnorderedAlternate, TypeSequence:
					for element := node.(List).Front(); element != nil; element = element.Next() {
						countRules(element.Value.(Node))
					}
				case TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
					countRules(node.(Fix).GetNode())
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
					return checkRecursion(node.(Fix).GetNode())
				case TypeCharacter, TypeString:
					return len(node.String()) > 0
				case TypeDot, TypeClass:
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
				b := node.String()[0]
				if b == '\\' {
					b = node.String()[1]
					switch b {
					case 'a':
						b = '\a' /* bel */
					case 'b':
						b = '\b' /* bs */
					case 'f':
						b = '\f' /* ff */
					case 'n':
						b = '\n' /* nl */
					case 'r':
						b = '\r' /* cr */
					case 't':
						b = '\t' /* ht */
					case 'v':
						b = '\v' /* vt */
					}
				}
				class.add(b)
			case TypeClass:
				consumes, class = true, t.classes[node.String()]
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
							class := &token{Type: TypeClass, string: properties[c].class.String(), class: properties[c].class}
							sequence, predicate :=
								&nodeList{Type: TypeSequence}, &fix{Type: TypePeekFor, node: class}
							sequence.PushBack(predicate)
							sequence.PushBack(element.Value)
							length := properties[c].class.len()
							if length > max {
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
				for ; !consumes && element != nil; element, c = element.Next(), c+1 {
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
				_, eof, _, class = optimizeAlternates(node.(Fix).GetNode())
				eof = !eof
				class = class.copy()
				class.complement()
			case TypePeekFor:
				peek = true
				fallthrough
			case TypeQuery, TypeStar:
				_, eof, _, class = optimizeAlternates(node.(Fix).GetNode())
			case TypePlus:
				consumes, eof, peek, class = optimizeAlternates(node.(Fix).GetNode())
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
	print := func(format string, a ...interface{}) { fmt.Fprintf(out, format, a...) }
	printSave := func(n uint) { print("\n   position%d := position", n) }
	printRestore := func(n uint) { print("   position = position%d", n) }

	print(
		`package %v
import "fmt"
type %v struct {%v
 Buffer string
 Min, Max int
 rules [%d]func() bool
}
func (p *%v) Parse() bool {
 if p.rules[0]() {
  return true
 }
 return false
}
func (p *%v) PrintError() {
 line := 1
 character := 0
 for i, c := range p.Buffer[0:] {
  if c == '\n' {
   line++
   character = 0
  } else {
   character++
  }
  if i == p.Min {
   if p.Min != p.Max {
    fmt.Printf("parse error after line %%v character %%v\n", line, character)
   } else {break}
  } else if i == p.Max {break}
 }
 fmt.Printf("parse error: unexpected ")
 if p.Max >= len(p.Buffer) {
  fmt.Printf("end of file found\n")
 } else {
  fmt.Printf("'%%c' at line %%v character %%v\n", p.Buffer[p.Max], line, character)
 }
}
func (p *%v) Init() {
 var position int`, _package, name, state, len(t.rules), name, name, name)

	hasActions := t.actions.Len() != 0
	if hasActions {
		bits := 0
		for length := t.actions.Len(); length != 0; length >>= 1 {
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
		print("\n actions := [...]func(buffer string, begin, end int) {\n")
		for i := t.actions.Front(); i != nil; i = i.Next() {
			a := i.Value.(Action)
			print("  /* %v %v */\n", a.GetId(), a.GetRule())
			print("  func(buffer string, begin, end int) {\n")
			print("   %v\n", a)
			print("  },\n")
		}
		print(
			` }
 var thunkPosition, begin, end int
 thunks := make([]struct {action uint%d; begin, end int}, 32)
 do := func(action uint%d) {
  if thunkPosition == len(thunks) {
   newThunks := make([]struct {action uint%d; begin, end int}, 2 * len(thunks))
   copy(newThunks, thunks)
   thunks = newThunks
  }
  thunks[thunkPosition].action = action
  thunks[thunkPosition].begin = begin
  thunks[thunkPosition].end = end
  thunkPosition++
 }`, bits, bits, bits)
		if counts[TypeCommit] > 0 {
			print(
				`
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
 }`)
		}
		printSave = func(n uint) { print("\n   position%d,  thunkPosition%d := position, thunkPosition", n, n) }
		printRestore = func(n uint) { print("   position, thunkPosition = position%d, thunkPosition%d", n, n) }
	}

	if counts[TypeDot] > 0 {
		print(
			`
 matchDot := func() bool {
  if position < len(p.Buffer) {
   position++
   return true
  } else if position >= p.Max {
   p.Max = position
  }
  return false
 }`)
	}
	if counts[TypeCharacter] > 0 {
		print(
			`
 matchChar := func(c byte) bool {
  if (position < len(p.Buffer)) && (p.Buffer[position] == c) {
   position++
   return true
  } else if position >= p.Max {
   p.Max = position
  }
  return false
 }`)
	}
	if counts[TypeString] > 0 {
		print(
			`
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
 }`)
	}

	classes := make(map[string]uint)
	if len(t.classes) != 0 {
		print("\n classes := [...][32]uint8 {\n")
		var index uint
		for className, classBitmap := range t.classes {
			classes[className] = index
			print("  [32]uint8{")
			for _, b := range *classBitmap {
				print("%d, ", b)
			}
			print("},\n")
			index++
		}
		print(
			` }
 matchClass := func(class uint) bool {
  if (position < len(p.Buffer)) &&
     ((classes[class][p.Buffer[position] >> 3] & (1 << (p.Buffer[position] & 7))) != 0) {
   position++
   return true
  } else if position >= p.Max {
   p.Max = position
  }
  return false
 }`)
	}

	var printRule func(node Node)
	var compile func(expression Node, ko uint)
	var label uint
	labels := make(map[uint]bool)
	printBegin := func() { print("\n   {") }
	printEnd := func() { print("\n   }") }
	printLabel := func(n uint, loop bool) {
		print("\n")
		if loop || labels[n] {
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
			expression := node.(Rule).GetExpression()
			if expression != nilNode {
				printRule(expression)
			}
		case TypeDot:
			print(".")
		case TypeName:
			print("%v", node)
		case TypeCharacter,
			TypeString:
			print("'%v'", node)
		case TypeClass:
			print("[%v]", node)
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
			if list.GetLastIsEmpty() {
				print(" /")
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
			printRule(node.(Fix).GetNode())
		case TypePeekNot:
			print("!")
			printRule(node.(Fix).GetNode())
		case TypeQuery:
			printRule(node.(Fix).GetNode())
			print("?")
		case TypeStar:
			printRule(node.(Fix).GetNode())
			print("*")
		case TypePlus:
			printRule(node.(Fix).GetNode())
			print("+")
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
		case TypeCharacter:
			print("\n   if !matchChar('%v') {", node)
			printJump(ko)
			print("}")
		case TypeString:
			print("\n   if !matchString(\"%v\") {", node)
			printJump(ko)
			print("}")
		case TypeClass:
			print("\n   if !matchClass(%d) {", classes[node.String()])
			printJump(ko)
			print("}")
		case TypePredicate:
			print("\n   if !(%v) {", node)
			printJump(ko)
			print("}")
		case TypeAction:
			print("\n   do(%d)", node.(Action).GetId())
		case TypeCommit:
			print("\n   if !(commit(thunkPosition0)) {")
			printJump(ko)
			print("}")
		case TypeBegin:
			if hasActions {
				print("\n   begin = position")
			}
		case TypeEnd:
			if hasActions {
				print("\n   end = position")
			}
		case TypeAlternate:
			list := node.(List)
			ok := label
			label++
			printBegin()
			element := list.Front()
			if element.Next() != nil || list.GetLastIsEmpty() {
				printSave(ok)
			}
			for element.Next() != nil {
				next := label
				label++
				compile(element.Value.(Node), next)
				printJump(ok)
				printLabel(next, false)
				printRestore(ok)
				element = element.Next()
			}
			if list.GetLastIsEmpty() {
				done := label
				label++
				compile(element.Value.(Node), done)
				printJump(ok)
				printLabel(done, false)
				printRestore(ok)
			} else {
				compile(element.Value.(Node), ko)
			}
			printEnd()
			printLabel(ok, false)
		case TypeUnorderedAlternate:
			list := node.(List)
			done, ok := ko, label
			label++
			printBegin()
			if list.GetLastIsEmpty() {
				done = label
				label++
				printSave(ok)
			}
			print("\n   if position == len(p.Buffer) {")
			printJump(done)
			print("}")
			print("\n   switch p.Buffer[position] {")
			element := list.Front()
			for ; element.Next() != nil; element = element.Next() {
				sequence := element.Value.(List).Front()
				class := sequence.Value.(Fix).GetNode().(Token).GetClass()
				sequence = sequence.Next()
				print("\n   case")
				comma := false
				for d := 0; d < 256; d++ {
					if class.has(uint8(d)) {
						if comma {
							print(",")
						}
						s := ""
						switch uint8(d) {
						case '\a':
							s = `\a` /* bel */
						case '\b':
							s = `\b` /* bs */
						case '\f':
							s = `\f` /* ff */
						case '\n':
							s = `\n` /* nl */
						case '\r':
							s = `\r` /* cr */
						case '\t':
							s = `\t` /* ht */
						case '\v':
							s = `\v` /* vt */
						case '\'':
							s = `\'` /* ' */
						default:
							s = fmt.Sprintf("%c", d)
						}
						print(" '%s'", s)
						comma = true
					}
				}
				print(":")
				compile(sequence.Value.(Node), done)
			}
			print("\n   default:")
			compile(element.Value.(List).Front().Next().Value.(Node), done)
			print("\n   }")
			if list.GetLastIsEmpty() {
				printJump(ok)
				printLabel(done, false)
				printRestore(ok)
			}
			printEnd()
			printLabel(ok, false)
		case TypeSequence:
			for element := node.(List).Front(); element != nil; element = element.Next() {
				compile(element.Value.(Node), ko)
			}
		case TypePeekFor:
			ok := label
			label++
			printBegin()
			printSave(ok)
			compile(node.(Fix).GetNode(), ko)
			printRestore(ok)
			printEnd()
		case TypePeekNot:
			ok := label
			label++
			printBegin()
			printSave(ok)
			compile(node.(Fix).GetNode(), ok)
			printJump(ko)
			printLabel(ok, false)
			printRestore(ok)
			printEnd()
		case TypeQuery:
			qko := label
			label++
			qok := label
			label++
			printBegin()
			printSave(qko)
			compile(node.(Fix).GetNode(), qko)
			printJump(qok)
			printLabel(qko, false)
			printRestore(qko)
			printEnd()
			printLabel(qok, false)
		case TypeStar:
			again := label
			label++
			out := label
			label++
			printLabel(again, true)
			printBegin()
			printSave(out)
			compile(node.(Fix).GetNode(), out)
			printJump(again)
			printLabel(out, false)
			printRestore(out)
			printEnd()
		case TypePlus:
			again := label
			label++
			out := label
			label++
			compile(node.(Fix).GetNode(), ko)
			printLabel(again, true)
			printBegin()
			printSave(out)
			compile(node.(Fix).GetNode(), out)
			printJump(again)
			printLabel(out, false)
			printRestore(out)
			printEnd()
		case TypeNil:
		default:
			fmt.Fprintf(os.Stderr, "illegal node type: %v\n", node.GetType())
		}
	}

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
		if !expression.GetType().IsSafe() {
			printSave(0)
		}
		compile(expression, ko)
		print("\n   return true")
		if !expression.GetType().IsSafe() {
			printLabel(ko, false)
			printRestore(0)
			print("\n   return false")
		}
		print("\n  },")
	}
	print("\n }")
	print("\n}\n")
}
