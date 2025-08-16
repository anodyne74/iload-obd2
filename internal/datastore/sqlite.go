package datastore

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"iload-obd2/internal/vehicle"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

// initialize creates the necessary database tables
func (s *SQLiteStore) initialize() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS vehicles (
			vin TEXT PRIMARY KEY,
			make TEXT NOT NULL,
			model TEXT NOT NULL,
			year INTEGER NOT NULL,
			capabilities JSON,
			last_updated TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS vehicle_profiles (
			make TEXT NOT NULL,
			model TEXT NOT NULL,
			profile JSON NOT NULL,
			PRIMARY KEY (make, model)
		)`,
		`CREATE TABLE IF NOT EXISTS performance_reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vin TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			duration INTEGER NOT NULL,
			report JSON NOT NULL,
			FOREIGN KEY (vin) REFERENCES vehicles(vin)
		)`,
		`CREATE TABLE IF NOT EXISTS service_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vin TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			service_type TEXT NOT NULL,
			description TEXT,
			mileage REAL,
			technician TEXT,
			parts JSON,
			cost REAL,
			FOREIGN KEY (vin) REFERENCES vehicles(vin)
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vin TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			alert_type TEXT NOT NULL,
			severity TEXT NOT NULL,
			message TEXT NOT NULL,
			value REAL,
			threshold REAL,
			pids JSON,
			FOREIGN KEY (vin) REFERENCES vehicles(vin)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_performance_vin_time 
			ON performance_reports(vin, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_service_vin_time 
			ON service_records(vin, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_vin_time 
			ON alerts(vin, timestamp)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

func (s *SQLiteStore) SaveVehicle(v *vehicle.Vehicle) error {
	capabilities, err := json.Marshal(v.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO vehicles (
			vin, make, model, year, capabilities, last_updated
		) VALUES (?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query, v.VIN, v.Make, v.Model, v.Year,
		capabilities, v.LastUpdated)
	if err != nil {
		return fmt.Errorf("failed to save vehicle: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetVehicle(vin string) (*vehicle.Vehicle, error) {
	query := `SELECT vin, make, model, year, capabilities, last_updated 
		FROM vehicles WHERE vin = ?`

	var v vehicle.Vehicle
	var capabilitiesJSON []byte

	err := s.db.QueryRow(query, vin).Scan(
		&v.VIN, &v.Make, &v.Model, &v.Year, &capabilitiesJSON, &v.LastUpdated)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("vehicle not found: %s", vin)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicle: %w", err)
	}

	if err := json.Unmarshal(capabilitiesJSON, &v.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	return &v, nil
}

func (s *SQLiteStore) SaveProfile(make, model string, profile *vehicle.Profile) error {
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	query := `INSERT OR REPLACE INTO vehicle_profiles (make, model, profile) 
		VALUES (?, ?, ?)`

	_, err = s.db.Exec(query, make, model, profileJSON)
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

func (s *SQLiteStore) SavePerformanceReport(vin string, report *vehicle.PerformanceReport) error {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	query := `INSERT INTO performance_reports (
		vin, timestamp, duration, report
	) VALUES (?, ?, ?, ?)`

	_, err = s.db.Exec(query, vin, report.Timestamp,
		int64(report.Duration.Seconds()), reportJSON)
	if err != nil {
		return fmt.Errorf("failed to save performance report: %w", err)
	}

	return nil
}

func (s *SQLiteStore) SaveAlert(vin string, alert *vehicle.Alert) error {
	pidsJSON, err := json.Marshal(alert.PIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal PIDs: %w", err)
	}

	query := `INSERT INTO alerts (
		vin, timestamp, alert_type, severity, message, value, threshold, pids
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query, vin, alert.Timestamp, alert.Type,
		alert.Severity, alert.Message, alert.Value, alert.Threshold, pidsJSON)
	if err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
