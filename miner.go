package main

import (
	"github.com/1lann/krist-miner/permuter"
	_ "github.com/1lann/krist-miner/permuter/ascii"
	// _ "github.com/1lann/krist-miner/permuter/number"
	// _ "github.com/1lann/krist-miner/permuter/urandom"
	"github.com/1lann/krist-miner/sha2"
	// _ "github.com/1lann/krist-miner/sha2/asm"
	_ "github.com/1lann/krist-miner/sha2/go"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"runtime/debug"
)

var maxWork int64
var lastBlock string

const address = "k3be4p30lb"

var hashesThisSecond int64

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

		previousBlcok := lastBlock
		lastBlock = string(data)

		if previousBlcok != lastBlock {
			log.Println("last block updated to:", lastBlock)
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

func main() {
	updateWork()
	updateLastBlock()

	debug.SetGCPercent(100)

	var result = make(chan bool)

	for thread := 0; thread < 7; thread++ {
		go func(thread int) {
			header := []byte(address + lastBlock + strconv.Itoa(thread))

			sha2algo := sha2.NewAlgorithmInstance("go")
			permalgo := permuter.NewAlgorithmInstance("ascii")

			var nonce []byte

			for {
				for i := 0; i < 100000; i++ {
					nonce = permalgo.Next()
					if sha2algo.Sum256NumberCmp(append(header, nonce...),
						maxWork) {
						log.Println("Found result!")
						log.Println(nonce)
						result <- true
						return
					}
				}

				hashesThisSecond++
			}
		}(thread)
	}

	go func() {
		for {
			time.Sleep(time.Second * 5)
			log.Println(float64(hashesThisSecond)/50.0, "MH/s")
			hashesThisSecond = 0
		}

		result <- true
	}()

	<-result
}
