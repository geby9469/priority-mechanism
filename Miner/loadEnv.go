package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() (string, string, string, string, string, string, string, string) {
	// Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Read variables from the environment
	value1 := os.Getenv("MY_IP")
	value2 := os.Getenv("MyTCPPort")
	value3 := os.Getenv("MyHTTPPort")
	value4 := os.Getenv("Leader_IP")
	value5 := os.Getenv("Leader_Port")
	value6 := os.Getenv("ClosestNodeIP")
	value7 := os.Getenv("ClosestNodePort") //LeadernodeInfoVersionPort
	value8 := os.Getenv("LeadernodeInfoVersionPort")
	return value1, value2, value3, value4, value5, value6, value7, value8

}
