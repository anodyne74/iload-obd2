package analysis

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/anodyne74/iload-obd2/capture"
)

type SessionAnalyzer struct {
	Session *capture.Session
}

type AnalysisMetrics struct {
	TotalFrames     int
	UniqueIDs       map[uint32]int
	MaxRPM          float64
	AvgRPM          float64
	MaxSpeed        float64
	AvgSpeed        float64
	TempRange       [2]float64 // [min, max]
	DTCFrequency    map[string]int
	DataRatePerSec  float64
	IdlePercentage  float64
	AccelEvents     int    // Rapid acceleration events
	DecelEvents     int    // Rapid deceleration events
	DrivingDuration string // Total time vehicle was moving
}

func NewAnalyzer(session *capture.Session) *SessionAnalyzer {
	return &SessionAnalyzer{Session: session}
}

func (sa *SessionAnalyzer) AnalyzeSession() (*AnalysisMetrics, error) {
	metrics := &AnalysisMetrics{
		UniqueIDs:    make(map[uint32]int),
		DTCFrequency: make(map[string]int),
	}

	var (
		rpmSum, speedSum, tempSum       float64
		rpmCount, speedCount, tempCount int
		lastSpeed                       float64
		lastTime                        int64
	)

	metrics.TotalFrames = len(sa.Session.Frames)

	for i, frame := range sa.Session.Frames {
		// Count unique CAN IDs
		metrics.UniqueIDs[frame.ID]++

		// Analyze frame data based on ID
		switch frame.ID {
		case 0x7E8: // RPM data
			if rpm := decodeRPM(frame.Data); rpm > 0 {
				metrics.MaxRPM = math.Max(metrics.MaxRPM, rpm)
				rpmSum += rpm
				rpmCount++
			}
		case 0x7E9: // Speed data
			if speed := decodeSpeed(frame.Data); speed >= 0 {
				metrics.MaxSpeed = math.Max(metrics.MaxSpeed, speed)
				speedSum += speed
				speedCount++

				// Detect acceleration/deceleration events
				if i > 0 {
					timeDiff := float64(frame.Timestamp-lastTime) / float64(time.Second)
					speedDiff := speed - lastSpeed
					if timeDiff > 0 {
						acceleration := speedDiff / timeDiff
						if acceleration > 7.0 { // More than 7 m/s²
							metrics.AccelEvents++
						} else if acceleration < -7.0 {
							metrics.DecelEvents++
						}
					}
				}
				lastSpeed = speed
				lastTime = frame.Timestamp
			}
		}
	}

	// Calculate averages
	if rpmCount > 0 {
		metrics.AvgRPM = rpmSum / float64(rpmCount)
	}
	if speedCount > 0 {
		metrics.AvgSpeed = speedSum / float64(speedCount)
	}

	// Calculate data rate
	duration := float64(sa.Session.EndTime - sa.Session.StartTime)
	if duration > 0 {
		metrics.DataRatePerSec = float64(metrics.TotalFrames) / duration
	}

	// Calculate idle percentage (RPM < 1000)
	idleTime := 0
	for _, frame := range sa.Session.Frames {
		if frame.ID == 0x7E8 {
			if rpm := decodeRPM(frame.Data); rpm > 0 && rpm < 1000 {
				idleTime++
			}
		}
	}
	metrics.IdlePercentage = float64(idleTime) / float64(metrics.TotalFrames) * 100

	return metrics, nil
}

func (sa *SessionAnalyzer) ExportToCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Timestamp", "Frame ID", "Data Type", "Value", "Unit"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, frame := range sa.Session.Frames {
		timestamp := time.Unix(0, frame.Timestamp).Format(time.RFC3339)

		switch frame.ID {
		case 0x7E8: // RPM
			rpm := decodeRPM(frame.Data)
			record := []string{timestamp, fmt.Sprintf("0x%X", frame.ID), "RPM", fmt.Sprintf("%.2f", rpm), "rpm"}
			if err := writer.Write(record); err != nil {
				return err
			}
		case 0x7E9: // Speed
			speed := decodeSpeed(frame.Data)
			record := []string{timestamp, fmt.Sprintf("0x%X", frame.ID), "Speed", fmt.Sprintf("%.2f", speed), "km/h"}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

// Helper functions to decode CAN data
func decodeRPM(data []byte) float64 {
	if len(data) < 2 {
		return 0
	}
	return float64(uint16(data[0])<<8|uint16(data[1])) / 4
}

func decodeSpeed(data []byte) float64 {
	if len(data) < 1 {
		return -1
	}
	return float64(data[0])
}

// Advanced analysis features
func (sa *SessionAnalyzer) GenerateDrivingProfile() (map[string]interface{}, error) {
	profile := make(map[string]interface{})

	// Analyze driving patterns
	var (
		accelerationPhases int
		decelerationPhases int
		cruisingPhases     int
		idlePhases         int
		lastSpeed          float64
		phaseStart         time.Time
		currentPhase       string
	)

	for i, frame := range sa.Session.Frames {
		if frame.ID != 0x7E9 { // Speed frame
			continue
		}

		speed := decodeSpeed(frame.Data)
		timestamp := time.Unix(0, frame.Timestamp)

		if i == 0 {
			phaseStart = timestamp
			if speed < 1 {
				currentPhase = "idle"
			} else {
				currentPhase = "cruising"
			}
			continue
		}

		// Detect phase changes
		switch {
		case speed > lastSpeed+5:
			if currentPhase != "acceleration" {
				accelerationPhases++
				updatePhaseStats(profile, currentPhase, phaseStart, timestamp)
				currentPhase = "acceleration"
				phaseStart = timestamp
			}
		case speed < lastSpeed-5:
			if currentPhase != "deceleration" {
				decelerationPhases++
				updatePhaseStats(profile, currentPhase, phaseStart, timestamp)
				currentPhase = "deceleration"
				phaseStart = timestamp
			}
		case speed < 1:
			if currentPhase != "idle" {
				idlePhases++
				updatePhaseStats(profile, currentPhase, phaseStart, timestamp)
				currentPhase = "idle"
				phaseStart = timestamp
			}
		default:
			if currentPhase != "cruising" {
				cruisingPhases++
				updatePhaseStats(profile, currentPhase, phaseStart, timestamp)
				currentPhase = "cruising"
				phaseStart = timestamp
			}
		}

		lastSpeed = speed
	}

	profile["acceleration_phases"] = accelerationPhases
	profile["deceleration_phases"] = decelerationPhases
	profile["cruising_phases"] = cruisingPhases
	profile["idle_phases"] = idlePhases

	return profile, nil
}

func updatePhaseStats(profile map[string]interface{}, phase string, start, end time.Time) {
	key := phase + "_duration"
	duration := end.Sub(start).Seconds()
	if existing, ok := profile[key].(float64); ok {
		profile[key] = existing + duration
	} else {
		profile[key] = duration
	}
}
