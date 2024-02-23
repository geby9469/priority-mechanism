package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

var jsonFileMutex2 sync.Mutex

func readBlockchainFile(filePath string) ([]Block, error) {
	var data []Block
	jsonFileMutex2.Lock()
	defer jsonFileMutex2.Unlock()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist, return empty data
		return data, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func writeBlockchainFile(filePath string, data []Block) error {

	jsonFileMutex2.Lock()
	defer jsonFileMutex2.Unlock()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, jsonData, 0644)
}
