#!/bin/bash

set -Eeuo pipefail


(cd bootstrap && go build && rm -f bootstrap/bootstrap.peg.go)


cd cmd/peg-bootstrap

# Build peg0
./../../bootstrap/bootstrap
go build -tags bootstrap -o peg0
rm -f bootstrap.peg.go

# Build peg1
./peg0 < bootstrap.peg > peg1.peg.go
go build -tags bootstrap -o peg1
rm -f peg1.peg.go

# Build peg2
./peg1 < peg.bootstrap.peg > peg2.peg.go
go build -tags bootstrap -o peg2
rm -f peg2.peg.go

# Build peg3
./peg2 < ../../peg.peg > peg3.peg.go
go build -tags bootstrap -o peg3
rm -f peg3.peg.go

# Build peg-bootstrap
./peg3 < ../../peg.peg > peg-bootstrap.peg.go
go build -tags bootstrap -o peg-bootstrap
rm -f peg-bootstrap.peg.go

# Build peg
cd ../..
./cmd/peg-bootstrap/peg-bootstrap < peg.peg > peg.peg.go
go build
./peg -inline -switch peg.peg
