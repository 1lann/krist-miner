package urandom

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/1lann/krist-miner/permuter"
)

type generator struct{}

func (g *generator) Next() []byte {
	var result = make([]byte, 5)
	_, err := rand.Read(result)
	if err != nil {
		panic(err)
	}

	var dst = make([]byte, hex.EncodedLen(len(result)))

	hex.Encode(dst, result)

	return dst
}

func (g *generator) Reset() {

}

func init() {
	permuter.RegisterAlgorithm("urandom", func() permuter.PermuterAlgorithm {
		return &generator{}
	})
}
