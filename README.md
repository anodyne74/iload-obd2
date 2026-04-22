# iload-obd2
![iload](/static/images/my-vehicle-dark.png)

A comprehensive vehicle telemetry and diagnostics platform for the Hyundai iLoad/H-1.

## Overview
This application provides real-time monitoring, diagnostics, and data analysis through OBD2 and CANBUS interfaces. While optimized for the Hyundai iLoad/H-1, the architecture supports multiple vehicle types.

## Core Features
- Real-time monitoring of:
  - Engine RPM
  - Vehicle Speed
  - Engine Temperature
  - Engine Maps (Fuel and Timing)
  - ECU Information
  - DTCs (Diagnostic Trouble Codes)
- Multiple Transport Options:
  - Serial OBD-II Connection
  - Direct CAN Bus Access
  - TCP Connection (for testing/simulation)

## Requirements

### Hardware
- Raspberry Pi 4 (recommended)
- OBD2 adapter (USB) or CAN interface
- Web browser for dashboard access

### Software Prerequisites
- Go 1.21 or later
- SQLite 3
- InfluxDB 2.x

## Installation

### Raspberry Pi Deployment (Recommended)

1. **Install Required Packages**
```bash
# Update system
sudo apt update
sudo apt upgrade -y

# Install SQLite and CAN utils
sudo apt install -y sqlite3 can-utils

# Install InfluxDB
# Add InfluxData repository
curl -s https://repos.influxdata.com/influxdata-archive_compat.key | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/influxdb.gpg

echo "deb [signed-by=/etc/apt/trusted.gpg.d/influxdb.gpg] https://repos.influxdata.com/debian $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/influxdb.list

# Update and install InfluxDB
sudo apt update
sudo apt install -y influxdb

# Start InfluxDB service
sudo systemctl enable influxdb
sudo systemctl start influxdb

# Configure InfluxDB (run these commands once)
# Start influx CLI
influx

# Create database and user (run these in the influx CLI)
CREATE DATABASE vehicle_telemetry
CREATE USER admin WITH PASSWORD 'your-secure-password' WITH ALL PRIVILEGES
EXIT

# Update your config.yaml to match these settings:
datastore:
  influxdb:
    url: "http://localhost:8086"
    database: "vehicle_telemetry"
    username: "admin"
    password: "your-secure-password"
```

2. **Configure CAN Interface**
```bash
# Load kernel modules
sudo modprobe can
sudo modprobe can_raw

# Set up CAN interface
sudo ip link set can0 type can bitrate 500000
sudo ip link set up can0
```

3. **Install Application**
```bash
# Clone repository
git clone https://github.com/anodyne74/iload-obd2.git
cd iload-obd2

# Run installation script
sudo bash scripts/install.sh
```

4. **Start Service**
```bash
sudo systemctl enable iload-obd2
sudo systemctl start iload-obd2
```

### Configuration

Edit `/opt/iload-obd2/config/config.yaml`:
```yaml
datastore:
  sqlite:
    path: "/opt/iload-obd2/data/sqlite/vehicles.db"
  influxdb:
    url: "http://localhost:8086"
    org: "your-org"
    bucket: "vehicle-telemetry"
    token: "your-token"

transport:
  type: "can"  # or "serial" for OBD2 adapter
  device: "can0"  # or "/dev/ttyUSB0" for serial
  baudrate: 500000
```

## Development

### VS Code Development
1. **Prerequisites**
   - VS Code installed
   - Go extension installed
   - Remote-SSH extension installed

2. **Build and Deploy**
   - Press `Ctrl+Shift+B` to build for Raspberry Pi
   - Press `F1`, type "Tasks: Run Task", select "deploy-pi"
   - Enter your Raspberry Pi's IP address when prompted

3. **Remote Development**
   - Press `F1`, type "Remote-SSH: Connect to Host"
   - Enter `pi@raspberrypi.local` (or your Pi's IP)
   - Open `/opt/iload-obd2` folder
   - Use integrated terminal for commands

### Testing Environment
```bash
# Run with mock data
go run main.go --mock-data

# Run unit tests
go test ./...
```

### Building
```bash
# For Raspberry Pi (ARM64)
GOOS=linux GOARCH=arm64 go build -o iload-obd2
```

## Troubleshooting

1. **Check Service Status**
```bash
systemctl status iload-obd2
```

2. **View Logs**
```bash
journalctl -u iload-obd2 -f
```

3. **Test CAN Interface**
```bash
candump can0
```

## Contributing
Contributions are welcome! Please submit pull requests for:
- Additional vehicle support
- Enhanced diagnostics
- UI improvements
- Documentation updates

## License
[Add your license information here]
