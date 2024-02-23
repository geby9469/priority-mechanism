package main

import (
	"sync"
	"time"
)

type ClientInfo struct {
	IP   string `json:"nodeip"`
	Port string `json:"port"`
}

type Message struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type Transaction struct {
	TransactionID int `json:"transactionid"`
	RemainingTime int `json:"remainingtime"`
	Threshold     int `json:"threshold"`
	// Sender         string  `json:"sender"`
	// Receiver       string  `json:"receiver"`
	// Amount         float64 `json:"amount"`
	// PriorityWeight float64 `json:"priority"`
	// DeadlineTime   string
}

type Block struct {
	Index            int           `json:"index"`
	Timestamp        time.Time     `json:"timestamp"`
	PreviousHash     string        `json:"previousHash"`
	Hash             string        `json:"hash"`
	Nonce            int           `json:"nonce"`
	Difficulty       int           `json:"difficulty"`
	DifficultyTarget string        `json:"difficultyTarget"`
	TxCount          int           `json:"txCount"`
	Transactions     []Transaction `json:"transactions"`
}

type Blockchain struct {
	Chain   []Block
	Mutex   sync.Mutex
	Mempool []Transaction
}
