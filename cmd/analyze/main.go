package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/anodyne74/iload-obd2/internal/analysis"
	"github.com/anodyne74/iload-obd2/internal/capture"

	"github.com/anodyne74/iload-obd2/pkg/logger"
)

var log = logger.Get("analyze")

func main() {
	var (
		inputFile    string
		exportCsv    string
		fullAnalysis bool
	)

	flag.StringVar(&inputFile, "file", "", "Capture file to analyze")
	flag.StringVar(&exportCsv, "export-csv", "", "Export data to CSV file")
	flag.BoolVar(&fullAnalysis, "full", false, "Perform full analysis including driving profile")
	flag.Parse()

	if inputFile == "" {
		fmt.Println("Please specify a capture file with -file")
		os.Exit(1)
	}

	// Load the session
	session, err := capture.LoadSession(inputFile)
	if err != nil {
		log.Fatalw("Failed to load session", "error", err, "inputfile", inputFile)
	}

	analyzer := analysis.NewAnalyzer(session, analysis.DefaultOptions())

	result, err := analyzer.Analyze()
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	// Print basic metrics
	fmt.Printf("\nSession Analysis for %s\n", filepath.Base(inputFile))
	fmt.Printf("=================================\n")
	fmt.Printf("Duration: %s\n", result.SessionInfo.Duration)
	fmt.Printf("Total Frames: %d\n", result.SessionInfo.TotalFrames)
	fmt.Printf("Unique CAN IDs: %d\n", result.CANActivity.UniqueIDs)
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("- Max RPM: %.2f\n", result.Performance.RPM.Max)
	fmt.Printf("- Average RPM: %.2f\n", result.Performance.RPM.Mean)
	fmt.Printf("- Max Speed: %.2f km/h\n", result.Performance.Speed.Max)
	fmt.Printf("- Average Speed: %.2f km/h\n", result.Performance.Speed.Mean)
	fmt.Printf("- Data Rate: %.2f frames/sec\n", result.Performance.DataRate)
	fmt.Printf("\nDriving Behavior:\n")
	fmt.Printf("- Idle Time: %.1f%%\n", result.DrivingBehavior.IdleTime)
	fmt.Printf("- Rapid Accelerations: %d\n", result.DrivingBehavior.RapidAccel)
	fmt.Printf("- Rapid Decelerations: %d\n", result.DrivingBehavior.RapidDecel)

	if fullAnalysis {
		accelCount := 0
		decelCount := 0
		cruiseCount := 0
		idleCount := 0
		var cruiseDuration time.Duration
		var idleDuration time.Duration

		for _, phase := range result.DrivingBehavior.Phases {
			switch phase.Type {
			case "acceleration":
				accelCount++
			case "deceleration":
				decelCount++
			case "cruise":
				cruiseCount++
				cruiseDuration += phase.Duration
			case "idle":
				idleCount++
				idleDuration += phase.Duration
			}
		}

		fmt.Printf("\nDriving Profile:\n")
		fmt.Printf("- Acceleration Phases: %d\n", accelCount)
		fmt.Printf("- Deceleration Phases: %d\n", decelCount)
		fmt.Printf("- Cruising Phases: %d\n", cruiseCount)
		fmt.Printf("- Idle Phases: %d\n", idleCount)
		fmt.Printf("- Total Cruising Time: %.1f minutes\n", cruiseDuration.Minutes())
		fmt.Printf("- Total Idle Time: %.1f minutes\n", idleDuration.Minutes())
	}

	// Export to CSV if requested
	if exportCsv != "" {
		fmt.Printf("\nExporting data to %s...\n", exportCsv)
		if err := exportSessionToCSV(session, exportCsv); err != nil {
			log.Fatalf("Failed to export CSV: %v", err)
		}
		fmt.Println("Export complete!")
	}
}

func exportSessionToCSV(session *capture.Session, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Timestamp", "Frame ID", "Data Type", "Value", "Unit"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, frame := range session.Frames {
		if frame.Type != "OBD2" {
			continue
		}

		decoded, ok := frame.Decoded.(map[string]interface{})
		if !ok {
			continue
		}

		timestamp := frame.Timestamp.Format(time.RFC3339)

		if rpm, ok := asFloat64(decoded["rpm"]); ok {
			record := []string{timestamp, fmt.Sprintf("0x%X", frame.ID), "RPM", fmt.Sprintf("%.2f", rpm), "rpm"}
			if err := writer.Write(record); err != nil {
				return err
			}
		}

		if speed, ok := asFloat64(decoded["speed"]); ok {
			record := []string{timestamp, fmt.Sprintf("0x%X", frame.ID), "Speed", fmt.Sprintf("%.2f", speed), "km/h"}
			if err := writer.Write(record); err != nil {
				return err
			}
		}

		if temp, ok := asFloat64(decoded["temp"]); ok {
			record := []string{timestamp, fmt.Sprintf("0x%X", frame.ID), "CoolantTemp", fmt.Sprintf("%.2f", temp), "C"}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

func asFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return parsed, true
		}
	}

	return 0, false
}
