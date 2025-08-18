package vehicle

import (
	"testing"
)

func TestVehicleManager(t *testing.T) {
	manager := NewManager()

	// Test vehicle registration
	vin := "1HGCM82633A123456"
	v, err := manager.RegisterVehicle(vin, "Honda", "Accord", 2023)
	if err != nil {
		t.Fatalf("Failed to register vehicle: %v", err)
	}
	if v.VIN != vin {
		t.Errorf("Expected VIN %s, got %s", vin, v.VIN)
	}

	// Test duplicate registration
	_, err = manager.RegisterVehicle(vin, "Honda", "Accord", 2023)
	if err == nil {
		t.Error("Expected error on duplicate registration")
	}

	// Test vehicle retrieval
	v2, err := manager.GetVehicle(vin)
	if err != nil {
		t.Fatalf("Failed to get vehicle: %v", err)
	}
	if v2.VIN != vin {
		t.Errorf("Expected VIN %s, got %s", vin, v2.VIN)
	}

	// Test state update
	state := State{
		EngineRunning:    true,
		Speed:            60.0,
		RPM:              2500.0,
		ThrottlePosition: 25.0,
		EngineLoad:       40.0,
		CoolantTemp:      85.0,
	}
	err = manager.UpdateVehicleState(vin, state)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	v3, _ := manager.GetVehicle(vin)
	if v3.State.Speed != state.Speed {
		t.Errorf("Expected speed %.1f, got %.1f", state.Speed, v3.State.Speed)
	}

	// Test profile management
	profile := Profile{
		MaxRPM:           6500,
		RedlineRPM:       6000,
		IdleRPM:          800,
		OptimalShiftRPM:  2500,
		FuelType:         "gasoline",
		TransmissionType: "automatic",
		GearRatios:       []float64{2.995, 1.759, 1.171, 0.870, 0.707},
		WeightKg:         1500,
		EngineSize:       2.0,
		CustomThresholds: map[string]float64{
			"01 05": 100.0, // Coolant temp threshold
		},
	}
	manager.RegisterProfile("Honda", "Accord", profile)

	p, err := manager.GetProfile("Honda", "Accord")
	if err != nil {
		t.Fatalf("Failed to get profile: %v", err)
	}
	if p.MaxRPM != profile.MaxRPM {
		t.Errorf("Expected MaxRPM %.1f, got %.1f", profile.MaxRPM, p.MaxRPM)
	}

	// Test anomaly detection
	state.RPM = 6200 // Above redline
	err = manager.UpdateVehicleState(vin, state)
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	alerts, err := manager.DetectAnomalies(vin)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}
	if len(alerts) == 0 {
		t.Error("Expected at least one alert for high RPM")
	}

	found := false
	for _, alert := range alerts {
		if alert.Type == "RPM" && alert.Severity == "critical" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected critical RPM alert")
	}
}

func TestServiceSchedule(t *testing.T) {
	schedule := DefaultServiceSchedule()
	if len(schedule.Items) == 0 {
		t.Error("Expected default service schedule to have items")
	}

	// Find oil change service
	var oilChange *ServiceItem
	for i := range schedule.Items {
		if schedule.Items[i].Name == "Oil Change" {
			oilChange = &schedule.Items[i]
			break
		}
	}

	if oilChange == nil {
		t.Fatal("Expected to find oil change service")
	}

	if oilChange.IntervalMiles != 5000 {
		t.Errorf("Expected oil change interval of 5000 miles, got %.1f", oilChange.IntervalMiles)
	}

	if oilChange.Priority != "required" {
		t.Errorf("Expected oil change priority 'required', got '%s'", oilChange.Priority)
	}
}
