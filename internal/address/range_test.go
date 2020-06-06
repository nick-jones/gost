package address_test

import (
	"testing"

	"github.com/nick-jones/gost/internal/address"
	"github.com/stretchr/testify/assert"
)

func TestRange_Size(t *testing.T) {
	r := address.Range{Start: 10, End: 20}
	assert.Equal(t, 10, r.Size())
}

func TestRange_Equal(t *testing.T) {
	testCases := []struct {
		name       string
		addrRange1 address.Range
		addrRange2 address.Range
		expect     bool
	}{
		{
			name:       "equal start and end values",
			addrRange1: address.Range{Start: 10, End: 20},
			addrRange2: address.Range{Start: 10, End: 20},
			expect:     true,
		},
		{
			name:       "start value is higher",
			addrRange1: address.Range{Start: 10, End: 20},
			addrRange2: address.Range{Start: 11, End: 20},
			expect:     false,
		},
		{
			name:       "start value is lower",
			addrRange1: address.Range{Start: 10, End: 20},
			addrRange2: address.Range{Start: 9, End: 20},
			expect:     false,
		},
		{
			name:       "end value is higher",
			addrRange1: address.Range{Start: 10, End: 21},
			addrRange2: address.Range{Start: 10, End: 20},
			expect:     false,
		},
		{
			name:       "end value is lower",
			addrRange1: address.Range{Start: 10, End: 19},
			addrRange2: address.Range{Start: 10, End: 20},
			expect:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			assert.Equal(tt, tc.expect, tc.addrRange1.Equal(tc.addrRange2))
		})
	}
}

func TestRange_Contains(t *testing.T) {
	testCases := []struct {
		name      string
		addr      uint64
		addrRange address.Range
		expect    bool
	}{
		{
			name:      "address within range",
			addr:      11,
			addrRange: address.Range{Start: 10, End: 12},
			expect:    true,
		},
		{
			name:      "address within range (start boundary)",
			addr:      10,
			addrRange: address.Range{Start: 10, End: 12},
			expect:    true,
		},
		{
			name:      "address within range (end boundary)",
			addr:      12,
			addrRange: address.Range{Start: 10, End: 12},
			expect:    true,
		},
		{
			name:      "address outside of range (too low)",
			addr:      9,
			addrRange: address.Range{Start: 10, End: 12},
			expect:    false,
		},
		{
			name:      "address outside of range (too high)",
			addr:      13,
			addrRange: address.Range{Start: 10, End: 12},
			expect:    false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			assert.Equal(tt, tc.expect, tc.addrRange.Contains(tc.addr))
		})
	}
}
