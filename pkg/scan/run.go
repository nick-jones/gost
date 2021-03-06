package scan

import (
	"bytes"
	"debug/gosym"
	"fmt"
	"io"
	"sort"

	"github.com/nick-jones/gost/internal/address"
	"github.com/nick-jones/gost/internal/analysis"
	"github.com/nick-jones/gost/internal/exe"
	"github.com/nick-jones/gost/internal/strtable"
)

// Result encapsulates a single located string
type Result struct {
	Addr  uint64      // address where the string resides
	Value string      // raw value of the string
	Refs  []Reference // references (if known)
}

// References carries information relating to a reference to a string
type Reference struct {
	Addr         uint64 // address where the reference is made
	SymbolName   string // closest symbol
	SymbolOffset int    // offset from the closes symbol
	File         string // file that contains the reference
	Line         int    // line number of the above file
}

// Run performs analysis over data read from the supplied reader and returns potential strings
func Run(r io.ReaderAt, opts ...Option) ([]Result, error) {
	runOptions := &RunOptions{}
	for _, o := range opts {
		o(runOptions)
	}

	f, err := exe.New(r)
	if err != nil {
		return nil, fmt.Errorf("invalid file: %w", err)
	}

	var strRange *address.Range
	if !runOptions.stringTableIgnore {
		// locate address range for go.string.*
		located, err := strtable.Locate(f, runOptions.stringTableGuess)
		if err != nil {
			return nil, fmt.Errorf("failed to locate string table: %w", err)
		}
		strRange = &located
	}

	// search for strings referenced in instructions
	candidates1, err := analysis.EvaluateDirectReferences(f, strRange)
	if err != nil {
		return nil, fmt.Errorf("failed to analyse instructions: %w", err)
	}

	// search for strings referenced from statictmp
	candidates2, err := analysis.EvaluateIndirectReferences(f, strRange)
	if err != nil {
		return nil, fmt.Errorf("failed to analyse statictmp: %w", err)
	}

	// merge candidates
	candidates := dedupeCandidates(append(candidates1, candidates2...))

	return buildResults(candidates, f, runOptions)
}

func buildResults(candidates []analysis.Candidate, f *exe.File, opts *RunOptions) ([]Result, error) {
	sect, err := f.RODataSection()
	if err != nil {
		return nil, err
	}

	symtab, err := createSymtab(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create symtab: %w", err)
	}

	results := make([]Result, 0, len(candidates))
	for _, candidate := range candidates {
		if !sect.AddrRange.Contains(candidate.Addr) || !sect.AddrRange.Contains(candidate.Addr+candidate.Len) {
			continue // ignore if the address isn't in __rodata
		}
		if candidate.Len == 0 {
			continue // ignore empty strings - all observed cases are false positives (real empty strings manifest differently)
		}
		buf := make([]byte, candidate.Len)
		if _, err := sect.ReadAt(buf, int64(candidate.Addr-sect.AddrRange.Start)); err != nil {
			return nil, fmt.Errorf("failed to read data: %w", err)
		}
		if !opts.permitNulls && bytes.IndexByte(buf, 0x00) != -1 {
			continue // string contains nulls, ignore
		}
		res := Result{
			Addr:  candidate.Addr,
			Value: string(buf),
		}
		for _, addr := range candidate.RefAddrs {
			file, line, _ := symtab.PCToLine(addr)
			ref := Reference{
				Addr: addr,
				File: file,
				Line: line,
			}
			res.Refs = append(res.Refs, ref)
		}
		results = append(results, res)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Addr < results[j].Addr
	})

	return enrichWithSymbols(results, f)
}

func dedupeCandidates(candidates []analysis.Candidate) []analysis.Candidate {
	addrToRes := make(map[uint64]analysis.Candidate)
	for _, res := range candidates {
		if dupe, found := addrToRes[res.Addr]; found {
			dupe.RefAddrs = append(dupe.RefAddrs, res.RefAddrs...)
			addrToRes[res.Addr] = dupe
		} else {
			addrToRes[res.Addr] = res
		}
	}
	deduped := make([]analysis.Candidate, 0, len(addrToRes))
	for _, res := range addrToRes {
		deduped = append(deduped, res)
	}
	return deduped
}

func enrichWithSymbols(results []Result, f *exe.File) ([]Result, error) {
	// extract reference addresses
	addrs := make([]uint64, 0)
	for _, res := range results {
		for _, ref := range res.Refs {
			addrs = append(addrs, ref.Addr)
		}
	}

	// resolve symbols for all addresses
	syms, err := f.SymbolsForAddresses(addrs)
	if err != nil {
		return nil, err
	}

	// enrich references with symbols
	for i, res := range results {
		for j, ref := range res.Refs {
			if sym, found := syms[ref.Addr]; found {
				ref.SymbolName = sym.Name
				ref.SymbolOffset = int(ref.Addr) - int(sym.AddrRange.Start)
				res.Refs[j] = ref
			}
		}
		results[i] = res
	}
	return results, nil
}

func createSymtab(f *exe.File) (*gosym.Table, error) {
	txt, err := f.TextSection()
	if err != nil {
		return nil, err
	}

	pclntab, err := f.PCLNTabSection()
	if err != nil {
		return nil, err
	}

	data, err := pclntab.Data()
	if err != nil {
		return nil, err
	}

	// `gosym.LineTable` doesn't provide file information. So we have to wrap it with `gosym.Table`, which does. Not
	// need to provide symtab data - and in fact, the symtab section is zero size in Mach-O binaries, so I'm assuming
	// it is no longer populated.
	return gosym.NewTable(nil, gosym.NewLineTable(data, txt.AddrRange.Start))
}
