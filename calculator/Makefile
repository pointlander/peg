# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=calculator

GOFILES=\
	calculator.go\
	calculator.peg.go\
        main.go\

PREREQ+=calculator.peg.go
CLEANFILES+=calculator.peg.go

include $(GOROOT)/src/Make.cmd

calculator.peg.go: calculator.peg
	../peg calculator.peg
