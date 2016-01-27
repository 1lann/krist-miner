package asm

// #include "sha256.h"
import "C"

import (
	"github.com/1lann/krist-miner/sha2"
)

type generator struct{}

func (g *generator) Sum256Number(data []byte) int64 {
	var result [8]uint32

	C.asm_sha256_hash((*C.uint8_t)(&data[0]),
		C.uint32_t(len(data)),
		(*C.uint32_t)(&result[0]))

	return int64(result[7]) + ((int64(result[6]) & 0x0000ffff) << 0x20)
}

func (g *generator) Sum256NumberCmp(data []byte, work int64) bool {
	var result [8]uint32

	C.asm_sha256_hash((*C.uint8_t)(&data[0]),
		C.uint32_t(len(data)),
		(*C.uint32_t)(&result[0]))

	value := int64(result[7])
	if value > work {
		return false
	}

	value += ((int64(result[6]) & 0x0000ffff) << 0x20)
	if value > work {
		return false
	}

	return true
}

func init() {
	sha2.RegisterAlgorithm("asm", func() sha2.SumNumberAlgorithm {
		return &generator{}
	})
}
