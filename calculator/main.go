// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
 "calculator"
 "fmt"
 "os"
)

func main() {
 if len(os.Args) < 2 {
  name := os.Args[0]
  fmt.Printf("Usage: %v \"EXPRESSION\"\n", name)
  fmt.Printf("Example: %v \"( 1 - -3 ) / 3 + 2 * ( 3 + -4 ) + 3 %% 2^2\"\n         =2\n", name)
  return
 }
 expression := os.Args[1]
 calc := &calculator.Calculator{Buffer: expression}
 calc.Init()
 calc.Expression.Init(expression)
 if !calc.Parse() {calc.PrintError(); return}
 fmt.Printf("= %v\n", calc.Evaluate())
}
