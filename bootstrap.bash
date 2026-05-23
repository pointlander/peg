#!/bin/bash

set -Eeuo pipefail


# Clean the file produced by ./bootstrap
rm -f bootstrap/bootstrap.peg.go


cd cmd/peg-bootstrap

# Remove artefacts from a previous incomplete build
rm -f peg[0123].peg.go peg-bootstrap.peg.go
# Remove files produced by previous versions of the build
rm -f peg[0123] peg-bootstrap bootstrap.peg.go

go run ../../bootstrap                                                   > peg0.peg.go
go run -tags bootstrap  main.go peg0.peg.go          < bootstrap.peg     > peg1.peg.go
go run -tags bootstrap  main.go peg1.peg.go          < peg.bootstrap.peg > peg2.peg.go
go run -tags bootstrap  main.go peg2.peg.go          < ../../peg.peg     > peg3.peg.go
go run -tags bootstrap  main.go peg3.peg.go          < ../../peg.peg     > peg-bootstrap.peg.go
go run -tags bootstrap  main.go peg-bootstrap.peg.go < ../../peg.peg     > ../../peg.peg.go

# Remove artefacts from the build
rm -f peg[0123].peg.go peg-bootstrap.peg.go

# Final rebuild
cd ../..
go tool peg -inline -switch peg.peg


