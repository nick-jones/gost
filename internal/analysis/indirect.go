package analysis

import (
	"fmt"
	"reflect"

	"github.com/nick-jones/gost/internal/address"
	"github.com/nick-jones/gost/internal/exe"
	"github.com/nick-jones/gost/internal/pattern"
)

// EvaluateIndirectReferences scans for indirect references to the supplied address range and returns candidates
func EvaluateIndirectReferences(f *exe.File, strRange address.Range) ([]Candidate, error) {
	refs, err := findInterfaceReferences(f)
	if err != nil {
		return nil, err
	}

	sect, err := f.RODataSection()
	if err != nil {
		return nil, err
	}

	typeBuf := make([]byte, 1)
	strPtrBuf := make([]byte, 8)
	strLenBuf := make([]byte, 8)
	candidates := make([]Candidate, 0)
	for _, ref := range refs {
		if !sect.AddrRange.Contains(ref.typeAddr) || !sect.AddrRange.Contains(ref.valueHeaderAddr) {
			continue
		}

		// check type
		if _, err := sect.ReadAt(typeBuf, int64(ref.typeAddr-sect.AddrRange.Start+23)); err != nil {
			return nil, err
		}
		if reflect.Kind(typeBuf[0]) != reflect.String {
			continue
		}

		// read pointer and check address
		if _, err := sect.ReadAt(strPtrBuf, int64(ref.valueHeaderAddr-sect.AddrRange.Start)); err != nil {
			return nil, err
		}
		strPtr := readUint64(strPtrBuf, f.ByteOrder())
		if !strRange.Contains(strPtr) {
			continue
		}

		// read len
		if _, err := sect.ReadAt(strLenBuf, int64(ref.valueHeaderAddr-sect.AddrRange.Start+8)); err != nil {
			return nil, err
		}
		strLen := readUint64(strLenBuf, f.ByteOrder())

		candidates = append(candidates, Candidate{
			Addr:     strPtr,
			Len:      strLen,
			RefAddrs: []uint64{ref.addr},
		})
	}
	return candidates, nil
}

type interfaceReference struct {
	addr            uint64
	typeAddr        uint64
	valueHeaderAddr uint64
}

func findInterfaceReferences(f *exe.File) ([]interfaceReference, error) {
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
	for i, matcher := range indirectMatchers {
		patterns[i] = matcher.pattern
	}

	// search __text section for string candidates
	matched := pattern.MatchBytes(data, patterns)

	references := make([]interfaceReference, 0)
	for _, m := range matched {
		matcher := indirectMatchers[m.Pattern] // locate original indirectMatcher

		var arg1, arg2 int
		if matcher.arg1Pos >= 0 {
			// extract the stack pointer offset for the first argument (pointer to string value). If a position is not
			// supplied, we assume zero offset
			arg1 = int(data[m.Index+matcher.arg1Pos])
		}
		// extract the stack pointer offset for the second argument (string length)
		arg2 = int(data[m.Index+matcher.arg2Pos])

		if arg1%8 == 0 && arg2 == arg1+8 {
			refAddr := txt.AddrRange.Start + uint64(m.Index+matcher.insPos)
			typeRelAddr := txt.AddrRange.Start + uint64(m.Index+matcher.typeOffsetPos+matcher.typeOffsetLen)
			typeOffset := uint64(readUint32(data[m.Index+matcher.typeOffsetPos:m.Index+matcher.typeOffsetPos+matcher.typeOffsetLen], f.ByteOrder()))
			valueHeaderRelAddr := txt.AddrRange.Start + uint64(m.Index+matcher.valueHeaderOffsetPos+matcher.valueHeaderOffsetLen)
			valueHeaderOffset := uint64(readUint32(data[m.Index+matcher.valueHeaderOffsetPos:m.Index+matcher.valueHeaderOffsetPos+matcher.valueHeaderOffsetLen], f.ByteOrder()))

			references = append(references, interfaceReference{
				addr:            refAddr,
				typeAddr:        typeRelAddr + typeOffset,
				valueHeaderAddr: valueHeaderRelAddr + valueHeaderOffset,
			})
		}
	}

	return references, nil
}
