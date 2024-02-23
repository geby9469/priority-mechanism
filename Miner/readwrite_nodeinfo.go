package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

var jsonFileMutex sync.Mutex

func readJSONFile(filePath string) ([]ClientInfo, error) {
	var data []ClientInfo

	jsonFileMutex.Lock()
	defer jsonFileMutex.Unlock()

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

func writeJSONFile(filePath string, data []ClientInfo) error {

	jsonFileMutex.Lock()
	defer jsonFileMutex.Unlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, jsonData, 0644)
}
