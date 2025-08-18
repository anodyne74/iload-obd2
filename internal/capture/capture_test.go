package capture

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	vehicleInfo := "Test Vehicle"
	session := NewSession(vehicleInfo)

	if session.VehicleInfo != vehicleInfo {
		t.Errorf("Expected vehicle info %s, got %s", vehicleInfo, session.VehicleInfo)
	}

	if session.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}

	if len(session.Frames) != 0 {
		t.Error("Expected empty frames slice")
	}
}

func TestAddFrame(t *testing.T) {
	session := NewSession("Test Vehicle")
	frame := Frame{
		Timestamp: time.Now(),
		Type:      "TEST",
		Data:      []byte{0x01, 0x02, 0x03},
	}

	session.AddFrame(frame)

	if len(session.Frames) != 1 {
		t.Error("Expected one frame in session")
	}

	if session.Frames[0].Type != frame.Type {
		t.Errorf("Expected frame type %s, got %s", frame.Type, session.Frames[0].Type)
	}
}

func TestSaveSession(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "capture_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create session with known file path
	session := NewSession("Test Vehicle")
	session.filePath = filepath.Join(tempDir, "test_session.json")

	// Add some test data
	session.AddFrame(Frame{
		Timestamp: time.Now(),
		Type:      "TEST",
		Data:      []byte{0x01, 0x02, 0x03},
	})

	// Save session
	if err := session.Save(); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(session.filePath); os.IsNotExist(err) {
		t.Error("Expected session file to exist")
	}
}

func TestRecorder(t *testing.T) {
	recorder := NewRecorder("Test Vehicle")

	// Test starting recorder
	if err := recorder.Start(); err != nil {
		t.Fatalf("Failed to start recorder: %v", err)
	}

	if !recorder.IsRunning() {
		t.Error("Expected recorder to be running")
	}

	// Test recording frames
	frame := Frame{
		Timestamp: time.Now(),
		Type:      "TEST",
		Data:      []byte{0x01, 0x02, 0x03},
	}

	if err := recorder.Record(frame); err != nil {
		t.Errorf("Failed to record frame: %v", err)
	}

	// Test stopping recorder
	if err := recorder.Stop(); err != nil {
		t.Errorf("Failed to stop recorder: %v", err)
	}

	if recorder.IsRunning() {
		t.Error("Expected recorder to be stopped")
	}
}
