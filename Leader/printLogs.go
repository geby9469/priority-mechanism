package main

import (
	"fmt"
)

func printLogs(text, message string) {

	fmt.Println("===========================================")
	fmt.Println("|                                         |")
	fmt.Println(text + " " + message)
	fmt.Println("===========================================")

}
