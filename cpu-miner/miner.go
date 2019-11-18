package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/1lann/krist-miner/cpu-miner/cpuid"
)

var maxWork uint32
var lastBlock string

const (
	version        = "2.1"
	fullHeaderSize = 30
	endpoint       = "https://krist.ceriat.net"
)

var address string
var privateKey string
var namedAddr string
var workerSpeeds []time.Duration
var newLastBlock = make(chan bool)
var client = new(http.Client)

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

var tradAddrRegex = regexp.MustCompile(`^([a-f0-9]{10}|k[a-z0-9]{9})$`)
var namedAddrRegex = regexp.MustCompile(`^((?:[a-z0-9-_]{1,32}@)?([a-z0-9]{1,64})\.kst)$`)

func main() {
	numProcs := runtime.NumCPU()

	fmt.Println("krist-miner v" + version +
		" by 1lann (github.com/1lann/krist-miner)")

	if len(os.Args) == 1 {
		fmt.Println("Usage: " + os.Args[0] + " address [num processes]")
		fmt.Println("By default, the number of processes used will be the\n" +
			"number of CPU cores available on this system (" +
			strconv.Itoa(numProcs) + ").")
		fmt.Println("An address can be a v1, v2 or named address (like me@name.kst)")

		optimisations := ""
		if cpuid.SHA && cpuid.SSSE3 && cpuid.SSE41 {
			optimisations += " SHA"
		}
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
			fmt.Println("No optimisations are supported on your CPU")
		} else {
			fmt.Println("Optimisations supported:" + optimisations)
		}

		return
	}

	if tradAddrRegex.Match([]byte(os.Args[1])) {
		address = os.Args[1]
	} else if namedAddrRegex.Match([]byte(os.Args[1])) {
		namedAddr = os.Args[1]
		if !setupNamedAddress() {
			return
		}
	} else {
		fmt.Println("Invalid address, check that you entered your address correctly")
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

func mine(numProcs int) {
	runtime.GOMAXPROCS(numProcs)
	workerSpeeds = make([]time.Duration, numProcs)

	updateWork()
	updateLastBlock()

	debug.SetGCPercent(-1)

	log.Println("using", numProcs, "processes")

	if os.Getenv("MINER_COMPAT") == "1" {
		cpuid.AVX2 = false
		cpuid.AVX = false
		cpuid.SSSE3 = false
		cpuid.ArmSha = false
		cpuid.SHA = false
		log.Println("using compatibility mode optimisations")
	} else {
		switch {
		case cpuid.SHA && cpuid.SSSE3 && cpuid.SSE41:
			log.Println("using SHA optimisations")
		case cpuid.AVX2:
			log.Println("using AVX2 optimisations")
		case cpuid.AVX:
			log.Println("using AVX optimisations")
		case cpuid.SSSE3:
			log.Println("using SSSE3 optimisations")
		case cpuid.ArmSha:
			log.Println("using ARMSHA optimisations")
		default:
			log.Println("using no optimisations")
		}
	}

	for proc := 0; proc < numProcs; proc++ {
		// decide on miner and execute
		switch {
		case cpuid.SHA && cpuid.SSSE3 && cpuid.SSE41:
			go mineSHA(proc)
		case cpuid.AVX2:
			go mineAVX2(proc)
		case cpuid.AVX:
			go mineAVX(proc)
		case cpuid.SSSE3:
			go mineSSSE3(proc)
		case cpuid.ArmSha:
			go mineARM(proc)
		default:
			go mineSlow(proc)
		}
	}

	log.Println("mining for address " + address + "...")

	for {
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second * 5)

			hashes := 0.0
			for _, speed := range workerSpeeds {
				hashes += 5 / speed.Seconds()
			}

			log.Printf("%.2f MH/s\n", hashes)

			updateWork()
			updateLastBlock()
		}

		debug.FreeOSMemory()
	}
}

func incrementNonce(na []byte) {
	for place := 10; place >= 0; place-- {
		if na[place] < 'z' {
			na[place] = na[place] + 1
			return
		}
		na[place] = 'A'
	}
}

func generateInstanceID() string {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatal("crypto/rand not supported on this system: ", err)
	}

	return hex.EncodeToString(bytes)
}
