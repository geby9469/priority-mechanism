package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() (string, string, string, string, string) {
	// Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Read variables from the environment
	value1 := os.Getenv("MY_IP")
	value2 := os.Getenv("HttpPort")
	value3 := os.Getenv("Leader_IP")
	value4 := os.Getenv("Leader_TCP_Port") //nodeInfoVersionPort
	value5 := os.Getenv("nodeInfoVersionPort")
	return value1, value2, value3, value4, value5

}
