# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include ../Make.common

calculator: main.$(O)
	$(LD) -o calculator main.$(O)

main.$(O): calculator.$(O)

calculator.$(O): calculator.go calculator.peg.go
	$(GC) -o calculator.$(O) calculator.go calculator.peg.go

calculator.peg.go: calculator.peg
	../peg calculator.peg

.PHONY: clean
clean:
	rm -f *.6 *.8 calculator calculator.peg.go
