# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include ../Make.common

bootstrap: main.$(O)
	$(LD) -o bootstrap main.$(O)

main.$(O): peg.$(O)

peg.$(O):
	$(GC) ../peg.go

.PHONY: clean
clean:
	rm -f *.6 *.8 bootstrap
