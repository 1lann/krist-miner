package main

import (
	"time"

	sha256 "github.com/1lann/sha256-simd"
)

func mineSlow(proc int) {
	instanceID := generateInstanceID()

	var full = make([]byte, 64)

	copy(full[:fullHeaderSize], []byte(address+lastBlock+instanceID))

	if len(address+lastBlock+instanceID) != fullHeaderSize {
		panic("miner: incorrect header size. report this to 1lann.")
	}

	threadBlock := lastBlock
	na := full[fullHeaderSize : fullHeaderSize+11]
	na[0], na[1], na[2], na[3], na[4], na[5], na[6], na[7], na[8], na[9], na[10] =
		'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A'

	if full[41] != 0 || full[40] == 0 {
		panic("overwrite! report this to 1lann.")
	}

	full[41] = 128
	full[62] = 1
	full[63] = 72

	for {
		start := time.Now()
		for i := 0; i < 5000000; i++ {
			incrementNonce(na)
			if sha256.SumCmp256(full, maxWork) {
				submitResult(lastBlock, string(full[22:41]))
			}
		}

		workerSpeeds[proc] = time.Since(start)

		if threadBlock != lastBlock {
			threadBlock = lastBlock
			copy(full[:fullHeaderSize], []byte(address+lastBlock+instanceID))

			if len(address+lastBlock+instanceID) != fullHeaderSize {
				panic("miner: incorrect header size. report this to 1lann.")
			}
		}
	}
}
