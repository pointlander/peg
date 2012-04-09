# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

peg: bootstrap.peg.go peg.go main.go
	go build

bootstrap.peg.go: bootstrap/main.go peg.go
	cd bootstrap; go build
	bootstrap/bootstrap

clean:
	rm -f bootstrap/bootstrap peg peg.peg.go
