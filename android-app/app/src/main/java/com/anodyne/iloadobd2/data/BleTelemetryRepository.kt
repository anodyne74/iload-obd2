package com.anodyne.iloadobd2.data

import android.Manifest
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothSocket
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import androidx.core.content.ContextCompat
import com.anodyne.iloadobd2.model.TelemetryData
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import java.io.BufferedReader
import java.io.BufferedWriter
import java.io.InputStreamReader
import java.io.OutputStreamWriter
import java.util.UUID

class BleTelemetryRepository(
    private val context: Context,
    private val preferredDeviceAddress: String? = null,
    private val json: Json = Json { ignoreUnknownKeys = true },
) : TelemetryRepository {

    companion object {
        private val sppUuid: UUID = UUID.fromString("00001101-0000-1000-8000-00805F9B34FB")
        private val deviceNameHints = listOf("obd", "mx", "vlink", "elm", "scan")
    }

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    private var socket: BluetoothSocket? = null
    private var writer: BufferedWriter? = null

    private val telemetryState = MutableStateFlow(TelemetryData())
    private val captureEventStream = MutableSharedFlow<CaptureControlResponse>(extraBufferCapacity = 32)

    override val telemetry: Flow<TelemetryData> = telemetryState.asStateFlow()
    override val captureEvents: Flow<CaptureControlResponse> = captureEventStream.asSharedFlow()

    override suspend fun connect() {
        if (socket?.isConnected == true) {
            return
        }

        if (!hasBluetoothConnectPermission()) {
            emitError("Missing Bluetooth permission (BLUETOOTH_CONNECT)")
            return
        }

        val adapter = BluetoothAdapter.getDefaultAdapter()
        if (adapter == null) {
            emitError("Bluetooth is not available on this device")
            return
        }

        if (!adapter.isEnabled) {
            emitError("Bluetooth is disabled")
            return
        }

        val device = selectCandidateDevice(adapter.bondedDevices)
        if (device == null) {
            emitError("No bonded OBD/BLE adapter found")
            return
        }

        runCatching {
            adapter.cancelDiscovery()
            val nextSocket = device.createRfcommSocketToServiceRecord(sppUuid)
            nextSocket.connect()

            socket = nextSocket
            writer = BufferedWriter(OutputStreamWriter(nextSocket.outputStream))

            captureEventStream.tryEmit(
                CaptureControlResponse(
                    type = "capture_control",
                    status = "connected",
                    message = "Connected to ${device.name ?: device.address}",
                ),
            )

            startReader(nextSocket)
        }.onFailure {
            emitError(it.message ?: "Bluetooth connection failed")
            disconnectInternal()
        }
    }

    override suspend fun disconnect() {
        disconnectInternal()
    }

    override suspend fun startCapture(label: String?, metadata: Map<String, String>?) {
        sendCaptureCommand(CaptureCommand(command = "start", label = label, metadata = metadata))
    }

    override suspend fun stopCapture() {
        sendCaptureCommand(CaptureCommand(command = "stop"))
    }

    override suspend fun requestCaptureStatus() {
        sendCaptureCommand(CaptureCommand(command = "status"))
    }

    private fun hasBluetoothConnectPermission(): Boolean {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.S) {
            return true
        }

        return ContextCompat.checkSelfPermission(
            context,
            Manifest.permission.BLUETOOTH_CONNECT,
        ) == PackageManager.PERMISSION_GRANTED
    }

    private fun selectCandidateDevice(devices: Set<BluetoothDevice>): BluetoothDevice? {
        if (devices.isEmpty()) {
            return null
        }

        val preferred = preferredDeviceAddress?.trim()?.lowercase()
        if (!preferred.isNullOrEmpty()) {
            devices.firstOrNull { device ->
                device.address.lowercase() == preferred
            }?.let { return it }
        }

        return devices.firstOrNull { device ->
            val name = (device.name ?: "").lowercase()
            deviceNameHints.any { hint -> name.contains(hint) }
        } ?: devices.firstOrNull()
    }

    private fun startReader(activeSocket: BluetoothSocket) {
        scope.launch {
            runCatching {
                BufferedReader(InputStreamReader(activeSocket.inputStream)).use { reader ->
                    while (true) {
                        val line = reader.readLine() ?: break
                        if (line.isNotBlank()) {
                            handleInboundMessage(line)
                        }
                    }
                }
            }.onFailure {
                emitError(it.message ?: "Bluetooth stream closed")
            }

            disconnectInternal()
        }
    }

    private suspend fun sendCaptureCommand(command: CaptureCommand) {
        val payload = json.encodeToString(CaptureCommand.serializer(), command)
        val activeWriter = writer
        if (activeWriter == null) {
            emitError("Bluetooth transport is not connected")
            return
        }

        runCatching {
            withContext(Dispatchers.IO) {
                activeWriter.write(payload)
                activeWriter.newLine()
                activeWriter.flush()
            }
        }.onFailure {
            emitError(it.message ?: "Failed to send BLE command")
        }
    }

    private fun handleInboundMessage(payload: String) {
        runCatching {
            json.decodeFromString(CaptureControlResponse.serializer(), payload)
        }.onSuccess { response ->
            if (response.type == "capture_control") {
                captureEventStream.tryEmit(response)
                return
            }
        }

        runCatching {
            json.decodeFromString(TelemetryData.serializer(), payload)
        }.onSuccess { data ->
            telemetryState.value = data
        }
    }

    private suspend fun emitError(message: String) {
        captureEventStream.emit(
            CaptureControlResponse(
                type = "capture_control",
                status = "error",
                error = message,
            ),
        )
    }

    private fun disconnectInternal() {
        runCatching { writer?.close() }
        writer = null

        runCatching { socket?.close() }
        socket = null
    }

    fun clear() {
        disconnectInternal()
        scope.cancel()
    }
}
