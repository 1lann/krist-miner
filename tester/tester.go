package main

import (
	"fmt"

	"github.com/1lann/krist-miner/sha2"
	// _ "github.com/1lann/krist-miner/sha2/asm"
	_ "github.com/1lann/krist-miner/sha2/simd"
)

func main() {
	sha2algo := sha2.NewAlgorithmInstance("simd")
	result := sha2algo.Sum256Number([]byte("a"))
	if result != 222752054364699 {
		fmt.Println("-- Fail --")
		fmt.Println("Expected:", 222752054364699)
		fmt.Println("Got:     ", result)
	} else {
		fmt.Println("-- Success --")
	}
}
