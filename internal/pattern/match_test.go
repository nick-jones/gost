package pattern_test

import (
	"testing"

	"github.com/nick-jones/gost/internal/pattern"
	"github.com/stretchr/testify/assert"
)

func TestMatchBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}

	patterns := [][]byte{
		{
			0x01, 0x02, 0x03,
		},
		{
			0x03, 0x04,
		},
		{
			0x03, 0x04, 0x05,
		},
		{
			0x03, pattern.Wildcard,
		},
	}

	expected := []pattern.Match{
		{
			Index:   0,
			Pattern: 0,
		},
		{
			Index:   2,
			Pattern: 1,
		},
		{
			Index:   2,
			Pattern: 3,
		},
	}

	assert.Equal(t, expected, pattern.MatchBytes(data, patterns))
}
