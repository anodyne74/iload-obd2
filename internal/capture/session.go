package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Frame represents a captured data frame
type Frame struct {
	Timestamp time.Time   `json:"timestamp"`
	Type      string      `json:"type"`         // "OBD2" or "CAN"
	ID        uint32      `json:"id,omitempty"` // CAN ID if applicable
	Data      []byte      `json:"data"`         // Raw frame data
	Decoded   interface{} `json:"decoded"`      // Decoded data (if available)
}

// Session represents a capture session
type Session struct {
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time,omitempty"`
	VehicleInfo string            `json:"vehicle_info"`
	Frames      []Frame           `json:"frames"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	filePath    string            // Path where session will be saved
}

// NewSession creates a new capture session
func NewSession(vehicleInfo string) *Session {
	return &Session{
		StartTime:   time.Now(),
		VehicleInfo: vehicleInfo,
		Frames:      make([]Frame, 0),
		Metadata:    make(map[string]string),
	}
}

// AddFrame adds a frame to the session
func (s *Session) AddFrame(frame Frame) {
	s.Frames = append(s.Frames, frame)
}

// SetMetadata adds or updates metadata
func (s *Session) SetMetadata(key, value string) {
	s.Metadata[key] = value
}

// Save writes the session to disk
func (s *Session) Save() error {
	if s.filePath == "" {
		// Generate default filename if none specified
		timestamp := time.Now().Format("20060102_150405")
		s.filePath = filepath.Join("captures", fmt.Sprintf("session_%s.json", timestamp))
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Set end time
	s.EndTime = time.Now()

	// Marshal to JSON
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}
