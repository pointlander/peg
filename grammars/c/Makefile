# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

c: c.peg.go main.go
	go build

c.peg.go: c.peg
	../../peg -switch -inline c.peg

clean:
	rm -f c c.peg.go

test:
	go test
