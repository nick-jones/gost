package exe

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/nick-jones/gost/internal/address"
)

var (
	// ErrSymbolNotFound is returned when a symbol is requested that cannot be located
	ErrSymbolNotFound = errors.New("symbol not found")
	// ErrSectionNotFound is returned when a section is requested that cannot be located
	ErrSectionNotFound = errors.New("section not found")

	machoMagicLE = []byte{0xcf, 0xfa, 0xed, 0xfe}
	machoMagicBE = []byte{0xfe, 0xed, 0xfa, 0xcf}
	elfMagic     = []byte{0x7f, 0x45, 0x4c, 0x46}
)

// File represents an executable file
type File struct {
	adapt adapter
	f     *os.File
}

type adapter interface {
	ByteOrder() binary.ByteOrder
	TextSection() (Section, error)
	RODataSection() (Section, error)
	PCLNTabSection() (Section, error)
	Sections() ([]Section, error)
	Symbols() ([]Symbol, error)
}

// Open opens the named executable file
func Open(filePath string) (*File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}

	ident := make([]byte, 4)
	if _, err := f.ReadAt(ident, 0); err != nil {
		return nil, err
	}

	var adapt adapter
	switch {
	case bytes.Equal(ident, machoMagicLE) || bytes.Equal(ident, machoMagicBE):
		adapt, err = newMachoFile(f)
	case bytes.Equal(ident, elfMagic):
		adapt, err = newELFFile(f)
	default:
		err = fmt.Errorf("could not determine exe type for %s", filePath)
	}
	if err != nil {
		return nil, err
	}

	return &File{adapt: adapt}, nil
}

// ByteOrder returns the byte order (little or big endian)
func (e *File) ByteOrder() binary.ByteOrder {
	return e.adapt.ByteOrder()
}

// SectionContainingRange returns the section that fully contains the supplied range
func (e *File) SectionContainingRange(addrRange address.Range) (Section, error) {
	sects, err := e.adapt.Sections()
	if err != nil {
		return Section{}, err
	}
	for _, s := range sects {
		if s.AddrRange.Contains(addrRange.Start) && s.AddrRange.Contains(addrRange.End) {
			return s, nil
		}
	}
	return Section{}, fmt.Errorf("failed to locate section for address range %s", addrRange)
}

// Symbol locates a symbol by name
func (e *File) Symbol(name string) (Symbol, error) {
	sects, err := e.adapt.Symbols()
	if err != nil {
		return Symbol{}, err
	}
	for _, s := range sects {
		if s.Name == name {
			return s, nil
		}
	}
	return Symbol{}, ErrSymbolNotFound
}

// SymbolsForAddresses locates at most one symbol for each address. Not every address may resolve to a symbol; in such
// cases the address will not feature in the returned map.
func (e *File) SymbolsForAddresses(addrs []uint64) (map[uint64]Symbol, error) {
	sort.Slice(addrs, func(i, j int) bool {
		return addrs[i] < addrs[j]
	})

	syms, err := e.adapt.Symbols()
	if err != nil {
		return nil, err
	}

	results := make(map[uint64]Symbol)
	i := 0
	for _, addr := range addrs {
		for ; i < len(syms) && !syms[i].AddrRange.Contains(addr); i++ {
			// keep moving the index forward until we find a match (or run out of symbols)
		}
		if i == len(syms) {
			// reached the end of symbols
			break
		}
		results[addr] = syms[i]
	}
	return results, nil
}

// TextSection returns the text section
func (e *File) TextSection() (Section, error) {
	return e.adapt.TextSection()
}

// RODataSection returns the read-only data section
func (e *File) RODataSection() (Section, error) {
	return e.adapt.RODataSection()
}

// PCLNTabSection returns the Go PCLN table section
func (e *File) PCLNTabSection() (Section, error) {
	return e.adapt.PCLNTabSection()
}

// Close closes the underlying file
func (e *File) Close() error {
	return e.f.Close()
}
