# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.inc

TARG=stl

GOFILES=\
	stl.go\
	stl.peg.go\

PREREQ+=stl.peg.go
CLEANFILES+=stl.peg.go

include $(GOROOT)/src/Make.pkg

stl.peg.go: stl.peg
	../peg -inline=true -switch=true stl.peg
