package main

import (
	"fmt"

	"github.com/1lann/krist-miner/deprecated/sha2"
	// _ "github.com/1lann/krist-miner/sha2/asm"
	_ "github.com/1lann/krist-miner/deprecated/sha2/simd"
)

func main() {
	sha2algo := sha2.NewAlgorithmInstance("simd")
	result := sha2algo.Sum256Number([]byte("uptotwentytwocrdctuptotwentytwochracterer"))
	if result != 175960619416471 {
		fmt.Println("-- Fail --")
		fmt.Println("Expected:", 175960619416471)
		fmt.Println("Got:     ", result)
	} else {
		fmt.Println("-- Success --")
	}
}
