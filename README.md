# Gost

WIP experiments in extracting string constants from Go compiled binaries.

At the moment it has a number of limitations:
- It only works with x86-64 Mach-O executables (so ELF, etc, will not work)
- Since this is heuristic driven, not all cases will be captured
- Unreferenced strings likely will not feature in compiled binaries
- This relies on certain characteristics of how Go compiles binaries; these are liable to change between versions
- Functions can get inlined, making some references a little imperfect

## About

_Why not use bin-utils/strings?_ Good question. Go handles strings a little differently to languages such as C in that 
it doesn't use NULL terminated strings. Instead, it carries the length of a given string, along with a pointer. Due to
that, `strings` doesn't perform well with Go binaries; it'll print long incoherent strings and joined up strings that
aren't easily parsed.  

Gost uses a number of heuristics to extract strings from binaries; these are imperfect by nature, so be warned, results
may vary! For more information about how Go handles strings, see [docs/strings.md]

## Installing

```
GO111MODULE=on go get github.com/nick-jones/gost
```

## Running

Simply supply a path to a binary as an argument. Note that if the binary must have been compiled with symbols (which is
the default, but can be prevented).

As a quick measure we can run `gost` in itself and obtain strings referenced in `main.go`:

```
$ go build            
$ ./gost gost | rg 'main\.'
121b656: "gost" → main.main+149(11b1da5) 
121c96a: "format" → main.run+274(11b2012) 
121e5a1: "template" → main.main+67(11b1d53) 
122374b: "failed to open file: %w" → main.run+586(11b214a) 
1225734: "failed to execute template: %w" → main.run+1329(11b2431) 
1225bb4: "failed to parse format flag: %w" → main.run+386(11b2082) 
12266c2: "failed to search instructions: %w" → main.run+791(11b2217) 
122aae5: "template string for printing the results (format is text/template)" → main.main+85(11b1d65) 
122b06c: "{{printf \"%x: %q\" .Addr .Value}} → {{range $i, $e := .Refs}}\n{{- if le $i 5}}{{ printf \"%s+%d(%x) \" .Symbol.Name .Offset .Addr }}{{end}}\n{{- end}}\n{{- if gt (len .Refs) 5}}... (truncated, {{len .Refs}} total){{- end -}}\n" → main.main+104(11b1d78)
```
