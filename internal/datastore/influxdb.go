package datastore

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
)

// InfluxDBStore implements telemetry storage using InfluxDB
type InfluxDBStore struct {
	client   influxdb2.Client
	org      string
	bucket   string
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
}

// NewInfluxDBStore creates a new InfluxDB-backed store
func NewInfluxDBStore(url, token, org, bucket string) (*InfluxDBStore, error) {
	client := influxdb2.NewClient(url, token)

	store := &InfluxDBStore{
		client:   client,
		org:      org,
		bucket:   bucket,
		writeAPI: client.WriteAPIBlocking(org, bucket),
		queryAPI: client.QueryAPI(org),
	}

	// Test connection
	if _, err := client.Ping(context.Background()); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	return store, nil
}

func (s *InfluxDBStore) SaveTelemetry(vin string, data *TelemetryData) error {
	point := influxdb2.NewPoint(
		"vehicle_telemetry",
		map[string]string{
			"vin": vin,
		},
		map[string]interface{}{
			"engine_running":    data.EngineRunning,
			"speed":             data.Speed,
			"rpm":               data.RPM,
			"throttle_position": data.ThrottlePos,
			"engine_load":       data.EngineLoad,
			"coolant_temp":      data.CoolantTemp,
			"intake_temp":       data.IntakeTemp,
			"maf":               data.MAF,
			"map":               data.MAP,
			"o2_voltage":        data.O2Voltage,
			"fuel_level":        data.FuelLevel,
		},
		data.Timestamp,
	)

	if data.Location != nil {
		geoPoint := influxdb2.NewPoint(
			"vehicle_location",
			map[string]string{
				"vin": vin,
			},
			map[string]interface{}{
				"latitude":    data.Location.Latitude,
				"longitude":   data.Location.Longitude,
				"altitude":    data.Location.Altitude,
				"speed":       data.Location.Speed,
				"heading":     data.Location.Heading,
				"satellites":  data.Location.Satellites,
				"hdop":        data.Location.HDOP,
				"fix_quality": data.Location.FixQuality,
			},
			data.Location.Timestamp,
		)
		if err := s.writeAPI.WritePoint(context.Background(), geoPoint); err != nil {
			return fmt.Errorf("failed to write location data: %w", err)
		}
	}

	if err := s.writeAPI.WritePoint(context.Background(), point); err != nil {
		return fmt.Errorf("failed to write telemetry data: %w", err)
	}

	return nil
}

func (s *InfluxDBStore) GetTelemetry(vin string, start, end time.Time) ([]*TelemetryData, error) {
	query := fmt.Sprintf(`
		from(bucket:"%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "vehicle_telemetry" and r["vin"] == "%s")
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
	`, s.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), vin)

	result, err := s.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry: %w", err)
	}
	defer result.Close()

	var data []*TelemetryData
	for result.Next() {
		record := result.Record()
		td := &TelemetryData{
			Timestamp:     record.Time(),
			VIN:           vin,
			EngineRunning: record.ValueByKey("engine_running").(bool),
			Speed:         record.ValueByKey("speed").(float64),
			RPM:           record.ValueByKey("rpm").(float64),
			ThrottlePos:   record.ValueByKey("throttle_position").(float64),
			EngineLoad:    record.ValueByKey("engine_load").(float64),
			CoolantTemp:   record.ValueByKey("coolant_temp").(float64),
			IntakeTemp:    record.ValueByKey("intake_temp").(float64),
			MAF:           record.ValueByKey("maf").(float64),
			MAP:           record.ValueByKey("map").(float64),
			O2Voltage:     record.ValueByKey("o2_voltage").(float64),
			FuelLevel:     record.ValueByKey("fuel_level").(float64),
		}
		data = append(data, td)
	}

	// Query location data
	locQuery := fmt.Sprintf(`
		from(bucket:"%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "vehicle_location" and r["vin"] == "%s")
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
	`, s.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), vin)

	locResult, err := s.queryAPI.Query(context.Background(), locQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query location data: %w", err)
	}
	defer locResult.Close()

	// Create a map of timestamps to location data
	locations := make(map[time.Time]*Location)
	for locResult.Next() {
		record := locResult.Record()
		timestamp := record.Time()
		locations[timestamp] = &Location{
			Timestamp:  timestamp,
			Latitude:   record.ValueByKey("latitude").(float64),
			Longitude:  record.ValueByKey("longitude").(float64),
			Altitude:   record.ValueByKey("altitude").(float64),
			Speed:      record.ValueByKey("speed").(float64),
			Heading:    record.ValueByKey("heading").(float64),
			Satellites: int(record.ValueByKey("satellites").(int64)),
			HDOP:       record.ValueByKey("hdop").(float64),
			FixQuality: int(record.ValueByKey("fix_quality").(int64)),
		}
	}

	// Merge location data with telemetry data
	for _, td := range data {
		if loc, exists := locations[td.Timestamp]; exists {
			td.Location = loc
		}
	}

	return data, nil
}

func (s *InfluxDBStore) GetLatestTelemetry(vin string) (*TelemetryData, error) {
	query := fmt.Sprintf(`
		from(bucket:"%s")
			|> range(start: -1h)
			|> filter(fn: (r) => r["_measurement"] == "vehicle_telemetry" and r["vin"] == "%s")
			|> last()
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
	`, s.bucket, vin)

	result, err := s.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest telemetry: %w", err)
	}
	defer result.Close()

	if !result.Next() {
		return nil, fmt.Errorf("no telemetry data found for VIN: %s", vin)
	}

	record := result.Record()
	td := &TelemetryData{
		Timestamp:     record.Time(),
		VIN:           vin,
		EngineRunning: record.ValueByKey("engine_running").(bool),
		Speed:         record.ValueByKey("speed").(float64),
		RPM:           record.ValueByKey("rpm").(float64),
		ThrottlePos:   record.ValueByKey("throttle_position").(float64),
		EngineLoad:    record.ValueByKey("engine_load").(float64),
		CoolantTemp:   record.ValueByKey("coolant_temp").(float64),
		IntakeTemp:    record.ValueByKey("intake_temp").(float64),
		MAF:           record.ValueByKey("maf").(float64),
		MAP:           record.ValueByKey("map").(float64),
		O2Voltage:     record.ValueByKey("o2_voltage").(float64),
		FuelLevel:     record.ValueByKey("fuel_level").(float64),
	}

	return td, nil
}

func (s *InfluxDBStore) Close() error {
	s.client.Close()
	return nil
}
