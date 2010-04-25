# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include $(GOROOT)/src/Make.$(GOARCH)

.PHONY: all
all: peg

peg: bootstrap.go peg_main.go
	$(GC) peg.go bootstrap.go
	$(GC) -I ./ peg_main.go
	$(LD) -L ./ -o peg peg_main.6

bootstrap.go: bootstrap
	./bootstrap

bootstrap: main.6
	$(LD) -L ./ -o bootstrap main.6

main.6: peg.6 main.go
	$(GC) -I ./ main.go

peg.6: peg.go
	$(GC) peg.go

.PHONY: clean
clean:
	rm -f *.6 bootstrap.go bootstrap peg
