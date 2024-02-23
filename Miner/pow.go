package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// minTxs           = 5
	difficulty       = 4      // the difficulty can be changed after 2016 blocks like in bitcoin protocol
	difficultyTarget = "0000" // the target bits or the hash prefix
)

func increaseNonce(prevHash string) int {
	nonce := 0
	for {
		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%s%d", prevHash, nonce)))
		hashSum := hex.EncodeToString(hash.Sum(nil))
		prefix := ""
		for i := 0; i < difficulty; i++ {
			prefix += "0"
		}
		if hashSum[:difficulty] == prefix {
			return nonce
		}
		nonce++
	}
}

func getDifficulty() int {
	return difficulty
}

func getTarget() string {
	return difficultyTarget
}

func proofOfWork(prevHash string) string {
	nonce := 0
	now := time.Now()
	timestamp := now.Unix()
	for {
		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%s%d", prevHash, nonce, timestamp))) // given the same input, the output is deterministic, so make sure there is a timestamp
		hashSum := hex.EncodeToString(hash.Sum(nil))
		prefix := ""
		for i := 0; i < difficulty; i++ {
			prefix += "0"
		}
		if hashSum[:difficulty] == prefix {
			return hashSum
		}
		// fmt.Println("Finding block hash via PoW", hashSum)
		fmt.Printf("\r%s", hashSum)
		nonce++
	}
}
