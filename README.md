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
123608e: "gost" → /Users/nicholas/Dev/gost/main.go:23 
1239401: "template" → /Users/nicholas/Dev/gost/main.go:26 
123eed1: "failed to open file: %w" → /Users/nicholas/Dev/gost/main.go:50 
1240f88: "failed to execute template: %w" → /Users/nicholas/Dev/gost/main.go:63 
1241408: "failed to parse format flag: %w" → /Users/nicholas/Dev/gost/main.go:45 
1241f57: "failed to search instructions: %w" → /Users/nicholas/Dev/gost/main.go:57 
1246495: "template string for printing the results (format is text/template)" → /Users/nicholas/Dev/gost/main.go:27 
1246a1c: "{{printf \"%x: %q\" .Addr .Value}} → {{range $i, $e := .Refs}}\n{{- if le $i 5}}{{ printf \"%s:%d \" .File .Line }}{{end}}\n{{- end}}\n{{- if gt (len .Refs) 5}}... (truncated, {{len .Refs}} total){{- end -}}\n" → /Users/nicholas/Dev/gost/main.go:28 
```
