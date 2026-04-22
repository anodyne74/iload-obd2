# Android App (Migration WIP)

This folder contains the native Android replacement for the previous Raspberry Pi web UI.

## Current status

Implemented:
- Kotlin + Compose app scaffold
- WebSocket telemetry client for existing backend stream at /ws
- Tabbed app shell with Dashboard, ECU Info, Engine Maps, DTC, and Settings tabs
- Dashboard screen showing RPM, speed, temperature, DTC count, and capture state
- Capture control buttons for start, stop, and status via WebSocket commands
- In-app runtime settings for host, port, and telemetry mode with reconnect
- Connection settings persisted across app restarts
- Engine maps parity screen with fuel/timing matrix visualization
- DTC details view with built-in code description dictionary
- BLE direct mode transport with bonded-device Bluetooth connection and JSON line parsing
- Runtime BLE permission request flow when switching to BLE mode
- Basic BLE device picker (paired-device list + manual address entry)
- BLE settings notices for permission denial and empty paired-device state
- Manual paired-device refresh action in Settings
- BLE diagnostics panel (permission/adapter/Bluetooth/selected-device status)
- Active nearby Bluetooth discovery scan from Settings
- Explicit scan stop control and scan duration indicator in Settings
- BLE MAC address normalization and format validation before apply
- RSSI capture and strongest-signal-first ordering for discovered BLE devices
- Signal quality badge and last-seen age labels in BLE picker
- Automatic pruning of stale discovered BLE devices during longer sessions

Not implemented yet:
- Full production-grade map visualization and interaction tooling

## Open in Android Studio

1. Open folder android-app as a standalone project.
2. Let Gradle sync complete.
3. Run app module on a device/emulator.

## Backend expectation

Default app settings connect to host 192.168.1.100 port 8080 in WEBSOCKET mode.

You can change host, port, and mode inside the Settings tab and apply without rebuilding.

## Runtime contract

See docs/telemetry-contract.md for payload and command contract details.
