package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var blocksInfoFile = "blockchain.json"
var nodesInfoFile = "allnodes.json"
var csvFilePath = "../logs/result.csv"

var MY_IP, MyTCPPort, MyHTTPPort, Leader_IP, Leader_Port, closestNodeIP, closestNodePort, LeadernodeInfoVersionPort = loadEnv()
var MyTCPaddress = fmt.Sprintf("%s:%s", MY_IP, MyTCPPort)
var MyHTTPaddress = fmt.Sprintf("%s:%s", MY_IP, MyHTTPPort)
var LeaderAddress = fmt.Sprintf("%s:%s", Leader_IP, Leader_Port)
var ClosestNode = fmt.Sprintf("%s:%s", closestNodeIP, closestNodePort) //

var LeaderNodeVersion = fmt.Sprintf("%s:%s", Leader_IP, LeadernodeInfoVersionPort)

var isGenesisBlock = true

var priority1Txs = []TransactionPool{} // temporary storage for transactions with priority 1

func main() {

	listener, err := net.Listen("tcp", MyTCPaddress)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	// Leader node와 TCP 통신
	dialLeaderNode()
	if closestNodePort == MyTCPPort {
		fmt.Println("There is not any blockchain to download")
	} else {
		dialClosestNode()
	}

	fmt.Println("Communicating with other nodes on port", MyTCPPort)

	// Leader node와 Heartbeat 통신
	go dialLeaderNode_Continouslly()
	go handleTCPConnections(listener)

	// Leader node로 부터 전달받은 트랜잭션
	go handleHTTPRequests()
	//go handleReceiveBlockchain()

	select {}
}

func handleTCPConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleHTTPRequests() {
	fmt.Println("ready to receive blocks on", MyHTTPPort)
	http.HandleFunc("/sendmessage", handleMessage)
	http.HandleFunc("/receivenewblockchain", handleReceiveChain)
	err := http.ListenAndServe(MyHTTPaddress, nil)
	if err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}

/*func handleReceiveBlockchain() {
	fmt.Println("ready to receive blocks on", MyHTTPPort)
	http.HandleFunc("/receivenewblockchain", handleReceiveChain)
	err := http.ListenAndServe(MyHTTPaddress, nil)
	if err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}*/

func handleReceiveChain(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	defer r.Body.Close()
	fmt.Println("Update a block in 'blockchain.json' file")

	var receivedBlockchainHere []Block

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&receivedBlockchainHere); err != nil {
		http.Error(w, "Error decoding JSON:", http.StatusBadRequest)
		return
	}

	if err := writeBlockchainFile(blocksInfoFile, receivedBlockchainHere); err != nil {
		log.Println("Error writing to JSON file:", err)
		return
	}
	fmt.Println("Propagation complete!!")
	// fmt.Println("Message appended to message.json.")

	// Respond to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Message received and appended successfully.")
}

func saveTransactions(transaction Transaction) {
	fmt.Println("\n[Miner received transaction]")
	// fmt.Println("transaction -->> ", transaction)

	transformTxToTxpool := TransactionPool{}
	transformTxToTxpool.TransactionID = transaction.TransactionID
	transformTxToTxpool.RemainingTime = transaction.RemainingTime
	transformTxToTxpool.Priority = 0
	transformTxToTxpool.Threshold = transaction.Threshold
	// fmt.Println("transformTxToTxpool -->> ", transaction)

	transactionPools[transaction.TransactionID] = transformTxToTxpool

	// fmt.Println("transactionPools[transaction.TransactionID] -->> ", transactionPools[transaction.TransactionID])

	// transaction pool iteration
	// go setEpochTimeRoutine()
	if len(transactionPools) != 0 {
		// epochTime for genesis block
		if epochTime == 0 {
			// epochTime = transactionPools[1].RemainingTime
			// because of 5 miner
			epochTime = 9999
			for _, element := range transactionPools {
				// fmt.Printf("Key: %d, txid: %d, remainingtime: %d threshold: %d\n", key, element.TransactionID, element.RemainingTime, element.Threshold)
				// 재전송하는 것은 고려하지 않음 (threshold > 0)
				if element.RemainingTime <= epochTime && element.Threshold > 0 {
					epochTime = element.RemainingTime
				}
			}

		}
		// fmt.Println("Set epochTime -->> ", epochTime)
		// 현재 트랜잭션 풀 중에서 remaining time이 가장 작은 값으로 epochTime 설정
		for _, element := range transactionPools {
			// fmt.Printf("Key: %d, txid: %d, remainingtime: %d threshold: %d\n", key, element.TransactionID, element.RemainingTime, element.Threshold)
			// 재전송하는 것은 고려하지 않음 (threshold > 0)
			if element.RemainingTime <= epochTime && element.Threshold > 0 {
				epochTime = element.RemainingTime
			}
		}
		// fmt.Println("epochTime : ", epochTime)
	}
}

func myProposalAlgorithm() Block {

	// block creation alogirhtm
	if len(transactionPools) != 0 {
		// fmt.Println("len(transactionPools) -->> ", len(transactionPools))
		// START CREATING A BLOCK!!
		// Step 1. Block generation
		// elapsedGenerationPlusValidationTime := time.Now()
		elapsedGenerationTime := time.Now()

		// 처음에는 epochtime(가장작은remainingtime)으로 설정하고
		// 두번째부터는 previous total block time으로 설정.
		if isGenesisBlock {
			blockGenerationEpochTime = epochTime
			fmt.Println("Start GenesisBlock")
		} else {
			fmt.Print("totalBlockTimeInt -->> ", totalBlockTimeInt)
			// case: previousTotalBlockTime < 1s
			if totalBlockTimeInt == 0 {
				fmt.Println("\nPrevious total block time under 1s")
				blockGenerationEpochTime = epochTime
			} else {
				fmt.Println("\nPrevious total block time between 1s and 59s")
				blockGenerationEpochTime = totalBlockTimeInt
			}
		}

		fmt.Printf("\nEpoch Time in one block cycle : %ds\n", blockGenerationEpochTime)

		// According to deadline of success or failure, Processing transactions
		if IsDeadlineSuccess {
			fmt.Println("Remove completed transaction in the transaction pool")
			fmt.Println("Completed transaction count : ", len(priority1Txs))
			for i := 0; i < len(priority1Txs); i++ {
				delete(transactionPools, priority1Txs[i].TransactionID)
			}
			priority1Txs = []TransactionPool{} // initialization
			// fmt.Println("priority1Txs -->> ", priority1Txs)
		} else {
			fmt.Println("Add incompleted transaction in the transaction pool")
			fmt.Println("Incompleted transaction count : ", len(priority1Txs))
			for key, element := range priority1Txs {
				var priority1TxPool TransactionPool
				priority1TxPool.TransactionID = element.TransactionID
				priority1TxPool.RemainingTime = 0
				priority1TxPool.Priority = element.Threshold
				priority1TxPool.Threshold = 0
				transactionPools[key] = priority1TxPool
			}
			priority1Txs = []TransactionPool{} // initialization
			// fmt.Println("priority1Txs -->> ", priority1Txs)
		}

		// calculate priority and copy transaction pool
		TrashTransactionCount = 0                      // initialize count of previous trashtransaction.
		var copyTxPool = make(map[int]TransactionPool) // temporary priority driven transaction pool
		for key, element := range transactionPools {
			var oneTxInTxPool TransactionPool
			oneTxInTxPool.TransactionID = element.TransactionID
			// genesis block
			if isGenesisBlock {
				fmt.Println("Calculate priority for GenesisBlock")
				oneTxInTxPool.RemainingTime = element.RemainingTime
				oneTxInTxPool.Priority = int(element.RemainingTime / blockGenerationEpochTime)
				oneTxInTxPool.Threshold = element.Threshold
				isGenesisBlock = false // genesis block 해제
				// fmt.Println("oneTxInTxPool.Priority", oneTxInTxPool.Priority)
				// not genesis block
			} else {
				// fmt.Println("Calculate priority for a normal block")
				// oneTxInTxPool.RemainingTime = element.RemainingTime
				// oneTxInTxPool.Priority = element.Priority
				// oneTxInTxPool.Threshold = element.Threshold
				// deadline success
				if IsDeadlineSuccess {
					// fmt.Println("Deadline Success: Calculate Priority")

					// threshold가 존재x + deadline 보장x
					if element.Threshold == 0 {
						oneTxInTxPool.RemainingTime = 0
						oneTxInTxPool.Threshold = 0
						oneTxInTxPool.Priority = element.Priority - 1
						if element.Priority == 0 {
							// Discard transaction.
							delete(transactionPools, key)
							TrashTransactionCount++
							continue
						}
					}
					// threshold가 존재o + deadline 보장x
					notGuaranteedTx := element.RemainingTime - blockGenerationEpochTime
					if notGuaranteedTx <= 0 {
						oneTxInTxPool.RemainingTime = 0
						oneTxInTxPool.Priority = element.Threshold
						oneTxInTxPool.Threshold = 0
					} else {
						oneTxInTxPool.RemainingTime = notGuaranteedTx
						oneTxInTxPool.Priority = int(oneTxInTxPool.RemainingTime / blockGenerationEpochTime)
						oneTxInTxPool.Threshold = element.Threshold
					}
					// deadli fail
				} else {
					// fmt.Println("Deadline Fail: Calculate Priority")
					notGuaranteedFailedTx := element.RemainingTime - blockGenerationEpochTime

					// threshold가 존재x + deadline 보장x
					if element.Threshold == 0 {
						oneTxInTxPool.RemainingTime = 0
						oneTxInTxPool.Threshold = 0
						oneTxInTxPool.Priority = element.Priority - 1
						if element.Priority == 0 {
							// Discard transaction.
							delete(transactionPools, key)
							TrashTransactionCount++
							continue
						}
					}

					// threshold가 존재o + deadline 보장x
					if notGuaranteedFailedTx <= 0 {
						oneTxInTxPool.RemainingTime = 0
						oneTxInTxPool.Priority = element.Threshold
						oneTxInTxPool.Threshold = 0
					}
				}
			}

			copyTxPool[key] = oneTxInTxPool
		}

		// fmt.Println("copyTxPool --> ", copyTxPool)
		// fmt.Println("copyTxPool[1].Priority --> ", copyTxPool[1].Priority)
		fmt.Printf("\nBefore generating a block, Priority driven transaction pool size : %d\n", len(transactionPools))
		// fmt.Println("copyTxPool :: ", copyTxPool)
		// elapsedCreationTime := time.Now().UnixNano() / int64(time.Millisecond)

		// Insert transactions with priority 1 in body of a block.
		for _, element := range copyTxPool {
			if element.Priority == 1 {
				tempBlockTx := TransactionPool{
					TransactionID: element.TransactionID,
					RemainingTime: element.RemainingTime,
					Priority:      element.Priority,
					Threshold:     element.Threshold,
				}
				priority1Txs = append(priority1Txs, tempBlockTx)
				// delete transaction with priority 1 in transaction pool
				// delete(transactionPools, tempBlockTx.TransactionID)
			}
		}
		// block structure
		// fmt.Println("After priority1Txs -->> ", priority1Txs)
		fmt.Printf("\nNumber of Transactions with priority 1: %d\n", len(priority1Txs))
		fmt.Printf("\nAfter generating a block, Priority driven transaction pool size : %d\n", len(transactionPools))
		if len(priority1Txs) == 0 {
			// fmt.Println("It doesn't have transactions in a block")
			var noBlock = Block{}
			return noBlock
		}

		var prevHash = getLastBlockchainHash()
		var previousID = getLastID()

		// 2023-09-08 Block Generation Time 추가
		blockGenerationTime = time.Since(elapsedGenerationTime)

		elapsedValidationTime := time.Now()
		block := Block{
			Index:            previousID + 1,
			Timestamp:        time.Now(),
			TransactionPool:  priority1Txs,
			PreviousHash:     prevHash,
			Hash:             proofOfWork(prevHash), // Step2. Block validation
			Nonce:            increaseNonce(prevHash),
			Difficulty:       getDifficulty(),
			DifficultyTarget: getTarget(),
			TxCount:          len(priority1Txs),
		}
		blockValidationTime = time.Since(elapsedValidationTime)

		// 2023-09-11 Add logic of transaction validation
		elapsedTxValidationTime := time.Now()
		transactionValidationCount = 0
		for i := 0; i < len(priority1Txs); i++ {
			priority1Data := bytes.Join([][]byte{
				[]byte(strconv.Itoa(priority1Txs[i].TransactionID)),
				[]byte(strconv.Itoa(priority1Txs[i].RemainingTime)),
				[]byte(strconv.Itoa(priority1Txs[i].Priority)),
				[]byte(strconv.Itoa(priority1Txs[i].Threshold)),
			}, []byte{})
			for j := 0; j < len(block.TransactionPool); j++ {
				blockData := bytes.Join([][]byte{
					[]byte(strconv.Itoa(block.TransactionPool[j].TransactionID)),
					[]byte(strconv.Itoa(block.TransactionPool[j].RemainingTime)),
					[]byte(strconv.Itoa(block.TransactionPool[j].Priority)),
					[]byte(strconv.Itoa(block.TransactionPool[j].Threshold)),
				}, []byte{})
				p1Hash := sha256.Sum256(priority1Data)
				txHash := sha256.Sum256(blockData)
				if p1Hash == txHash {
					transactionValidationCount++
					break
				}
			}
		}
		txInblockPropagationTime = time.Since(elapsedTxValidationTime)

		fmt.Println("\nElapsed Time to transaction validation time : ", txInblockPropagationTime)
		fmt.Println("\nElapsed Time to block generation time : ", blockGenerationTime)
		fmt.Println("\nElapsed Time to block validation time : ", blockValidationTime)

		// writing for .csv file
		TxCount = strconv.Itoa(len(copyTxPool))               // 트랜잭션 풀에 있는 트랜잭션 수
		Prirotiy1TxCount = strconv.Itoa(len(priority1Txs))    // 트랜잭션 풀에 있는 prirotiy1 트랜잭션 수
		TxCountInBlock = strconv.Itoa(block.TxCount)          // 블록에 있는 트랜잭션 수
		StrblockGenerationTime = blockGenerationTime.String() // 블록 생성 시간
		StrBlockValidationTime = blockValidationTime.String() // 블록 검증 시간

		return block
	}
	var noBlock = Block{}
	return noBlock
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	defer r.Body.Close()

	var transactionReceived Transaction

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&transactionReceived); err != nil {
		http.Error(w, "Error decoding JSON:", http.StatusBadRequest)
		return
	}

	// Leader node한테 받은 transactions 저장
	saveTransactions(transactionReceived)

	// block creation algorithm
	if isBlockCycle {
		fmt.Println("========= block cylce lock =========")
		isBlockCycle = false

		// Execute my proposal algorithm.
		// var block Block
		block := myProposalAlgorithm()

		// Block creation complete.
		// var checkBlock Block
		if !reflect.ValueOf(block).IsZero() {
			fmt.Println("\nGenerating a block complete!!")
			// fmt.Println("block index -->> ", block.Index)
			// fmt.Println("block PreviousHash -->> ", block.PreviousHash)
			// fmt.Println("block hash -->> ", block.Hash)
			// Append the new client info to the JSON file
			existingData, err := readBlockchainFile(blocksInfoFile)
			if err != nil {
				log.Println("Error reading JSON file:", err)
				return
			}

			existingData = append(existingData, block)

			if err := writeBlockchainFile(blocksInfoFile, existingData); err != nil {
				log.Println("Error writing to JSON file:", err)
				return
			}

			// Read the JSON file content and include it in the response
			_, err = ioutil.ReadFile(blocksInfoFile)
			if err != nil {
				log.Println("Error reading JSON file:", err)
				return
			}

			// fmt.Println("New Block added")

			// Respond to the client
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Message received and appended successfully.")
			// fmt.Println("Propagating new block")
			alwaySendBlockchainAtAnyTime() // propagate the block, basically sending the entire blockchain version
		} else {
			// No block
			fmt.Println("It doesn't have transactions in a block")
			isBlockCycle = true
		}

	}
}

func dialLeaderNode() {
	clientIP, clientPort := getClientIPAndPort()

	// Send the client's IP address and port to the server
	clientInfo := ClientInfo{
		IP:   clientIP,
		Port: clientPort,
	}

	var conn net.Conn
	var err error

	// Keep trying to connect to the leader until it becomes available
	for {
		conn, err = net.DialTimeout("tcp", LeaderAddress, time.Second*20)
		if err != nil {
			fmt.Println("Wainting for the leader node to connect...:", err)
			// Retry after a short delay (e.g., 5 seconds)
			time.Sleep(time.Second * 5)
		} else {
			break
		}
	}

	defer conn.Close()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(clientInfo); err != nil {
		log.Fatal("Error sending client info:", err)
	}

	// Receive the JSON content from the server
	var serverResponse ClientInfo
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&serverResponse); err != nil {
		log.Fatal("Error decoding response:", err)
	}

	fmt.Println("Received response from server:")
	// fmt.Println("Server IP:", serverResponse.IP)
	fmt.Println("Server Port:", serverResponse.Port)
	// fmt.Println("JSON Content:")
	// fmt.Println("IP and port well received", serverResponse.IP)

	// Save the JSON content to the client's file
	if err := ioutil.WriteFile(nodesInfoFile, []byte(serverResponse.IP), 0644); err != nil {
		log.Fatal("Error saving JSON file:", err)
	}

	fmt.Println("Node info received")

}

func dialLeaderNode_Continouslly() {

	for {
		clientIP, clientPort := getClientIPAndPort()

		// Send the client's IP address and port to the server
		clientInfo := ClientInfo{
			IP:   clientIP,
			Port: clientPort,
		}

		var conn net.Conn
		var err error

		// Keep trying to connect to the leader until it becomes available
		for {
			conn, err = net.DialTimeout("tcp", LeaderNodeVersion, time.Second*20)
			if err != nil {
				fmt.Println("Wainting for the leader node to connect...:", err)
				// Retry after a short delay (e.g., 5 seconds)
				time.Sleep(time.Second * 5)
			} else {
				break
			}
		}

		defer conn.Close()

		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(clientInfo); err != nil {
			log.Fatal("Error sending client info:", err)
		}

		// Receive the JSON content from the server
		var serverResponse ClientInfo
		decoder := json.NewDecoder(conn)
		if err := decoder.Decode(&serverResponse); err != nil {
			log.Fatal("Error decoding response:", err)
		}

		fmt.Println("\n[Miner Heartbeat] Connected miner node")
		// fmt.Printf("Miner IP: %s\n", clientInfo.IP)
		// fmt.Printf("Miner Port: %s\n", clientInfo.Port)

		// Save the JSON content to the client's file
		if err := ioutil.WriteFile(nodesInfoFile, []byte(serverResponse.IP), 0644); err != nil {
			log.Fatal("Error saving JSON file:", err)
		}

		// fmt.Println("Node info received")

		time.Sleep(1 * time.Second)
	}

}

func dialClosestNode() {

	clientIP, clientPort := getClientIPAndPort()

	// Send the client's IP address and port to the server
	clientInfo := ClientInfo{
		IP:   clientIP,
		Port: clientPort,
	}

	var conn net.Conn
	var err error

	// Keep trying to connect to the leader until it becomes available
	for {
		conn, err = net.DialTimeout("tcp", ClosestNode, time.Second*2)
		if err != nil {
			fmt.Println("Wainting for the leader node to connect...:", err)
			// Retry after a short delay (e.g., 5 seconds)
			time.Sleep(time.Second * 5)
		} else {
			break
		}
	}

	defer conn.Close()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(clientInfo); err != nil {
		log.Fatal("Error sending client info:", err)
	}

	// Receive the JSON content from the server
	var serverResponse []Block
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&serverResponse); err != nil {
		log.Fatal("Error decoding response:", err)
	}

	fmt.Println("JSON Content:", serverResponse[0])

	// Save the JSON content to the client's file
	if err := ioutil.WriteFile(blocksInfoFile, []byte(serverResponse[0].PreviousHash), 0644); err != nil {
		log.Fatal("Error saving JSON file:", err)
	}

	fmt.Println("Node info received")

}

func getClientIPAndPort() (string, string) {
	/*conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal("Error getting client IP:", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)*/
	return MY_IP, MyHTTPPort //
}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	var clientInfo ClientInfo

	blockchainchain := []Block{}

	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&clientInfo); err != nil {
		log.Println("Error decoding JSON:", err)
		return
	}

	printLogs("A new connection from:", clientInfo.IP)

	// Append the new client info to the JSON file
	existingData, err := readBlockchainFile(blocksInfoFile)
	if err != nil {
		log.Println("Error reading JSON file:", err)
		return
	}

	// Read the JSON file content and include it in the response
	jsonContent, err := ioutil.ReadFile(blocksInfoFile)
	if err != nil {
		log.Println("Error reading JSON file:", err)
		return
	}

	blockchainchain = existingData

	blockchainchain[0].PreviousHash = string(jsonContent)

	// Send the response (including the JSON content) back to the client
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(blockchainchain); err != nil {
		log.Println("Error encoding JSON response:", err)
		return
	}
}
