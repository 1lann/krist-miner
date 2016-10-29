package sha2

type SumNumberAlgorithm interface {
	Sum256Number(data []byte) int64
	Sum256NumberCmp(data []byte, work int64) bool
}

type namedAlgorithm struct {
	name    string
	factory func() SumNumberAlgorithm
}

var algorithms []namedAlgorithm

func RegisterAlgorithm(name string, factory func() SumNumberAlgorithm) {
	algorithms = append(algorithms, namedAlgorithm{name, factory})
}

func ListAlgorithms() []string {
	var results = make([]string, len(algorithms))
	for i, algorithm := range algorithms {
		results[i] = algorithm.name
	}

	return results
}

func NewAlgorithmInstance(name string) SumNumberAlgorithm {
	for _, algorithm := range algorithms {
		if algorithm.name == name {
			return algorithm.factory()
		}
	}

	panic("sha2: no such algorithm")
}
