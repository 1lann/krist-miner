// +build !noasm

package main

import (
	"runtime"
	"time"
)

//go:noescape
func blockArm(h []uint32, message []uint8)

func mineAVX2(proc int)  {}
func mineAVX(proc int)   {}
func mineSSSE3(proc int) {}

func mineARM(proc int) {
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

	h := []uint32{0, 0, 0, 0, 0, 0, 0, 0}

	for {
		start := time.Now()
		for i := 0; i < 5000000; i++ {
			incrementNonce(na)

			h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7] =
				init0, init1, init2, init3, init4, init5, init6, init7
			blockArm(h, full)

			if h[0] > 1 {
				continue
			}

			if (h[0]<<16)|(h[1]>>16) > maxWork {
				continue
			}

			submitResult(lastBlock, string(full[22:41]))
		}

		workerSpeeds[proc] = time.Since(start)

		runtime.Gosched()

		if threadBlock != lastBlock {
			threadBlock = lastBlock
			copy(full[:fullHeaderSize], []byte(address+lastBlock+instanceID))

			if len(address+lastBlock+instanceID) != fullHeaderSize {
				panic("miner: incorrect header size. report this to 1lann.")
			}
		}
	}
}
