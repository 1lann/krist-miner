// +build !noasm

package main

import "unsafe"

//go:noescape
func blockSsse(h []uint32, message []uint8, reserved0, reserved1, reserved2, reserved3 uint64)

func mineSSSE3() {
	instanceID := generateInstanceID()

	var full = make([]byte, 64)

	copy(full[:fullHeaderSize], append([]byte(address+lastBlock), instanceID...))

	if len(address+lastBlock)+len(instanceID) != fullHeaderSize {
		panic("miner: incorrect header size. report this to 1lann.")
	}

	threadBlock := lastBlock
	noncePtr := (*uint64)(unsafe.Pointer(&full[fullHeaderSize]))

	if full[41] != 0 {
		panic("overwrite! report this to 1lann.")
	}

	full[41] = 128
	full[62] = 1
	full[63] = 72

	h := []uint32{0, 0, 0, 0, 0, 0, 0, 0}

	for {
		for i := 0; i < 1000000; i++ {
			(*noncePtr)++

			h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7] =
				init0, init1, init2, init3, init4, init5, init6, init7
			blockSsse(h, full, 0, 0, 0, 0)

			if h[0] > 16 {
				continue
			}

			if (h[0]<<16)+(h[1]>>16) > maxWork {
				continue
			}

			submitResult(lastBlock, string(full[22:41]))
		}

		hashesThisPeriod++

		if threadBlock != lastBlock {
			threadBlock = lastBlock
			copy(full[:fullHeaderSize], append([]byte(address+lastBlock), instanceID...))

			if len(address+lastBlock)+len(instanceID) != fullHeaderSize {
				panic("miner: incorrect header size. report this to 1lann.")
			}
		}
	}
}
