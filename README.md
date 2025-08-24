# iload-obd2
![iload](/static/images/my-vehicle-dark.png)

A comprehensive vehicle telemetry and diagnostics platform for the Hyundai iLoad/H-1.

## Overview

This application provides real-time monitoring, diagnostics, and data analysis through OBD2 and CANBUS interfaces. While optimized for the Hyundai iLoad/H-1, the architecture supports multiple vehicle types and can be extended for other manufacturers.

The platform combines modern web technologies with robust vehicle communication protocols to deliver:
- Real-time performance monitoring
- Comprehensive diagnostic capabilities
- Data logging and analysis
- Custom vehicle profiles and configurations

## Quick Start

### Container Deployment (Recommended)
```bash
# Clone the repository
git clone https://github.com/anodyne74/iload-obd2.git
cd iload-obd2

# Start the application
podman-compose up -d

# Access the dashboard
http://localhost:8080
```

## Core Features
- Real-time monitoring of:
  - Engine RPM
  - Vehicle Speed
  - Engine Temperature
  - CAN Bus Messages
  - ECU Information
  - Engine Maps (Fuel and Timing)
- Multiple Transport Options:
  - Serial OBD-II Connection
  - TCP Connection (for testing/simulation)
  - Direct CAN Bus Access
- Cross-Platform Support:
  - Windows (with compatible CAN adapter)
  - Linux (Raspberry Pi/standard distributions)
- Web Interface:
  - Real-time WebSocket updates
  - Responsive design
  - JSON-based telemetry data
  - Automatic reconnection
- Extensible Architecture:
  - Modular transport layer
  - Support for custom CAN frame handling
  - Configurable update intervals

## Requirements

### Hardware Requirements
- Raspberry Pi 4 (recommended)
- OBD2 adapter (USB or Bluetooth)
- Tablet or device with web browser
- Vehicle compatibility:
  - Hyundai iLoad
  - Hyundai H-1
  - Compatible with 2.5L CRDi diesel engine

### Software Prerequisites
- Podman and Podman Compose installed
- OBD2 adapter connected (USB or Bluetooth)
- CAN interface configured (for direct CAN access)

## Installation

### Container Deployment (Recommended)

#### First-time Setup
```bash
# Install Podman (if not already installed)
sudo dnf install -y podman podman-compose   # Fedora/RHEL
# or
sudo apt install -y podman podman-compose   # Ubuntu/Debian

# Create required directories with correct permissions
mkdir -p data/sqlite data/influxdb logs
sudo chown -R 1000:1000 data logs
sudo semanage fcontext -a -t container_file_t "data(/.*)?"
sudo semanage fcontext -a -t container_file_t "logs(/.*)?"
sudo restorecon -Rv data logs
```
```bash
# Clone the repository
git clone https://github.com/anodyne74/iload-obd2.git
cd iload-obd2

# Configure the application
cp config.docker.yaml config.yaml
# Edit config.yaml with your settings

# Build and start the services with Podman
podman-compose up -d

# View logs
podman-compose logs -f

# For rootless containers, add your user to required groups
sudo usermod -aG dialout $USER
sudo usermod -aG can $USER

# Access the dashboard
# http://localhost:8080 or http://raspberry-pi-ip:8080
```

### Troubleshooting Podman Setup

1. **Permission Issues**
```bash
# If you see permission errors, try:
sudo chown -R $USER:$USER data logs
sudo chmod -R 755 data logs
```

2. **Device Access Issues**
```bash
# Check if your user has proper device access
ls -l /dev/ttyUSB0
ls -l /dev/can0

# Add current user to required groups
sudo usermod -aG dialout,can $USER
# Log out and back in for changes to take effect
```

3. **SELinux Issues**
```bash
# Temporarily disable SELinux for testing
sudo setenforce 0

# For permanent solution, create proper SELinux context
sudo semanage fcontext -a -t container_file_t "/path/to/iload-obd2/data(/.*)?"
sudo restorecon -Rv /path/to/iload-obd2/data
```

## Platform-Specific Configuration

### Raspberry Pi Setup
```bash
# Configure CAN interface
sudo ip link set can0 up type can bitrate 500000

# Make CAN interface persistent
sudo nano /etc/network/interfaces.d/can0
# Add these lines:
# auto can0
# iface can0 inet manual
#     pre-up ip link set $IFACE type can bitrate 500000
#     up ip link set $IFACE up
#     down ip link set $IFACE down
```


## Technical Architecture

### System Components

#### Transport Layer
- Serial OBD-II interface via `github.com/rzetterberg/elmobd`
- TCP connection for testing/simulation
- Direct CAN bus access using `github.com/brutella/can`
- Modular design for easy extension

#### Data Collection (`capture` package)
- Real-time OBD-II parameter monitoring
- CAN frame capture and processing
- Engine map data collection
- ECU information retrieval
- Session management
- Data validation and preprocessing

#### Analysis Engine (`analysis` package)
- Real-time performance analysis
- Driving behavior classification
- Statistical computations
- Anomaly detection
- Efficiency scoring
- Pattern recognition

#### Vehicle Management (`vehicle` package)
- Multi-vehicle support
- Vehicle profiles and configurations
- State management
- Maintenance tracking
- Alert management
- Service scheduling

#### Data Persistence (`datastore` package)
- Hybrid storage (SQLite + InfluxDB)
- Query optimization
- Data retention policies
- Backup management
- Performance metrics
- Historical analysis

#### Communication
- WebSocket server for real-time updates
- JSON-based telemetry protocol
- Configurable update intervals
- Automatic connection management

### Frontend
- Real-time Dashboard:
  - Engine performance metrics
  - CAN bus message monitor
  - Engine map visualization
  - ECU information display
- WebSocket Client:
  - Automatic reconnection
  - Buffered updates
  - Error handling

## DTC Coverage
The system includes comprehensive diagnostic code coverage:

### Powertrain (P) Codes
- Engine Management
- Transmission
- Emissions Systems
- Fuel Systems

### Hyundai iLoad Specific Systems
1. Fuel and Air System
   - Common Rail Pressure
   - Injector Circuits
   - Turbocharger Systems

2. EGR System
   - Flow Monitoring
   - Circuit Diagnostics
   - Sensor Validation

3. Diesel-Specific
   - Glow Plug Systems
   - DPF Monitoring
   - Exhaust Temperature

4. Transmission
   - Fluid Temperature
   - Gear Ratio Monitoring
   - Speed Sensors

5. Vehicle Sensors
   - Fuel Level
   - Cooling Systems
   - Speed Sensors
   - Brake Systems

6. Body Systems
   - Door Controls
   - Module Programming
   - EEPROM Systems

## Installation

### Prerequisites
1. System Requirements:
   - Go 1.21 or later
   - SQLite 3
   - InfluxDB 2.x
   - For CAN bus support:
     - Windows: Compatible CAN adapter and drivers
     - Linux: CAN interface configured (`can0`)

2. Database Setup:
   ```bash
   # Install InfluxDB
   wget https://dl.influxdata.com/influxdb/releases/influxdb2-latest-amd64.deb
   sudo dpkg -i influxdb2-latest-amd64.deb
   sudo service influxdb start

   # Configure InfluxDB
   influx setup \
     --org your-org \
     --bucket vehicle-telemetry \
     --username admin \
     --password your-password \
     --token your-token \
     --force
   ```

3. Clone and Build:
   ```bash
   # Clone repository
   git clone https://github.com/anodyne74/iload-obd2.git
   cd iload-obd2

   # Install dependencies
   go mod tidy

   # Build application
   go build ./...
   ```

4. Configuration:
   Create `config.yaml`:
   ```yaml
   datastore:
     sqlite:
       path: "./data/vehicles.db"
     influxdb:
       url: "http://localhost:8086"
       org: "your-org"
       bucket: "vehicle-telemetry"
       token: "your-token"
   
   vehicle:
     default_thresholds:
       rpm_redline: 6000
       coolant_temp_max: 105
       engine_load_max: 90
   ```

5. Run the Application:
   ```bash
   # Start with default settings
   ./iload-obd2

   # Use specific configuration
   ./iload-obd2 --config ./config.yaml

   # Test mode with simulated data
   ./iload-obd2 --test-mode

   # Development mode with debug logging
   ./iload-obd2 --debug
   ```

The web interface will be available at `http://localhost:8080`

## Configuration

The application supports several transport options:

### Serial Connection
```bash
./iload-obd2 --port COM1  # Windows
./iload-obd2 --port /dev/ttyUSB0  # Linux
```

### TCP Connection (Testing)
```bash
./iload-obd2 --test-tcp --tcp-addr localhost:6789
```

### Mock Data
```bash
./iload-obd2 --mock-data
```

## Data Storage

The application uses a hybrid storage approach to efficiently handle different types of vehicle data:

### SQLite Storage (Structured Data)
- Vehicle information and profiles
- Maintenance records and service history
- Performance reports and analytics
- Alert history and diagnostics
- Custom vehicle configurations

### InfluxDB Storage (Time-Series Data)
- Real-time telemetry data
- GPS location tracking
- Performance metrics
- Sensor readings
- Engine parameters

## WebSocket API

The application provides WebSocket endpoints for real-time data:

### Telemetry Data (`/ws/telemetry`)
```json
{
  "timestamp": "2025-08-16T10:00:00Z",
  "vin": "1HGCM82633A123456",
  "telemetry": {
    "rpm": 2500,
    "speed": 60,
    "engine_load": 45.5,
    "coolant_temp": 90,
    "throttle_position": 25.0,
    "dtcs": []
  },
  "location": {
    "latitude": 51.5074,
    "longitude": -0.1278,
    "altitude": 100,
    "speed": 60.5,
    "heading": 180,
    "fix_quality": 1
  }
}
```

### Analysis Data (`/ws/analysis`)
```json
{
  "timestamp": "2025-08-16T10:00:00Z",
  "vin": "1HGCM82633A123456",
  "analysis": {
    "efficiency_score": 85.5,
    "driving_phase": "cruise",
    "alerts": [{
      "type": "engine_temp",
      "severity": "warning",
      "message": "Engine temperature approaching threshold"
    }]
  }
}
```

4. Access the dashboard:
Open a web browser and navigate to `http://<raspberry-pi-ip>:8080`

## Port Configuration
The application uses the following default ports:
- Web Interface: 8080
- OBD2/CANBUS: /dev/ttyUSB0 (adjustable in configuration)

## Development Guide

### Testing Environment Setup
You can test the application without an actual vehicle using the virtual CAN interface (vcan):

1. Set up virtual CAN interface (Linux/WSL):
```bash
# Load the vcan kernel module
sudo modprobe vcan

# Create a virtual CAN interface
sudo ip link add dev vcan0 type vcan
sudo ip link set up vcan0
```

2. For Windows development:
```powershell
# Install com0com (Null-modem emulator) for virtual serial ports
# Download from: http://com0com.sourceforge.net/

# Launch the setup utility and create a virtual port pair:
# COM10 <-> COM11 (You can use these for testing OBD communications)
```

3. Run the simulator (included in ./testing/simulator.go):
```bash
go run testing/simulator.go
```

The simulator provides:
- Simulated OBD2 responses
- Mock vehicle data:
  - RPM: Varies between 800-3000
  - Speed: 0-120 km/h
  - Temperature: 80-95°C
  - Random DTCs for testing
- Virtual CAN bus messages

### Testing Features
- Mock data generation
- Simulated fault conditions
- Network disconnection testing
- DTC code injection
- Performance testing under load

### Running Tests
```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run the test suite with mock data
go run main.go --mock-data
```

### Test Coverage
```bash
go test -cover ./...
```

## Data Capture and Replay

### Capturing CAN Bus Data
The application can record all CAN bus and OBD2 data for later analysis:

```bash
# Start the application with data capture enabled
go run main.go --capture

# Specify a custom capture file name
go run main.go --capture --capture-file="my_session.json"
```

Captured data includes:
- Timestamp for each frame
- CAN bus IDs and data
- OBD2 responses
- Vehicle information
- Session metadata

### Replaying Captured Data
Use the replay tool to analyze captured sessions:

```bash
# List available capture files
go run cmd/replay/main.go --list

# Replay a specific capture file
go run cmd/replay/main.go --file captures/capture_1234567890.json

# Adjust replay speed (2x faster)
go run cmd/replay/main.go --file captures/capture_1234567890.json --speed 2.0
```

### Capture File Format
Captures are stored in JSON format with the following structure:
```json
{
  "start_time": 1628697600,
  "end_time": 1628697900,
  "vehicle_info": "Hyundai iLoad 2.5 CRDi",
  "frames": [
    {
      "timestamp": 1628697600000000000,
      "id": 2024,
      "data": [0x02, 0x01, 0x0C, 0x00, 0x00, 0x00, 0x00, 0x00],
      "type": "OBD2"
    }
    // ... more frames
  ]
}
```

### Analysis Tools
The application includes comprehensive analysis and visualization capabilities.

#### Command Line Analysis
```bash
# Basic session analysis
go run cmd/analyze/main.go --file captures/my_session.json

# Export detailed data to CSV
go run cmd/analyze/main.go --file captures/my_session.json --export-csv=analysis.csv

# Full analysis with driving profile
go run cmd/analyze/main.go --file captures/my_session.json --full
```

#### Analysis Output Example
```
Session Analysis for session_20250811.json
=================================
Duration: 1h15m30s
Total Frames: 45300
Unique CAN IDs: 12

Performance Metrics:
- Max RPM: 4200.00
- Average RPM: 2150.75
- Max Speed: 110.50 km/h
- Average Speed: 65.30 km/h
- Data Rate: 10.05 frames/sec

Driving Behavior:
- Idle Time: 15.5%
- Rapid Accelerations: 23
- Rapid Decelerations: 18

Driving Profile:
- Acceleration Phases: 45
- Deceleration Phases: 42
- Cruising Phases: 38
- Idle Phases: 12
- Total Cruising Time: 45.5 minutes
- Total Idle Time: 11.2 minutes
```

#### Analysis Features

1. **Performance Analysis**
   - Engine Performance
     * RPM statistics and distribution
     * Engine temperature patterns
     * Power output estimates
     * Idle stability analysis
   
   - Vehicle Dynamics
     * Speed profiling
     * Acceleration/deceleration rates
     * Gear shifting patterns
     * Performance benchmarking

2. **Diagnostic Intelligence**
   - DTC Analysis
     * Occurrence frequency and patterns
     * Correlation with driving conditions
     * Historical comparison
     * Root cause analysis
   
   - System Health
     * Sensor response times
     * Data quality validation
     * System latency measurements
     * Communication reliability

3. **Driving Behavior Analytics**
   - Pattern Recognition
     * Driving style classification
     * Efficiency scoring
     * Risk factor analysis
     * Behavioral trends
   
   - Trip Analysis
     * Route profiling
     * Stop/start patterns
     * Urban vs highway detection
     * Fuel efficiency estimation

4. **Data Export and Integration**
   - Export Formats
     * CSV for spreadsheet analysis
     * JSON for programmatic access
     * Time-series formatting
     * Custom data filtering
   
   - Integration Options
     * Excel/Google Sheets templates
     * Python analysis scripts
     * Grafana dashboards
     * Jupyter notebook examples

#### Visualization Tools

1. **Real-time Dashboard**
   - Live data monitoring
   - Customizable layouts
   - Mobile-responsive design
   - Alert configuration

2. **Historical Analysis**
   - Time-series plotting
   - Trend analysis
   - Comparison tools
   - Pattern highlighting

3. **Report Generation**
   - PDF summary reports
   - Custom metrics
   - Graph generation
   - Data annotations

4. **Integration Examples**
   ```python
   # Python visualization example
   import pandas as pd
   import matplotlib.pyplot as plt

   # Load captured data
   data = pd.read_csv('analysis.csv')
   
   # Create performance plot
   plt.figure(figsize=(12,6))
   plt.plot(data['Timestamp'], data['RPM'], label='Engine RPM')
   plt.plot(data['Timestamp'], data['Speed'], label='Vehicle Speed')
   plt.title('Vehicle Performance Analysis')
   plt.xlabel('Time')
   plt.ylabel('Value')
   plt.legend()
   plt.show()
   ```

#### Advanced Usage

1. **Custom Analysis**
   ```bash
   # Analyze specific metrics
   go run cmd/analyze/main.go --file capture.json --metrics=rpm,speed,temp

   # Generate detailed report
   go run cmd/analyze/main.go --file capture.json --report --output=report.pdf
   ```

2. **Batch Processing**
   ```bash
   # Analyze multiple sessions
   go run cmd/analyze/main.go --dir captures/ --batch --export-csv

   # Compare sessions
   go run cmd/analyze/main.go --compare session1.json session2.json
   ```

3. **Real-time Analysis**
   ```bash
   # Start live analysis
   go run main.go --analyze-live --threshold-rpm=4000
   ```

## Vehicle Data Querying

### Query Commands
```bash
# Query all vehicle data
go run cmd/query/main.go --query all

# Query ECU details
go run cmd/query/main.go --query ecu --json

# Query engine maps
go run cmd/query/main.go --query maps --output maps.json

# Monitor live data
go run cmd/query/main.go --query live --continuous
```

### Available Query Types

1. **ECU Information**
   - Hardware/Software versions
   - Calibration IDs
   - Manufacturer data
   - Protocol information
   - Support status

2. **Engine Maps**
   - Fuel injection maps
   - Ignition timing maps
   - Boost control maps
   - Volumetric efficiency maps
   - Calibration data

3. **Vehicle Systems**
   - Transmission data
   - Fuel system parameters
   - Emissions controls
   - Sensor calibrations
   - System capabilities

4. **Live Data Monitoring**
   - Real-time sensor readings
   - Performance metrics
   - System status
   - Fault monitoring
   - Operating conditions

### Sample Output
```json
{
  "vin": "KMFWBX7KPHU123456",
  "ecus": {
    "ENGINE": {
      "id": "ECM",
      "hardwareVersion": "H-1.2.3",
      "softwareVersion": "S-2.4.5",
      "manufacturer": "Hyundai",
      "protocol": "ISO 15765-4 (CAN)",
      "calibrationId": "CAL-123456"
    }
  },
  "engineMaps": {
    "fuelMaps": {
      "idle": [1.0, 1.2, 1.3],
      "low": [1.5, 1.7, 1.9],
      "mid": [2.0, 2.2, 2.4],
      "high": [2.5, 2.7, 2.9]
    },
    "ignitionMap": {
      "idle": [10, 12, 14],
      "low": [16, 18, 20],
      "mid": [22, 24, 26],
      "high": [28, 30, 32]
    }
  }
}
```

## Contributing
Contributions are welcome! Please feel free to submit pull requests, particularly for:
- Additional Hyundai iLoad-specific DTCs
- Enhanced sensor support
- UI improvements
- Documentation updates
