package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"time"

	"github.com/go-daq/canbus"
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
	// Create virtual CAN interface
	send, err := canbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer send.Close()

	err = send.Bind("vcan0")
	if err != nil {
		log.Fatalf("could not bind send socket: %+v", err)
	}

	// Initialize simulated data
	data := SimulatedData{
		RPM:         800,
		Speed:       0,
		Temperature: 85,
		DTCs:        []string{},
	}

	// Periodically inject random DTCs
	go func() {
		for {
			time.Sleep(30 * time.Second)
			if rand.Float64() < 0.3 { // 30% chance to add a DTC
				randomDTC := testDTCs[rand.Intn(len(testDTCs))]
				if !contains(data.DTCs, randomDTC) {
					data.DTCs = append(data.DTCs, randomDTC)
					log.Printf("Added DTC: %s\n", randomDTC)
				}
			}
		}
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Main simulation loop
	for range ticker.C {
		// Simulate RPM changes
		data.RPM += (rand.Float64() - 0.5) * 100
		if data.RPM < 800 {
			data.RPM = 800
		}
		if data.RPM > 3000 {
			data.RPM = 3000
		}

		// Simulate speed changes
		data.Speed += (rand.Float64() - 0.5) * 2
		if data.Speed < 0 {
			data.Speed = 0
		}
		if data.Speed > 120 {
			data.Speed = 120
		}

		// Simulate temperature changes
		data.Temperature += (rand.Float64() - 0.5) * 0.5
		if data.Temperature < 80 {
			data.Temperature = 80
		}
		if data.Temperature > 95 {
			data.Temperature = 95
		}

		// Send RPM data
		sendCANFrame(send, 0x7E8, encodeRPM(data.RPM))

		// Send Speed data
		sendCANFrame(send, 0x7E9, encodeSpeed(data.Speed))

		// Send Temperature data
		sendCANFrame(send, 0x7EA, encodeTemp(data.Temperature))

		// Send DTCs if any exist
		if len(data.DTCs) > 0 {
			sendCANFrame(send, 0x7EB, encodeDTCs(data.DTCs))
		}
	}
}

func sendCANFrame(send *canbus.Socket, id uint32, data []byte) {
	frame := canbus.Frame{
		ID:   id,
		Data: data,
		Kind: canbus.SFF,
	}
	if _, err := send.Send(frame); err != nil {
		log.Printf("Error sending frame: %v\n", err)
	}
}

func encodeRPM(rpm float64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint16(data[0:2], uint16(rpm*4))
	return data
}

func encodeSpeed(speed float64) []byte {
	data := make([]byte, 8)
	data[0] = byte(speed)
	return data
}

func encodeTemp(temp float64) []byte {
	data := make([]byte, 8)
	data[0] = byte(temp + 40) // OBD2 temperature offset
	return data
}

func encodeDTCs(dtcs []string) []byte {
	data := make([]byte, 8)
	if len(dtcs) > 0 {
		// Just encode the first DTC for demonstration
		dtc := dtcs[0]
		// Convert the DTC string to bytes (simplified)
		copy(data, []byte(dtc))
	}
	return data
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
