package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type CANFrame struct {
	Timestamp int64  `json:"timestamp"`
	ID        uint32 `json:"id"`
	Data      []byte `json:"data"`
	Type      string `json:"type"` // "CAN" or "OBD2"
}

type Session struct {
	StartTime   int64      `json:"start_time"`
	EndTime     int64      `json:"end_time"`
	VehicleInfo string     `json:"vehicle_info"`
	Frames      []CANFrame `json:"frames"`
	filename    string
	file        *os.File
	encoder     *json.Encoder
}

func NewSession(vehicleInfo string) (*Session, error) {
	timestamp := time.Now().Unix()
	filename := filepath.Join("captures", fmt.Sprintf("capture_%d.json", timestamp))

	// Ensure captures directory exists
	if err := os.MkdirAll("captures", 0755); err != nil {
		return nil, fmt.Errorf("failed to create captures directory: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create capture file: %v", err)
	}

	session := &Session{
		StartTime:   timestamp,
		VehicleInfo: vehicleInfo,
		filename:    filename,
		file:        file,
		encoder:     json.NewEncoder(file),
	}

	// Write initial session info
	if err := session.encoder.Encode(session); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write session header: %v", err)
	}

	return session, nil
}

func (s *Session) CaptureFrame(id uint32, data []byte, frameType string) error {
	frame := CANFrame{
		Timestamp: time.Now().UnixNano(),
		ID:        id,
		Data:      data,
		Type:      frameType,
	}

	if err := s.encoder.Encode(frame); err != nil {
		return fmt.Errorf("failed to encode frame: %v", err)
	}

	s.Frames = append(s.Frames, frame)
	return nil
}

func (s *Session) Close() error {
	s.EndTime = time.Now().Unix()
	if err := s.encoder.Encode(s); err != nil {
		return fmt.Errorf("failed to write session footer: %v", err)
	}
	return s.file.Close()
}

func LoadSession(filename string) (*Session, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open capture file: %v", err)
	}
	defer file.Close()

	var session Session
	if err := json.NewDecoder(file).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode session: %v", err)
	}

	return &session, nil
}
