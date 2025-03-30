# PEG file syntax

## Examples of PEG grammars

Here are some projects that use `peg` to provide further examples of PEG grammars:

* https://github.com/tj/go-naturaldate - natural date/time parsing
* https://github.com/gnames/gnparser - scientific names parsing

## Go package and imports

First declare the package name and any import(s) required.

```
package <package name>

import <import name>
```

## Parser

Then declare the parser:

```
type <parser name> Peg {
	<parser state variables>
}
```

## Rules

Next declare the rules. Note that the main rules are described below but are based on the [peg/leg rules](https://www.piumarta.com/software/peg/peg.1.html) which provide additional documentation.

The first rule is the entry point into the parser:

```
<rule name> <- <rule body>
```

The first rule should probably end with `!.` to indicate no more input follows:

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

## Naming convention

Use caution when picking your names to avoid overwriting existing `.go` files. Since only one PEG grammar is allowed per Go package (currently) the use of the name `grammar.peg` is suggested as a convention.

```
grammar.peg
grammar.go
```