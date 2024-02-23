package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func getLastID() int {
	filePath := "blockchain.json"

	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// If the file doesn't exist, create it with an empty array
		err = ioutil.WriteFile(filePath, []byte("[]"), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Read JSON data from the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var blocks []Block

	// Unmarshal JSON data into a slice of blocks
	err = json.Unmarshal(data, &blocks)
	if err != nil {
		log.Fatal(err)
	}

	lastID := 0

	if len(blocks) == 0 {
		fmt.Println("Array is empty")
		lastID = 0
	} else {
		// Get the last block's hash
		lastBlock := blocks[len(blocks)-1]
		lastID = lastBlock.Index
	}

	return lastID
}
