package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

func setupNamedAddress() bool {
	results := namedAddrRegex.FindAllStringSubmatch(namedAddr, -1)

	var result struct {
		OK bool `json:"ok"`
	}
	err := makeGet(endpoint+"/names/"+results[0][2], &result)
	if err != nil {
		log.Println("failed to check if name exists:", err)
		return false
	}

	if !result.OK {
		log.Println("the name \"" + results[0][2] + "\" does not exist")
		log.Println("check that you entered your address correctly")
		return false
	}

	log.Println("mined krist will be relayed to", namedAddr)
	rawKey := make([]byte, 8)
	_, err = rand.Read(rawKey)
	if err != nil {
		log.Println("failed to generate relay private key:", err)
		log.Println("try mining directly to an address instead")
		return false
	}

	privateKey = hex.EncodeToString(rawKey)
	err = ioutil.WriteFile("krist_key.txt",
		[]byte(privateKey+"\n"), 0600)
	if err != nil {
		log.Println("failed to write to \"krist_key.txt\" to store relay private key:", err)
		log.Println("try mining directly to an address instead")
		return false
	}

	log.Println("relay private key generated and stored to \"krist_key.txt\"")

	var addrResult struct {
		OK      bool   `json:"ok"`
		Address string `json:"address"`
	}
	err = makePost(endpoint+"/v2", url.Values{
		"privatekey": []string{privateKey},
	}, &addrResult)
	if err != nil {
		log.Println("failed to get address from private key:", err)
		return false
	}

	if !addrResult.OK {
		log.Println("server responded with not OK on get address from private key")
		return false
	}

	address = addrResult.Address
	log.Println("the relay address is", address)

	return true
}

func makePost(url string, data url.Values, dst interface{}) error {
	req, _ := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "github.com/1lann/krist-miner v"+version)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dst)
}

func makeGet(url string, dst interface{}) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "github.com/1lann/krist-miner v"+version)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dst)
}

func updateLastBlock() {
	var result struct {
		OK    bool `json:"ok"`
		Block struct {
			ShortHash string `json:"short_hash"`
		} `json:"block"`
	}
	err := makeGet(endpoint+"/blocks/last", &result)
	if err != nil {
		log.Println("failed to update last block:", err)
		return
	}

	if !result.OK {
		log.Println("server responded with not OK on get last block")
		return
	}

	previousBlock := lastBlock
	lastBlock = result.Block.ShortHash

	if previousBlock != lastBlock {
		log.Println("last block updated to", lastBlock)

		select {
		case newLastBlock <- true:
		default:
		}
	}
}

func updateWork() {
	var result struct {
		OK   bool   `json:"ok"`
		Work uint32 `json:"work"`
	}

	err := makeGet(endpoint+"/work", &result)
	if err != nil {
		log.Println("failed to update work:", err)
		return
	}

	if !result.OK {
		log.Println("server responded with not OK on get work")
		return
	}

	previousWork := maxWork
	maxWork = result.Work

	if maxWork != previousWork {
		log.Println("work updated to", maxWork)
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

	log.Println("submitting solved block", blockUsed, "with nonce", nonce)

	recentlySubmittedBlocks[0] = blockUsed

	var result struct {
		OK      bool   `json:"ok"`
		Success bool   `json:"success"`
		Work    uint32 `json:"work"`
		Address struct {
			Balance int `json:"balance"`
		} `json:"address"`
		Block struct {
			ShortHash string `json:"short_hash"`
		} `json:"block"`
	}

	err := makePost(endpoint+"/submit", url.Values{
		"address": []string{address},
		"nonce":   []string{nonce},
	}, &result)
	if err != nil {
		log.Println("failed to submit block:", err)
		return
	}

	if !result.OK || !result.Success {
		log.Println("submission was unsuccessful :(")
		return
	}

	log.Println("submitted successfully!")

	maxWork = result.Work
	log.Println("work updated to", maxWork)
	lastBlock = result.Block.ShortHash
	log.Println("last block updated to", lastBlock)

	select {
	case newLastBlock <- true:
	default:
	}

	if namedAddr != "" {
		var transResult struct {
			OK bool `json:"ok"`
		}
		err := makePost(endpoint+"/transactions", url.Values{
			"privatekey": []string{privateKey},
			"to":         []string{namedAddr},
			"amount":     []string{strconv.Itoa(result.Address.Balance)},
		}, &transResult)
		if err != nil {
			log.Println("failed to relay to named address, please try to relay it manually:", err)
		}

		if !transResult.OK {
			log.Println("failed to relay to named address! please try to relay it manually")
		}
	}
}
