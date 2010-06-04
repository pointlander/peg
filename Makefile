# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include Make.common

.PHONY: all
all: peg

peg: main.$(O)
	$(LD) -o peg main.$(O)

main.$(O): peg.$(O)

peg.$(O): bootstrap.go
	$(GC) peg.go bootstrap.go

bootstrap.go: peg.go bootstrap/main.go
	$(MAKE) -C bootstrap/ bootstrap
	./bootstrap/bootstrap

.PHONY: clean
clean:
	$(MAKE) -C bootstrap/ clean
	rm -f *.6 *.8 bootstrap.go peg
