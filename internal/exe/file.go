package exe

import (
	"debug/gosym"
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
	SectionContainingRange(address.Range) (Section, error)
	Symbol(name string) (Symbol, error)
	SymbolForAddress(addr uint64) (Symbol, error)
	SymbolsInRange(address.Range) ([]Symbol, error)
	GoSymbolTable() (*gosym.Table, error)
	io.Closer
}

// Open opens the named file
func Open(filePath string) (File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	// TODO: check for Mach-O / ELF / etc
	return newMacho(f)
}
