package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/1lann/krist-miner/cpu-miner/cpuid"
)

var maxWork uint32
var lastBlock string

const version = "2.0"
const fullHeaderSize = 30

var address string
var hashesThisPeriod int64
var newLastBlock = make(chan bool)

const (
	init0 = 0x6A09E667
	init1 = 0xBB67AE85
	init2 = 0x3C6EF372
	init3 = 0xA54FF53A
	init4 = 0x510E527F
	init5 = 0x9B05688C
	init6 = 0x1F83D9AB
	init7 = 0x5BE0CD19
)

func main() {
	numProcs := runtime.NumCPU()

	if len(os.Args) == 1 {
		fmt.Println("krist-miner v" + version +
			" by 1lann (github.com/1lann/krist-miner)")
		fmt.Println("Usage: " + os.Args[0] + " address [num processes]")
		fmt.Println("By default, the number of processes used will be the\n" +
			"number of CPU cores available on this system (" +
			strconv.Itoa(numProcs) + ").")

		optimisations := ""
		if cpuid.AVX2 {
			optimisations += " AVX2"
		}
		if cpuid.AVX {
			optimisations += " AVX"
		}
		if cpuid.SSSE3 {
			optimisations += " SSSE3"
		}
		if cpuid.ArmSha {
			optimisations += " ARMSHA"
		}

		if optimisations == "" {
			fmt.Println("No optimisations are supported on your CPU! This version will not work on your CPU.")
			fmt.Println("Either use v1.1 or get a new CPU.")
		} else {
			fmt.Println("Optimisations supported:" + optimisations)
		}

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

	values := url.Values{}

	values.Set("address", address)
	values.Set("nonce", nonce)

	resp, err := http.Get("https://krist.ceriat.net/?submitblock&" + values.Encode())
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
		log.Fatal("crypto/rand not supported on this system: ", err)
	}

	return hex.EncodeToString(bytes)
}

func mine(numProcs int) {
	runtime.GOMAXPROCS(numProcs)

	updateWork()
	updateLastBlock()

	debug.SetGCPercent(-1)

	log.Println("using", numProcs, "processes")

	switch {
	case cpuid.AVX2:
		log.Println("using AVX2 optimisations")
	case cpuid.AVX:
		log.Println("using AVX optimisations")
	case cpuid.SSSE3:
		log.Println("using SSSE3 optimisations")
	case cpuid.ArmSha:
		log.Println("using ARMSHA optimisations")
	default:
		log.Println("your CPU isn't supported for optimised mining")
		log.Println("please use v1.1 or get a new CPU.")
		os.Exit(1)
	}

	for proc := 0; proc < numProcs; proc++ {
		// decide on miner and execute
		switch {
		case cpuid.AVX2:
			go mineAVX2()
		case cpuid.AVX:
			go mineAVX()
		case cpuid.SSSE3:
			go mineSSSE3()
		case cpuid.ArmSha:
			go mineARM()
		}
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

func incrementNonce(na []byte) {
	for place := 10; place >= 0; place-- {
		if na[place] < 'z' {
			na[place] = na[place] + 1
			return
		} else {
			na[place] = 'A'
		}
	}
}
