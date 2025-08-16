package datastore

import (
	"fmt"
	"time"

	"iload-obd2/internal/vehicle"
)

// Config holds datastore configuration
type Config struct {
	SQLitePath     string
	InfluxDBURL    string
	InfluxDBOrg    string
	InfluxDBToken  string
	InfluxDBBucket string
}

// CombinedStore implements Store using both SQLite and InfluxDB
type CombinedStore struct {
	sqlite *SQLiteStore
	influx *InfluxDBStore
}

// NewStore creates a new combined datastore
func NewStore(config *Config) (Store, error) {
	sqlite, err := NewSQLiteStore(config.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLite store: %w", err)
	}

	influx, err := NewInfluxDBStore(
		config.InfluxDBURL,
		config.InfluxDBToken,
		config.InfluxDBOrg,
		config.InfluxDBBucket,
	)
	if err != nil {
		sqlite.Close()
		return nil, fmt.Errorf("failed to create InfluxDB store: %w", err)
	}

	return &CombinedStore{
		sqlite: sqlite,
		influx: influx,
	}, nil
}

// Vehicle management methods
func (s *CombinedStore) SaveVehicle(v *vehicle.Vehicle) error {
	return s.sqlite.SaveVehicle(v)
}

func (s *CombinedStore) GetVehicle(vin string) (*vehicle.Vehicle, error) {
	return s.sqlite.GetVehicle(vin)
}

func (s *CombinedStore) ListVehicles() ([]*vehicle.Vehicle, error) {
	return s.sqlite.ListVehicles()
}

func (s *CombinedStore) DeleteVehicle(vin string) error {
	return s.sqlite.DeleteVehicle(vin)
}

// Profile management methods
func (s *CombinedStore) SaveProfile(make, model string, profile *vehicle.Profile) error {
	return s.sqlite.SaveProfile(make, model, profile)
}

func (s *CombinedStore) GetProfile(make, model string) (*vehicle.Profile, error) {
	return s.sqlite.GetProfile(make, model)
}

func (s *CombinedStore) ListProfiles() (map[string]*vehicle.Profile, error) {
	return s.sqlite.ListProfiles()
}

// Telemetry methods
func (s *CombinedStore) SaveTelemetry(vin string, data *TelemetryData) error {
	return s.influx.SaveTelemetry(vin, data)
}

func (s *CombinedStore) GetTelemetry(vin string, start, end time.Time) ([]*TelemetryData, error) {
	return s.influx.GetTelemetry(vin, start, end)
}

func (s *CombinedStore) GetLatestTelemetry(vin string) (*TelemetryData, error) {
	return s.influx.GetLatestTelemetry(vin)
}

// Performance metrics methods
func (s *CombinedStore) SavePerformanceReport(vin string, report *vehicle.PerformanceReport) error {
	return s.sqlite.SavePerformanceReport(vin, report)
}

func (s *CombinedStore) GetPerformanceReports(vin string, start, end time.Time) ([]*vehicle.PerformanceReport, error) {
	return s.sqlite.GetPerformanceReports(vin, start, end)
}

// Maintenance methods
func (s *CombinedStore) SaveServiceRecord(vin string, record *vehicle.ServiceRecord) error {
	return s.sqlite.SaveServiceRecord(vin, record)
}

func (s *CombinedStore) GetServiceHistory(vin string) ([]*vehicle.ServiceRecord, error) {
	return s.sqlite.GetServiceHistory(vin)
}

// Alert methods
func (s *CombinedStore) SaveAlert(vin string, alert *vehicle.Alert) error {
	return s.sqlite.SaveAlert(vin, alert)
}

func (s *CombinedStore) GetAlerts(vin string, start, end time.Time) ([]*vehicle.Alert, error) {
	return s.sqlite.GetAlerts(vin, start, end)
}

// Close both stores
func (s *CombinedStore) Close() error {
	sqliteErr := s.sqlite.Close()
	influxErr := s.influx.Close()

	if sqliteErr != nil {
		return sqliteErr
	}
	return influxErr
}
