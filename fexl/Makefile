# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=fexl

GOFILES=\
	fexl.peg.go\
        main.go\

PREREQ+=fexl.peg.go
CLEANFILES+=fexl.peg.go

include $(GOROOT)/src/Make.cmd

fexl.peg.go: fexl.peg
	../peg -switch -inline fexl.peg
