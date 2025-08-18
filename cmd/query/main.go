package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/anodyne74/iload-obd2/vehicle"
	"github.com/rzetterberg/elmobd"
)

func main() {
	var (
		queryType  string
		outputFile string
		continuous bool
		formatJSON bool
	)

	flag.StringVar(&queryType, "query", "all", "Type of query: all, ecu, maps, live")
	flag.StringVar(&outputFile, "output", "", "Output file for the query results")
	flag.BoolVar(&continuous, "continuous", false, "Enable continuous monitoring")
	flag.BoolVar(&formatJSON, "json", false, "Output in JSON format")
	flag.Parse()

	// Initialize OBD connection
	dev, err := elmobd.NewDevice("/dev/ttyUSB0", false)
	if err != nil {
		log.Fatal(err)
	}

	querier := vehicle.NewVehicleQuerier(dev)

	switch queryType {
	case "all":
		data, err := querier.QueryAllData()
		if err != nil {
			log.Fatalf("Failed to query vehicle data: %v", err)
		}
		outputData(data, outputFile, formatJSON)

	case "ecu":
		ecus, err := querier.QueryECUs()
		if err != nil {
			log.Fatalf("Failed to query ECU data: %v", err)
		}
		outputData(ecus, outputFile, formatJSON)

	case "maps":
		maps, err := querier.QueryEngineMaps()
		if err != nil {
			log.Fatalf("Failed to query engine maps: %v", err)
		}
		outputData(maps, outputFile, formatJSON)

	case "live":
		if continuous {
			fmt.Println("Starting continuous monitoring...")
			querier.MonitorLiveData(func(data map[string]interface{}) {
				if formatJSON {
					json, _ := json.MarshalIndent(data, "", "  ")
					fmt.Println(string(json))
				} else {
					fmt.Printf("\rRPM: %.2f, Speed: %.2f km/h",
						data["RPM"], data["Speed"])
				}
			})
		} else {
			data, err := querier.QueryAllData()
			if err != nil {
				log.Fatalf("Failed to query live data: %v", err)
			}
			outputData(data.LiveData, outputFile, formatJSON)
		}
	}
}

func outputData(data interface{}, outputFile string, formatJSON bool) {
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(data); err != nil {
			log.Fatalf("Failed to write data: %v", err)
		}
		return
	}

	if formatJSON {
		json, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal data: %v", err)
		}
		fmt.Println(string(json))
		return
	}

	// Pretty print the data based on type
	switch v := data.(type) {
	case *vehicle.VehicleData:
		fmt.Printf("Vehicle Information:\n")
		fmt.Printf("VIN: %s\n", v.VIN)
		fmt.Printf("\nECU Information:\n")
		for id, ecu := range v.ECUs {
			fmt.Printf("  %s:\n", id)
			fmt.Printf("    Hardware Version: %s\n", ecu.HardwareVersion)
			fmt.Printf("    Software Version: %s\n", ecu.SoftwareVersion)
			fmt.Printf("    Manufacturer: %s\n", ecu.Manufacturer)
		}
		// Add more pretty printing as needed
	}
}
