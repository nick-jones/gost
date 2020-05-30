package analysis

import (
	"encoding/binary"
	"fmt"

	"github.com/nick-jones/gost/internal/address"
	"github.com/nick-jones/gost/internal/exe"
	"github.com/nick-jones/gost/internal/pattern"
)

// FindDirectReferences scans for direct references to the supplied address range and returns candidates
func FindDirectReferences(f exe.File, strRange address.Range) ([]Candidate, error) {
	// the __text section contains executable instructions
	txt, err := f.TextSection()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve text section: %w", err)
	}

	// read data for this section
	data, err := txt.Data()
	if err != nil {
		return nil, fmt.Errorf("could not read data from text section: %w", err)
	}

	// copy out the patterns
	patterns := make([][]byte, len(directMatchers))
	for i, matcher := range directMatchers {
		patterns[i] = matcher.pattern
	}

	// search __text section for string candidates
	matched := pattern.MatchBytes(data, patterns)

	var candidates []Candidate
	for _, m := range matched {
		matcher := directMatchers[m.Pattern] // locate original directMatcher

		var arg1, arg2 int
		if matcher.arg1Pos >= 0 {
			// extract the stack pointer offset for the first argument (pointer to string value). If a position is not
			// supplied, we assume zero offset
			arg1 = int(data[m.Index+matcher.arg1Pos])
		}
		// extract the stack pointer offset for the second argument (string length)
		arg2 = int(data[m.Index+matcher.arg2Pos])

		// the string and length are always passed around together. Since the Go compiler uses the stack to pass
		// arguments (i.e. doesn't use System V), we can use this as an additional heuristic; a point to the string
		// value should be set into the stack. The length should be set +8 bytes from that.
		if arg1%8 == 0 && arg2 == arg1+8 {
			relAddr := txt.AddrRange.Start + uint64(m.Index+matcher.offsetPos+matcher.offsetLen)
			offset := uint64(readUint32(data[m.Index+matcher.offsetPos:m.Index+matcher.offsetPos+matcher.offsetLen], f.ByteOrder()))
			checkAddr := relAddr + offset

			if strRange.Contains(checkAddr) {
				length := uint64(readUint32(data[m.Index+matcher.lenPos:m.Index+matcher.lenPos+matcher.lenSize], f.ByteOrder()))
				candidates = append(candidates, Candidate{
					Addr:     checkAddr,
					Len:      length,
					RefAddrs: []uint64{txt.AddrRange.Start + uint64(m.Index+matcher.insPos)},
				})
			}
		}
	}
	return candidates, nil
}

// readUint32 will return a uint32 from the supplied bytes, taking the byte order into account
func readUint32(src []byte, bo binary.ByteOrder) uint32 {
	switch len(src) {
	case 1:
		return uint32(src[0])
	case 2:
		return uint32(bo.Uint16(src))
	case 4:
		return bo.Uint32(src)
	default:
		panic(fmt.Sprintf("unexpected src size %d", len(src)))
	}
}

// readUint64 will return a uint64 from the supplied bytes, taking the byte order into account
func readUint64(src []byte, bo binary.ByteOrder) uint64 {
	switch len(src) {
	case 1:
		return uint64(src[0])
	case 2:
		return uint64(bo.Uint16(src))
	case 4:
		return uint64(bo.Uint32(src))
	case 8:
		return bo.Uint64(src)
	default:
		panic(fmt.Sprintf("unexpected src size %d", len(src)))
	}
}
