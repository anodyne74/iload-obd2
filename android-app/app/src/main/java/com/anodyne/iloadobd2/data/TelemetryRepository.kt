package com.anodyne.iloadobd2.data

import com.anodyne.iloadobd2.model.TelemetryData
import kotlinx.coroutines.flow.Flow

interface TelemetryRepository {
    val telemetry: Flow<TelemetryData>
    val captureEvents: Flow<CaptureControlResponse>

    suspend fun connect()
    suspend fun disconnect()
    suspend fun startCapture(label: String? = null, metadata: Map<String, String>? = null)
    suspend fun stopCapture()
    suspend fun requestCaptureStatus()
}
