package main

import (
	"log"
	"time"

	"iload-obd2/testing/simulator"
)

func main() {
	writer, err := simulator.NewSerialWriter("COM10", 38400)
	if err != nil {
		log.Fatal(err)
	}

	sim := simulator.NewSimulator(writer, 100*time.Millisecond)
	sim.Start()
}
