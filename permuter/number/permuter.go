package number

import (
	"github.com/1lann/krist-miner/permuter"
	"strconv"
)

type generator struct {
	lastNumber int64
}

func (g *generator) Next() []byte {
	g.lastNumber++
	var dst []byte
	return strconv.AppendInt(dst, g.lastNumber, 10)
}

func (g *generator) Reset() {
	g.lastNumber = 0
}

func init() {
	permuter.RegisterAlgorithm("number", func() permuter.PermuterAlgorithm {
		return &generator{0}
	})
}
