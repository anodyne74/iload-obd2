package vehicle

import (
	"fmt"
	"time"

	"github.com/rzetterberg/elmobd"
)

// ECUInfo contains detailed information about an ECU
type ECUInfo struct {
	ID              string
	HardwareVersion string
	SoftwareVersion string
	Manufacturer    string
	Protocol        string
	CalibrationID   string
}

// EngineMap contains engine tuning parameters
type EngineMap struct {
	FuelMaps    map[string][]float64
	IgnitionMap map[string][]float64
	BoostMap    map[string][]float64
	VEMap       map[string][]float64 // Volumetric Efficiency
}

// VehicleData contains comprehensive vehicle information
type VehicleData struct {
	VIN           string
	ECUs          map[string]ECUInfo
	EngineMaps    EngineMap
	FuelSystem    map[string]interface{}
	Transmission  map[string]interface{}
	Sensors       map[string]interface{}
	LiveData      map[string]float64
	Capabilities  []string
	SupportedPIDs map[string][]string
}

// PIDs for Hyundai iLoad/H-1
const (
	PID_ECU_INFO    = "09 02"
	PID_CAL_ID      = "09 04"
	PID_VIN         = "09 02"
	PID_FUEL_MAP    = "2C"
	PID_TIMING_MAP  = "2D"
	PID_BOOST_MAP   = "2E"
	PID_VE_MAP      = "2F"
	PID_TRANS_DATA  = "A4"
	PID_SENSOR_DATA = "01 00"
)

type VehicleQuerier struct {
	dev *elmobd.Device
}

func NewVehicleQuerier(device *elmobd.Device) *VehicleQuerier {
	return &VehicleQuerier{
		dev: device,
	}
}

func (vq *VehicleQuerier) QueryAllData() (*VehicleData, error) {
	data := &VehicleData{
		ECUs:          make(map[string]ECUInfo),
		FuelSystem:    make(map[string]interface{}),
		Transmission:  make(map[string]interface{}),
		Sensors:       make(map[string]interface{}),
		LiveData:      make(map[string]float64),
		SupportedPIDs: make(map[string][]string),
	}

	// Query VIN
	vin, err := vq.queryVIN()
	if err != nil {
		return nil, fmt.Errorf("failed to query VIN: %v", err)
	}
	data.VIN = vin

	// Query ECU information
	ecus, err := vq.queryECUs()
	if err != nil {
		return nil, fmt.Errorf("failed to query ECUs: %v", err)
	}
	data.ECUs = ecus

	// Query Engine Maps
	engineMaps, err := vq.queryEngineMaps()
	if err != nil {
		return nil, fmt.Errorf("failed to query engine maps: %v", err)
	}
	data.EngineMaps = engineMaps

	// Query supported PIDs
	supportedPIDs, err := vq.querySupportedPIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to query supported PIDs: %v", err)
	}
	data.SupportedPIDs = supportedPIDs

	return data, nil
}

func (vq *VehicleQuerier) queryVIN() (string, error) {
	// Implementation for VIN query using Mode 09 PID 02
	return "Sample VIN Query", nil
}

func (vq *VehicleQuerier) queryECUs() (map[string]ECUInfo, error) {
	ecus := make(map[string]ECUInfo)

	// Query each ECU (Engine, Transmission, ABS, etc.)
	engineECU := ECUInfo{
		ID:              "ECM",
		HardwareVersion: "H-1.2.3",
		SoftwareVersion: "S-2.4.5",
		Manufacturer:    "Hyundai",
		Protocol:        "ISO 15765-4 (CAN)",
		CalibrationID:   "CAL-123456",
	}
	ecus["ENGINE"] = engineECU

	return ecus, nil
}

func (vq *VehicleQuerier) queryEngineMaps() (EngineMap, error) {
	maps := EngineMap{
		FuelMaps:    make(map[string][]float64),
		IgnitionMap: make(map[string][]float64),
		BoostMap:    make(map[string][]float64),
		VEMap:       make(map[string][]float64),
	}

	// Query fuel maps
	fuelMap, err := vq.queryFuelMap()
	if err != nil {
		return maps, fmt.Errorf("failed to query fuel map: %v", err)
	}
	maps.FuelMaps = fuelMap

	return maps, nil
}

func (vq *VehicleQuerier) queryFuelMap() (map[string][]float64, error) {
	// Implementation for fuel map query
	return map[string][]float64{
		"idle": {1.0, 1.2, 1.3},
		"low":  {1.5, 1.7, 1.9},
		"mid":  {2.0, 2.2, 2.4},
		"high": {2.5, 2.7, 2.9},
	}, nil
}

func (vq *VehicleQuerier) querySupportedPIDs() (map[string][]string, error) {
	// Query supported PIDs for each mode
	return map[string][]string{
		"01": {"0C", "0D", "0E", "0F"},
		"09": {"02", "04", "06"},
	}, nil
}

// MonitorLiveData starts continuous monitoring of vehicle data
func (vq *VehicleQuerier) MonitorLiveData(callback func(map[string]interface{})) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		data := make(map[string]interface{})

		// Query real-time data
		if rpm, err := vq.dev.RunOBDCommand(elmobd.NewEngineRPM()); err == nil {
			data["RPM"] = rpm
		}

		callback(data)
	}

	return nil
}
