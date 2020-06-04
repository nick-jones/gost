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
	size := s.AddrRange.Size()
	if size == 0 {
		return nil, nil
	}
	buf := make([]byte, size)
	if _, err := s.ReaderAt.ReadAt(buf, 0); err != nil {
		return nil, err
	}
	return buf, nil
}
