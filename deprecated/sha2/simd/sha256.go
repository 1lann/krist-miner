package shasimd

import (
	"github.com/1lann/krist-miner/deprecated/sha2"
	"github.com/1lann/sha256-simd"
)

type generator struct{}

func (g *generator) Sum256Number(data []byte) int64 {
	var start [64]byte
	start[41] = 128
	start[63] = 72
	start[62] = 1

	if len(data) != 41 {
		panic("must be exactly 41 long")
	}
	copy(start[:], data)

	return sha256.SumToNum256(start[:])
}

func (g *generator) Sum256NumberCmp(data []byte, work int64) bool {
	var start [64]byte
	start[41] = 128
	start[63] = 72
	start[62] = 1

	if len(data) != 41 {
		panic("must be exactly 41 long")
	}
	copy(start[:], data)

	return sha256.SumCmp256(start[:], uint32(work))
}

func init() {
	sha2.RegisterAlgorithm("simd", func() sha2.SumNumberAlgorithm {
		return &generator{}
	})
}
