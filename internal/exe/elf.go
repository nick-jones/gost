package exe

import (
	"debug/elf"
	"encoding/binary"
	"errors"
	"io"
	"sort"

	"github.com/nick-jones/gost/internal/address"
)

// elfFile covers Executable and Linkable Format (elfFile) type binaries
type elfFile struct {
	byteOrder binary.ByteOrder
	symbols   []Symbol
	sections  []Section
}

// newELFFile initialises the elfFile type
func newELFFile(r io.ReaderAt) (*elfFile, error) {
	ef, err := elf.NewFile(r)
	if err != nil {
		return nil, err
	}

	syms, err := mapELFSymbols(ef)
	if err != nil {
		return nil, err
	}

	return &elfFile{
		byteOrder: ef.ByteOrder,
		symbols:   syms,
		sections:  mapELFSections(ef),
	}, nil
}

// ByteOrder returns the byte order (little or big endian)
func (e *elfFile) ByteOrder() binary.ByteOrder {
	return e.byteOrder
}

// TextSection locates and returns .text
func (e *elfFile) TextSection() (Section, error) {
	return e.section(".text")
}

// TextSection locates and returns .rodata
func (e *elfFile) RODataSection() (Section, error) {
	return e.section(".rodata")
}

// TextSection locates and returns .gopclntab
func (e *elfFile) PCLNTabSection() (Section, error) {
	return e.section(".gopclntab")
}

// section searches for a section by name
func (e *elfFile) section(name string) (Section, error) {
	for _, s := range e.sections {
		if s.Name == name {
			return s, nil
		}
	}
	return Section{}, ErrSectionNotFound
}

// Sections returns all known sections
func (e *elfFile) Sections() ([]Section, error) {
	return e.sections, nil
}

// Symbols returns all known symbols
func (e *elfFile) Symbols() ([]Symbol, error) {
	return e.symbols, nil
}

// mapELFSymbols maps ELF symbols to our standard type
func mapELFSymbols(f *elf.File) ([]Symbol, error) {
	// read symbols
	orig, err := f.Symbols()
	if err != nil {
		if errors.Is(err, elf.ErrNoSymbols) {
			// symbols are not guaranteed to be included, so this error can be ignored
			return nil, nil
		}
		return nil, err
	}

	// copy symbols & sort
	syms := make([]elf.Symbol, len(orig))
	copy(syms, orig)
	sort.Slice(syms, func(i, j int) bool {
		return syms[i].Value < syms[j].Value
	})

	// Map symbols to our standard Symbol type. ELF symbols can carry size, with Go compiled binaries some important
	// symbols don't have size set. So we use similar tricks to Mach-O where the size is zero: guess the size based on
	// the next symbol. This may not be perfect but it works well enough for what we need.
	mapped := make([]Symbol, 0, len(syms))
	buffered := make([]elf.Symbol, 0)
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
		if s.Size > 0 {
			// if we have size we can map the symbol straight away
			mapped = append(mapped, Symbol{
				Name: s.Name,
				AddrRange: address.Range{
					Start: s.Value,
					End:   s.Value + s.Size,
				},
			})
		} else {
			// if we have no size we'll buffer it until we have a symbol with a greater address
			buffered = append(buffered, s)
			anchor = s.Value
		}
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
	return mapped, nil
}

// mapELFSections maps ELF sections to our standard type
func mapELFSections(f *elf.File) []Section {
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
