package pwgen

import (
	"github.com/broadinstitute/thelma/internal/thelma/utils/set"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Pwgen(t *testing.T) {
	testCases := []struct {
		name      string
		pwgen     Pwgen
		minCounts counts
		len       int
	}{
		{
			name: "length 0",
			len:  defaultLength,
		},
		{
			name: "length rounded up to absolute min",
			pwgen: Pwgen{
				MinLength: 2,
			},
			len: absoluteMinimumLength,
		},
		{
			name: "length respected if greater than absolute min",
			pwgen: Pwgen{
				MinLength: absoluteMinimumLength + 10,
			},
			len: absoluteMinimumLength + 10,
		},
		{
			name: "all upper",
			pwgen: Pwgen{
				MinLength: 12,
				MinUpper:  12,
			},
			minCounts: counts{
				lower:   0,
				upper:   12,
				num:     0,
				special: 0,
			},
			len: 12,
		},
		{
			name: "all num",
			pwgen: Pwgen{
				MinLength: 20,
				MinNum:    20,
			},
			minCounts: counts{
				lower:   0,
				upper:   0,
				num:     20,
				special: 0,
			},
			len: 20,
		},
		{
			name: "all special",
			pwgen: Pwgen{
				MinLength:  47,
				MinSpecial: 47,
			},
			minCounts: counts{
				lower:   0,
				upper:   0,
				num:     0,
				special: 47,
			},
			len: 47,
		},
		{
			name: "mixed",
			pwgen: Pwgen{
				MinLength:  32,
				MinLower:   6,
				MinUpper:   7,
				MinNum:     9,
				MinSpecial: 10,
			},
			minCounts: counts{
				lower:   6,
				upper:   7,
				num:     9,
				special: 10,
			},
			len: 32,
		},
		{
			name: "mixed with room for random chars",
			pwgen: Pwgen{
				MinLength:  12,
				MinLower:   1,
				MinUpper:   1,
				MinNum:     1,
				MinSpecial: 1,
			},
			minCounts: counts{
				lower:   1,
				upper:   1,
				num:     1,
				special: 1,
			},
			len: 12,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pass := tc.pwgen.Generate()

			c := countKinds(pass)

			assert.LessOrEqualf(t, tc.minCounts.lower, c.lower, "expect >= %d lowercase chars, have %d (%q)", tc.minCounts.lower, c.lower, pass)
			assert.LessOrEqual(t, tc.minCounts.upper, c.upper, "expect >= %d uppercase chars, have %d (%q)", tc.minCounts.upper, c.upper, pass)
			assert.LessOrEqual(t, tc.minCounts.num, c.num, "expect >= %d numeric chars, have %d (%q)", tc.minCounts.num, c.num, pass)
			assert.LessOrEqual(t, tc.minCounts.special, c.special, "expect >= %d special chars, have %d (%q)", tc.minCounts.special, c.special, pass)
			assert.Equal(t, tc.len, len(pass))
		})
	}
}

type counts struct {
	lower   int
	upper   int
	num     int
	special int
}

type runeSets struct {
	lower   set.Set[rune]
	upper   set.Set[rune]
	num     set.Set[rune]
	special set.Set[rune]
}

var rs = func() runeSets {
	var s runeSets
	s.lower = set.NewSet(lower...)
	s.upper = set.NewSet(upper...)
	s.num = set.NewSet(num...)
	s.special = set.NewSet(special...)
	return s
}()

func countKinds(s string) counts {
	var c counts
	for _, r := range s {
		if rs.lower.Exists(r) {
			c.lower++
		} else if rs.upper.Exists(r) {
			c.upper++
		} else if rs.num.Exists(r) {
			c.num++
		} else if rs.special.Exists(r) {
			c.special++
		} else {
			panic(errors.Errorf("unrecognized character: %c", r))
		}
	}

	return c
}
