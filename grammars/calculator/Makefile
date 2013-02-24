# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

calculator: calculator.peg.go calculator.go main.go
	go build

calculator.peg.go: calculator.peg
	../../peg -switch -inline calculator.peg

clean:
	rm -f calculator calculator.peg.go
