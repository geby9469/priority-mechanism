package main

import (
	"sync"
	"time"
)

var isBlockCycle = true                    // one block cycle
var IsPropagted = 0                        // propagation mutex
var epochTime = 0                          // Genesis blockchain 초기 블록 생성 시간
var blockGenerationTime time.Duration      // block generation time
var blockValidationTime time.Duration      // block validation time
var blockPropagationTime time.Duration     // block propagation time
var txInblockPropagationTime time.Duration // tx in a block validation time
var totalBlockTime time.Duration           // total block time
var totalBlockTimeInt int                  // total block time for calculation
var blockGenerationEpochTime int           // blockchain one block cycle time
var IsDeadlineSuccess bool                 // deadline guaranteed

var transactionPools = make(map[int]TransactionPool) // priority driven transaction pool

// for csv
var allOutputForCSV = make(map[int]OutputForCSV) // For csv
var TxCount string
var Prirotiy1TxCount string
var TxCountInBlock string
var StrblockGenerationTime string
var StrBlockValidationTime string
var StrTxInBlockValidationTime string
var TrashTransactionCount = 0
var transactionValidationCount = 0

type ClientInfo struct {
	IP           string `json:"nodeip"`
	Port         string `json:"port"`
	IsPropagated bool   `json:"ispropagated"`
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
}

type Block struct {
	Index            int               `json:"index"`
	Timestamp        time.Time         `json:"timestamp"`
	PreviousHash     string            `json:"previousHash"`
	Hash             string            `json:"hash"`
	Nonce            int               `json:"nonce"`
	Difficulty       int               `json:"difficulty"`
	DifficultyTarget string            `json:"difficultyTarget"`
	TxCount          int               `json:"txCount"`
	TransactionPool  []TransactionPool `json:"transactions"`
}

type Blockchain struct {
	Chain   []Block
	Mutex   sync.Mutex
	Mempool []Transaction
}

type TransactionPool struct {
	TransactionID int
	RemainingTime int
	Priority      int
	Threshold     int
}

type BlockTransaction struct {
	TransactionID int
	Priority      int
	Threshold     int
}

type OutputForCSV struct {
	TxCount                    string
	Prirotiy1TxCount           string
	TxCountInBlock             string
	StrblockGenerationTime     string
	StrBlockValidationTime     string
	BlockPropagationTime       string
	StrTxInBlockValidationTime string
	TotalBlockTime             string
	EpochTime                  string
	IsGuaranteedDeadline       string
	TrashTransactionCount      string
	TransactionValidationCount string
}
