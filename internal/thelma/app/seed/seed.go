package seed

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"github.com/rs/zerolog/log"
	mathrand "math/rand"
	"time"
)

// Rand seeds the math/rand package's pseudo random number generator
// Via stack overflow: https://stackoverflow.com/a/54491783
func Rand() {
	var seed [8]byte
	_, err := cryptorand.Read(seed[:])
	if err != nil {
		log.Warn().Msgf("Failed to seed math/rand with crypto/rand, falling back to time seed: %v", err)
		mathrand.Seed(time.Now().UnixNano())
	}
	mathrand.Seed(int64(binary.LittleEndian.Uint64(seed[:])))
}
