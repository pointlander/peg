# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

all: dirbin peg leg

dirbin:
	mkdir -p bin

peg: src/peg/bootstrap.peg.go src/peg/peg.go src/peg/main.go
	cd src/peg/; go build
	mv src/peg/peg bin/

leg: src/leg/bootstrap.leg.go src/leg/leg.go src/leg/main.go
	cd src/leg/; go build
	mv src/leg/leg bin/

bootstrap.peg.go: src/bootstrap/peg/main.go src/peg/peg.go
	cd src/bootstrap/peg; go build
	src/bootstrap/peg/peg
	mv src/bootstrap/peg/bootstrap.peg.go ./

bootstrap.leg.go: src/bootstrap/leg/main.go src/leg/leg.go
	cd src/bootstrap/leg; go build
	src/bootstrap/leg/leg
	mv src/bootstrap/leg/bootstrap.leg.go ./

clean:
	rm -f src/bootstrap/peg/peg src/bootstrap/leg/leg bin/peg bin/leg
