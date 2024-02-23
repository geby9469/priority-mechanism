package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup
var jsonFilePath = "allnodes.json"

func writeResultCSV(trimStrBGT, trimStrBVT, trimStrBPT, trimStrTBT, trimStrTBVT string) {
	file, err := os.Create(csvFilePath)
	defer file.Close()
	if err != nil {
		log.Fatal("failed to open file", err)
	}
	w := csv.NewWriter(file)
	defer w.Flush()

	// IsDeadlineSuccess
	var tempAOFC OutputForCSV
	previousBlockID := getLastID()
	headers := []string{"TxCount", "Prirotiy1TxCount", "TxCountInBlock", "BlockGenerationTime", "BlockValidationTime", "BlockPropagationTime", "TxValidationTime", "TotalBlockTime", "EpochTime", "IsGuaranteedDeadline", "TrashTransactionCount", "TransactionValidationCount"}
	tempAOFC.TxCount = TxCount                                                     // 트랜잭션 풀에 있는 트랜잭션 수
	tempAOFC.Prirotiy1TxCount = Prirotiy1TxCount                                   // 트랜잭션 풀에 있는 prirotiy1 트랜잭션 수
	tempAOFC.TxCountInBlock = TxCountInBlock                                       // 블록에 있는 트랜잭션 수
	tempAOFC.StrblockGenerationTime = trimStrBGT                                   // 블록 생성 시간
	tempAOFC.StrBlockValidationTime = trimStrBVT                                   // 블록 검증 시간
	tempAOFC.BlockPropagationTime = trimStrBPT                                     // 블록 전파 시간
	tempAOFC.StrTxInBlockValidationTime = trimStrTBVT                              // 트랜잭션 검증 시간
	tempAOFC.TotalBlockTime = trimStrTBT                                           // 전체 블록 생성 시간
	tempAOFC.EpochTime = strconv.Itoa(blockGenerationEpochTime)                    // 블록 에폭 시간 (블록 타임아웃)
	tempAOFC.IsGuaranteedDeadline = strconv.FormatBool(IsDeadlineSuccess)          // 데드라인 보장
	tempAOFC.TrashTransactionCount = strconv.Itoa(TrashTransactionCount)           // 버려지는 트랜잭션 수
	tempAOFC.TransactionValidationCount = strconv.Itoa(transactionValidationCount) // 검증한 트랜잭션 수
	allOutputForCSV[previousBlockID] = tempAOFC

	w.Write(headers)
	for _, value := range allOutputForCSV {
		line := []string{value.TxCount, value.Prirotiy1TxCount, value.TxCountInBlock, value.StrblockGenerationTime, value.StrBlockValidationTime, value.BlockPropagationTime, value.StrTxInBlockValidationTime, value.TotalBlockTime, value.EpochTime, value.IsGuaranteedDeadline, value.TrashTransactionCount, value.TransactionValidationCount}
		err := w.Write(line)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Complete writting data of output in .csv file")
}

func alwaySendBlockchainAtAnyTime() {
	// fmt.Println("blockTransaction -->> ", &blockTransaction)

	// Read the JSON file content to get all nodes' info
	jsonContent, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Println("Error reading JSON file:", err)
		return
	}

	var allNodes []ClientInfo
	if err := json.Unmarshal(jsonContent, &allNodes); err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return
	}

	filePath := "blockchain.json"

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	var blocks []Block
	err = json.Unmarshal(data, &blocks)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	if len(blocks) == 0 {
		fmt.Println("Array is empty")
		return
	}

	// Send the message to all node IPs and ports
	// Concurrent programming - wait group
	wg.Add(len(allNodes))
	elapsedPropagationTime := time.Now()
	for _, node := range allNodes {
		go sendBlockchain(node.IP, node.Port, blocks)
		// fmt.Println("before isBlockCycle -->> ", isBlockCycle)
	}
	wg.Wait()
	blockPropagationTime = time.Since(elapsedPropagationTime)
	/************* total block time 계산 *************/
	// fmt.Println("Elapsed time to block generation time : ", blockGenerationTime)
	// fmt.Println("Elapsed time to block validation time : ", blockValidationTime)
	// fmt.Println("Elapsed Time to block propagation time : ", blockPropagationTime)
	totalBlockTime = blockGenerationTime + blockValidationTime + blockPropagationTime + txInblockPropagationTime
	// fmt.Println("Elapsed Time to total block time : ", totalBlockTime)

	/* CSV file에 데이터 넣기 전에 ms 단위로 변환하기 */
	// time -> string
	strBGT := blockGenerationTime.String()
	strBVT := blockValidationTime.String()
	strBPT := blockPropagationTime.String()
	strTBT := totalBlockTime.String()
	strTBVT := txInblockPropagationTime.String()

	// ms or s?
	msOrSBGT := strBGT[len(strBGT)-2:]
	msOrSBVT := strBVT[len(strBVT)-2:]
	msOrSBPT := strBPT[len(strBPT)-2:]
	msOrSTBT := strTBT[len(strTBT)-2:]
	msOrSTBVT := strTBVT[len(strTBVT)-2:]

	// 생성시간
	var trimStrBGT string
	if strings.Contains(msOrSBGT, "ms") {
		// 특정 문자열 제거
		trimStrBGT = strings.TrimRight(strBGT, "ms")
		msfBGT, err := strconv.ParseFloat(trimStrBGT, 64)
		if err != nil {
			fmt.Println("FAILED: MS convert string -> float about validation time")
		}
		trimStrBGT = strconv.FormatFloat(msfBGT, 'f', -1, 64)
	} else {
		removesBGT := strings.TrimRight(strBGT, "s")
		fBGT, err := strconv.ParseFloat(removesBGT, 64)
		if err != nil {
			fmt.Println("FAILED: convert string -> float about validation time")
		}
		trimStrBGT = strconv.FormatFloat(fBGT*1000, 'f', -1, 64)
	}

	// 검증시간
	var trimStrBVT string
	if strings.Contains(msOrSBVT, "ms") {
		// 특정 문자열 제거
		trimStrBVT = strings.TrimRight(strBVT, "ms")
		msfBVT, err := strconv.ParseFloat(trimStrBVT, 64)
		if err != nil {
			fmt.Println("FAILED: MS convert string -> float about validation time")
		}
		trimStrBVT = strconv.FormatFloat(msfBVT, 'f', -1, 64)
	} else {
		removesBVT := strings.TrimRight(strBVT, "s")
		fBVT, err := strconv.ParseFloat(removesBVT, 64)
		if err != nil {
			fmt.Println("FAILED: convert string -> float about validation time")
		}
		trimStrBVT = strconv.FormatFloat(fBVT*1000, 'f', -1, 64)
	}

	// 전파시간
	var trimStrBPT string
	if strings.Contains(msOrSBPT, "ms") {
		// 특정 문자열 제거
		trimStrBPT = strings.TrimRight(strBPT, "ms")
		msfBPT, err := strconv.ParseFloat(trimStrBPT, 64)
		if err != nil {
			fmt.Println("FAILED: MS convert string -> float about propagation time")
		}
		trimStrBPT = strconv.FormatFloat(msfBPT, 'f', -1, 64)
		// fmt.Println("trimStrBPT : ", trimStrBPT)
	} else {
		removesBPT := strings.TrimRight(strBPT, "s")
		fBPT, err := strconv.ParseFloat(removesBPT, 64)
		if err != nil {
			fmt.Println("FAILED: convert string -> float about propagation time")
		}
		trimStrBPT = strconv.FormatFloat(fBPT*1000, 'f', -1, 64)
	}

	// 전체시간
	var trimStrTBT string
	if strings.Contains(msOrSTBT, "ms") {
		// 특정 문자열 제거
		trimStrTBT = strings.TrimRight(strTBT, "ms")
		msfTBT, err := strconv.ParseFloat(trimStrTBT, 64)
		if err != nil {
			fmt.Println("FAILED: MS convert string -> float about total block time")
		}
		trimStrTBT = strconv.FormatFloat(msfTBT, 'f', -1, 64)
		// fmt.Println("trimStrTBT : ", trimStrTBT)
	} else {
		removesTBT := strings.TrimRight(strTBT, "s")
		fTBT, err := strconv.ParseFloat(removesTBT, 64)
		if err != nil {
			fmt.Println("FAILED: convert string -> float about total block time")
		}
		trimStrTBT = strconv.FormatFloat(fTBT*1000, 'f', -1, 64)
	}

	// 트랜잭션 검증 시간
	var trimStrTBVT string
	if strings.Contains(msOrSTBVT, "ms") {
		// 특정 문자열 제거
		trimStrTBVT = strings.TrimRight(strTBVT, "ms")
		msfTBVT, err := strconv.ParseFloat(trimStrTBVT, 64)
		if err != nil {
			fmt.Println("FAILED: MS convert string -> float about transaction validation time")
		}
		trimStrTBVT = strconv.FormatFloat(msfTBVT, 'f', -1, 64)
	} else {
		removesTBVT := strings.TrimRight(strTBVT, "s")
		fTBVT, err := strconv.ParseFloat(removesTBVT, 64)
		if err != nil {
			fmt.Println("FAILED: convert string -> float about transaction validation time")
		}
		trimStrTBVT = strconv.FormatFloat(fTBVT*1000, 'f', -1, 64)
	}

	// 블록 생성, 검증, 전파, 전체 시간을 ms/s 단위로 변환하고 string으로 변경한 값
	// fmt.Println("trimStrBGT -->> ", trimStrBGT)
	// fmt.Println("trimStrBVT -->> ", trimStrBVT)
	// fmt.Println("trimStrBPT -->> ", trimStrBPT)
	// fmt.Println("trimStrTBT -->> ", trimStrTBT)

	strTotalBlockTime := totalBlockTime.String()
	strSplit := strings.Split(strTotalBlockTime, ".")
	// fmt.Println("seconds strSplit[0]-->> ", strSplit[0])
	// fmt.Println("seconds strSplit[1]-->> ", strSplit[1])
	twoSecond := strSplit[1]
	// ms(x) + s(o)
	if strings.LastIndex(twoSecond, "ms") < 0 && strings.LastIndex(twoSecond, "s") > 0 {
		totalBlockTimeInt, err = strconv.Atoi(strSplit[0])
		if err != nil {
			log.Fatal("failed to convert from string to int.")
		}
		// fmt.Println("Set s, i.e., totalBlockTimeInt = previousTotalBlockTime")
	}
	// ms(o)
	if strings.LastIndex(twoSecond, "ms") > 0 {
		totalBlockTimeInt = 0
		// fmt.Println("Set ms, i.e., totalBlockTimeInt = 0")
	}
	// fmt.Println("totalBlockTime to int -->> ", totalBlockTimeInt)
	// fmt.Println("blockGenerationEpochTime -->> ", blockGenerationEpochTime)

	// one block cycle time과 total block time과 비교하여 deadline 보장 확인
	if blockGenerationEpochTime >= totalBlockTimeInt {
		fmt.Println("Guarantee deadline")
		IsDeadlineSuccess = true

	} else {
		// 실패했을때 transaction pool로 이동하고 priority 계산하고 threshold 0으로 설정
		IsDeadlineSuccess = false
		fmt.Println("Not guarantee deadline")
	}

	// write .csv file
	writeResultCSV(trimStrBGT, trimStrBVT, trimStrBPT, trimStrTBT, trimStrTBVT)

	fmt.Println("========= block cylce unlock =========")
	isBlockCycle = true
	// fmt.Println("after isBlockCycle -->> ", isBlockCycle)

}

func sendBlockchain(ip, port string, blockchain []Block) {
	url := fmt.Sprintf("http://%s:%s/receivenewblockchain", ip, port)
	fmt.Println("[Propagate a block to node]")
	fmt.Printf("Node IP: %s\n", ip)
	fmt.Printf("Node Port: %s\n", port)
	// printLogs("Sending the blockchain to", url)

	jsonData, err := json.Marshal(blockchain)
	if err != nil {
		log.Println("Error marshaling transactionsInfo:", err)
		return
	}

	// 블록을 blockchain.json에 기록.
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending transactionsInfo : %v", err)

	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Message sent successfully.")

		// IsPropagated Flag 설정
		wg.Done()
		return
	}

	log.Println("Max retries reached. Unable to send message.")
}
