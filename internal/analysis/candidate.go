package analysis

// Candidate is a potential candidate string reference
type Candidate struct {
	Addr     uint64   // address where the string resides
	Len      uint64   // length of the string
	RefAddrs []uint64 // addresses that reference the string
}
