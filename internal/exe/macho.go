package exe

import (
	"debug/macho"
	"encoding/binary"
	"os"
	"sort"

	"github.com/nick-jones/gost/internal/address"
)

// machoFile covers Mach-O type executables
type machoFile struct {
	byteOrder binary.ByteOrder
	symbols   []Symbol
	sections  []Section
}

// newMachoFile initialises the machoFile type
func newMachoFile(f *os.File) (*machoFile, error) {
	mf, err := macho.NewFile(f)
	if err != nil {
		return nil, err
	}

	return &machoFile{
		byteOrder: mf.ByteOrder,
		symbols:   mapMachoSymbols(mf),
		sections:  mapMachoSections(mf),
	}, nil
}

// ByteOrder returns the byte order (little or big endian)
func (m *machoFile) ByteOrder() binary.ByteOrder {
	return m.byteOrder
}

// TextSection locates and returns __text
func (m *machoFile) TextSection() (Section, error) {
	return m.section("__text")
}

// TextSection locates and returns __rodata
func (m *machoFile) RODataSection() (Section, error) {
	return m.section("__rodata")
}

// PCLNTabSection locates and returns __gopclntab
func (m *machoFile) PCLNTabSection() (Section, error) {
	return m.section("__gopclntab")
}

// section searches for a section by name
func (m *machoFile) section(name string) (Section, error) {
	for _, s := range m.sections {
		if s.Name == name {
			return s, nil
		}
	}
	return Section{}, ErrSectionNotFound
}

// Sections returns all sections
func (m *machoFile) Sections() ([]Section, error) {
	return m.sections, nil
}

// Symbols returns all symbols
func (m *machoFile) Symbols() ([]Symbol, error) {
	return m.symbols, nil
}

// mapELFSymbols maps Mach-O symbols to our standard type
func mapMachoSymbols(f *macho.File) []Symbol {
	// copy symbols & sort
	syms := make([]macho.Symbol, len(f.Symtab.Syms))
	copy(syms, f.Symtab.Syms)
	sort.Slice(syms, func(i, j int) bool {
		return syms[i].Value < syms[j].Value
	})

	// Mach-O symbols do not carry size, so the following performs a best guess at address ranges. It's imperfect but
	// good enough for what we need.
	mapped := make([]Symbol, 0, len(syms))
	buffered := make([]macho.Symbol, 0)
	var anchor uint64
	for _, s := range syms {
		if len(buffered) > 0 && s.Value > anchor {
			for _, b := range buffered {
				mapped = append(mapped, Symbol{
					Name: b.Name,
					AddrRange: address.Range{
						Start: b.Value,
						End:   s.Value - 1,
					},
				})
			}
			buffered = buffered[:0]
		}
		buffered = append(buffered, s)
		anchor = s.Value
	}
	for _, b := range buffered {
		mapped = append(mapped, Symbol{
			Name: b.Name,
			AddrRange: address.Range{
				Start: b.Value,
				End:   b.Value, // since we don't know where to end this we'll just use the same address
			},
		})
	}
	return mapped
}

// mapMachoSections maps Mach-O sections to our standard type
func mapMachoSections(f *macho.File) []Section {
	sects := make([]Section, len(f.Sections))
	for i, s := range f.Sections {
		sects[i] = Section{
			Name: s.Name,
			AddrRange: address.Range{
				Start: s.Addr,
				End:   s.Addr + s.Size,
			},
			ReaderAt: s.ReaderAt,
		}
	}
	return sects
}
