# Telemetry and Capture Contract

This document defines the Android-side payload contract mirrored from the Go runtime.

## Telemetry Payload

The runtime broadcasts one telemetry payload per second over WebSocket (`/ws`).

Fields:
- `rpm` number
- `speed` number
- `temp` number
- `dtcs` string array
- `ecuInfo` object
- `engineMaps` object
- `canFrames` array
- `capture` object

## Capture Control Commands

Android sends one of the following command payloads over WebSocket:

- `{"command":"start"}`
- `{"command":"start","label":"trip-name"}`
- `{"command":"stop"}`
- `{"command":"status"}`

Supported aliases in runtime:
- `start_capture`
- `stop_capture`
- `capture_status`

## Capture Control Response

The runtime responds on the same WebSocket connection:

- `type` (always `capture_control`)
- `status` (`recording`, `stopped`, `idle`, or `error`)
- `message` optional
- `error` optional
- `filePath` optional
- `frameCount` optional
- `durationSeconds` optional
