# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

java: java_1_7.peg.go main.go
	go build

java_1_7.peg.go: java_1_7.peg
	../../peg -switch -inline java_1_7.peg

clean:
	rm -f java java_1_7.peg.go
