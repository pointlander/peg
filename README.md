# PEG, an Implementation of a Packrat Parsing Expression Grammar in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/pointlander/peg.svg)](https://pkg.go.dev/github.com/pointlander/peg)
[![Go Report Card](https://goreportcard.com/badge/github.com/pointlander/peg)](https://goreportcard.com/report/github.com/pointlander/peg)

A [Parsing Expression Grammar](https://en.wikipedia.org/wiki/Parsing_expression_grammar) ( hence `peg`) is a way to create grammars similar in principle to [regular expressions](https://en.wikipedia.org/wiki/Regular_expression) but which allow better code integration. Specifically, `peg` is an implementation of the [Packrat](https://en.wikipedia.org/wiki/Parsing_expression_grammar#Implementing_parsers_from_parsing_expression_grammars) parser generator originally implemented as [peg/leg](https://www.piumarta.com/software/peg/) by [Ian Piumarta](https://www.piumarta.com/cv/) in C. A Packrat parser is a "descent recursive parser" capable of backtracking and negative look-ahead assertions which are problematic for regular expression engines . 

## See Also

* https://en.wikipedia.org/wiki/Parsing_expression_grammar
* https://bford.info/packrat/
* https://piumarta.com/software/peg/

## Installation

```
go install github.com/pointlander/peg@latest
```

## Usage

```
peg [<option>]... <file>

Usage of peg:
  -inline
      parse rule inlining
  -noast
      disable AST
  -output string
      specify name of output file
  -print
      directly dump the syntax tree
  -strict
      treat compiler warnings as errors
  -switch
      replace if-else if-else like blocks with switch blocks
  -syntax
      print out the syntax tree
  -version
      print the version and exit
```


## Sample Makefile

This sample `Makefile` will convert any file ending with `.peg` into a `.go` file with the same name. Adjust as needed.

```make
.SUFFIXES: .peg .go

.peg.go:
	peg -noast -switch -inline -strict -output $@ $<

all: grammar.go
```

Use caution when picking your names to avoid overwriting existing `.go` files. Since only one PEG grammar is allowed per Go package (currently) the use of the name `grammar.peg` is suggested as a convention:

```
grammar.peg
grammar.go
```

## PEG File Syntax

First declare the package name and any import(s) required:

```
package <package name>

import <import name>
```

Then declare the parser:

```
type <parser name> Peg {
	<parser state variables>
}
```

Next declare the rules. Note that the main rules are described below but are based on the [peg/leg rules](https://www.piumarta.com/software/peg/peg.1.html) which provide additional documentation.

The first rule is the entry point into the parser:

```
<rule name> <- <rule body>
```

The first rule should probably end with `!.` to indicate no more input follows. 

```
first <- . !.
```

This is often set to `END` to make PEG rules more readable:

```
END <- !.
```

`.` means any character matches. For zero or more character matches, use:

```
repetition <- .*
```

For one or more character matches, use:

```
oneOrMore <- .+
```

For an optional character match, use:

```
optional <- .?
```

If specific characters are to be matched, use single quotes:

```
specific <- 'a'* 'bc'+ 'de'?
```

This will match the string `"aaabcbcde"`.

For choosing between different inputs, use alternates:

```
prioritized <- 'a' 'a'* / 'bc'+ / 'de'?
```

This will match `"aaaa"` or `"bcbc"` or `"de"` or `""`. The matches are attempted in order.

If the characters are case-insensitive, use double quotes:

```
insensitive <- "abc"
```

This will match `"abc"` or `"Abc"` or `"ABc"` and so on.

For matching a set of characters, use a character class:

```
class <- [a-z]
```

This will match `"a"` or `"b"` or all the way to `"z"`.

For an inverse character class, start with a caret:

```
inverse <- [^a-z]
```

This will match anything but `"a"` or `"b"` or all the way to `"z"`.

If the character class is case-insensitive, use double brackets:

```
insensitive <- [[A-Z]]
```

(Note that this is not available in regular expression syntax.)

Use parentheses for grouping:

```
grouping <- (rule1 / rule2) rule3
```

For looking ahead a match (predicate), use:

```
lookAhead <- &rule1 rule2
```

For inverse look ahead, use:

```
inverse <- !rule1 rule2
```

Use curly braces for Go code:

```
gocode <- { fmt.Println("hello world") }
```

For string captures, use less than and greater than:

```
capture <- <'capture'> { fmt.Println(text) }
```

Will print out `"capture"`. The captured string is stored in `buffer[begin:end]`.

## Development

### Requirements

* [Golang](https://golang.org/doc/install), see [go.mod](go.mod) for version
* [golangci-lint latest version](https://github.com/golangci/golangci-lint#install)
* gofumpt
  ```
  go install mvdan.cc/gofumpt@latest
  ```

### Build

```
go build
```


### Generate
Bootstrap and generate grammar *.peg.go
```
go generate
```

* `bootstrap/main.go` - bootstrap syntax tree of peg
* `tree/peg.go` - syntax tree and code generator
* `peg.peg` - peg in its own language

### Test

Run [Generate](#Generate) before

```
go test -short ./...
```

### Lint

```
golangci-lint run
```

### Format

```
gofumpt -l -w .
```

### Benchmark
```
go test -benchmem -bench .
```

### Build with version

Use the version from the tag if the current commit has a tag. If not use the current commit hash.
```
go build -ldflags "-X main.Version=$(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)"
```

Additionally, since [Go 1.18](https://go.dev/doc/go1.18) the go command embeds version control information. Read the information:
```
go version -m peg
```

## Author

Andrew Snodgrass

## Projects That Use `peg`

Here are some projects that use `peg` to provide further examples of PEG grammars:

* https://github.com/tj/go-naturaldate - natural date/time parsing
* https://github.com/gnames/gnparser - scientific names parsing
