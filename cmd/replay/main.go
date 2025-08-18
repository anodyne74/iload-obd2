package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/anodyne74/iload-obd2/capture"
)

func main() {
	var (
		captureFile string
		speed       float64
		list        bool
	)

	flag.StringVar(&captureFile, "file", "", "Capture file to replay")
	flag.Float64Var(&speed, "speed", 1.0, "Replay speed multiplier (1.0 = real-time)")
	flag.BoolVar(&list, "list", false, "List available capture files")
	flag.Parse()

	if list {
		listCaptureFiles()
		return
	}

	if captureFile == "" {
		fmt.Println("Please specify a capture file with -file")
		os.Exit(1)
	}

	session, err := capture.LoadSession(captureFile)
	if err != nil {
		log.Fatalf("Failed to load session: %v", err)
	}

	replayer := capture.NewReplayer(session)
	replayer.SetSpeed(speed)

	fmt.Printf("Replaying session from %s\n", time.Unix(session.StartTime, 0))
	fmt.Printf("Vehicle Info: %s\n", session.VehicleInfo)
	fmt.Printf("Total frames: %d\n", len(session.Frames))

	replayer.Play(func(frame capture.CANFrame) {
		fmt.Printf("Frame ID: 0x%X, Type: %s, Data: %X\n",
			frame.ID, frame.Type, frame.Data)
	})
}

func listCaptureFiles() {
	files, err := filepath.Glob("captures/*.json")
	if err != nil {
		log.Fatalf("Failed to list capture files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No capture files found")
		return
	}

	fmt.Println("Available capture files:")
	for _, file := range files {
		session, err := capture.LoadSession(file)
		if err != nil {
			fmt.Printf("  %s (error: %v)\n", file, err)
			continue
		}

		duration := time.Unix(session.EndTime, 0).Sub(time.Unix(session.StartTime, 0))
		fmt.Printf("  %s:\n", filepath.Base(file))
		fmt.Printf("    Date: %s\n", time.Unix(session.StartTime, 0))
		fmt.Printf("    Duration: %s\n", duration)
		fmt.Printf("    Vehicle: %s\n", session.VehicleInfo)
		fmt.Printf("    Frames: %d\n", len(session.Frames))
		fmt.Println()
	}
}
