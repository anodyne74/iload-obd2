package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/brutella/can"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rzetterberg/elmobd"

	"iload-obd2/internal/transport"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

type ECUInfo struct {
	Version     string `json:"version,omitempty"`
	Hardware    string `json:"hardware,omitempty"`
	Software    string `json:"software,omitempty"`
	Calibration string `json:"calibration,omitempty"`
	VIN         string `json:"vin,omitempty"`
	BuildDate   string `json:"buildDate,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
}

type MapData struct {
	Values [][]float64 `json:"values,omitempty"`
	XAxis  []float64   `json:"xAxis,omitempty"`
	YAxis  []float64   `json:"yAxis,omitempty"`
}

type EngineMaps struct {
	Fuel   *MapData `json:"fuel,omitempty"`
	Timing *MapData `json:"timing,omitempty"`
}

type TelemetryData struct {
	RPM        float64     `json:"rpm,omitempty"`
	Speed      float64     `json:"speed,omitempty"`
	Temp       float64     `json:"temp,omitempty"`
	DTCs       []string    `json:"dtcs,omitempty"`
	ECUInfo    *ECUInfo    `json:"ecuInfo,omitempty"`
	EngineMaps *EngineMaps `json:"engineMaps,omitempty"`
	CANFrames  []CANFrame  `json:"canFrames,omitempty"`
}

// CANFrame represents a CAN bus frame
type CANFrame struct {
	ID        uint32    `json:"id"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

// CANHandler handles incoming CAN frames
type CANHandler struct {
	frameChan chan<- CANFrame
}

func (h *CANHandler) Handle(frame can.Frame) {
	data := make([]byte, len(frame.Data))
	copy(data, frame.Data[:])
	h.frameChan <- CANFrame{
		ID:        uint32(frame.ID),
		Data:      data,
		Timestamp: time.Now(),
	}
}

// parseDTCs converts a hex string response into DTC codes
func parseDTCs(hexString string) []string {
	var dtcs []string
	// Remove spaces and process in pairs
	hexString = strings.ReplaceAll(hexString, " ", "")
	for i := 0; i < len(hexString); i += 4 {
		if i+4 <= len(hexString) {
			dtc := hexString[i : i+4]
			if dtc != "0000" {
				dtcs = append(dtcs, fmt.Sprintf("P%s", dtc))
			}
		}
	}
	return dtcs
}

var (
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket upgrade error: %v", err)
		return
	}

	clientsMux.Lock()
	clients[ws] = true
	clientsMux.Unlock()

	defer func() {
		clientsMux.Lock()
		delete(clients, ws)
		clientsMux.Unlock()
		ws.Close()
	}()

	// Keep connection alive
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
}

func broadcastTelemetry(data TelemetryData) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling telemetry: %v", err)
		return
	}

	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("Error sending to client: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

var (
	useMockData   bool
	useTestTCP    bool
	tcpAddr       string
	enableCapture bool
	captureFile   string
)

func init() {
	flag.BoolVar(&useMockData, "mock-data", false, "Use mock data instead of real OBD2/CAN connection")
	flag.BoolVar(&useTestTCP, "test-tcp", false, "Use TCP connection for testing")
	flag.StringVar(&tcpAddr, "tcp-addr", "localhost:6789", "TCP address for test connection")
	flag.BoolVar(&enableCapture, "capture", false, "Enable data capture")
	flag.StringVar(&captureFile, "capture-file", "", "Specify custom capture file name")
	flag.Parse()
}

// Custom commands currently not supported by elmobd
func executeCustomCommand(dev *elmobd.Device, mode, pid string) (string, error) {
	return "", fmt.Errorf("custom commands not supported")
}

func getECUInfo(dev *elmobd.Device) (*ECUInfo, error) {
	info := &ECUInfo{}

	// Get VIN
	if vin, err := executeCustomCommand(dev, "09", "02"); err == nil {
		info.VIN = strings.TrimSpace(vin)
	}

	// Get ECU info
	if ecuData, err := executeCustomCommand(dev, "09", "0A"); err == nil {
		parts := strings.Split(ecuData, "\n")
		if len(parts) >= 3 {
			info.Version = parts[0]
			info.Calibration = parts[1]
			info.Software = parts[2]
		}
	}

	// Get hardware number
	if hwNum, err := executeCustomCommand(dev, "09", "0B"); err == nil {
		info.Hardware = strings.TrimSpace(hwNum)
	}

	// Get protocol info
	if protocol, err := executeCustomCommand(dev, "09", "0C"); err == nil {
		info.Protocol = strings.TrimSpace(protocol)
	}

	// Get build date
	if date, err := executeCustomCommand(dev, "09", "0D"); err == nil {
		info.BuildDate = strings.TrimSpace(date)
	}

	return info, nil
}

func getEngineMaps(dev *elmobd.Device) (*EngineMaps, error) {
	maps := &EngineMaps{
		Fuel: &MapData{
			Values: make([][]float64, 16),
			XAxis:  make([]float64, 16),
			YAxis:  make([]float64, 16),
		},
		Timing: &MapData{
			Values: make([][]float64, 16),
			XAxis:  make([]float64, 16),
			YAxis:  make([]float64, 16),
		},
	}

	// Get fuel map data
	for i := 0; i < 16; i++ {
		maps.Fuel.XAxis[i] = float64(i) * 500  // RPM steps
		maps.Fuel.YAxis[i] = float64(i) * 6.25 // Load steps

		maps.Fuel.Values[i] = make([]float64, 16)
		for j := 0; j < 16; j++ {
			// Command format: 090E XX YY (where XX is RPM index and YY is load index)
			if val, err := executeCustomCommand(dev, "09", fmt.Sprintf("0E%02X%02X", i, j)); err == nil {
				if f, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
					maps.Fuel.Values[i][j] = f
				}
			}
		}
	}

	// Get timing map data
	for i := 0; i < 16; i++ {
		maps.Timing.XAxis[i] = float64(i) * 500  // RPM steps
		maps.Timing.YAxis[i] = float64(i) * 6.25 // Load steps

		maps.Timing.Values[i] = make([]float64, 16)
		for j := 0; j < 16; j++ {
			// Command format: 090F XX YY (where XX is RPM index and YY is load index)
			if val, err := executeCustomCommand(dev, "09", fmt.Sprintf("0F%02X%02X", i, j)); err == nil {
				if f, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
					maps.Timing.Values[i][j] = f
				}
			}
		}
	}

	return maps, nil
}

func main() {
	// Initialize HTTP server
	router := mux.NewRouter()
	router.HandleFunc("/ws", wsHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	go func() {
		log.Printf("Starting web server on http://localhost:8080")
		if err := http.ListenAndServe(":8080", router); err != nil {
			log.Fatal(err)
		}
	}()

	// Initialize OBD connection
	var config transport.Config

	if useTestTCP {
		config = transport.Config{
			Type:    "tcp",
			Address: tcpAddr,
		}
	} else if useMockData {
		config = transport.Config{
			Type: "mock",
		}
	} else {
		config = transport.Config{
			Type:     "serial",
			Address:  "COM1", // Adjust port based on your Windows setup
			BaudRate: 38400,
		}
	}

	device, err := elmobd.NewDevice(config.Address, true) // true = debug mode
	if err != nil {
		log.Fatal(err)
	}

	// Initialize CAN bus if available
	var canBus *can.Bus
	var frameChan chan CANFrame

	if bus, err := can.NewBusForInterfaceWithName("can0"); err == nil {
		canBus = bus
		frameChan = make(chan CANFrame, 100) // Buffer up to 100 frames

		// Create and subscribe the CAN frame handler
		handler := &CANHandler{frameChan: frameChan}
		canBus.Subscribe(handler)

		// Start processing received frames
		go func() {
			defer canBus.Disconnect()
			log.Printf("CAN bus handler started")
		}()
	} else {
		log.Printf("CAN bus not available: %v", err)
	}

	// Get initial ECU info and engine maps
	ecuInfo, err := getECUInfo(device)
	if err != nil {
		log.Printf("Warning: Failed to get ECU info: %v", err)
	}

	engineMaps, err := getEngineMaps(device)
	if err != nil {
		log.Printf("Warning: Failed to get engine maps: %v", err)
	}

	// Start periodic ECU info and maps update
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if info, err := getECUInfo(device); err == nil {
				ecuInfo = info
			}
			if maps, err := getEngineMaps(device); err == nil {
				engineMaps = maps
			}
		}
	}()

	// Start telemetry collection in a separate goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			telemetry := TelemetryData{}

			// Read RPM
			if cmd, err := device.RunOBDCommand(elmobd.NewEngineRPM()); err == nil {
				if rpm, ok := cmd.(*elmobd.EngineRPM); ok {
					telemetry.RPM = float64(rpm.Value)
				}
			}

			// Read Speed
			if cmd, err := device.RunOBDCommand(elmobd.NewVehicleSpeed()); err == nil {
				if speed, ok := cmd.(*elmobd.VehicleSpeed); ok {
					telemetry.Speed = float64(speed.Value)
				}
			}

			// Read Engine Temperature
			if cmd, err := device.RunOBDCommand(elmobd.NewCoolantTemperature()); err == nil {
				if temp, ok := cmd.(*elmobd.CoolantTemperature); ok {
					telemetry.Temp = float64(temp.Value)
				}
			}

			// Read DTCs - not currently supported directly by elmobd
			telemetry.DTCs = []string{}

			// Add ECU info and engine maps to telemetry
			telemetry.ECUInfo = ecuInfo
			telemetry.EngineMaps = engineMaps

			// Add any received CAN frames
			if frameChan != nil {
				// Non-blocking read of all available frames
				for {
					select {
					case frame := <-frameChan:
						telemetry.CANFrames = append(telemetry.CANFrames, frame)
					default:
						goto done
					}
				}
			done:
			}

			broadcastTelemetry(telemetry)
		}
	}()

	// If CAN bus is available, send some test frames
	if canBus != nil {
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				frame := can.Frame{
					ID:    0x123,                           // Example ID
					Data:  [8]byte{0x54, 0x45, 0x53, 0x54}, // "TEST" in hex
					Flags: 0,                               // Standard frame
				}
				if err := canBus.Publish(frame); err != nil {
					log.Printf("Error sending CAN frame: %v", err)
				}
			}
		}()
	}

	// Block main thread
	select {}
}
