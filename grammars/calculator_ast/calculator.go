// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build grammars

package main

import (
	"math/big"
)

func (c *Calculator) Eval() *big.Int {
	return c.Rulee(c.AST())
}

func (c *Calculator) Rulee(node *node32) *big.Int {
	node = node.up
	for node != nil {
		switch node.pegRule {
		case rulee1:
			return c.Rulee1(node)
		}
		node = node.next
	}
	return nil
}

func (c *Calculator) Rulee1(node *node32) *big.Int {
	node = node.up
	var a *big.Int
	for node != nil {
		switch node.pegRule {
		case rulee2:
			a = c.Rulee2(node)
		case ruleadd:
			node = node.next
			b := c.Rulee2(node)
			a.Add(a, b)
		case ruleminus:
			node = node.next
			b := c.Rulee2(node)
			a.Sub(a, b)
		}
		node = node.next
	}
	return a
}

func (c *Calculator) Rulee2(node *node32) *big.Int {
	node = node.up
	var a *big.Int
	for node != nil {
		switch node.pegRule {
		case rulee3:
			a = c.Rulee3(node)
		case rulemultiply:
			node = node.next
			b := c.Rulee3(node)
			a.Mul(a, b)
		case ruledivide:
			node = node.next
			b := c.Rulee3(node)
			a.Div(a, b)
		case rulemodulus:
			node = node.next
			b := c.Rulee3(node)
			a.Mod(a, b)
		}
		node = node.next
	}
	return a
}

func (c *Calculator) Rulee3(node *node32) *big.Int {
	node = node.up
	var a *big.Int
	for node != nil {
		switch node.pegRule {
		case rulee4:
			a = c.Rulee4(node)
		case ruleexponentiation:
			node = node.next
			b := c.Rulee4(node)
			a.Exp(a, b, nil)
		}
		node = node.next
	}
	return a
}

func (c *Calculator) Rulee4(node *node32) *big.Int {
	node = node.up
	minus := false
	for node != nil {
		switch node.pegRule {
		case rulevalue:
			a := c.Rulevalue(node)
			if minus {
				a.Neg(a)
			}
			return a
		case ruleminus:
			minus = true
		}
		node = node.next
	}
	return nil
}

func (c *Calculator) Rulevalue(node *node32) *big.Int {
	node = node.up
	for node != nil {
		switch node.pegRule {
		case rulenumber:
			a := big.NewInt(0)
			a.SetString(string(c.buffer[node.begin:node.end]), 10)
			return a
		case rulesub:
			return c.Rulesub(node)
		}
		node = node.next
	}
	return nil
}

func (c *Calculator) Rulesub(node *node32) *big.Int {
	node = node.up
	for node != nil {
		switch node.pegRule {
		case rulee1:
			return c.Rulee1(node)
		}
		node = node.next
	}
	return nil
}
