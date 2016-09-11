package openssl

// #cgo CFLAGS: -I/usr/local/Cellar/openssl/1.0.2e_1/include
// #cgo LDFLAGS: -L/usr/local/Cellar/openssl/1.0.2e_1/lib -lssl -lcrypto
// #include <openssl/sha.h>
import "C"

import (
	"github.com/1lann/krist-miner/sha2"
	"unsafe"
)

type generator struct{}

func sum256(data []byte) []byte {
	var hash = make([]byte, 65)
	var sha256 C.SHA256_CTX
	C.SHA256_Init(&sha256)
	C.SHA256_Update(&sha256, unsafe.Pointer(&data[0]), C.size_t(len(data)))
	C.SHA256_Final((*C.uchar)(&hash[0]), &sha256)
	return hash[:64]
}

func (g *generator) Sum256Number(data []byte) int64 {
	result := sum256(data)
	// Turn first 6 bytes to int64
	return int64(result[5]) + int64(result[4])<<(8*1) +
		int64(result[3])<<(8*2) + int64(result[2])<<(8*3) +
		int64(result[1])<<(8*4) + int64(result[0])<<(8*5)
}

func (g *generator) Sum256NumberCmp(data []byte, work int64) bool {
	result := sum256(data)

	value := int64(result[5])
	if value > work {
		return false
	}

	value += int64(result[4]) << (8 * 1)
	if value > work {
		return false
	}

	value += int64(result[3]) << (8 * 2)
	if value > work {
		return false
	}

	value += int64(result[2]) << (8 * 3)
	if value > work {
		return false
	}

	value += int64(result[1]) << (8 * 4)
	if value > work {
		return false
	}

	value += int64(result[0]) << (8 * 5)
	if value > work {
		return false
	}

	return true
}

func init() {
	sha2.RegisterAlgorithm("openssl", func() sha2.SumNumberAlgorithm {
		return &generator{}
	})
}
