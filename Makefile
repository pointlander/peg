# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.$(GOARCH)

.PHONY: all
all: peg

peg: bootstrap.go peg_main.go
	$(GC) peg.go bootstrap.go
	$(GC) -I ./ peg_main.go
	$(LD) -L ./ -o peg peg_main.$(O)

bootstrap.go: bootstrap
	./bootstrap

bootstrap: main.$(O)
	$(LD) -L ./ -o bootstrap main.$(O)

main.$(O): peg.$(O) main.go
	$(GC) -I ./ main.go

peg.$(O): peg.go
	$(GC) peg.go

.PHONY: clean
clean:
	rm -f *.6 *.8 bootstrap.go bootstrap peg
