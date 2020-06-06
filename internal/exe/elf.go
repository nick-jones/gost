package exe

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"sort"

	"github.com/nick-jones/gost/internal/address"
)

// ELF covers Executable and Linkable Format (ELF) type binaries
type ELF struct {
	f       *elf.File
	symbols []elf.Symbol
}

// newELF initialises the ELF type
func newELF(f *os.File) (*ELF, error) {
	ef, err := elf.NewFile(f)
	if err != nil {
		return nil, err
	}

	// as with Mach-O, symbols are not returned in address order
	var syms []elf.Symbol
	orig, err := ef.Symbols()
	if err == nil {
		syms = make([]elf.Symbol, len(orig))
		copy(syms, orig)
		sort.Slice(syms, func(i, j int) bool {
			return syms[i].Value < syms[j].Value
		})
	}

	return &ELF{
		f:       ef,
		symbols: syms,
	}, nil
}

// ByteOrder returns the byte order (little or big endian)
func (e *ELF) ByteOrder() binary.ByteOrder {
	return e.f.ByteOrder
}

// TextSection locates and returns .text
func (e *ELF) TextSection() (Section, error) {
	return e.section(".text")
}

// TextSection locates and returns .rodata
func (e *ELF) RODataSection() (Section, error) {
	return e.section(".rodata")
}

// TextSection locates and returns .gopclntab
func (e *ELF) PCLNTabSection() (Section, error) {
	return e.section(".gopclntab")
}

// section searches for a section by name
func (e *ELF) section(name string) (Section, error) {
	sect := e.f.Section(name)
	if sect == nil {
		return Section{}, fmt.Errorf("failed to locate section %s", name)
	}
	return Section{
		Name: sect.Name,
		AddrRange: address.Range{
			Start: sect.Addr,
			End:   sect.Addr + sect.Size,
		},
		ReaderAt: sect.ReaderAt,
	}, nil
}

// SectionContainingRange locates a section containing a given range. It returns an error if one cannot be found, or
// the range spans over the boundary of a section.
func (e *ELF) SectionContainingRange(addrRange address.Range) (Section, error) {
	for _, s := range e.f.Sections {
		if addrRange.Start >= s.Addr && addrRange.Start <= s.Addr+s.Size {
			if addrRange.End < s.Addr && addrRange.End > s.Addr+s.Size {
				return Section{}, fmt.Errorf("range overflows from section %s (%s)", s.Name, addrRange)
			}
			return Section{
				Name: s.Name,
				AddrRange: address.Range{
					Start: s.Addr,
					End:   s.Addr + s.Size,
				},
				ReaderAt: s.ReaderAt,
			}, nil
		}
	}
	return Section{}, fmt.Errorf("failed to locate section for address range (%s)", addrRange)
}

// Symbol locates a symbol by name. For Go binaries, not all symbols are returned with associated size. If size is
// returned, we use it. Otherwise we apply similar guess work seen with Mach-O.
func (e *ELF) Symbol(name string) (Symbol, error) {
	var (
		matched Symbol
		found   bool
	)
	for _, sym := range e.symbols {
		if found {
			matched.Range.End = sym.Value - 1
			return matched, nil
		}
		if sym.Name == name {
			// If the symbol carries size, return straight away
			if sym.Size > 0 {
				return Symbol{
					Name: sym.Name,
					Range: address.Range{
						Start: sym.Value,
						End:   sym.Size,
					},
				}, nil
			}
			// Unfortunately not all symbols have size (including go.string.*) - so we have to use the same trickery
			// that we use for Mach-O, base the end address on the next symbol.
			matched = Symbol{
				Name:  sym.Name,
				Range: address.Range{Start: sym.Value},
			}
			found = true
		}
	}
	return Symbol{}, ErrSymbolNotFound
}

// SymbolForAddress locates a symbol that is closest to the supplied address.
func (e *ELF) SymbolForAddress(addr uint64) (Symbol, error) {
	var previous elf.Symbol
	for _, sym := range e.symbols {
		// If the symbol size is > 0 and this condition holds true, we're good
		if sym.Value < addr && sym.Value + sym.Size > addr {
			return Symbol{
				Name:  sym.Name,
				Range: address.Range{
					Start: sym.Value,
					End:   sym.Value + sym.Size,
				},
			}, nil
		}
		// Otherwise the previous was likely 0 size
		if sym.Value > addr {
			return Symbol{
				Name: previous.Name,
				Range: address.Range{
					Start: previous.Value,
					End:   sym.Value - 1,
				},
			}, nil
		}
		previous = sym
	}
	return Symbol{}, ErrSymbolNotFound
}

// Close closes the underlying file
func (e *ELF) Close() error {
	return e.f.Close()
}
