# Copyright 2010 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

peg: peg.peg.go main.go
	go build

peg.peg.go: bootstrap
	cmd/peg-bootstrap/peg-bootstrap <peg.peg >peg.peg.go
	go build
	peg -inline -switch peg.peg

# Use peg-bootstrap to compile peg from peg.peg
.PHONY: bootstrap
bootstrap: cmd/peg-bootstrap/peg-bootstrap

bootstrap/bootstrap: bootstrap/main.go tree/peg.go
	cd bootstrap; go build

# peg0 <- bootstrap.peg.go from bootstrap/main.go
# peg0 accepts inputs compiled from main.go matching bootstrap.peg
cmd/peg-bootstrap/peg0: cmd/peg-bootstrap/main.go bootstrap/bootstrap
	cd cmd/peg-bootstrap/ &&		\
		rm -f *.peg.go &&		\
		../../bootstrap/bootstrap &&	\
		go build -o peg0

# peg1 <- bootstrap.peg.go from peg0 + bootstrap.peg
# peg1 accepts inputs compiled from bootstrap.peg
cmd/peg-bootstrap/peg1: cmd/peg-bootstrap/peg0 cmd/peg-bootstrap/bootstrap.peg
	cd cmd/peg-bootstrap/ &&				\
		rm -f *.peg.go &&				\
		./peg0 <bootstrap.peg >peg1.peg.go &&		\
		go build -o peg1

# peg2 <- peg.bootstrap.peg.go from peg1 + peg.bootstrap.peg
# peg2 accepts inputs compiled from peg.bootstrap.peg matching peg.peg
cmd/peg-bootstrap/peg2: cmd/peg-bootstrap/peg1 cmd/peg-bootstrap/peg.bootstrap.peg
	cd cmd/peg-bootstrap/ &&				\
		rm -f *.peg.go &&				\
		peg1 <peg.bootstrap.peg >peg2.peg.go && \
		go build -o peg2

# peg3 <- peg.peg.go from peg2 + peg.peg
# peg3 accepts inputs compiled from peg.peg
cmd/peg-bootstrap/peg3: cmd/peg-bootstrap/peg2 peg.peg
	cd cmd/peg-bootstrap/ &&			\
		rm -f *.peg.go &&			\
		./peg2 <../../peg.peg >peg3.peg.go  &&	\
		go build -o peg3

cmd/peg-bootstrap/peg-bootstrap: cmd/peg-bootstrap/peg3
	cd cmd/peg-bootstrap/ &&			\
		rm -f *.peg.go &&			\
		./peg3 <../../peg.peg >peg-bootstrap.peg.go  &&	\
		go build -o peg-bootstrap

.PHONY:clean
clean:
	rm -f bootstrap/bootstrap peg peg.peg.go cmd/peg-bootstrap/{*peg.go,peg{[0-3],-bootstrap}}

.PHONY:test
test: peg
	go test

.PHONY:bench
bench:
	go test -benchmem -bench .
