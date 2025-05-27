// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"iter"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"unicode"

	"github.com/pointlander/peg/set"
)

//go:embed peg.go.tmpl
var pegHeaderTemplate string

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
	TypeSpace
	TypeComment
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

/*
var TypeMap = [...]string{
	"TypeUnknown",
	"TypeRule",
	"TypeName",
	"TypeDot",
	"TypeCharacter",
	"TypeRange",
	"TypeString",
	"TypePredicate",
	"TypeStateChange",
	"TypeCommit",
	"TypeAction",
	"TypeSpace",
	"TypeComment",
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
	"TypeLast",
}

func (n *node) debug() {
	if len(n.string) == 1 {
		fmt.Printf("%v %v '%v' %d\n", n.id, TypeMap[n.Type], n.string, n.string[0])
		return
	}
	fmt.Printf("%v %v '%v'\n", n.id, TypeMap[n.Type], n.string)
}
*/

func (t Type) GetType() Type {
	return t
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

	parentDetect      bool
	parentMultipleKey bool
}

func (n *node) String() string {
	return n.string
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

func (n *node) GetID() int {
	return n.id
}

func (n *node) SetID(id int) {
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

func (n *node) Iterator() iter.Seq[*node] {
	element := n.Front()
	return func(yield func(*node) bool) {
		for element != nil {
			if !yield(element) {
				return
			}
			element = element.Next()
		}
	}
}

func (n *node) Iterator2() iter.Seq2[int, *node] {
	element := n.Front()
	return func(yield func(int, *node) bool) {
		i := 0
		for element != nil {
			if !yield(i, element) {
				return
			}
			i++
			element = element.Next()
		}
	}
}

func (n *node) ParentDetect() bool {
	return n.parentDetect
}

func (n *node) SetParentDetect(detect bool) {
	n.parentDetect = detect
}

func (n *node) ParentMultipleKey() bool {
	return n.parentMultipleKey
}

func (n *node) SetParentMultipleKey(multipleKey bool) {
	n.parentMultipleKey = multipleKey
}

// Tree is a tree data structure into which a PEG can be parsed.
type Tree struct {
	Rules      map[string]*node
	rulesCount map[string]uint
	node
	inline, _switch, Ast bool
	Strict               bool
	werr                 error

	Generator       string
	RuleNames       []*node
	Comments        string
	PackageName     string
	Imports         []string
	EndSymbol       rune
	PegRuleType     string
	StructName      string
	StructVariables string
	RulesCount      int
	HasActions      bool
	Actions         []*node
	HasPush         bool
	HasCommit       bool
	HasDot          bool
	HasCharacter    bool
	HasString       bool
	HasRange        bool
}

func New(inline, _switch, noast bool) *Tree {
	return &Tree{
		Rules:      make(map[string]*node),
		rulesCount: make(map[string]uint),
		inline:     inline,
		_switch:    _switch,
		Ast:        !noast,
	}
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
	t.PushFront(&node{Type: TypeCharacter, string: string(rune(hexa))})
}

func (t *Tree) AddOctalCharacter(text string) {
	octal, _ := strconv.ParseInt(text, 8, 8)
	t.PushFront(&node{Type: TypeCharacter, string: string(rune(octal))})
}
func (t *Tree) AddPredicate(text string)   { t.PushFront(&node{Type: TypePredicate, string: text}) }
func (t *Tree) AddStateChange(text string) { t.PushFront(&node{Type: TypeStateChange, string: text}) }
func (t *Tree) AddNil()                    { t.PushFront(&node{Type: TypeNil, string: "<nil>"}) }
func (t *Tree) AddAction(text string)      { t.PushFront(&node{Type: TypeAction, string: text}) }
func (t *Tree) AddPackage(text string)     { t.PushBack(&node{Type: TypePackage, string: text}) }
func (t *Tree) AddSpace(text string)       { t.PushBack(&node{Type: TypeSpace, string: text}) }
func (t *Tree) AddComment(text string)     { t.PushBack(&node{Type: TypeComment, string: text}) }
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

func (t *Tree) countRules(n *node, ruleReached []bool) {
	switch n.GetType() {
	case TypeRule:
		name, id := n.String(), n.GetID()
		if count, ok := t.rulesCount[name]; ok {
			t.rulesCount[name] = count + 1
		} else {
			t.rulesCount[name] = 1
		}
		if ruleReached[id] {
			return
		}
		ruleReached[id] = true
		t.countRules(n.Front(), ruleReached)
	case TypeName:
		t.countRules(t.Rules[n.String()], ruleReached)
	case TypeImplicitPush, TypePush:
		t.countRules(n.Front(), ruleReached)
	case TypeAlternate, TypeUnorderedAlternate, TypeSequence,
		TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
		for element := range n.Iterator() {
			t.countRules(element, ruleReached)
		}
	}
}

func (t *Tree) checkRecursion(n *node, ruleReached []bool) bool {
	switch n.GetType() {
	case TypeRule:
		id := n.GetID()
		if ruleReached[id] {
			t.warn(fmt.Errorf("possible infinite left recursion in rule '%v'", n))
			return false
		}
		ruleReached[id] = true
		consumes := t.checkRecursion(n.Front(), ruleReached)
		ruleReached[id] = false
		return consumes
	case TypeAlternate:
		for element := range n.Iterator() {
			if !t.checkRecursion(element, ruleReached) {
				return false
			}
		}
		return true
	case TypeSequence:
		return slices.ContainsFunc(slices.Collect(n.Iterator()), func(n *node) bool {
			return t.checkRecursion(n, ruleReached)
		})
	case TypeName:
		return t.checkRecursion(t.Rules[n.String()], ruleReached)
	case TypePlus, TypePush, TypeImplicitPush:
		return t.checkRecursion(n.Front(), ruleReached)
	case TypeCharacter, TypeString:
		return len(n.String()) > 0
	case TypeDot, TypeRange:
		return true
	}
	return false
}

func (t *Tree) warn(e error) {
	if t.werr == nil {
		t.werr = fmt.Errorf("warning: %w", e)
		return
	}
	t.werr = fmt.Errorf("%w\nwarning: %w", t.werr, e)
}

func (t *Tree) link(countsForRule *[TypeLast]uint, n *node, counts *[TypeLast]uint, countsByRule *[]*[TypeLast]uint, rule *node) {
	nodeType := n.GetType()
	id := counts[nodeType]
	counts[nodeType]++
	countsForRule[nodeType]++
	switch nodeType {
	case TypeAction:
		n.SetID(int(id))
		cp := n.Copy()
		name := fmt.Sprintf("Action%v", id)
		t.Actions = append(t.Actions, cp)
		n.Init()
		n.SetType(TypeName)
		n.SetString(name)
		n.SetID(t.RulesCount)

		emptyRule := &node{Type: TypeRule, string: name, id: t.RulesCount}
		implicitPush := &node{Type: TypeImplicitPush}
		emptyRule.PushBack(implicitPush)
		implicitPush.PushBack(cp)
		implicitPush.PushBack(emptyRule.Copy())
		t.PushBack(emptyRule)
		t.RulesCount++

		t.Rules[name] = emptyRule
		t.RuleNames = append(t.RuleNames, emptyRule)
		*countsByRule = append(*countsByRule, &[TypeLast]uint{})
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
			*countsByRule = append(*countsByRule, &[TypeLast]uint{})
		}
	case TypePush:
		cp := rule.Copy()
		name := "PegText"
		cp.SetString(name)
		if _, ok := t.Rules[name]; !ok {
			emptyRule := &node{Type: TypeRule, string: name, id: t.RulesCount}
			emptyRule.PushBack(&node{Type: TypeNil, string: "<nil>"})
			t.PushBack(emptyRule)
			t.RulesCount++

			t.Rules[name] = emptyRule
			t.RuleNames = append(t.RuleNames, emptyRule)
			*countsByRule = append(*countsByRule, &[TypeLast]uint{})
		}
		n.PushBack(cp)
		fallthrough
	case TypeImplicitPush:
		t.link(countsForRule, n.Front(), counts, countsByRule, rule)
	case TypeRule, TypeAlternate, TypeUnorderedAlternate, TypeSequence,
		TypePeekFor, TypePeekNot, TypeQuery, TypeStar, TypePlus:
		for node := range n.Iterator() {
			t.link(countsForRule, node, counts, countsByRule, rule)
		}
	}
}

func (t *Tree) Compile(file string, args []string, out io.Writer) (err error) {
	t.AddImport("fmt")
	if t.Ast {
		t.AddImport("io")
		t.AddImport("os")
		t.AddImport("bytes")
	}
	t.AddImport("slices")
	t.AddImport("strconv")
	t.EndSymbol = 0x110000
	t.RulesCount++

	t.Generator = strings.Join(slices.Concat([]string{"peg"}, args[1:]), " ")

	counts := [TypeLast]uint{}
	countsByRule := make([]*[TypeLast]uint, t.RulesCount)

	/* first pass */
	for n := range t.Iterator() {
		switch n.GetType() {
		case TypePackage:
			t.PackageName = n.String()
		case TypeImport:
			t.Imports = append(t.Imports, n.String())
		case TypePeg:
			t.StructName = n.String()
			t.StructVariables = n.Front().String()
		case TypeRule:
			if _, ok := t.Rules[n.String()]; !ok {
				expression := n.Front()
				cp := expression.Copy()
				expression.Init()
				expression.SetType(TypeImplicitPush)
				expression.PushBack(cp)
				expression.PushBack(n.Copy())

				t.Rules[n.String()] = n
				t.RuleNames = append(t.RuleNames, n)
			}
		}
	}
	/* sort imports to satisfy gofmt */
	slices.Sort(t.Imports)

	/* second pass */
	for _, n := range slices.Collect(t.Iterator()) {
		if n.GetType() == TypeRule {
			countsForRule := [TypeLast]uint{}
			countsByRule[n.GetID()] = &countsForRule
			t.link(&countsForRule, n, &counts, &countsByRule, n)
		}
	}

	usage := [TypeLast]uint{}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		ruleReached := make([]bool, t.RulesCount)
		for n := range t.Iterator() {
			if n.GetType() == TypeRule {
				t.countRules(n, ruleReached)
				break
			}
		}
		for id, reached := range ruleReached {
			if reached {
				for i, count := range countsByRule[id] {
					usage[i] += count
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		ruleReached := make([]bool, t.RulesCount)
		for n := range t.Iterator() {
			if n.GetType() == TypeRule {
				t.checkRecursion(n, ruleReached)
			}
		}
	}()

	wg.Wait()

	if t._switch {
		var optimizeAlternates func(node *node) (consumes bool, s *set.Set)
		cache := make([]struct {
			reached  bool
			consumes bool
			s        *set.Set
		}, t.RulesCount)

		firstPass := true
		for i := range cache {
			cache[i].s = set.NewSet()
		}
		optimizeAlternates = func(n *node) (consumes bool, s *set.Set) {
			s = set.NewSet()
			/*n.debug()*/
			switch n.GetType() {
			case TypeRule:
				cache := &cache[n.GetID()]
				if cache.reached {
					consumes = cache.consumes
					s = cache.s
					return
				}

				cache.reached = true
				consumes, s = optimizeAlternates(n.Front())
				cache.consumes = consumes
				cache.s = s
			case TypeName:
				consumes, s = optimizeAlternates(t.Rules[n.String()])
			case TypeDot:
				consumes = true
				/* TypeDot set doesn't include the EndSymbol */
				s.Add(t.EndSymbol)
				s = s.Complement(t.EndSymbol - 1)
			case TypeString, TypeCharacter:
				consumes = true
				s.Add([]rune(n.String())[0])
			case TypeRange:
				consumes = true
				element := n.Front()
				lower := []rune(element.String())[0]
				element = element.Next()
				upper := []rune(element.String())[0]
				s.AddRange(lower, upper)
			case TypeAlternate:
				consumes = true
				properties := make([]struct {
					intersects bool
					s          *set.Set
				}, n.Len())

				for i := range properties {
					properties[i].s = set.NewSet()
				}
				for i, element := range n.Iterator2() {
					consumes, properties[i].s = optimizeAlternates(element)
					s = s.Union(properties[i].s)
				}

				if firstPass {
					break
				}

				intersections := 2
				for ai, a := range properties[:len(properties)-1] {
					for _, b := range properties[ai+1:] {
						if a.s.Intersects(b.s) {
							intersections++
							properties[ai].intersects = true
							break
						}
					}
				}
				if intersections >= len(properties) {
					break
				}

				unordered := &node{Type: TypeUnorderedAlternate}
				ordered := &node{Type: TypeAlternate}
				maxVal := 0
				for i, element := range n.Iterator2() {
					if properties[i].intersects {
						ordered.PushBack(element.Copy())
					} else {
						class := &node{Type: TypeUnorderedAlternate}
						for d := range unicode.MaxRune {
							if properties[i].s.Has(d) {
								class.PushBack(&node{Type: TypeCharacter, string: string(d)})
							}
						}

						sequence := &node{Type: TypeSequence}
						predicate := &node{Type: TypePeekFor}
						length := properties[i].s.Len()
						if length == 0 {
							class.PushBack(&node{Type: TypeNil, string: "<nil>"})
						}
						predicate.PushBack(class)
						sequence.PushBack(predicate)
						sequence.PushBack(element.Copy())

						if element.GetType() == TypeNil {
							unordered.PushBack(sequence)
						} else if length > maxVal {
							unordered.PushBack(sequence)
							maxVal = length
						} else {
							unordered.PushFront(sequence)
						}
					}
				}
				n.Init()
				if ordered.Front() == nil {
					n.SetType(TypeUnorderedAlternate)
					for element := range unordered.Iterator() {
						n.PushBack(element.Copy())
					}
				} else {
					for element := range ordered.Iterator() {
						n.PushBack(element.Copy())
					}
					n.PushBack(unordered)
				}
			case TypeSequence:
				classes := make([]struct {
					s *set.Set
				}, n.Len())
				for i := range classes {
					classes[i].s = set.NewSet()
				}
				elements := slices.Collect(n.Iterator())
				for c, element := range elements {
					consumes, classes[c].s = optimizeAlternates(element)
					if consumes {
						elements, classes = elements[c+1:], classes[:c+1]
						break
					}
				}

				for c := range slices.Backward(classes) {
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
				// empty
			}
			return
		}
		for element := range t.Iterator() {
			if element.GetType() == TypeRule {
				optimizeAlternates(element)
				break
			}
		}

		for i := range cache {
			cache[i].reached = false
		}
		firstPass = false
		for element := range t.Iterator() {
			if element.GetType() == TypeRule {
				optimizeAlternates(element)
				break
			}
		}
	}

	var buffer bytes.Buffer

	_print := func(format string, a ...any) { _, _ = fmt.Fprintf(&buffer, format, a...) }
	printSave := func(n uint) { _print("\n   position%d, tokenIndex%d := position, tokenIndex", n, n) }
	printRestore := func(n uint) { _print("\n   position, tokenIndex = position%d, tokenIndex%d", n, n) }
	printMemoSave := func(rule int, n uint64, ret bool) {
		_print("\n   memoize(%d, position%d, tokenIndex%d, %t)", rule, n, n, ret)
	}
	printMemoCheck := func(rule int) {
		_print("\n   if memoized, ok := memoization[memoKey[U]{%d, position}]; ok {", rule)
		_print("\n       return memoizedResult(memoized)")
		_print("\n   }")
	}

	t.HasActions = usage[TypeAction] > 0
	t.HasPush = usage[TypePush] > 0
	t.HasCommit = usage[TypeCommit] > 0
	t.HasDot = usage[TypeDot] > 0
	t.HasCharacter = usage[TypeCharacter] > 0
	t.HasString = usage[TypeString] > 0
	t.HasRange = usage[TypeRange] > 0

	var printRule func(n *node)
	var compile func(expression *node, ko uint) (labelLast bool)
	var label uint
	labels := make(map[uint]bool)
	printBegin := func() { _print("\n   {") }
	printEnd := func() { _print("\n   }") }
	printLabel := func(n uint) bool {
		_print("\n")
		if labels[n] {
			_print("   l%d:\t", n)
			return true
		}
		return false
	}
	printJump := func(n uint) {
		_print("\n   goto l%d", n)
		labels[n] = true
	}
	printRule = func(n *node) {
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
			elements := slices.Collect(n.Iterator())
			printRule(elements[0])
			for _, element := range elements[1:] {
				_print(" / ")
				printRule(element)
			}
			_print(")")
		case TypeUnorderedAlternate:
			_print("(")
			elements := slices.Collect(n.Iterator())
			printRule(elements[0])
			for _, element := range elements[1:] {
				_print(" | ")
				printRule(element)
			}
			_print(")")
		case TypeSequence:
			_print("(")
			elements := slices.Collect(n.Iterator())
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
		case TypeComment:
		case TypeNil:
		default:
			t.warn(fmt.Errorf("illegal node type: %v", n.GetType()))
		}
	}
	dryCompile := true

	compile = func(n *node, ko uint) (labelLast bool) {
		switch n.GetType() {
		case TypeRule:
			t.warn(fmt.Errorf("internal error #1 (%v)", n))
		case TypeDot:
			if n.ParentDetect() {
				break
			}
			_print("\n   if !matchDot() {")
			/*print("\n   if buffer[position] == endSymbol {")*/
			printJump(ko)
			/*print("}\nposition++")*/
			_print("}")
		case TypeName:
			name := n.String()
			rule := t.Rules[name]
			if t.inline && t.rulesCount[name] == 1 {
				element := rule.Front()
				element.SetParentDetect(n.ParentDetect())
				element.SetParentMultipleKey(n.ParentMultipleKey())
				compile(element, ko)
				return
			}
			_print("\n   if !_rules[rule%v]() {", name /*rule.GetID()*/)
			printJump(ko)
			_print("}")
		case TypeRange:
			if n.ParentDetect() {
				_print("\nposition++")
				break
			}
			element := n.Front()
			lower := element
			element = element.Next()
			upper := element
			/*print("\n   if !matchRange('%v', '%v') {", escape(lower.String()), escape(upper.String()))*/
			_print("\n   if c := buffer[position]; c < '%v' || c > '%v' {", escape(lower.String()), escape(upper.String()))
			printJump(ko)
			_print("}\nposition++")
		case TypeCharacter:
			if n.ParentDetect() && !n.ParentMultipleKey() {
				_print("\nposition++")
				break
			}
			/*print("\n   if !matchChar('%v') {", escape(n.String()))*/
			_print("\n   if buffer[position] != '%v' {", escape(n.String()))
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
			element.SetParentDetect(n.ParentDetect())
			element.SetParentMultipleKey(n.ParentMultipleKey())
			label++
			nodeType, rule := element.GetType(), element.Next()
			printBegin()
			if nodeType == TypeAction {
				if t.Ast {
					_print("\nadd(rule%v, position)", rule)
				} else {
					// There is no AST support, so inline the rule code
					_print("\n%v", element)
				}
			} else {
				_print("\nposition%d := position", ok)
				compile(element, ko)
				if n.GetType() == TypePush && !t.Ast {
					// This is TypePush and there is no AST support,
					// so inline capture to text right here
					_print("\nbegin := position%d", ok)
					_print("\nend := position")
					_print("\ntext = string(buffer[begin:end])")
				} else {
					_print("\nadd(rule%v, position%d)", rule, ok)
				}
			}
			printEnd()
		case TypeAlternate:
			ok := label
			label++
			printBegin()
			elements := slices.Collect(n.Iterator())
			elements[0].SetParentDetect(n.ParentDetect())
			elements[0].SetParentMultipleKey(n.ParentMultipleKey())
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
			labelLast = printLabel(ok)
		case TypeUnorderedAlternate:
			done := ko
			ok := label
			label++
			printBegin()
			_print("\n   switch buffer[position] {")
			elements := slices.Collect(n.Iterator())
			elements, last := elements[:len(elements)-1], elements[len(elements)-1].Front().Next()
			for _, element := range elements {
				sequence := element.Front()
				class := sequence.Front()
				sequence = sequence.Next()
				_print("\n   case")
				comma := false
				for character := range class.Iterator() {
					if comma {
						_print(",")
					} else {
						comma = true
					}
					_print(" '%s'", escape(character.String()))
				}
				_print(":")
				if !dryCompile {
					sequence.SetParentDetect(true)
					if class.Len() > 1 {
						sequence.SetParentMultipleKey(true)
					}
				}
				if compile(sequence, done) {
					_print("\nbreak")
				}
			}
			_print("\n   default:")
			if compile(last, done) {
				_print("\nbreak")
			}
			_print("\n   }")
			printEnd()
			labelLast = printLabel(ok)
		case TypeSequence:
			elements := slices.Collect(n.Iterator())
			elements[0].SetParentDetect(n.ParentDetect())
			elements[0].SetParentMultipleKey(n.ParentMultipleKey())
			for _, element := range elements {
				labelLast = compile(element, ko)
			}
		case TypePeekFor:
			ok := label
			label++
			printBegin()
			printSave(ok)
			element := n.Front()
			element.SetParentDetect(n.ParentDetect())
			element.SetParentMultipleKey(n.ParentMultipleKey())
			compile(element, ko)
			printRestore(ok)
			printEnd()
		case TypePeekNot:
			ok := label
			label++
			printBegin()
			printSave(ok)
			element := n.Front()
			element.SetParentDetect(n.ParentDetect())
			element.SetParentMultipleKey(n.ParentMultipleKey())
			compile(element, ok)
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
			element := n.Front()
			element.SetParentDetect(n.ParentDetect())
			element.SetParentMultipleKey(n.ParentMultipleKey())
			compile(element, qko)
			printJump(qok)
			printLabel(qko)
			printRestore(qko)
			printEnd()
			labelLast = printLabel(qok)
		case TypeStar:
			again := label
			label++
			out := label
			label++
			printLabel(again)
			printBegin()
			printSave(out)
			element := n.Front()
			element.SetParentDetect(n.ParentDetect())
			element.SetParentMultipleKey(n.ParentMultipleKey())
			compile(element, out)
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
		case TypeComment:
		case TypeNil:
		default:
			t.warn(fmt.Errorf("illegal node type: %v", n.GetType()))
		}
		return labelLast
	}

	/* let's figure out which jump labels are going to be used with this dry compile */
	printTemp, _print := _print, func(_ string, _ ...any) {}
	for element := range t.Iterator() {
		if element.GetType() == TypeComment {
			t.Comments += "//" + element.String() + "\n"
		} else if element.GetType() == TypeSpace {
			t.Comments += element.String()
		}
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
	_print = printTemp
	label = 0
	dryCompile = false

	/* now for the real compile pass */
	t.PegRuleType = "uint8"
	if length := int64(t.Len()); length > math.MaxUint32 {
		t.PegRuleType = "uint64"
	} else if length > math.MaxUint16 {
		t.PegRuleType = "uint32"
	} else if length > math.MaxUint8 {
		t.PegRuleType = "uint16"
	}

	err = template.Must(template.New("peg").Parse(pegHeaderTemplate)).Execute(&buffer, t)
	if err != nil {
		return err
	}

	for element := range t.Iterator() {
		if element.GetType() != TypeRule {
			continue
		}
		expression := element.Front()
		if implicit := expression.Front(); expression.GetType() == TypeNil || implicit.GetType() == TypeNil {
			if element.String() != "PegText" {
				t.warn(fmt.Errorf("rule '%v' used but not defined", element))
			}
			_print("\n  nil,")
			continue
		}
		ko := label
		label++
		_print("\n  /* %v ", element.GetID())
		printRule(element)
		_print(" */")
		if count, ok := t.rulesCount[element.String()]; !ok {
			t.warn(fmt.Errorf("rule '%v' defined but not used", element))
			_print("\n  nil,")
			continue
		} else if t.inline && count == 1 && ko != 0 {
			_print("\n  nil,")
			continue
		}
		_print("\n  func() bool {")
		if t.Ast {
			printMemoCheck(element.GetID())
		}
		if t.Ast || labels[ko] {
			printSave(ko)
		}
		compile(expression, ko)
		// print("\n  fmt.Printf(\"%v\\n\")", element.String())
		if t.Ast {
			printMemoSave(element.GetID(), uint64(ko), true)
		}
		_print("\n   return true")
		if labels[ko] {
			printLabel(ko)
			if t.Ast {
				printMemoSave(element.GetID(), uint64(ko), false)
			}
			printRestore(ko)
			_print("\n   return false")
		}
		_print("\n  },")
	}
	_print("\n }\n p.rules = _rules")
	_print("\n return nil")
	_print("\n}\n")

	if t.Strict && t.werr != nil {
		// Treat warnings as errors.
		err = t.werr
	}
	if !t.Strict && t.werr != nil {
		// Display warnings.
		_, _ = fmt.Fprintln(os.Stderr, t.werr)
	}
	if err != nil {
		return
	}
	fileSet := token.NewFileSet()
	code, err := parser.ParseFile(fileSet, file, &buffer, parser.ParseComments)
	if err != nil {
		_, _ = buffer.WriteTo(out)
		return
	}
	formatter := printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}
	err = formatter.Fprint(out, fileSet, code)
	if err != nil {
		_, _ = buffer.WriteTo(out)
		return
	}

	return nil
}
