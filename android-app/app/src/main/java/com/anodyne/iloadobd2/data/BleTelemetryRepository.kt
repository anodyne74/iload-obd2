package com.anodyne.iloadobd2.data

import com.anodyne.iloadobd2.model.TelemetryData
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow

class BleTelemetryRepository : TelemetryRepository {
    private val telemetryState = MutableStateFlow(TelemetryData())
    private val captureEventStream = MutableSharedFlow<CaptureControlResponse>(extraBufferCapacity = 32)

    override val telemetry: Flow<TelemetryData> = telemetryState.asStateFlow()
    override val captureEvents: Flow<CaptureControlResponse> = captureEventStream.asSharedFlow()

    override suspend fun connect() {
        captureEventStream.emit(
            CaptureControlResponse(
                type = "capture_control",
                status = "error",
                error = "BLE direct mode is not implemented yet",
            )
        )
    }

    override suspend fun disconnect() {
        // No active transport in placeholder implementation.
    }

    override suspend fun startCapture(label: String?) {
        captureEventStream.emit(
            CaptureControlResponse(
                type = "capture_control",
                status = "error",
                error = "BLE direct mode is not implemented yet",
            )
        )
    }

    override suspend fun stopCapture() {
        captureEventStream.emit(
            CaptureControlResponse(
                type = "capture_control",
                status = "error",
                error = "BLE direct mode is not implemented yet",
            )
        )
    }

    override suspend fun requestCaptureStatus() {
        captureEventStream.emit(
            CaptureControlResponse(
                type = "capture_control",
                status = "idle",
            )
        )
    }
}
