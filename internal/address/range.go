package address

import "fmt"

// Range represents a range of addresses
type Range struct {
	Start uint64
	End   uint64
}

// Size returns the size of the range
func (r Range) Size() int {
	return int(r.End - r.Start)
}

// Equal returns true if both ranges start and end at the same address
func (r Range) Equal(other Range) bool {
	return r.Start == other.Start && r.End == other.End
}

// Contains returns true if the supplied address is within the range boundaries
func (r Range) Contains(addr uint64) bool {
	return addr >= r.Start && addr <= r.End
}

// String returns a string representation of the range
func (r Range) String() string {
	return fmt.Sprintf("0x%x - 0x%x", r.Start, r.End)
}
