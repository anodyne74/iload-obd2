package com.anodyne.iloadobd2.viewmodel

import com.anodyne.iloadobd2.data.CaptureControlResponse
import com.anodyne.iloadobd2.data.TelemetryRepository
import com.anodyne.iloadobd2.model.TelemetryData
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

class DashboardViewModel(
    private val repository: TelemetryRepository,
) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)

    private val _telemetry = MutableStateFlow(TelemetryData())
    val telemetry: StateFlow<TelemetryData> = _telemetry.asStateFlow()

    private val _captureResponse = MutableStateFlow<CaptureControlResponse?>(null)
    val captureResponse: StateFlow<CaptureControlResponse?> = _captureResponse.asStateFlow()

    private val _connectionState = MutableStateFlow("connecting")
    val connectionState: StateFlow<String> = _connectionState.asStateFlow()

    init {
        scope.launch {
            runCatching {
                repository.connect()
            }.onSuccess {
                _connectionState.value = "connected"
            }.onFailure {
                _connectionState.value = "error"
                _captureResponse.value = CaptureControlResponse(
                    type = "capture_control",
                    status = "error",
                    error = it.message ?: "connection failed",
                )
            }
        }

        scope.launch {
            repository.telemetry.collect { data ->
                _telemetry.value = data
            }
        }

        scope.launch {
            repository.captureEvents.collect { event ->
                _captureResponse.value = event
            }
        }
    }

    fun startCapture(label: String?) {
        scope.launch { repository.startCapture(label) }
    }

    fun stopCapture() {
        scope.launch { repository.stopCapture() }
    }

    fun requestCaptureStatus() {
        scope.launch { repository.requestCaptureStatus() }
    }

    fun clear() {
        scope.launch {
            repository.disconnect()
            _connectionState.value = "disconnected"
        }
        scope.cancel()
    }
}
