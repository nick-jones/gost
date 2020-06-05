package exe

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/nick-jones/gost/internal/address"
)

type File interface {
	ByteOrder() binary.ByteOrder
	TextSection() (Section, error)
	RODataSection() (Section, error)
	PCLNTabSection() (Section, error)
	SectionContainingRange(address.Range) (Section, error)
	Symbol(name string) (Symbol, error)
	SymbolForAddress(addr uint64) (Symbol, error)
	io.Closer
}

var (
	machoMagicLE = []byte{0xcf, 0xfa, 0xed, 0xfe}
	machoMagicBE = []byte{0xfe, 0xed, 0xfa, 0xcf}
)

// Open opens the named file
func Open(filePath string) (File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}

	ident := make([]byte, 4)
	if _, err := f.ReadAt(ident, 0); err != nil {
		return nil, err
	}

	if bytes.Equal(ident, machoMagicLE) || bytes.Equal(ident, machoMagicBE) {
		return newMacho(f)
	}

	return nil, fmt.Errorf("could not determine exe type for %s", filePath)
}
