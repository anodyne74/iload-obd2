package analysis

import (
	"fmt"
	"math"
	"time"

	"github.com/anodyne74/iload-obd2/internal/capture"
)

// Analyzer processes capture sessions to generate analysis results
type Analyzer struct {
	session  *capture.Session
	analysis *Analysis
	options  AnalyzerOptions
}

// AnalyzerOptions configures the analysis process
type AnalyzerOptions struct {
	RapidAccelThreshold float64       // km/h/s for rapid acceleration detection
	RapidDecelThreshold float64       // km/h/s for rapid deceleration detection
	IdleSpeedThreshold  float64       // km/h below which is considered idle
	CruiseThreshold     float64       // km/h/s variance for cruise detection
	MinPhaseTime        time.Duration // minimum duration for a driving phase
}

// DefaultOptions returns sensible default analyzer options
func DefaultOptions() AnalyzerOptions {
	return AnalyzerOptions{
		RapidAccelThreshold: 10.0, // 10 km/h per second
		RapidDecelThreshold: -8.0, // -8 km/h per second
		IdleSpeedThreshold:  3.0,  // 3 km/h
		CruiseThreshold:     2.0,  // 2 km/h/s variance
		MinPhaseTime:        3 * time.Second,
	}
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(session *capture.Session, options AnalyzerOptions) *Analyzer {
	return &Analyzer{
		session:  session,
		analysis: &Analysis{},
		options:  options,
	}
}

// Analyze processes the session and returns analysis results
func (a *Analyzer) Analyze() (*Analysis, error) {
	if err := a.analyzeSessionInfo(); err != nil {
		return nil, fmt.Errorf("session info analysis failed: %w", err)
	}

	if err := a.analyzePerformance(); err != nil {
		return nil, fmt.Errorf("performance analysis failed: %w", err)
	}

	if err := a.analyzeDrivingBehavior(); err != nil {
		return nil, fmt.Errorf("driving behavior analysis failed: %w", err)
	}

	if err := a.analyzeCANActivity(); err != nil {
		return nil, fmt.Errorf("CAN activity analysis failed: %w", err)
	}

	if err := a.analyzeDiagnostics(); err != nil {
		return nil, fmt.Errorf("diagnostics analysis failed: %w", err)
	}

	return a.analysis, nil
}

func (a *Analyzer) analyzeSessionInfo() error {
	a.analysis.SessionInfo.StartTime = a.session.StartTime
	a.analysis.SessionInfo.EndTime = a.session.EndTime
	a.analysis.SessionInfo.Duration = a.session.EndTime.Sub(a.session.StartTime)
	a.analysis.SessionInfo.VehicleInfo = fmt.Sprintf("%v", a.session.VehicleInfo)
	a.analysis.SessionInfo.TotalFrames = len(a.session.Frames)
	return nil
}

func (a *Analyzer) analyzePerformance() error {
	var rpmValues, speedValues, tempValues []float64

	for _, frame := range a.session.Frames {
		switch frame.Type {
		case "OBD2":
			if decoded, ok := frame.Decoded.(map[string]interface{}); ok {
				if rpm, ok := decoded["rpm"].(float64); ok {
					rpmValues = append(rpmValues, rpm)
				}
				if speed, ok := decoded["speed"].(float64); ok {
					speedValues = append(speedValues, speed)
				}
				if temp, ok := decoded["temp"].(float64); ok {
					tempValues = append(tempValues, temp)
				}
			}
		}
	}

	a.analysis.Performance.RPM = CalculateStats(rpmValues)
	a.analysis.Performance.Speed = CalculateStats(speedValues)
	a.analysis.Performance.Temperature = CalculateStats(tempValues)

	// Calculate data rate
	duration := a.analysis.SessionInfo.Duration.Seconds()
	if duration > 0 {
		a.analysis.Performance.DataRate = float64(len(a.session.Frames)) / duration
	}

	return nil
}

func (a *Analyzer) analyzeDrivingBehavior() error {
	var currentPhase *DrivingPhase
	var lastSpeed float64
	var lastTime time.Time

	for _, frame := range a.session.Frames {
		if frame.Type != "OBD2" {
			continue
		}

		decoded, ok := frame.Decoded.(map[string]interface{})
		if !ok {
			continue
		}

		speed, ok := decoded["speed"].(float64)
		if !ok {
			continue
		}

		// Calculate acceleration
		if !lastTime.IsZero() {
			timeDiff := frame.Timestamp.Sub(lastTime).Seconds()
			if timeDiff > 0 {
				acceleration := (speed - lastSpeed) / timeDiff

				// Detect driving phase
				phaseType := a.detectPhaseType(speed, acceleration)

				if currentPhase == nil || currentPhase.Type != phaseType {
					// Start new phase
					if currentPhase != nil {
						currentPhase.EndTime = frame.Timestamp
						currentPhase.Duration = currentPhase.EndTime.Sub(currentPhase.StartTime)
						if currentPhase.Duration >= a.options.MinPhaseTime {
							a.analysis.DrivingBehavior.Phases = append(a.analysis.DrivingBehavior.Phases, *currentPhase)
						}
					}

					currentPhase = &DrivingPhase{
						Type:      phaseType,
						StartTime: frame.Timestamp,
						Stats:     make(map[string]float64),
					}
				}

				// Count rapid acceleration/deceleration
				if acceleration >= a.options.RapidAccelThreshold {
					a.analysis.DrivingBehavior.RapidAccel++
				} else if acceleration <= a.options.RapidDecelThreshold {
					a.analysis.DrivingBehavior.RapidDecel++
				}
			}
		}

		lastSpeed = speed
		lastTime = frame.Timestamp
	}

	// Calculate idle time percentage
	var idleTime time.Duration
	for _, phase := range a.analysis.DrivingBehavior.Phases {
		if phase.Type == "idle" {
			idleTime += phase.Duration
		}
	}

	totalDuration := a.analysis.SessionInfo.Duration
	if totalDuration > 0 {
		a.analysis.DrivingBehavior.IdleTime = float64(idleTime) / float64(totalDuration) * 100
	}

	return nil
}

func (a *Analyzer) detectPhaseType(speed, acceleration float64) string {
	if speed < a.options.IdleSpeedThreshold {
		return "idle"
	}
	if acceleration >= a.options.RapidAccelThreshold {
		return "acceleration"
	}
	if acceleration <= a.options.RapidDecelThreshold {
		return "deceleration"
	}
	if math.Abs(acceleration) < a.options.CruiseThreshold {
		return "cruise"
	}
	return "unknown"
}

func (a *Analyzer) analyzeCANActivity() error {
	idCounts := make(map[uint32]int)

	for _, frame := range a.session.Frames {
		if frame.Type == "CAN" {
			idCounts[frame.ID]++
		}
	}

	a.analysis.CANActivity.UniqueIDs = len(idCounts)
	a.analysis.CANActivity.IDCounts = idCounts

	// Calculate bus load (assuming standard CAN frame size)
	totalBits := 0
	for _, frame := range a.session.Frames {
		if frame.Type == "CAN" {
			// Standard CAN frame: 108 bits (standard format)
			// Extended CAN frame: 128 bits
			totalBits += 108 + len(frame.Data)*8
		}
	}

	duration := a.analysis.SessionInfo.Duration.Seconds()
	if duration > 0 {
		bitsPerSecond := float64(totalBits) / duration
		a.analysis.CANActivity.BusLoad = bitsPerSecond / 1_000_000 * 100 // percentage of 1Mbps
	}

	return nil
}

func (a *Analyzer) analyzeDiagnostics() error {
	dtcs := make(map[string]int)

	for _, frame := range a.session.Frames {
		if frame.Type != "OBD2" {
			continue
		}

		decoded, ok := frame.Decoded.(map[string]interface{})
		if !ok {
			continue
		}

		if dtcList, ok := decoded["dtcs"].([]string); ok {
			for _, dtc := range dtcList {
				dtcs[dtc]++
			}
		}
	}

	a.analysis.Diagnostics.DTCCount = len(dtcs)
	for dtc := range dtcs {
		a.analysis.Diagnostics.UniqueDTCs = append(a.analysis.Diagnostics.UniqueDTCs, dtc)
	}

	// TODO: Implement DTC pattern analysis
	return nil
}
