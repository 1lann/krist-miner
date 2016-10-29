package permuterascii

import (
	"github.com/1lann/krist-miner/permuter"
)

func incrementString(text []byte) {
	for place := len(text) - 1; place >= 0; place-- {
		if text[place] < 'z' {
			text[place] = text[place] + 1
			return
		} else {
			text[place] = 'a'
		}
	}
}

type generator struct {
	lastString []byte
}

func (g *generator) Next() []byte {
	incrementString(g.lastString)
	return g.lastString
}

func (g *generator) Reset() {
	g.lastString = []byte("aaaaaaaaaaa")
}

func init() {
	permuter.RegisterAlgorithm("ascii", func() permuter.PermuterAlgorithm {
		return &generator{[]byte("aaaaaaaaaaa")}
	})
}
