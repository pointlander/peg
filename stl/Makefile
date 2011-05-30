# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include ../Make.common

stl: main.$(O)
	$(LD) -o stl main.$(O)

main.$(O): stl.$(O)

stl.$(O): stl.go stl.peg.go
	$(GC) -o stl.$(O) stl.go stl.peg.go

stl.peg.go: stl.peg
#	echo "WARNING: not generating stl.peg.go"
	../peg -switch=true -inline=true stl.peg

.PHONY: clean
clean:
	rm -f *.6 *.8 stl stl.peg.go
