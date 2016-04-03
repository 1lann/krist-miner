package main

import (
	"github.com/1lann/krist-miner/permuter"
	_ "github.com/1lann/krist-miner/permuter/ascii"
	// _ "github.com/1lann/krist-miner/permuter/number"
	// _ "github.com/1lann/krist-miner/permuter/urandom"
	"github.com/1lann/krist-miner/sha2"
	// _ "github.com/1lann/krist-miner/sha2/asm"
	"fmt"
	_ "github.com/1lann/krist-miner/sha2/go"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"runtime/pprof"
)

var maxWork int64
var lastBlock string

const version = "0.3"

var address string
var hashesThisPeriod int64
var newLastBlock = make(chan bool)

func main() {
	var numProcs int = runtime.NumCPU()

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
	resp, err := http.Get("http://krist.ceriat.net/?lastblock")
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
	resp, err := http.Get("http://krist.ceriat.net/?getwork")
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

		previousWork := maxWork
		maxWork, err = strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			log.Println("failed to convert work to int:", data)
			return
		}

		if maxWork != previousWork {
			log.Println("work updated to:", maxWork)
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

	resp, err := http.Get("http://krist.ceriat.net/?submitblock&address=" +
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

	resp, err = http.Get("http://krist.ceriat.net/?getbalance=" + address)
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

var alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func mine(numProcs int) {
	runtime.GOMAXPROCS(numProcs)

	updateWork()
	updateLastBlock()

	debug.SetGCPercent(-1)

	log.Println("using", numProcs, "processes")

	for proc := 0; proc < numProcs; proc++ {
		go func(proc int) {
			procId := string(alphabet[proc])
			header := []byte(address + lastBlock + procId)
			threadBlock := lastBlock
			headerLen := len(header)

			sha2algo := sha2.NewAlgorithmInstance("go")
			permalgo := permuter.NewAlgorithmInstance("ascii")

			var nonce []byte

			for {
				for i := 0; i < 100000; i++ {
					nonce = permalgo.Next()
					header = append(header[:headerLen], nonce...)
					if sha2algo.Sum256NumberCmp(header, maxWork) {
						submitResult(lastBlock, procId+string(nonce))
					}
				}

				hashesThisPeriod++

				if threadBlock != lastBlock {
					threadBlock = lastBlock
					header = []byte(address + lastBlock + procId)
				}
			}
		}(proc)
	}

	log.Println("mining for address " + address + "...")

	previousTime := time.Now()
	for {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second * 5)

			log.Printf("%.3f MH/s\n", float64(hashesThisPeriod)/
				(time.Now().Sub(previousTime).Seconds()*10.0))

			previousTime = time.Now()
			hashesThisPeriod = 0

			updateWork()
			updateLastBlock()
		}

		debug.SetGCPercent(10)
		debug.SetGCPercent(-1)
	}

	f, err := os.Create("mem.out")
	if err != nil {
		log.Fatal(err)
	}
	pprof.WriteHeapProfile(f)
	f.Close()
}
