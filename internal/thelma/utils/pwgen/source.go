package pwgen

import (
	crypto "crypto/rand"
	"encoding/binary"
	"fmt"
)

// Provide a cryptographic rand source - pulled from a PR to go standard lib
// https://github.com/golang/go/issues/25531

type reader struct{}

func (r *reader) Read(p []byte) (n int, err error) {
	return crypto.Read(p)
}

// random source that relies on crypto/rand
type source struct{}

func (s *source) Int63() (random int64) {
	err := binary.Read(&reader{}, binary.BigEndian, &random)
	if err != nil {
		panic(fmt.Sprintf("converting random bytes to an int64: %s", err.Error()))
	}
	if random < 0 {
		random = -random
	}
	return random
}

func (s *source) Seed(_ int64) {
	panic("you cannot seed the cryptographic source")
}
