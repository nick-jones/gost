package exe

import (
	"debug/macho"
	"encoding/binary"
	"fmt"
	"os"
	"sort"

	"github.com/nick-jones/gost/internal/address"
)

// Macho covers Mach-O type executables
type Macho struct {
	f       *macho.File
	symbols []macho.Symbol
}

// newMacho initialises the Macho type
func newMacho(f *os.File) (*Macho, error) {
	mf, err := macho.NewFile(f)
	if err != nil {
		return nil, err
	}

	// the symbols are not sorted by default. Since a number of functions benefit from them being ordered,
	// we copy out the symbols and sort them.
	syms := make([]macho.Symbol, len(mf.Symtab.Syms))
	copy(syms, mf.Symtab.Syms)
	sort.Slice(syms, func(i, j int) bool {
		return syms[i].Value < syms[j].Value
	})

	return &Macho{
		f:       mf,
		symbols: syms,
	}, nil
}

// ByteOrder returns the byte order (little or big endian)
func (m *Macho) ByteOrder() binary.ByteOrder {
	return m.f.ByteOrder
}

// TextSection locates and returns __text
func (m *Macho) TextSection() (Section, error) {
	return m.section("__text")
}

// TextSection locates and returns __rodata
func (m *Macho) RODataSection() (Section, error) {
	return m.section("__rodata")
}

// PCLNTabSection locates and returns __gopclntab
func (m *Macho) PCLNTabSection() (Section, error) {
	return m.section("__gopclntab")
}

// section searches for a section by name
func (m *Macho) section(name string) (Section, error) {
	sect := m.f.Section(name)
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
func (m *Macho) SectionContainingRange(addrRange address.Range) (Section, error) {
	for _, s := range m.f.Sections {
		if addrRange.Start >= s.Addr && addrRange.Start <= s.Addr+s.Size {
			if addrRange.End < s.Addr && addrRange.End > s.Addr+s.Size {
				return Section{}, fmt.Errorf("go.string.* unexpetedly overflows from section %s (%s)", s.Name, addrRange)
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

// Symbol locates a symbol by name. Because Mach-O symbols do not carry size, this returns a "best guess" at the
// address range by returning -1 of the closest subsequent symbol. This is imperfect but good enough.
func (m *Macho) Symbol(name string) (Symbol, error) {
	var (
		matched Symbol
		found   bool
	)
	for _, sym := range m.symbols {
		if found {
			matched.Range.End = sym.Value - 1
			return matched, nil
		}
		if sym.Name == name {
			matched = Symbol{
				Name:  sym.Name,
				Range: address.Range{Start: sym.Value},
			}
			found = true
		}
	}
	return Symbol{}, fmt.Errorf("symbol %s not found", name)
}

// SymbolForAddress locates a symbol that is closest to the supplied address. As with the Symbol above, it returns a
// best guess at the address range. If a symbol cannot be found for the address an error is returned.
func (m *Macho) SymbolForAddress(addr uint64) (Symbol, error) {
	var previous macho.Symbol
	for _, sym := range m.symbols {
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
	return Symbol{}, fmt.Errorf("symbol for address %x not found", addr)
}

// Close closes the underlying file
func (m *Macho) Close() error {
	return m.f.Close()
}
