package datastore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anodyne74/iload-obd2/internal/vehicle"

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

// ListVehicles returns all vehicles from the database
func (s *SQLiteStore) ListVehicles() ([]*vehicle.Vehicle, error) {
	rows, err := s.db.Query(`SELECT vin, make, model, year, capabilities, last_updated 
		FROM vehicles`)
	if err != nil {
		return nil, fmt.Errorf("failed to query vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []*vehicle.Vehicle
	for rows.Next() {
		var v vehicle.Vehicle
		var capabilitiesJSON []byte
		err := rows.Scan(&v.VIN, &v.Make, &v.Model, &v.Year, &capabilitiesJSON, &v.LastUpdated)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vehicle row: %w", err)
		}

		if err := json.Unmarshal(capabilitiesJSON, &v.Capabilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}

		vehicles = append(vehicles, &v)
	}

	return vehicles, rows.Err()
}

// DeleteVehicle removes a vehicle from the database
func (s *SQLiteStore) DeleteVehicle(vin string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete related records first
	tables := []string{"alerts", "performance_reports", "service_records"}
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE vin = ?", table)
		if _, err := tx.Exec(query, vin); err != nil {
			return fmt.Errorf("failed to delete from %s: %w", table, err)
		}
	}

	// Delete the vehicle record
	result, err := tx.Exec("DELETE FROM vehicles WHERE vin = ?", vin)
	if err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("vehicle not found: %s", vin)
	}

	return tx.Commit()
}

// GetProfile retrieves a vehicle profile from the database
func (s *SQLiteStore) GetProfile(make, model string) (*vehicle.Profile, error) {
	var profileJSON []byte
	err := s.db.QueryRow(
		"SELECT profile FROM vehicle_profiles WHERE make = ? AND model = ?",
		make, model,
	).Scan(&profileJSON)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("profile not found for %s %s", make, model)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	var profile vehicle.Profile
	if err := json.Unmarshal(profileJSON, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// ListProfiles returns all vehicle profiles
func (s *SQLiteStore) ListProfiles() (map[string]*vehicle.Profile, error) {
	rows, err := s.db.Query("SELECT make, model, profile FROM vehicle_profiles")
	if err != nil {
		return nil, fmt.Errorf("failed to query profiles: %w", err)
	}
	defer rows.Close()

	profiles := make(map[string]*vehicle.Profile)
	for rows.Next() {
		var make, model string
		var profileJSON []byte
		if err := rows.Scan(&make, &model, &profileJSON); err != nil {
			return nil, fmt.Errorf("failed to scan profile row: %w", err)
		}

		var profile vehicle.Profile
		if err := json.Unmarshal(profileJSON, &profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
		}

		key := fmt.Sprintf("%s-%s", make, model)
		profiles[key] = &profile
	}

	return profiles, rows.Err()
}

// GetPerformanceReports retrieves performance reports for a vehicle within a time range
func (s *SQLiteStore) GetPerformanceReports(vin string, start, end time.Time) ([]*vehicle.PerformanceReport, error) {
	rows, err := s.db.Query(`
		SELECT report FROM performance_reports 
		WHERE vin = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC`,
		vin, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query performance reports: %w", err)
	}
	defer rows.Close()

	var reports []*vehicle.PerformanceReport
	for rows.Next() {
		var reportJSON []byte
		if err := rows.Scan(&reportJSON); err != nil {
			return nil, fmt.Errorf("failed to scan report row: %w", err)
		}

		var report vehicle.PerformanceReport
		if err := json.Unmarshal(reportJSON, &report); err != nil {
			return nil, fmt.Errorf("failed to unmarshal report: %w", err)
		}

		reports = append(reports, &report)
	}

	return reports, rows.Err()
}

// SaveServiceRecord stores a service record in the database
func (s *SQLiteStore) SaveServiceRecord(vin string, record *vehicle.ServiceRecord) error {
	partsJSON, err := json.Marshal(record.Parts)
	if err != nil {
		return fmt.Errorf("failed to marshal parts: %w", err)
	}

	query := `INSERT INTO service_records (
		vin, timestamp, service_type, description, mileage, technician, parts, cost
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = s.db.Exec(query,
		vin,
		record.Date,
		record.Type,
		record.Description,
		record.Mileage,
		record.Technician,
		partsJSON,
		record.Cost,
	)
	if err != nil {
		return fmt.Errorf("failed to save service record: %w", err)
	}

	return nil
}

// GetServiceHistory retrieves all service records for a vehicle
func (s *SQLiteStore) GetServiceHistory(vin string) ([]*vehicle.ServiceRecord, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, service_type, description, mileage, technician, parts, cost
		FROM service_records 
		WHERE vin = ?
		ORDER BY timestamp DESC`,
		vin,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query service history: %w", err)
	}
	defer rows.Close()

	var records []*vehicle.ServiceRecord
	for rows.Next() {
		var record vehicle.ServiceRecord
		var partsJSON []byte
		if err := rows.Scan(
			&record.Date,
			&record.Type,
			&record.Description,
			&record.Mileage,
			&record.Technician,
			&partsJSON,
			&record.Cost,
		); err != nil {
			return nil, fmt.Errorf("failed to scan service record: %w", err)
		}

		if err := json.Unmarshal(partsJSON, &record.Parts); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parts: %w", err)
		}

		records = append(records, &record)
	}

	return records, rows.Err()
}

// GetAlerts retrieves alerts for a vehicle within a time range
func (s *SQLiteStore) GetAlerts(vin string, start, end time.Time) ([]*vehicle.Alert, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, alert_type, severity, message, value, threshold, pids
		FROM alerts 
		WHERE vin = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC`,
		vin, start, end,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*vehicle.Alert
	for rows.Next() {
		var alert vehicle.Alert
		var pidsJSON []byte
		if err := rows.Scan(
			&alert.Timestamp,
			&alert.Type,
			&alert.Severity,
			&alert.Message,
			&alert.Value,
			&alert.Threshold,
			&pidsJSON,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		if err := json.Unmarshal(pidsJSON, &alert.PIDs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal PIDs: %w", err)
		}

		alerts = append(alerts, &alert)
	}

	return alerts, rows.Err()
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
