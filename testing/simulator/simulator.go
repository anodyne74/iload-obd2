package simulator

import (
	"encoding/binary"
	"math/rand"
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
var TestDTCs = []string{
	"P0087", // Fuel Rail Pressure Too Low
	"P0088", // Fuel Rail Pressure Too High
	"P0191", // Fuel Rail Pressure Sensor Circuit Range/Performance
	"P0401", // EGR Flow Insufficient
	"P0234", // Turbocharger Overboost Condition
}

// Simulator handles vehicle data simulation
type Simulator struct {
	data     SimulatedData
	writer   DataWriter
	interval time.Duration
	done     chan struct{}
}

// DataWriter interface allows different transport implementations
type DataWriter interface {
	Write([]byte) (int, error)
	Close() error
}

// NewSimulator creates a new simulator instance
func NewSimulator(writer DataWriter, interval time.Duration) *Simulator {
	return &Simulator{
		data: SimulatedData{
			RPM:         800,
			Speed:       0,
			Temperature: 85,
			DTCs:        []string{},
		},
		writer:   writer,
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start begins the simulation loop
func (s *Simulator) Start() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateData()
			msg := s.createOBD2Message()
			if _, err := s.writer.Write(msg); err != nil {
				return
			}
		case <-s.done:
			return
		}
	}
}

// Stop halts the simulation
func (s *Simulator) Stop() {
	close(s.done)
	s.writer.Close()
}

func (s *Simulator) updateData() {
	// Update simulated values
	s.data.RPM = 800 + rand.Float64()*2200      // RPM between 800-3000
	s.data.Speed = rand.Float64() * 120         // Speed between 0-120
	s.data.Temperature = 80 + rand.Float64()*15 // Temp between 80-95

	// Randomly inject DTCs (5% chance each cycle)
	if rand.Float64() < 0.05 && len(s.data.DTCs) < 2 {
		newDTC := TestDTCs[rand.Intn(len(TestDTCs))]
		if !contains(s.data.DTCs, newDTC) {
			s.data.DTCs = append(s.data.DTCs, newDTC)
		}
	}
}

func (s *Simulator) createOBD2Message() []byte {
	// Basic OBD2 message format
	msg := make([]byte, 8)

	// Mode 1 PID format
	msg[0] = 0x02 // Length
	msg[1] = 0x01 // Mode 1

	// Rotate through PIDs
	switch time.Now().UnixNano() % 3 {
	case 0: // RPM (PID 0x0C)
		msg[2] = 0x0C
		rpm := uint16(s.data.RPM * 4) // OBD2 uses RPM/4
		binary.BigEndian.PutUint16(msg[3:5], rpm)
	case 1: // Speed (PID 0x0D)
		msg[2] = 0x0D
		msg[3] = byte(s.data.Speed)
	case 2: // Temperature (PID 0x05)
		msg[2] = 0x05
		msg[3] = byte(s.data.Temperature + 40) // OBD2 uses Temp+40
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
