package permuter

type PermuterAlgorithm interface {
	Next() []byte
	Reset()
}

type namedAlgorithm struct {
	name    string
	factory func() PermuterAlgorithm
}

var algorithms []namedAlgorithm

func RegisterAlgorithm(name string, factory func() PermuterAlgorithm) {
	algorithms = append(algorithms, namedAlgorithm{name, factory})
}

func ListAlgorithms() []string {
	var results = make([]string, len(algorithms))
	for i, algorithm := range algorithms {
		results[i] = algorithm.name
	}

	return results
}

func NewAlgorithmInstance(name string) PermuterAlgorithm {
	for _, algorithm := range algorithms {
		if algorithm.name == name {
			return algorithm.factory()
		}
	}

	panic("permuter: no such algorithm")
}
