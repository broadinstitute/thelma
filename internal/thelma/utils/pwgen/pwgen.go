package pwgen

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/app/seed"
	"math/rand"
)

const defaultLength = 24
const absoluteMinimumLength = 8

func init() {
	seed.Rand()
}

var lower = []rune(`abcdefghijklmnopqrstuvwxyz`)
var upper = []rune(`ABCDEFGHIJKLMNOPQRSTUVWXYZ`)
var num = []rune(`1234567890`)
var special = []rune(`!@#$%^&*()_-+={}[]/<>.,;?:|`) // '"\ excluded

var all = func() []rune {
	var a []rune
	a = append(a, lower...)
	a = append(a, upper...)
	a = append(a, num...)
	a = append(a, special...)
	return a
}()

// Pwgen is for generating random passwords that meet strict password policy requirements (eg.
// those that require symbols, letters, numbers, etc)
type Pwgen struct {
	MinLength         int
	MinLower          int
	MinUpper          int
	MinNum            int
	MinSpecial        int
	ExcludeCharacters []rune
}

func (p Pwgen) Generate() string {
	l := p.MinLength
	if l <= 0 {
		l = defaultLength
	}
	if l < absoluteMinimumLength {
		l = absoluteMinimumLength
	}

	minChars := p.MinLower + p.MinUpper + p.MinNum + p.MinSpecial
	if l < minChars {
		l = minChars
	}

	buf := make([]rune, l)

	i := 0

	n := p.MinLower
	for ; i < n; i++ {
		buf[i] = pick(lower)
	}

	n += p.MinUpper
	for ; i < n; i++ {
		buf[i] = pick(upper)
	}

	n += p.MinNum
	for ; i < n; i++ {
		buf[i] = pick(num)
	}

	n += p.MinSpecial
	for ; i < n; i++ {
		buf[i] = pick(special)
	}

	n = l
	for ; i < n; i++ {
		buf[i] = pick(all)
	}

	rand.Shuffle(n, func(a, b int) {
		buf[a], buf[b] = buf[b], buf[a]
	})

	password := string(buf)
	logging.MaskSecret(password)
	return password
}

func pick(from []rune) rune {
	return from[rand.Intn(len(from))]
}
