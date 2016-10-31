package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/1lann/sha256-simd"
)

var maxWork uint32
var lastBlock string

const version = "1.1"

var address string
var hashesThisPeriod int64
var newLastBlock = make(chan bool)

func main() {
	numProcs := runtime.NumCPU()

	if len(os.Args) == 1 {
		fmt.Println("krist-miner v" + version +
			" by 1lann (github.com/1lann/krist-miner)")
		fmt.Println("Usage: " + os.Args[0] + " address [num processes]")
		fmt.Println("By default, the number of processes used will be the\n" +
			"number of CPU cores available on this system (" +
			strconv.Itoa(numProcs) + ").")
		return
	}

	address = os.Args[1]
	if len(address) != 10 {
		fmt.Println("Invalid address specified.")
		return
	}

	if len(os.Args) >= 3 {
		readProcs, err := strconv.Atoi(os.Args[2])
		if err != nil || readProcs < 1 {
			fmt.Println("Invalid number, defaulting to using " +
				strconv.Itoa(numProcs) + " processes.")
		} else {
			numProcs = readProcs
		}
	}

	mine(numProcs)
}

func updateLastBlock() {
	resp, err := http.Get("https://krist.ceriat.net/?lastblock")
	if err != nil {
		log.Println("failed to update last block:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("fail to read last block response:", err)
			return
		}

		previousBlock := lastBlock
		lastBlock = string(data)

		if previousBlock != lastBlock {
			log.Println("last block updated to:", lastBlock)

			select {
			case newLastBlock <- true:
			default:
			}
		}
	} else {
		log.Println("failed to update last block:", resp.Status)
	}
}

func updateWork() {
	resp, err := http.Get("https://krist.ceriat.net/?getwork")
	if err != nil {
		log.Println("failed to update work:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("failed to read work response:", err)
			return
		}

		previousWork := int64(maxWork)
		newWork, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			log.Println("failed to convert work to int:", data)
			return
		}

		if newWork != previousWork {
			maxWork = uint32(newWork)
			log.Println("work updated to:", newWork)
		}
	} else {
		log.Println("failed to update work:", resp.Status)
		return
	}
}

var recentlySubmittedBlocks [5]string
var submissionLock = &sync.Mutex{}

func submitResult(blockUsed string, nonce string) {
	submissionLock.Lock()
	defer submissionLock.Unlock()

	for i := 0; i < 5; i++ {
		if recentlySubmittedBlocks[i] == blockUsed {
			if lastBlock != blockUsed {
				return
			}

			<-newLastBlock
			return
		}
	}

	for i := 1; i < 5; i++ {
		recentlySubmittedBlocks[i] = recentlySubmittedBlocks[i-1]
	}

	log.Println("submitting solved block", blockUsed, "with nonce:", nonce)

	recentlySubmittedBlocks[0] = blockUsed

	resp, err := http.Get("https://krist.ceriat.net/?submitblock&address=" +
		address + "&nonce=" + nonce)
	if err != nil {
		log.Println("failed to submit block:", err)
		return
	}

	if resp.StatusCode != 200 {
		log.Println("failed to submit block:", resp.Status)
	} else {
		log.Println("successfully submitted")
	}

	resp.Body.Close()

	resp, err = http.Get("https://krist.ceriat.net/?getbalance=" + address)
	if err != nil {
		log.Println("failed to check balance:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("failed to check balance:", resp.Status)
	} else {
		balance, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("failed to read balance response:", err)
			return
		}

		log.Println("balance:", string(balance))
	}

	if blockUsed != lastBlock {
		return
	}

	<-newLastBlock
}

func generateInstanceID() string {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatal("cyrpto/rand not supported on this system: ", err)
	}

	return hex.EncodeToString(bytes)
}

var alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func mine(numProcs int) {
	runtime.GOMAXPROCS(numProcs)

	updateWork()
	updateLastBlock()

	debug.SetGCPercent(-1)

	log.Println("using", numProcs, "processes")

	for proc := 0; proc < numProcs; proc++ {
		go func(proc int) {
			instanceID := generateInstanceID()

			var full = make([]byte, 64)

			copy(full[:30], []byte(address+lastBlock+instanceID))

			if len(address+lastBlock+instanceID) != 30 {
				panic("miner: incorrect header size. report this to 1lann.")
			}

			threadBlock := lastBlock

			// first byte of nonce is [30]
			copy(full[30:], []byte("aaaaaaaaaaa"))

			if full[41] != 0 || full[40] == 0 {
				panic("overwrite! report this to 1lann.")
			}

			full[41] = 128
			full[62] = 1
			full[63] = 72

			for {
				for i := 0; i < 1000000; i++ {
					if sha256.SumCmp256(full, maxWork) {
						submitResult(lastBlock, string(full[22:41]))
					}
					incrementString(full)
				}

				hashesThisPeriod++

				if threadBlock != lastBlock {
					threadBlock = lastBlock
					copy(full[:30], []byte(address+lastBlock+instanceID))

					if len(address+lastBlock+instanceID) != 30 {
						panic("miner: incorrect header size. report this to 1lann.")
					}
				}
			}
		}(proc)
	}

	log.Println("mining for address " + address + "...")

	previousTime := time.Now()
	for {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second * 5)

			log.Printf("%.2f MH/s\n", float64(hashesThisPeriod)/
				time.Now().Sub(previousTime).Seconds())

			previousTime = time.Now()
			hashesThisPeriod = 0

			updateWork()
			updateLastBlock()
		}

		debug.SetGCPercent(10)
		debug.SetGCPercent(-1)
	}
}

func incrementString(text []byte) {
	for place := 40; place >= 30; place-- {
		if text[place] < 'z' {
			text[place] = text[place] + 1
			return
		} else {
			text[place] = 'a'
		}
	}
}
