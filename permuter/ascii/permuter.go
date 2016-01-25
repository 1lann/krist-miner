package permuterascii

import (
	"github.com/1lann/krist-miner/permuter"
)

func incrementString(text []byte) []byte {
	for place := len(text) - 1; place >= 0; place-- {
		if text[place] < 'z' {
			text[place] = text[place] + 1
			return text
		} else {
			text[place] = 'a'
		}
	}

	text = append([]byte{'a'}, text...)

	return text
}

type generator struct {
	lastString []byte
}

func (g *generator) Next() []byte {
	g.lastString = incrementString(g.lastString)
	return g.lastString
}

func (g *generator) Reset() {
	g.lastString = []byte("aaaaaaaaaaaaaaaa")
}

func init() {
	permuter.RegisterAlgorithm("ascii", func() permuter.PermuterAlgorithm {
		return &generator{[]byte("aaaaaaaaaaaaaaaa")}
	})
}
