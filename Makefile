# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

peg: bootstrap.peg.go peg.go main.go
	go build

bootstrap.peg.go: bootstrap/peg/main.go peg.go
	cd bootstrap/peg; go build
	bootstrap/peg/peg

bootstrap.leg.go: bootstrap/leg/main.go leg.go
	cd bootstrap/leg; go build
	bootstrap/leg/leg

clean:
	rm -f bootstrap/peg/peg bootstrap/leg/leg peg peg.peg.go
