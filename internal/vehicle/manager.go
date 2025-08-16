package vehicle

import (
	"fmt"
	"sync"
	"time"

	"iload-obd2/internal/analysis"
)

// Manager handles vehicle connections and state management
type Manager struct {
	vehicles map[string]*Vehicle // VIN -> Vehicle mapping
	profiles map[string]*Profile // Make/Model -> Profile mapping
	mu       sync.RWMutex
}

// NewManager creates a new vehicle manager instance
func NewManager() *Manager {
	return &Manager{
		vehicles: make(map[string]*Vehicle),
		profiles: make(map[string]*Profile),
	}
}

// RegisterVehicle adds a new vehicle to the manager
func (m *Manager) RegisterVehicle(vin, make, model string, year int) (*Vehicle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.vehicles[vin]; exists {
		return nil, fmt.Errorf("vehicle with VIN %s already registered", vin)
	}

	v := &Vehicle{
		VIN:   vin,
		Make:  make,
		Model: model,
		Year:  year,
		Capabilities: Capabilities{
			SupportedPIDs: make(map[string]bool),
		},
		LastUpdated: time.Now(),
	}

	m.vehicles[vin] = v
	return v, nil
}

// GetVehicle retrieves a vehicle by VIN
func (m *Manager) GetVehicle(vin string) (*Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, exists := m.vehicles[vin]
	if !exists {
		return nil, fmt.Errorf("vehicle with VIN %s not found", vin)
	}
	return v, nil
}

// UpdateVehicleState updates the vehicle's state with new data
func (m *Manager) UpdateVehicleState(vin string, state State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, exists := m.vehicles[vin]
	if !exists {
		return fmt.Errorf("vehicle with VIN %s not found", vin)
	}

	v.State = state
	v.LastUpdated = time.Now()
	return nil
}

// RegisterProfile adds or updates a vehicle profile
func (m *Manager) RegisterProfile(make, model string, profile Profile) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s-%s", make, model)
	m.profiles[key] = &profile
}

// GetProfile retrieves a vehicle profile by make and model
func (m *Manager) GetProfile(make, model string) (*Profile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s-%s", make, model)
	profile, exists := m.profiles[key]
	if !exists {
		return nil, fmt.Errorf("profile for %s %s not found", make, model)
	}
	return profile, nil
}

// DetectAnomalies checks vehicle state against its profile and returns alerts
func (m *Manager) DetectAnomalies(vin string) ([]Alert, error) {
	v, err := m.GetVehicle(vin)
	if err != nil {
		return nil, err
	}

	profile, err := m.GetProfile(v.Make, v.Model)
	if err != nil {
		return nil, err
	}

	var alerts []Alert
	now := time.Now()

	// Check RPM thresholds
	if v.State.RPM > profile.RedlineRPM {
		alerts = append(alerts, Alert{
			Type:      "RPM",
			Severity:  "critical",
			Message:   fmt.Sprintf("Engine RPM exceeds redline (%.0f > %.0f)", v.State.RPM, profile.RedlineRPM),
			Timestamp: now,
			Value:     v.State.RPM,
			Threshold: profile.RedlineRPM,
			PIDs:      []string{"01 0C"}, // RPM PID
		})
	}

	// Check engine temperature
	if v.State.CoolantTemp > 105 { // degrees Celsius
		alerts = append(alerts, Alert{
			Type:      "Temperature",
			Severity:  "warning",
			Message:   fmt.Sprintf("Engine temperature too high: %.1f°C", v.State.CoolantTemp),
			Timestamp: now,
			Value:     v.State.CoolantTemp,
			Threshold: 105,
			PIDs:      []string{"01 05"}, // Coolant temp PID
		})
	}

	// Check engine load
	if v.State.EngineLoad > 90 {
		alerts = append(alerts, Alert{
			Type:      "Load",
			Severity:  "warning",
			Message:   fmt.Sprintf("High engine load: %.1f%%", v.State.EngineLoad),
			Timestamp: now,
			Value:     v.State.EngineLoad,
			Threshold: 90,
			PIDs:      []string{"01 04"}, // Engine load PID
		})
	}

	// Check custom thresholds
	for pid, threshold := range profile.CustomThresholds {
		if value, ok := getValueForPID(v.State, pid); ok {
			if value > threshold {
				alerts = append(alerts, Alert{
					Type:      "Custom",
					Severity:  "warning",
					Message:   fmt.Sprintf("Custom threshold exceeded for %s: %.1f > %.1f", pid, value, threshold),
					Timestamp: now,
					Value:     value,
					Threshold: threshold,
					PIDs:      []string{pid},
				})
			}
		}
	}

	return alerts, nil
}

// getValueForPID is a helper function to get state values by PID
func getValueForPID(state State, pid string) (float64, bool) {
	switch pid {
	case "01 0C":
		return state.RPM, true
	case "01 0D":
		return state.Speed, true
	case "01 04":
		return state.EngineLoad, true
	case "01 05":
		return state.CoolantTemp, true
	case "01 11":
		return state.ThrottlePosition, true
	default:
		return 0, false
	}
}

// AnalyzePerformance performs a detailed analysis of vehicle performance
func (m *Manager) AnalyzePerformance(analyzer *analysis.Analyzer) (*PerformanceReport, error) {
	results, err := analyzer.Analyze()
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	report := &PerformanceReport{
		Timestamp: time.Now(),
		Duration:  results.SessionInfo.Duration,
		Stats: PerformanceStats{
			AverageSpeed:    results.Performance.Speed.Mean,
			MaxSpeed:        results.Performance.Speed.Max,
			AverageRPM:      results.Performance.RPM.Mean,
			MaxRPM:          results.Performance.RPM.Max,
			IdleTimePercent: results.DrivingBehavior.IdleTime,
			RapidAccels:     results.DrivingBehavior.RapidAccel,
			RapidDecels:     results.DrivingBehavior.RapidDecel,
		},
		Alerts: make([]Alert, 0),
	}

	// Add efficiency metrics
	if results.Performance.Speed.Mean > 0 {
		report.Stats.EfficiencyScore = calculateEfficiencyScore(results)
	}

	return report, nil
}

// calculateEfficiencyScore generates a 0-100 score based on various metrics
func calculateEfficiencyScore(results *analysis.Analysis) float64 {
	// This is a simplified scoring model
	score := 100.0

	// Penalize for excessive idle time
	if results.DrivingBehavior.IdleTime > 20 {
		score -= (results.DrivingBehavior.IdleTime - 20) * 0.5
	}

	// Penalize for rapid accelerations/decelerations
	score -= float64(results.DrivingBehavior.RapidAccel) * 2
	score -= float64(results.DrivingBehavior.RapidDecel) * 2

	// Ensure score stays within 0-100 range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}
