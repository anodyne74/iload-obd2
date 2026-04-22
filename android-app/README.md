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

Not implemented yet:
- Direct BLE mode for MX201
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
