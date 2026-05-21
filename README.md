# PEG, an Implementation of a Packrat Parsing Expression Grammar in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/pointlander/peg.svg)](https://pkg.go.dev/github.com/pointlander/peg)
[![Go Report Card](https://goreportcard.com/badge/github.com/pointlander/peg)](https://goreportcard.com/report/github.com/pointlander/peg)

A [Parsing Expression Grammar](https://en.wikipedia.org/wiki/Parsing_expression_grammar) ( hence `peg`) is a way to create grammars similar in principle to [regular expressions](https://en.wikipedia.org/wiki/Regular_expression) but which allow better code integration. Specifically, `peg` is an implementation of the [Packrat](https://en.wikipedia.org/wiki/Parsing_expression_grammar#Implementing_parsers_from_parsing_expression_grammars) parser generator originally implemented as [peg/leg](https://www.piumarta.com/software/peg/) by [Ian Piumarta](https://www.piumarta.com/cv/) in C. A Packrat parser is a "descent recursive parser" capable of backtracking and negative look-ahead assertions which are problematic for regular expression engines.

## PEG file syntax

See [peg-file-syntax.md](docs/peg-file-syntax.md) for creating your own `*.peg` file.

## Usage

1. Add `peg` as a [tool dependency](https://go.dev/doc/modules/managing-dependencies#tools)
   ```
   go get -tool github.com/pointlander/peg@main
   ```

2. Create your `mygrammar.peg`

3. Add a `//go:generate` line in a Go source of your package
   ```go
   //go:generate go tool peg mygrammar.peg
   ```
   ---
   Widespread community design pattern: Add lines like this in `doc.go`.

4. Generate `mygrammar.peg.go` from `mygrammar.peg`
   ```
   go generate
   ```

### Example

This creates the file `peg.peg.go`
```
go tool peg -inline -switch peg.peg
```

### Help
```
go tool peg -h
```

## Development

### Requirements

* [Golang](https://go.dev/doc/install), see [go.mod](go.mod) for version
* [golangci-lint latest version](https://github.com/golangci/golangci-lint#install) (v2 or later)
* [Bash 3.2.x or higher](https://www.gnu.org/software/bash)


### Generate

Bootstrap and generate grammar *.peg.go. This commands should initially be executed once before other commands. 
```
go generate
```


### Build

```
go build
```

([`go generate`](#generate) required once beforehand)


#### Set version

Use the version from the tag if the current commit has a tag. If not use the current commit hash.
```
go build -ldflags "-X main.Version=$(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD)"
```

Additionally, since [Go 1.18](https://go.dev/doc/go1.18) the go command embeds version control information. Read the information:
```
go version -m peg
```


### Test

```
go test -short ./...
```

([`go generate`](#generate) required once beforehand)


### Lint

```
golangci-lint run
```

([`go generate`](#generate) required once beforehand)


### Format

```
golangci-lint fmt
```


### Benchmark
```
go test -benchmem -bench .
```

([`go generate`](#generate) required once beforehand)


## Author

Andrew Snodgrass
