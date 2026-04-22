# iload-obd2
![iload](/static/images/my-vehicle-dark.png)

A comprehensive vehicle telemetry and diagnostics platform for the Hyundai iLoad/H-1.

## Overview
This application provides real-time monitoring, diagnostics, capture control, and data analysis through a BLE OBD2 adapter workflow. While optimized for the Hyundai iLoad/H-1, the architecture supports multiple vehicle types.

The repository currently contains two primary operator surfaces:
- A Go backend that talks to vehicle transports, records telemetry/capture data, and exposes a WebSocket telemetry stream.
- A native Android app in [android-app](android-app) that is replacing the earlier browser-based dashboard.

## Core Features
- Real-time monitoring of:
  - Engine RPM
  - Vehicle Speed
  - Engine Temperature
  - Engine Maps (Fuel and Timing)
  - ECU Information
  - DTCs (Diagnostic Trouble Codes)
  - Capture state and frame counts
- OBD2 adapter transport support:
  - BLE only
- Live telemetry delivery over WebSocket for UI clients
- Capture control commands for start, stop, and status
- Capture metadata propagation for VIN/profile identity (`vin`, `profile_id`, `profile_version`, `profile_sync_state`, `make`, `model`, `year`)
- Local profile cloud-sync API in backend:
  - `POST /api/v1/vehicle-profiles/upsert`
  - `GET /api/v1/vehicle-profiles`
- Native Android client with:
  - Dashboard, ECU Info, Engine Maps, DTC, and Settings tabs
  - Runtime host/port/mode reconfiguration
  - WebSocket telemetry mode
  - BLE direct mode with paired-device picker, nearby device scan, RSSI sorting, and connection diagnostics
  - Automatic vehicle profile creation when a new VIN is detected
  - In-app vehicle profile editing (display name, notes, tags, make/model/year)
  - Background profile sync with manual `Sync Profiles Now` and `Retry Failed Only` actions
  - Profile sync diagnostics card (pending/failed/synced counts and last successful sync)

## Requirements

### Hardware
- BLE-capable OBD2 adapter
- Android device/emulator if using the native app

### Software Prerequisites
- Go 1.21 or later
- Android Studio (for Android app development)

Notes:
- SQLite/InfluxDB fields still exist in config for compatibility, but persistent vehicle profile and trip-history ownership is moving to cloud sync workflows.
- Raspberry Pi is not required.

## Installation

### Local Setup

1. **Install Application**
```bash
# Clone repository
git clone https://github.com/anodyne74/iload-obd2.git
cd iload-obd2
```

2. **Run backend locally (optional for WebSocket mode)**
```bash
go run .
```

### Configuration

Edit `config.yaml`:
```yaml
transport:
  type: "ble"
  ble:
    macAddress: "AA:BB:CC:DD:EE:FF"
```

## Development

### VS Code Development
1. **Prerequisites**
   - VS Code installed
   - Go extension installed

2. **Build and Run**
  - Use the default build/test tasks in VS Code
  - Or run `go build ./...` and `go test ./...` in terminal

### Testing Environment
```bash
# Run unit tests
go test ./...

# Run the backend locally
go run .
```

### Local Profile Sync API Smoke Test
```bash
# Upsert profile payload (local stub)
curl -sS -X POST http://localhost:8080/api/v1/vehicle-profiles/upsert \
  -H 'Content-Type: application/json' \
  -d '{"profile":{"id":"profile-1","vin":"KMH12345678901234","displayName":"My iLoad"}}'

# List stored profiles (local stub)
curl -sS http://localhost:8080/api/v1/vehicle-profiles
```

### Building
```bash
go build ./...
```

### Android App
The Android client lives in [android-app](android-app) and is designed to connect to the backend WebSocket stream or use BLE direct mode.

Current Android implementation includes:
- Dashboard telemetry and capture controls
- ECU info, engine maps, and DTC views
- BLE permission handling and diagnostics
- Paired-device refresh and nearby Bluetooth scan
- RSSI-aware device ordering and stale-device pruning
- Vehicle profile lifecycle:
  - auto-create on first vehicle detection
  - local JSON profile catalog persistence
  - in-app profile editing
  - periodic + manual cloud sync and failed-only retry
- Profile sync diagnostics and status feedback in Settings
- Capture start metadata enrichment with active profile identity fields

To work on the Android app:
1. Open [android-app](android-app) as a standalone project in Android Studio.
2. Let Gradle sync complete.
3. Run the `app` module on a device or emulator.

For Android-specific details, see [android-app/README.md](android-app/README.md).

## Troubleshooting

1. **Backend startup issues**
```bash
go run .
```

2. **Run tests**
```bash
go test ./...
```

3. **Profile sync API smoke test**
```bash
curl -sS -X POST http://localhost:8080/api/v1/vehicle-profiles/upsert \
  -H 'Content-Type: application/json' \
  -d '{"profile":{"id":"profile-1","vin":"KMH12345678901234","displayName":"My iLoad"}}'

curl -sS http://localhost:8080/api/v1/vehicle-profiles
```

## Contributing
Contributions are welcome! Please submit pull requests for:
- Additional vehicle support
- Enhanced diagnostics
- UI improvements
- Documentation updates

## License
[Add your license information here]
