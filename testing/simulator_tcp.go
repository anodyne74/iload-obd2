package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"time"
)

// SimulatedData represents the current state of our simulated vehicle
type SimulatedData struct {
	RPM         float64
	Speed       float64
	Temperature float64
	DTCs        []string
}

// Common DTCs for testing
var testDTCs = []string{
	"P0087", // Fuel Rail Pressure Too Low
	"P0088", // Fuel Rail Pressure Too High
	"P0191", // Fuel Rail Pressure Sensor Circuit Range/Performance
	"P0401", // EGR Flow Insufficient
	"P0234", // Turbocharger Overboost Condition
}

func main() {
	// Start TCP server
	listener, err := net.Listen("tcp", "localhost:6789")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println("Simulator listening on localhost:6789")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from %s", conn.RemoteAddr())

	// Initialize simulated data
	data := SimulatedData{
		RPM:         800,
		Speed:       0,
		Temperature: 85,
		DTCs:        []string{},
	}

	// Main simulation loop
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Update simulated values
		data.RPM = 800 + rand.Float64()*2200      // RPM between 800-3000
		data.Speed = rand.Float64() * 120         // Speed between 0-120
		data.Temperature = 80 + rand.Float64()*15 // Temp between 80-95

		// Randomly inject DTCs (5% chance each cycle)
		if rand.Float64() < 0.05 && len(data.DTCs) < 2 {
			newDTC := testDTCs[rand.Intn(len(testDTCs))]
			if !contains(data.DTCs, newDTC) {
				data.DTCs = append(data.DTCs, newDTC)
			}
		}

		// Create OBD2 message
		msg := createOBD2Message(data)

		// Send over TCP connection
		_, err := conn.Write(msg)
		if err != nil {
			log.Printf("Error writing to connection: %v", err)
			return
		}
	}
}

func createOBD2Message(data SimulatedData) []byte {
	// Basic OBD2 message format
	msg := make([]byte, 8)

	// Mode 1 PID format
	msg[0] = 0x02 // Length
	msg[1] = 0x01 // Mode 1

	// Rotate through PIDs
	switch time.Now().UnixNano() % 3 {
	case 0: // RPM (PID 0x0C)
		msg[2] = 0x0C
		rpm := uint16(data.RPM * 4) // OBD2 uses RPM/4
		binary.BigEndian.PutUint16(msg[3:5], rpm)
	case 1: // Speed (PID 0x0D)
		msg[2] = 0x0D
		msg[3] = byte(data.Speed)
	case 2: // Temperature (PID 0x05)
		msg[2] = 0x05
		msg[3] = byte(data.Temperature + 40) // OBD2 uses Temp+40
	}

	return msg
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
