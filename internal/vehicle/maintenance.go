package vehicle

import "time"

// PerformanceReport represents a detailed analysis of vehicle performance
type PerformanceReport struct {
	Timestamp time.Time
	Duration  time.Duration
	Stats     PerformanceStats
	Alerts    []Alert
}

// PerformanceStats contains calculated performance metrics
type PerformanceStats struct {
	AverageSpeed    float64
	MaxSpeed        float64
	AverageRPM      float64
	MaxRPM          float64
	IdleTimePercent float64
	RapidAccels     int
	RapidDecels     int
	EfficiencyScore float64
}

// Maintenance represents vehicle maintenance information
type Maintenance struct {
	LastService     time.Time
	NextService     time.Time
	Mileage         float64
	ServiceHistory  []ServiceRecord
	PendingServices []string
}

// ServiceRecord represents a maintenance service record
type ServiceRecord struct {
	Date        time.Time
	Type        string
	Description string
	Mileage     float64
	Technician  string
	Parts       []string
	Cost        float64
}

// ServiceSchedule represents maintenance intervals for a vehicle
type ServiceSchedule struct {
	Items []ServiceItem
}

// ServiceItem represents a scheduled maintenance item
type ServiceItem struct {
	Name           string
	IntervalMiles  float64
	IntervalMonths int
	Description    string
	EstimatedCost  float64
	Priority       string // "required", "recommended", "optional"
}

// DefaultServiceSchedule returns a basic service schedule
func DefaultServiceSchedule() ServiceSchedule {
	return ServiceSchedule{
		Items: []ServiceItem{
			{
				Name:           "Oil Change",
				IntervalMiles:  5000,
				IntervalMonths: 6,
				Description:    "Change engine oil and filter",
				EstimatedCost:  50,
				Priority:       "required",
			},
			{
				Name:           "Tire Rotation",
				IntervalMiles:  7500,
				IntervalMonths: 6,
				Description:    "Rotate and balance tires",
				EstimatedCost:  30,
				Priority:       "recommended",
			},
			{
				Name:           "Air Filter",
				IntervalMiles:  15000,
				IntervalMonths: 12,
				Description:    "Replace engine air filter",
				EstimatedCost:  20,
				Priority:       "recommended",
			},
			{
				Name:           "Brake Service",
				IntervalMiles:  30000,
				IntervalMonths: 24,
				Description:    "Inspect and service brake system",
				EstimatedCost:  200,
				Priority:       "required",
			},
		},
	}
}
