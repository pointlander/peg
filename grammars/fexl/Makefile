# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

fexl: fexl.peg.go main.go
	go build

fexl.peg.go: fexl.peg
	../../peg -switch -inline fexl.peg

clean:
	rm -f fexl fexl.peg.go
