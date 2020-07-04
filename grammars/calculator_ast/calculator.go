// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build grammars

package main

import (
	"math/big"
)

func Eval(buffer []rune, node *node32) *big.Int {
	switch node.pegRule {
	case rulee:
		node = node.up
		for node != nil {
			switch node.pegRule {
			case rulee1:
				return Eval(buffer, node)
			}
			node = node.next
		}
	case rulee1:
		node = node.up
		var a *big.Int
		for node != nil {
			switch node.pegRule {
			case rulee2:
				a = Eval(buffer, node)
			case ruleadd:
				node = node.next
				b := Eval(buffer, node)
				a.Add(a, b)
			case ruleminus:
				node = node.next
				b := Eval(buffer, node)
				a.Sub(a, b)
			}
			node = node.next
		}
		return a
	case rulee2:
		node = node.up
		var a *big.Int
		for node != nil {
			switch node.pegRule {
			case rulee3:
				a = Eval(buffer, node)
			case rulemultiply:
				node = node.next
				b := Eval(buffer, node)
				a.Mul(a, b)
			case ruledivide:
				node = node.next
				b := Eval(buffer, node)
				a.Div(a, b)
			case rulemodulus:
				node = node.next
				b := Eval(buffer, node)
				a.Mod(a, b)
			}
			node = node.next
		}
		return a
	case rulee3:
		node = node.up
		var a *big.Int
		for node != nil {
			switch node.pegRule {
			case rulee4:
				a = Eval(buffer, node)
			case ruleexponentiation:
				node = node.next
				b := Eval(buffer, node)
				a.Exp(a, b, nil)
			}
			node = node.next
		}
		return a
	case rulee4:
		node = node.up
		minus := false
		for node != nil {
			switch node.pegRule {
			case rulevalue:
				a := Eval(buffer, node)
				if minus {
					a.Neg(a)
				}
				return a
			case ruleminus:
				minus = true
			}
			node = node.next
		}
	case rulevalue:
		node = node.up
		for node != nil {
			switch node.pegRule {
			case rulenumber:
				a := big.NewInt(0)
				a.SetString(string(buffer[node.begin:node.end]), 10)
				return a
			case rulesub:
				return Eval(buffer, node)
			}
			node = node.next
		}
	case rulesub:
		node = node.up
		for node != nil {
			switch node.pegRule {
			case rulee1:
				return Eval(buffer, node)
			}
			node = node.next
		}
	}

	return nil
}
