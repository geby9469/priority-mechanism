package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

var jsonFilePath = "allnodes.json"

var MY_IP, HttpPort, Leader_IP, Leader_Port, nodeInfoVersionPort = loadEnv()
var HttpRequestPort = fmt.Sprintf("%s:%s", MY_IP, HttpPort)
var LeaderAddress = fmt.Sprintf("%s:%s", Leader_IP, Leader_Port)
var ResquestNodeInfo = fmt.Sprintf("%s:%s", Leader_IP, nodeInfoVersionPort)

func main() {
	listener, err := net.Listen("tcp", LeaderAddress)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()

	fmt.Println("Nodes communication Listening on port: Leader Port ==> ", Leader_Port)
	var myIP = getNodeIPAndPort()
	fmt.Println("Leader Node running on", myIP)

	go SendNodeInfNewversion()        // 새로운 client 접근시 해당 IP 출력
	go handleTCPConnections(listener) // 새로운 client 접근시 allnodes.json에 client 추가
	go handleHTTPRequests()           // /sendmessage로 오면 msg처리 MY_IP:HttpPort로 계속 대기

	select {}
}

// Leader와 connection 맺음
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

func SendNodeInfNewversion() {
	listener, err := net.Listen("tcp", ResquestNodeInfo)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleNodeInfo_New_version(conn)
	}
}

func handleHTTPRequests() {
	fmt.Println("ready to receive transactions on", HttpPort)
	http.HandleFunc("/sendmessage", handleMessage)
	err := http.ListenAndServe(HttpRequestPort, nil)
	if err != nil {
		log.Fatal("Error starting HTTP server:", err)
	}
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	// log.Println("received it")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Leader node에서 Transactions 정보 추출
	var transaction Transaction
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&transaction); err != nil {
		log.Println("Error decoding JSON:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	txID := int(transaction.TransactionID)
	txRemainingTime := int(transaction.RemainingTime)
	txThreshold := int(transaction.Threshold)

	fmt.Println("[Leader Transaction]")
	fmt.Printf("TransactionID: %d\nRemainingTime: %d\nThreshold: %d\n",
		txID, txRemainingTime, txThreshold)

	// Start the continuous function to handle "hi" messages in parallel
	go continuousFunction(transaction)
}

func continuousFunction(transactionsInfo Transaction) {
	// This function will now only run when the client sends a transactionsInfo to the /sendmessage endpoint.
	// It receives the transactionsInfo sent by the client and can process it accordingly.
	// printLogs("Transactions received ", string(transactionsInfo.Sender))

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

	// Shuffle the nodes array randomly
	rand.Seed(time.Now().UnixNano())
	// rand.Shuffle(len(allNodes), func(i, j int) { allNodes[i], allNodes[j] = allNodes[j], allNodes[i] })

	// Pick the first node after shuffling
	if len(allNodes) > 0 {
		node := allNodes[0]
		// fmt.Println("IP", node.IP)
		// fmt.Println("Port", node.Port)
		// fmt.Println("transactionInfo", transactionsInfo)
		sendMessage(node.IP, node.Port, transactionsInfo)
	} else {
		fmt.Println("No nodes available.")
	}
}

func sendMessage(ip, port string, transactionsInfo Transaction) {
	url := fmt.Sprintf("http://%s:%s/sendmessage", ip, port)

	// printLogs("Sending the blockchain mining node to", url)

	jsonData, err := json.Marshal(transactionsInfo)
	if err != nil {
		log.Println("Error marshaling transactionsInfo:", err)
		return
	}

	// Miner에게 최종 전달
	fmt.Println("[Leader send transaction to Miner]")
	fmt.Println("[Miner Info]")
	fmt.Printf("Miner IP: %s\nMiner Port: %s\n", ip, port)
	fmt.Println("Final transaction to be sent miner\n", string(jsonData))

	maxRetries := 50                 // Set the maximum number of retries you want to attempt
	retryInterval := 5 * time.Second // Set the interval between retries

	for retry := 0; retry <= maxRetries; retry++ {
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error sending transactionsInfo (retry %d): %v", retry, err)
			if retry < maxRetries {
				log.Printf("Retrying to sent the transactions in %v...", retryInterval)
				time.Sleep(retryInterval)
				continue
			}
			log.Println("Max retries reached. Unable to send transactionsInfo.")
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Message sent successfully.")
			return
		}

		log.Printf("Error (retry %d): None of the nodes is connected: %d", retry, resp.StatusCode)
		if retry < maxRetries {
			log.Printf("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	log.Println("Max retries reached. Unable to send message.")
}

func handleConnection(conn net.Conn) {
	// fmt.Println("start handleConnection!")
	defer conn.Close()
	// fmt.Println("conn", conn.LocalAddr())

	var clientInfo ClientInfo
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&clientInfo); err != nil {
		log.Println("Error decoding JSON:", err)
		return
	}

	printLogs("A new connection from:", clientInfo.IP)
	fmt.Println("client port : ", clientInfo.Port)

	// Append the new client info to the JSON file
	existingData, err := readJSONFile(jsonFilePath)
	if err != nil {
		log.Println("Error reading JSON file:", err)
		return
	}

	existingData = append(existingData, clientInfo)

	if err := writeJSONFile(jsonFilePath, existingData); err != nil {
		log.Println("Error writing to JSON file:", err)
		return
	}

	// Read the JSON file content and include it in the response
	jsonContent, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Println("Error reading JSON file:", err)
		return
	}

	clientInfo.IP = string(jsonContent)

	// Send the response (including the JSON content) back to the client
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(clientInfo); err != nil {
		log.Println("Error encoding JSON response:", err)
		return
	}
}
func handleNodeInfo_New_version(conn net.Conn) {
	defer conn.Close()

	var clientInfo ClientInfo

	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&clientInfo); err != nil {
		log.Println("Node info New version Error decoding JSON:", err)
		return
	}
	fmt.Println("\n[Leader Heartbeat] Connected miner node")
	// fmt.Printf("Miner IP: %s\n", clientInfo.IP)
	// fmt.Printf("Miner Port: %s\n", clientInfo.Port)

	// printLogs("Nodes list version request from:", clientInfo.IP)

	// Read the JSON file content and include it in the response
	jsonContent, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Println("Error reading JSON file:", err)
		return
	}

	clientInfo.IP = string(jsonContent)

	// Send the response (including the JSON content) back to the client
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(clientInfo); err != nil {
		log.Println("Error encoding JSON response:", err)
		return
	}
}

func readJSONFile(filePath string) ([]ClientInfo, error) {
	var data []ClientInfo

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
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, jsonData, 0644)
}

func getNodeIPAndPort() string {
	/*conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal("Error getting client IP:", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)*/
	return MY_IP //
}
