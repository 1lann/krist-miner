package shasimd

import (
	"github.com/1lann/krist-miner/sha2"
	"github.com/1lann/sha256-simd"
)

type generator struct{}

func (g *generator) Sum256Number(data []byte) int64 {
	return sha256.SumToNum256(data)
}

func (g *generator) Sum256NumberCmp(data []byte, work int64) bool {
	return sha256.SumCmp256(data, work)
}

func init() {
	sha2.RegisterAlgorithm("simd", func() sha2.SumNumberAlgorithm {
		return &generator{}
	})
}
