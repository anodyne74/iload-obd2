package datastore

import (
	"time"

	"iload-obd2/internal/vehicle"
)

// Store defines the interface for vehicle data storage
type Store interface {
	// Vehicle management
	SaveVehicle(v *vehicle.Vehicle) error
	GetVehicle(vin string) (*vehicle.Vehicle, error)
	ListVehicles() ([]*vehicle.Vehicle, error)
	DeleteVehicle(vin string) error

	// Profile management
	SaveProfile(make, model string, profile *vehicle.Profile) error
	GetProfile(make, model string) (*vehicle.Profile, error)
	ListProfiles() (map[string]*vehicle.Profile, error)

	// Telemetry storage
	SaveTelemetry(vin string, data *TelemetryData) error
	GetTelemetry(vin string, start, end time.Time) ([]*TelemetryData, error)
	GetLatestTelemetry(vin string) (*TelemetryData, error)

	// Performance metrics
	SavePerformanceReport(vin string, report *vehicle.PerformanceReport) error
	GetPerformanceReports(vin string, start, end time.Time) ([]*vehicle.PerformanceReport, error)

	// Maintenance records
	SaveServiceRecord(vin string, record *vehicle.ServiceRecord) error
	GetServiceHistory(vin string) ([]*vehicle.ServiceRecord, error)

	// Alert history
	SaveAlert(vin string, alert *vehicle.Alert) error
	GetAlerts(vin string, start, end time.Time) ([]*vehicle.Alert, error)

	// Database management
	Close() error
}

// TelemetryData represents a point-in-time vehicle state
type TelemetryData struct {
	Timestamp     time.Time `json:"timestamp"`
	VIN           string    `json:"vin"`
	EngineRunning bool      `json:"engine_running"`
	Speed         float64   `json:"speed"`
	RPM           float64   `json:"rpm"`
	ThrottlePos   float64   `json:"throttle_position"`
	EngineLoad    float64   `json:"engine_load"`
	CoolantTemp   float64   `json:"coolant_temp"`
	IntakeTemp    float64   `json:"intake_temp"`
	MAF           float64   `json:"maf"`
	MAP           float64   `json:"map"`
	O2Voltage     float64   `json:"o2_voltage"`
	FuelLevel     float64   `json:"fuel_level"`
	DTCs          []string  `json:"dtcs"`
	Location      *Location `json:"location,omitempty"`
}

// Location represents GPS coordinates and related data
type Location struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Altitude   float64   `json:"altitude"`
	Speed      float64   `json:"speed"`
	Heading    float64   `json:"heading"`
	Satellites int       `json:"satellites"`
	HDOP       float64   `json:"hdop"`
	FixQuality int       `json:"fix_quality"`
	Timestamp  time.Time `json:"timestamp"`
}
