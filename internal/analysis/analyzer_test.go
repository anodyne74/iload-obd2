package analysis

import (
	"math"
	"testing"
	"time"

	"iload-obd2/internal/capture"
)

func TestAnalyzer(t *testing.T) {
	// Create a test session
	now := time.Now()
	session := &capture.Session{
		StartTime:   now,
		EndTime:     now.Add(10 * time.Second),
		VehicleInfo: "TEST12345 Test Model 2023",
		Frames: []capture.Frame{
			// Idle phase
			{
				Type:      "OBD2",
				Timestamp: now,
				Decoded: map[string]interface{}{
					"rpm":   800.0,
					"speed": 0.0,
					"temp":  90.0,
				},
			},
			// Acceleration phase
			{
				Type:      "OBD2",
				Timestamp: now.Add(2 * time.Second),
				Decoded: map[string]interface{}{
					"rpm":   2500.0,
					"speed": 20.0,
					"temp":  92.0,
				},
			},
			// Cruise phase
			{
				Type:      "OBD2",
				Timestamp: now.Add(4 * time.Second),
				Decoded: map[string]interface{}{
					"rpm":   2000.0,
					"speed": 60.0,
					"temp":  95.0,
				},
			},
			// Deceleration phase
			{
				Type:      "OBD2",
				Timestamp: now.Add(6 * time.Second),
				Decoded: map[string]interface{}{
					"rpm":   1500.0,
					"speed": 30.0,
					"temp":  93.0,
				},
			},
			// CAN frame
			{
				Type:      "CAN",
				Timestamp: now.Add(8 * time.Second),
				ID:        0x7E8,
				Data:      []byte{0x02, 0x41, 0x0D, 0x45, 0x00, 0x00, 0x00, 0x00},
			},
		},
	}

	// Create analyzer with default options
	analyzer := NewAnalyzer(session, DefaultOptions())

	// Run analysis
	analysis, err := analyzer.Analyze()
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	// Test session info
	if analysis.SessionInfo.Duration != 10*time.Second {
		t.Errorf("Expected duration 10s, got %v", analysis.SessionInfo.Duration)
	}
	if analysis.SessionInfo.TotalFrames != 5 {
		t.Errorf("Expected 5 frames, got %d", analysis.SessionInfo.TotalFrames)
	}

	// Test performance stats
	if analysis.Performance.Speed.Max != 60.0 {
		t.Errorf("Expected max speed 60.0, got %f", analysis.Performance.Speed.Max)
	}
	if analysis.Performance.RPM.Min != 800.0 {
		t.Errorf("Expected min RPM 800.0, got %f", analysis.Performance.RPM.Min)
	}

	// Test driving behavior
	if analysis.DrivingBehavior.RapidAccel == 0 {
		t.Error("Expected at least one rapid acceleration")
	}
	if analysis.DrivingBehavior.RapidDecel == 0 {
		t.Error("Expected at least one rapid deceleration")
	}

	// Test CAN activity
	if analysis.CANActivity.UniqueIDs != 1 {
		t.Errorf("Expected 1 unique CAN ID, got %d", analysis.CANActivity.UniqueIDs)
	}
}

func TestCalculateStats(t *testing.T) {
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	stats := CalculateStats(values)

	expected := Stats{
		Min:    1.0,
		Max:    5.0,
		Mean:   3.0,
		StdDev: 1.5811388300841898,
	}

	if stats.Min != expected.Min {
		t.Errorf("Expected min %f, got %f", expected.Min, stats.Min)
	}
	if stats.Max != expected.Max {
		t.Errorf("Expected max %f, got %f", expected.Max, stats.Max)
	}
	if stats.Mean != expected.Mean {
		t.Errorf("Expected mean %f, got %f", expected.Mean, stats.Mean)
	}
	if math.Abs(stats.StdDev-expected.StdDev) > 0.0001 {
		t.Errorf("Expected stddev %f, got %f", expected.StdDev, stats.StdDev)
	}
}
