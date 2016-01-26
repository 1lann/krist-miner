package main

import (
	"fmt"
	"github.com/1lann/krist-miner/sha2"
	_ "github.com/1lann/krist-miner/sha2/asm"
	_ "github.com/1lann/krist-miner/sha2/go"
)

func main() {
	sha2algo := sha2.NewAlgorithmInstance("go")
	result := sha2algo.Sum256Number([]byte("a"))
	if result != 131416065984699 {
		fmt.Println("-- Fail --")
		fmt.Println("Expected:", 131416065984699)
		fmt.Println("Got:     ", result)
	} else {
		fmt.Println("-- Success --")
	}
}
