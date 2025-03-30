// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate ../../peg -switch -inline calculator.peg

package calculator

import (
	"math/big"
	"testing"
)

func TestCalculator(t *testing.T) {
	expression := "( 1 - -3 ) / 3 + 2 * ( 3 + -4 ) + 3 % 2^2"
	calc := &Calculator{Buffer: expression}
	err := calc.Init()
	if err != nil {
		t.Fatal(err)
	}
	calc.Expression.Init(expression)
	if err := calc.Parse(); err != nil {
		t.Fatal(err)
	}
	calc.Execute()
	if calc.Evaluate().Cmp(big.NewInt(2)) != 0 {
		t.Fatal("got incorrect result")
	}
}
