# About

Peg, Parsing Expression Grammar, is an implementation of a Packrat parser
generator. A Packrat parser is a descent recursive parser capable of
backtracking. The generated parser searches for the correct parsing of the
input.

For more information see:
* http://en.wikipedia.org/wiki/Parsing_expression_grammar
* http://pdos.csail.mit.edu/~baford/packrat/

This Go implementation is based on:
* http://piumarta.com/software/peg/


# Files

* bootstrap/main.go: bootstrap syntax tree of peg
* peg.go: syntax tree and code generator
* main.go: bootstrap main
* peg.peg: peg in its own language


# Testing

There should be no differences between the bootstrap and self compiled:

```
./peg -inline -switch peg.peg
diff bootstrap.peg.go peg.peg.go
```


# Usage

```
-inline
 Tells the parser generator to inline parser rules.
-switch
 Reduces the number of rules that have to be tried for some pegs.
```


# Author

Andrew Snodgrass
