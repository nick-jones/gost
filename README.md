# Gost

WIP experiments in extracting string constants from Go compiled binaries.

At the moment it has a number of limitations:
- It only works with x86-64 ELF and Mach-O executables
- Since this is heuristic driven, not all cases will be captured. In particular string comparisons are not well captured currently.
- This relies on certain characteristics of how Go compiles binaries; these are liable to change between versions
- Functions can get inlined, making some reference information a little imperfect

## About

_Why not use bin-utils/strings?_ Go handles strings a little differently to languages such as C in that 
it doesn't use NULL terminated strings. Instead, it carries the length of a given string, along with a pointer. Due to
that, `strings` doesn't perform well with Go binaries; it'll print long incoherent and joined up strings that aren't
easily parsed.  

Gost uses a number of heuristics to extract strings from binaries; these are imperfect by nature, so be warned, results
may vary! For more information about how Go handles strings, see [docs/strings.md](docs/strings.md)

## Installing

```
GO111MODULE=on go get github.com/nick-jones/gost
```

## Running

Simply supply a path to a binary as an argument. Note that if the binary must have been compiled with symbols (which is
the default, but can be prevented).

As a quick measure we can run `gost` against itself and obtain strings referenced in `main.go`:

```
$ go build
$ ./gost gost | rg 'main\.go'
123b852: "gost" → /Users/nicholas/Dev/gost/main.go:22 
123c1f6: "nulls" → /Users/nicholas/Dev/gost/main.go:34 
123ec66: "template" → /usr/local/Cellar/go/1.15.4/libexec/src/text/template/parse/lex.go:84 /Users/nicholas/Dev/gost/main.go:25 
12409a8: "string-table" → /Users/nicholas/Dev/gost/main.go:30 /Users/nicholas/Dev/gost/main.go:86 
12448af: "failed to open file: %w" → /Users/nicholas/Dev/gost/main.go:57 
1245257: "failed to parse flags: %w" → /Users/nicholas/Dev/gost/main.go:63 
124693a: "failed to execute template: %w" → /Users/nicholas/Dev/gost/main.go:75 
1246dba: "failed to parse format flag: %w" → /Users/nicholas/Dev/gost/main.go:52 
12473fb: "invalid str-table flag value: %s" → /Users/nicholas/Dev/gost/main.go:93 
124796a: "failed to search instructions: %w" → /Users/nicholas/Dev/gost/main.go:69 
124bdab: "string candidates containing null characters will be included" → /Users/nicholas/Dev/gost/main.go:35 
124c066: "template string for printing the results (format is text/template)" → /Users/nicholas/Dev/gost/main.go:26 
124c2a9: "if symbols are missing, use values \"guess\" or \"ignore\" to enable more fuzzy matching" → /Users/nicholas/Dev/gost/main.go:31 
124c641: "{{printf \"%x: %q\" .Addr .Value}} → {{range $i, $e := .Refs}}\n{{- if le $i 5}}{{ printf \"%s:%d \" .File .Line }}{{end}}\n{{- end}}\n{{- if gt (len .Refs) 5}}... (truncated, {{len .Refs}} total){{- end -}}\n" → /Users/nicholas/Dev/gost/main.go:27
```

## Fuzzing

Fuzzing of this tool is catered for in a separate repository - [gost-fuzz](https://github.com/nick-jones/gost-fuzz)