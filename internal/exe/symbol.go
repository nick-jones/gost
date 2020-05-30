package exe

import "github.com/nick-jones/gost/internal/address"

// Symbol represents a symbol compiled into an executable
type Symbol struct {
	Name  string
	Range address.Range
}
