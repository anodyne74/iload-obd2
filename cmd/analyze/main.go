package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/anodyne74/iload-obd2/analysis"
	"github.com/anodyne74/iload-obd2/capture"
)

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
		log.Fatalf("Failed to load session: %v", err)
	}

	analyzer := analysis.NewAnalyzer(session)

	// Basic metrics analysis
	metrics, err := analyzer.AnalyzeSession()
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	// Print basic metrics
	fmt.Printf("\nSession Analysis for %s\n", filepath.Base(inputFile))
	fmt.Printf("=================================\n")
	fmt.Printf("Duration: %s\n", time.Duration(session.EndTime-session.StartTime)*time.Second)
	fmt.Printf("Total Frames: %d\n", metrics.TotalFrames)
	fmt.Printf("Unique CAN IDs: %d\n", len(metrics.UniqueIDs))
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("- Max RPM: %.2f\n", metrics.MaxRPM)
	fmt.Printf("- Average RPM: %.2f\n", metrics.AvgRPM)
	fmt.Printf("- Max Speed: %.2f km/h\n", metrics.MaxSpeed)
	fmt.Printf("- Average Speed: %.2f km/h\n", metrics.AvgSpeed)
	fmt.Printf("- Data Rate: %.2f frames/sec\n", metrics.DataRatePerSec)
	fmt.Printf("\nDriving Behavior:\n")
	fmt.Printf("- Idle Time: %.1f%%\n", metrics.IdlePercentage)
	fmt.Printf("- Rapid Accelerations: %d\n", metrics.AccelEvents)
	fmt.Printf("- Rapid Decelerations: %d\n", metrics.DecelEvents)

	if fullAnalysis {
		// Generate and print driving profile
		profile, err := analyzer.GenerateDrivingProfile()
		if err != nil {
			log.Printf("Warning: Could not generate driving profile: %v", err)
		} else {
			fmt.Printf("\nDriving Profile:\n")
			fmt.Printf("- Acceleration Phases: %d\n", profile["acceleration_phases"])
			fmt.Printf("- Deceleration Phases: %d\n", profile["deceleration_phases"])
			fmt.Printf("- Cruising Phases: %d\n", profile["cruising_phases"])
			fmt.Printf("- Idle Phases: %d\n", profile["idle_phases"])

			if duration, ok := profile["cruising_duration"].(float64); ok {
				fmt.Printf("- Total Cruising Time: %.1f minutes\n", duration/60)
			}
			if duration, ok := profile["idle_duration"].(float64); ok {
				fmt.Printf("- Total Idle Time: %.1f minutes\n", duration/60)
			}
		}
	}

	// Export to CSV if requested
	if exportCsv != "" {
		fmt.Printf("\nExporting data to %s...\n", exportCsv)
		if err := analyzer.ExportToCSV(exportCsv); err != nil {
			log.Fatalf("Failed to export CSV: %v", err)
		}
		fmt.Println("Export complete!")
	}
}
