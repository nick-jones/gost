package analysis

import (
	"debug/gosym"
	"fmt"
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
	Addr         uint64     // address where the reference is made
	Symbol       exe.Symbol // closest symbol
	SymbolOffset int        // offset from the closes symbol
	File         string     // file that contains the reference
	Line         int        // line number of the above file
}

// Run performs analysis over the file and returns potential strings
func Run(f exe.File) ([]Result, error) {
	// locate address range for go.string.*
	strRange, err := strtable.Locate(f)
	if err != nil {
		return nil, fmt.Errorf("failed to locate string table: %w", err)
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

	return buildResults(candidates, f, strRange)
}

func buildResults(candidates []analysis.Candidate, f exe.File, strRange address.Range) ([]Result, error) {
	// find section the go.string.* range resides in (should be __rodata)
	sect, err := f.SectionContainingRange(strRange)
	if err != nil {
		return nil, fmt.Errorf("failed to locate section for range: %w", err)
	}

	symtab, err := createSymtab(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create symtab: %w", err)
	}

	results := make([]Result, len(candidates))
	for i, candidate := range candidates {
		buf := make([]byte, candidate.Len)
		if _, err := sect.ReadAt(buf, int64(candidate.Addr-sect.AddrRange.Start)); err != nil {
			return nil, fmt.Errorf("failed to read data: %w", err)
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
		results[i] = res
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Addr < results[j].Addr
	})

	return results, nil
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

func createSymtab(f exe.File) (*gosym.Table, error) {
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
