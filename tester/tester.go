package main

import (
	"fmt"

	"github.com/1lann/krist-miner/deprecated/sha2"
	_ "github.com/1lann/krist-miner/deprecated/sha2/go"
	_ "github.com/1lann/krist-miner/deprecated/sha2/simd"
)

func main() {
	sha2algo := sha2.NewAlgorithmInstance("go")
	result := sha2algo.Sum256Number([]byte("uptotwentytwocrdctuptotwentytwochractearr"))
	if result != 189624896564436 {
		fmt.Println("-- Fail --")
		fmt.Println("Expected:", 189624896564436)
		fmt.Println("Got:     ", result)
	} else {
		fmt.Println("-- Success --")
	}
}
