package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/brutella/can"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rzetterberg/elmobd"

	"github.com/anodyne74/iload-obd2/internal/capture"
	"github.com/anodyne74/iload-obd2/internal/config"
	"github.com/anodyne74/iload-obd2/internal/transport"
	"github.com/anodyne74/iload-obd2/pkg/logger"
)

var (
	version    = "dev"
	configPath = flag.String("config", "config.yaml", "path to config file")
	debug      = flag.Bool("debug", false, "enable debug logging")
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
	RPM        float64       `json:"rpm,omitempty"`
	Speed      float64       `json:"speed,omitempty"`
	Temp       float64       `json:"temp,omitempty"`
	DTCs       []string      `json:"dtcs,omitempty"`
	ECUInfo    *ECUInfo      `json:"ecuInfo,omitempty"`
	EngineMaps *EngineMaps   `json:"engineMaps,omitempty"`
	CANFrames  []CANFrame    `json:"canFrames,omitempty"`
	Capture    *CaptureState `json:"capture,omitempty"`
}

type CaptureState struct {
	Status          string `json:"status"`
	FrameCount      int    `json:"frameCount,omitempty"`
	DurationSeconds int64  `json:"durationSeconds,omitempty"`
}

type CaptureCommand struct {
	Command string `json:"command"`
	Label   string `json:"label,omitempty"`
}

type CaptureControlResponse struct {
	Type            string `json:"type"`
	Status          string `json:"status,omitempty"`
	Message         string `json:"message,omitempty"`
	Error           string `json:"error,omitempty"`
	FilePath        string `json:"filePath,omitempty"`
	FrameCount      int    `json:"frameCount,omitempty"`
	DurationSeconds int64  `json:"durationSeconds,omitempty"`
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

var (
	clients    = make(map[*websocket.Conn]bool)
	clientsMux sync.Mutex
	recorder   *capture.Recorder
	recorderMu sync.Mutex
	recordedAt time.Time
)

func currentCaptureState() *CaptureState {
	recorderMu.Lock()
	defer recorderMu.Unlock()

	if recorder == nil || !recorder.IsRunning() {
		return &CaptureState{Status: "idle"}
	}

	return &CaptureState{
		Status:          "recording",
		FrameCount:      recorder.FrameCount(),
		DurationSeconds: int64(time.Since(recordedAt).Seconds()),
	}
}

func writeCaptureResponse(ws *websocket.Conn, resp CaptureControlResponse) {
	payload, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling capture response: %v", err)
		return
	}

	if err := ws.WriteMessage(websocket.TextMessage, payload); err != nil {
		log.Printf("Error sending capture response: %v", err)
	}
}

func handleCaptureCommand(ws *websocket.Conn, cmd CaptureCommand) {
	command := strings.ToLower(strings.TrimSpace(cmd.Command))

	switch command {
	case "start", "start_capture":
		recorderMu.Lock()
		if recorder != nil && recorder.IsRunning() {
			recorderMu.Unlock()
			writeCaptureResponse(ws, CaptureControlResponse{
				Type:   "capture_control",
				Status: "error",
				Error:  "capture is already running",
			})
			return
		}

		newRecorder := capture.NewRecorder("Hyundai iLoad")
		newRecorder.SetMetadata("source", "app")
		newRecorder.SetMetadata("transport", "ble")
		if cmd.Label != "" {
			newRecorder.SetMetadata("label", cmd.Label)
		}

		if err := newRecorder.Start(); err != nil {
			recorderMu.Unlock()
			writeCaptureResponse(ws, CaptureControlResponse{
				Type:   "capture_control",
				Status: "error",
				Error:  err.Error(),
			})
			return
		}

		recorder = newRecorder
		recordedAt = time.Now()
		recorderMu.Unlock()

		writeCaptureResponse(ws, CaptureControlResponse{
			Type:    "capture_control",
			Status:  "recording",
			Message: "capture started",
		})

	case "stop", "stop_capture":
		recorderMu.Lock()
		active := recorder
		if active == nil || !active.IsRunning() {
			recorderMu.Unlock()
			writeCaptureResponse(ws, CaptureControlResponse{
				Type:   "capture_control",
				Status: "error",
				Error:  "capture is not running",
			})
			return
		}
		frameCount := active.FrameCount()
		duration := int64(time.Since(recordedAt).Seconds())
		recorder = nil
		recorderMu.Unlock()

		if err := active.Stop(); err != nil {
			writeCaptureResponse(ws, CaptureControlResponse{
				Type:   "capture_control",
				Status: "error",
				Error:  err.Error(),
			})
			return
		}

		writeCaptureResponse(ws, CaptureControlResponse{
			Type:            "capture_control",
			Status:          "stopped",
			Message:         "capture stopped",
			FilePath:        active.SessionPath(),
			FrameCount:      frameCount,
			DurationSeconds: duration,
		})

	case "status", "capture_status":
		state := currentCaptureState()
		writeCaptureResponse(ws, CaptureControlResponse{
			Type:            "capture_control",
			Status:          state.Status,
			FrameCount:      state.FrameCount,
			DurationSeconds: state.DurationSeconds,
		})

	default:
		writeCaptureResponse(ws, CaptureControlResponse{
			Type:   "capture_control",
			Status: "error",
			Error:  "unknown command",
		})
	}
}

func recordTelemetryFrame(frame capture.Frame) {
	recorderMu.Lock()
	active := recorder
	recorderMu.Unlock()

	if active == nil || !active.IsRunning() {
		return
	}

	if err := active.Record(frame); err != nil {
		log.Printf("Error recording frame: %v", err)
	}
}

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

	// Keep connection alive and process optional control commands.
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var command CaptureCommand
		if err := json.Unmarshal(message, &command); err != nil {
			continue
		}

		if command.Command == "" {
			continue
		}

		handleCaptureCommand(ws, command)
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
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
	flag.Parse()
}

// sendInfoRequest sends an OBD-II request for vehicle information
func sendInfoRequest(bus *can.Bus, mode, pid byte) error {
	frame := can.Frame{
		ID:    0x7DF, // Standard OBD-II diagnostic request
		Data:  [8]byte{0x02, mode, pid, 0x00, 0x00, 0x00, 0x00, 0x00},
		Flags: 0,
	}
	return bus.Publish(frame)
}

// processInfoResponse processes response frames for vehicle information
func processInfoResponse(frame can.Frame, mode byte) (string, error) {
	if frame.ID != 0x7E8 { // Standard ECU response ID
		return "", fmt.Errorf("unexpected response ID: %X", frame.ID)
	}

	// First byte is the number of additional bytes
	numBytes := frame.Data[0]
	if numBytes < 2 || frame.Data[1] != (0x40|mode) { // Response mode is request mode + 0x40
		return "", fmt.Errorf("invalid response format")
	}

	// Extract the data bytes
	data := make([]byte, 0, numBytes-2)
	for i := 2; i < int(numBytes); i++ {
		if frame.Data[i] != 0 {
			data = append(data, frame.Data[i])
		}
	}

	return string(data), nil
}

func getECUInfo(bus *can.Bus, frameChan chan CANFrame) (*ECUInfo, error) {
	if bus == nil {
		return nil, fmt.Errorf("CAN bus not available")
	}

	info := &ECUInfo{}

	// Helper function to request and receive info
	getInfo := func(mode, pid byte) (string, error) {
		if err := sendInfoRequest(bus, mode, pid); err != nil {
			return "", err
		}

		// Wait up to 100ms for response
		timeout := time.After(100 * time.Millisecond)
		select {
		case frame := <-frameChan:
			return processInfoResponse(can.Frame{
				ID:    frame.ID,
				Data:  [8]byte(frame.Data),
				Flags: 0,
			}, mode)
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for response")
		}
	}

	// Get VIN (Mode 09, PID 02)
	if vin, err := getInfo(0x09, 0x02); err == nil {
		info.VIN = strings.TrimSpace(vin)
	}

	// Get ECU info (Mode 09, PID 0A)
	if ecuVer, err := getInfo(0x09, 0x0A); err == nil {
		info.Version = strings.TrimSpace(ecuVer)
	}

	// Get calibration ID (Mode 09, PID 04)
	if calID, err := getInfo(0x09, 0x04); err == nil {
		info.Calibration = strings.TrimSpace(calID)
	}

	// Get ECU name (Mode 09, PID 0A)
	if ecuName, err := getInfo(0x09, 0x0A); err == nil {
		info.Software = strings.TrimSpace(ecuName)
	}

	// Get protocol version (Mode 09, PID 0C)
	if proto, err := getInfo(0x09, 0x0C); err == nil {
		info.Protocol = strings.TrimSpace(proto)
	}

	return info, nil
}

func getEngineMaps(bus *can.Bus, frameChan chan CANFrame) (*EngineMaps, error) {
	if bus == nil {
		return nil, fmt.Errorf("CAN bus not available")
	}

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

	// Helper function to request and receive map data
	getMapValue := func(pid byte, x, y byte) (float64, error) {
		frame := can.Frame{
			ID:    0x7DF,
			Data:  [8]byte{0x04, 0x09, pid, x, y, 0x00, 0x00, 0x00},
			Flags: 0,
		}

		if err := bus.Publish(frame); err != nil {
			return 0, err
		}

		timeout := time.After(100 * time.Millisecond)
		select {
		case frame := <-frameChan:
			if frame.ID != 0x7E8 {
				return 0, fmt.Errorf("unexpected response ID: %X", frame.ID)
			}
			if frame.Data[0] < 5 || frame.Data[1] != 0x49 { // 0x49 is response to mode 09, and need at least 5 bytes
				return 0, fmt.Errorf("invalid response format: insufficient data length")
			}
			// Convert 2 bytes to float64
			value := float64(uint16(frame.Data[3])<<8 | uint16(frame.Data[4]))
			return value / 100.0, nil // Scale factor for map values
		case <-timeout:
			return 0, fmt.Errorf("timeout waiting for response")
		}
	}

	// Initialize axis values
	for i := 0; i < 16; i++ {
		maps.Fuel.XAxis[i] = float64(i) * 500  // RPM steps
		maps.Fuel.YAxis[i] = float64(i) * 6.25 // Load steps
		maps.Timing.XAxis[i] = float64(i) * 500
		maps.Timing.YAxis[i] = float64(i) * 6.25
	}

	// Get fuel map data (PID 0E)
	for i := 0; i < 16; i++ {
		maps.Fuel.Values[i] = make([]float64, 16)
		for j := 0; j < 16; j++ {
			if val, err := getMapValue(0x0E, byte(i), byte(j)); err == nil {
				maps.Fuel.Values[i][j] = val
			}
		}
	}

	// Get timing map data (PID 0F)
	for i := 0; i < 16; i++ {
		maps.Timing.Values[i] = make([]float64, 16)
		for j := 0; j < 16; j++ {
			if val, err := getMapValue(0x0F, byte(i), byte(j)); err == nil {
				maps.Timing.Values[i][j] = val
			}
		}
	}

	return maps, nil
}

// DTCRequest sends a diagnostic trouble code request over CAN
func sendDTCRequest(bus *can.Bus) error {
	// Mode 03 request for DTCs
	frame := can.Frame{
		ID:    0x7DF, // Standard OBD-II diagnostic request
		Data:  [8]byte{0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Flags: 0,
	}
	return bus.Publish(frame)
}

// DTCResponse processes DTC response frames
func processDTCResponse(frame can.Frame) []string {
	if frame.ID != 0x7E8 { // Standard ECU response ID
		return nil
	}

	// First byte is the number of additional bytes
	numBytes := frame.Data[0]
	if numBytes < 2 || frame.Data[1] != 0x43 { // 0x43 is response to mode 03
		return nil
	}

	var dtcs []string
	// Process pairs of bytes starting from position 2
	for i := 2; i < int(numBytes) && i+1 < 8; i += 2 {
		if frame.Data[i] == 0 && frame.Data[i+1] == 0 {
			continue
		}

		// Convert two bytes into a DTC
		dtc := decodeDTC(frame.Data[i], frame.Data[i+1])
		if dtc != "" {
			dtcs = append(dtcs, dtc)
		}
	}
	return dtcs
}

// decodeDTC converts two bytes into a DTC string
func decodeDTC(b1, b2 byte) string {
	if b1 == 0 && b2 == 0 {
		return ""
	}

	// First nibble determines DTC type
	dtcType := ""
	switch b1 >> 6 {
	case 0:
		dtcType = "P" // Powertrain
	case 1:
		dtcType = "C" // Chassis
	case 2:
		dtcType = "B" // Body
	case 3:
		dtcType = "U" // Network
	}

	// Format remaining 14 bits as a single 4-digit hex value
	code := uint16(b1&0x3F)<<8 | uint16(b2)
	return fmt.Sprintf("%s%04X", dtcType, code)
}

func main() {
	flag.Parse()

	// Initialize logger
	if err := logger.Init(version, "/opt/iload-obd2/logs/iload-obd2.log", *debug); err != nil {
		os.Exit(1)
	}

	log := logger.Get("main")
	log.Infow("Starting iLoad-OBD2",
		"version", version,
		"config", *configPath,
	)

	// Initialize HTTP server
	router := mux.NewRouter()
	router.HandleFunc("/ws", wsHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Get server configuration
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	go func() {
		log.Printf("Starting web server on http://%s", serverAddr)
		if err := http.ListenAndServe(serverAddr, router); err != nil {
			log.Fatal(err)
		}
	}()

	// Initialize OBD connection
	transportConfig := cfg.GetTransportConfig()
	device, err := transport.NewDevice(transportConfig)
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

	// Get initial ECU info and engine maps if CAN is available
	var ecuInfo *ECUInfo
	var engineMaps *EngineMaps

	if canBus != nil {
		var err error
		ecuInfo, err = getECUInfo(canBus, frameChan)
		if err != nil {
			log.Printf("Warning: Failed to get ECU info: %v", err)
		}

		engineMaps, err = getEngineMaps(canBus, frameChan)
		if err != nil {
			log.Printf("Warning: Failed to get engine maps: %v", err)
		}

		// Start periodic ECU info and maps update
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				if info, err := getECUInfo(canBus, frameChan); err == nil {
					ecuInfo = info
				}
				if maps, err := getEngineMaps(canBus, frameChan); err == nil {
					engineMaps = maps
				}
			}
		}()
	} else {
		log.Println("Warning: CAN bus not available, ECU info and maps will not be available")
	}

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

			// Read DTCs via CAN if available
			if canBus != nil {
				dtcs := []string{}

				// Send DTC request
				if err := sendDTCRequest(canBus); err != nil {
					log.Printf("Error sending DTC request: %v", err)
				} else {
					// Wait up to 100ms for response
					timeout := time.After(100 * time.Millisecond)
					timeoutReached := false
					for !timeoutReached {
						select {
						case frame := <-frameChan:
							if newDTCs := processDTCResponse(can.Frame{
								ID:    frame.ID,
								Data:  [8]byte(frame.Data),
								Flags: 0,
							}); newDTCs != nil {
								dtcs = append(dtcs, newDTCs...)
							}
						case <-timeout:
							timeoutReached = true
						}
					}
				}
				telemetry.DTCs = dtcs
			}

			// Add ECU info and engine maps to telemetry
			telemetry.ECUInfo = ecuInfo
			telemetry.EngineMaps = engineMaps
			telemetry.Capture = currentCaptureState()

			recordTelemetryFrame(capture.Frame{
				Timestamp: time.Now(),
				Type:      "OBD2",
				Decoded: map[string]interface{}{
					"rpm":   telemetry.RPM,
					"speed": telemetry.Speed,
					"temp":  telemetry.Temp,
					"dtcs":  telemetry.DTCs,
				},
			})

			if telemetry.ECUInfo != nil && telemetry.ECUInfo.VIN != "" {
				recorderMu.Lock()
				if recorder != nil && recorder.IsRunning() {
					recorder.SetMetadata("vin", telemetry.ECUInfo.VIN)
				}
				recorderMu.Unlock()
			}

			// Add any received CAN frames
			if frameChan != nil {
				// Non-blocking read of all available frames
				for {
					select {
					case frame := <-frameChan:
						telemetry.CANFrames = append(telemetry.CANFrames, frame)
						recordTelemetryFrame(capture.Frame{
							Timestamp: frame.Timestamp,
							Type:      "CAN",
							ID:        frame.ID,
							Data:      frame.Data,
						})
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

	// Set up clean shutdown
	stop := make(chan struct{})
	done := make(chan struct{})

	// Handle graceful shutdown
	go func() {
		defer close(done)
		<-stop

		recorderMu.Lock()
		active := recorder
		recorder = nil
		recorderMu.Unlock()
		if active != nil && active.IsRunning() {
			if err := active.Stop(); err != nil {
				log.Printf("Error stopping active capture during shutdown: %v", err)
			} else {
				log.Printf("Capture finalized at %s", active.SessionPath())
			}
		}

		// Clean up websocket connections
		clientsMux.Lock()
		for client := range clients {
			client.Close()
			delete(clients, client)
		}
		clientsMux.Unlock()

		// Clean up CAN bus if available
		if canBus != nil {
			canBus.Disconnect()
		}

		// Note: elmobd.Device doesn't have a Close method,
		// but the underlying serial/TCP connection will be closed when the program exits

		log.Println("Cleanup completed")
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Initiate shutdown
	close(stop)
	<-done // Wait for cleanup to complete
}
