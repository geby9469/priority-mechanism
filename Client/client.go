package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const remainingTimeRange = 2 // remaining time 범위: 1~2
const thresholdRange = 2     // threshold 범위: 1~2
const leaderIP = "127.0.0.1" // leader ip
const leaderPort = "3012"    // lead port
const txCount = 801          // 50->51, 100->101, 200->201, 400->401, 800->801, 1000->1001, 1500->1501, 2000->2001

func main() {
	/* *************** */
	/* 트랜잭션 생성 및 제출 */
	/* *************** */
	transactionGenerator()
}

func transactionGenerator() {
	txSequence := 1
	cp1 := 0
	cp2 := 0
	for {
		tx := make(map[string]interface{})
		tx["transactionid"] = txSequence
		remainingTimeInt := rand.Intn(remainingTimeRange) + 1
		tx["remainingtime"] = remainingTimeInt
		thresholdInt := rand.Intn(thresholdRange) + 1
		tx["threshold"] = thresholdInt

		txByte, _ := json.Marshal(tx)
		httpRequestBody := bytes.NewBuffer([]byte(txByte))
		fmt.Printf("[Client Transaction %d]\n %s\n", txSequence, string(txByte))

		leadURL := fmt.Sprintf("http://%s:%s/sendmessage", leaderIP, leaderPort)
		response, err := http.Post(leadURL, "application/json", httpRequestBody)
		if err != nil {
			log.Fatal("ERROR: http post", err)
		}

		if response.Status == "200 OK" {
			fmt.Println("Success to submit transactions!")
			txSequence++
		}

		if remainingTimeInt == 1 {
			cp1++
		}
		if remainingTimeInt == 2 {
			cp2++
		}

		if txSequence == txCount {
			fmt.Printf("remaining time 1 count : %d\n", cp1)
			fmt.Printf("remaining time 2 count : %d\n", cp2)
			fmt.Printf("total transaction count : %d\n", cp1+cp2)
			break
		}

		// time.Sleep(time.Second * 1) // 1s -> 1tx/sec	-> 50
		// time.Sleep(time.Millisecond * 200) // 5tx/sec -> 100
		// time.Sleep(time.Millisecond * 100) // 0.1s -> 10tx/sec -> 200
		// time.Sleep(time.Millisecond * 50) // 0.05s -> 20tx/sec -> 400
		time.Sleep(time.Millisecond * 25) // 0.025s -> 40tx/sec -> 800
		// time.Sleep(time.Millisecond * 17) // 0.0107s -> 60tx/sec -> 1000
		// time.Sleep(time.Millisecond * 12) // 0.0125s -> 80tx/sec -> 1500
		// time.Sleep(time.Millisecond * 10) // 0.01s -> 100tx/sec -> 2000

	}

}
