package asm

// #cgo CFLAGS: -O3
// #include "sha2.h"
import "C"

import (
	"github.com/1lann/krist-miner/sha2"
)

type generator struct{}

func (g *generator) Sum256Number(data []byte) int64 {
	var result [32]byte

	C.cgminer_sha256_hash((*C.uchar)(&data[0]), C.uint(len(data)),
		(*C.uchar)(&result[0]))

	return int64(result[31]) + int64(result[30])<<(8*1) +
		int64(result[29])<<(8*2) + int64(result[28])<<(8*3) +
		int64(result[27])<<(8*4) + int64(result[26])<<(8*5)
}

func init() {
	sha2.RegisterAlgorithm("cgminer", func() sha2.SumNumberAlgorithm {
		return &generator{}
	})
}
