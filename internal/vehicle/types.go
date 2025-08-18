package vehicle

import "time"

// Vehicle represents a connected vehicle with its capabilities and state
type Vehicle struct {
	VIN          string
	Make         string
	Model        string
	Year         int
	Capabilities Capabilities
	State        State
	LastUpdated  time.Time
}

// Capabilities represents what the vehicle can report and control
type Capabilities struct {
	SupportedPIDs   map[string]bool // OBD-II PIDs supported
	ProtocolVersion string          // OBD-II protocol version
	HasCAN          bool            // Whether vehicle has CAN bus access
	ExtendedPIDs    bool            // Whether extended PIDs are supported
	RealTimePIDs    []string        // PIDs that can be queried in real-time
	ControlSystems  []string        // Available control systems (engine, transmission, etc.)
}

// State represents the current state of the vehicle
type State struct {
	EngineRunning    bool
	Speed            float64
	RPM              float64
	ThrottlePosition float64
	EngineLoad       float64
	CoolantTemp      float64
	IntakeTemp       float64
	MAF              float64
	MAP              float64
	O2Voltage        float64
	FuelLevel        float64
	DTCs             []string
	LastDiagnostic   time.Time
}

// Profile represents vehicle-specific configurations and thresholds
type Profile struct {
	MaxRPM           float64
	RedlineRPM       float64
	IdleRPM          float64
	OptimalShiftRPM  float64
	FuelType         string
	TransmissionType string
	GearRatios       []float64
	WeightKg         float64
	EngineSize       float64 // in liters
	CustomThresholds map[string]float64
}

// Alert represents a vehicle alert condition
type Alert struct {
	Type      string
	Severity  string // "info", "warning", "critical"
	Message   string
	Timestamp time.Time
	Value     float64
	Threshold float64
	PIDs      []string // Related PIDs that triggered the alert
}
