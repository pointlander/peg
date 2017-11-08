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


# Usage

```
-inline
 Tells the parser generator to inline parser rules.
-switch
 Use at your own peril!
 Reduces the number of rules that have to be tried for some pegs.
 If statements are replaced with switch statements.
```


# Syntax

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

Next declare the rules. The first rule is the entry point into the parser:
```
<rule name> <- <rule body>
```

The first rule should probably end with `!.` to indicate no more input follows:
```
first <- . !.
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
will match the string "aaabcbcde".

For choosing between different inputs, use alternates:
```
prioritized <- 'a' 'a'* / 'bc'+ / 'de'?
```
will match "aaaa" or "bcbc" or "de" or "". The matches are attempted in order.

If the characters are case insensitive, use double quotes:
```
insensitive <- "abc"
```
will match "abc" or "Abc" or "ABc" etc...

For matching a set of characters, use a character class:
```
class <- [a-z]
```
will match "a" or "b" or all the way to "z".

For an inverse character class, start with a caret:
```
inverse <- [^a-z]
```
will match anything but "a" or "b" or all the way to "z".

If the character class is case insensitive, use double brackets:
```
insensitive <- [[A-Z]]
```

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
capture <- <'capture'> { fmt.Println(buffer[begin:end]) }
```
Will print out "capture". The captured string is stored in `buffer[begin:end]`.


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


# Author

Andrew Snodgrass
