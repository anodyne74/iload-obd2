package analysis

import (
	"math"
	"sort"
	"time"
)

// Stats represents statistical analysis of numeric data
type Stats struct {
	Min      float64       `json:"min"`
	Max      float64       `json:"max"`
	Mean     float64       `json:"mean"`
	Median   float64       `json:"median"`
	StdDev   float64       `json:"std_dev"`
	Samples  int           `json:"samples"`
	Duration time.Duration `json:"duration"`
}

// DrivingPhase represents a distinct driving behavior period
type DrivingPhase struct {
	Type      string             `json:"type"` // "idle", "acceleration", "deceleration", "cruise"
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	Duration  time.Duration      `json:"duration"`
	Stats     map[string]float64 `json:"stats"`
}

// Analysis represents a complete analysis of a capture session
type Analysis struct {
	SessionInfo struct {
		StartTime   time.Time     `json:"start_time"`
		EndTime     time.Time     `json:"end_time"`
		Duration    time.Duration `json:"duration"`
		VehicleInfo string        `json:"vehicle_info"`
		TotalFrames int           `json:"total_frames"`
	} `json:"session_info"`

	Performance struct {
		RPM         Stats   `json:"rpm"`
		Speed       Stats   `json:"speed"`
		Temperature Stats   `json:"temperature"`
		DataRate    float64 `json:"data_rate"` // frames per second
	} `json:"performance"`

	DrivingBehavior struct {
		Phases     []DrivingPhase `json:"phases"`
		IdleTime   float64        `json:"idle_time_percent"`
		RapidAccel int            `json:"rapid_accelerations"`
		RapidDecel int            `json:"rapid_decelerations"`
		StopCount  int            `json:"stop_count"`
	} `json:"driving_behavior"`

	CANActivity struct {
		UniqueIDs  int            `json:"unique_ids"`
		IDCounts   map[uint32]int `json:"id_counts"`
		BusLoad    float64        `json:"bus_load_percent"`
		ErrorCount int            `json:"error_count"`
	} `json:"can_activity"`

	Diagnostics struct {
		DTCCount    int      `json:"dtc_count"`
		UniqueDTCs  []string `json:"unique_dtcs"`
		DTCPatterns []string `json:"dtc_patterns"`
	} `json:"diagnostics"`
}

// CalculateStats computes statistical measures from a slice of float64 values
func CalculateStats(values []float64) Stats {
	if len(values) == 0 {
		return Stats{}
	}

	// Calculate min, max, and mean
	min := values[0]
	max := values[0]
	sum := 0.0

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate standard deviation
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(values)-1))

	// Calculate median
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	var median float64
	if len(sorted)%2 == 0 {
		median = (sorted[len(sorted)/2-1] + sorted[len(sorted)/2]) / 2
	} else {
		median = sorted[len(sorted)/2]
	}

	return Stats{
		Min:     min,
		Max:     max,
		Mean:    mean,
		Median:  median,
		StdDev:  stdDev,
		Samples: len(values),
	}
}
