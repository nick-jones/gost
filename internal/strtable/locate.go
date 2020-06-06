package strtable

import (
	"errors"
	"fmt"

	"github.com/nick-jones/gost/internal/address"
	"github.com/nick-jones/gost/internal/exe"
)

// Locate returns the address range for the Go string table. This uses symbol information, falling back to plain old
// guess work if symbols are unavailable.
func Locate(f *exe.File) (address.Range, error) {
	// use the go.string.* symbol if available
	sym, err := f.Symbol("go.string.*")
	if err != nil {
		if errors.Is(err, exe.ErrSymbolNotFound) {
			// otherwise guess the address range
			return guessStringTableAddressRange(f)
		}
		return address.Range{}, fmt.Errorf("failed to locate go.string.* range: %w", err)
	}
	return sym.AddrRange, nil
}

// guessStringTableAddressRange is an imperfect attempt at guessing the address range for the Go string table. It looks
// for contiguous blocks of 7-bit ASCII. This is obviously defeated by UTF-8; further down it attempts to join blocks
// that are close by. Again this can be defeated by numerous UTF-8 encoded characters in a row. Unfortunately there is
// no simple way to guess the range.
func guessStringTableAddressRange(f *exe.File) (address.Range, error) {
	rodata, err := f.RODataSection()
	if err != nil {
		return address.Range{}, err
	}

	data, err := rodata.Data()
	if err != nil {
		return address.Range{}, err
	}

	// locate contiguous blocks of what look like ASCII.. this is of course not perfect.
	blocks := make([]address.Range, 0)
	current := address.Range{}
	inBlock := false
	for i, b := range data {
		// any printable 7-bit ASCII character along with newline and tab
		if (b >= 0x20 && b <= 0x7E) || b == 0x0A || b == 0x09 {
			if !inBlock {
				current.Start = rodata.AddrRange.Start + uint64(i)
				inBlock = true
			}
			continue
		}

		if inBlock {
			current.End = rodata.AddrRange.Start + uint64(i)
			blocks = append(blocks, current)
			current = address.Range{}
			inBlock = false
		}
	}

	if len(blocks) == 0 {
		return address.Range{}, fmt.Errorf("failed to find any contiguous large blocks of ASCII")
	}

	// merge nearby ranges and take the largest block
	var max address.Range
	for _, block := range mergeAddressRanges(blocks, 16) {
		if block.Size() > max.Size() {
			max = block
		}
	}
	return max, nil
}

// mergeAddressRanges merges ranges where the end and start addresses are no greater than the supplied max distances. The
// address ranges must be supplied ordered.
func mergeAddressRanges(addrRanges []address.Range, maxDistance int) []address.Range {
	if len(addrRanges) <= 1 {
		return addrRanges
	}
	merged := make([]address.Range, 0)
	var previous address.Range
	for i, addrRange := range addrRanges {
		if i == 0 {
			previous = addrRange
			continue
		}
		if addrRange.Start - previous.End < uint64(maxDistance) {
			previous.End = addrRange.End
			continue
		}
		merged = append(merged, previous)
		previous = addrRange
	}
	return append(merged, previous)
}