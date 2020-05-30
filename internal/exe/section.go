package exe

import (
	"io"

	"github.com/nick-jones/gost/internal/address"
)

// Section represents a particular section of a binary
type Section struct {
	Name      string
	AddrRange address.Range
	io.ReaderAt
}

// Data returns the raw bytes for this particular section
func (s Section) Data() ([]byte, error) {
	buf := make([]byte, s.AddrRange.Size())
	if _, err := s.ReaderAt.ReadAt(buf, 0); err != nil {
		return nil, err
	}
	return buf, nil
}
