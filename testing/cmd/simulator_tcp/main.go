package main

import (
	"log"

	"iload-obd2/testing/simulator"
)

func main() {
	err := simulator.StartTCPServer("localhost:6789")
	if err != nil {
		log.Fatal(err)
	}
}
