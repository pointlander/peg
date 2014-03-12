# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

calculator: calculator.leg.go main.go
	go build

calculator.leg.go: calculator.leg
	../../../bin/leg calculator.leg

clean:
	rm -f calculator calculator.leg.go
