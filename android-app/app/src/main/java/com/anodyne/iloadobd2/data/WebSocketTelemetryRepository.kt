package com.anodyne.iloadobd2.data

import com.anodyne.iloadobd2.model.TelemetryData
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener

class WebSocketTelemetryRepository(
    private val host: String,
    private val port: Int,
    private val client: OkHttpClient,
    private val json: Json = Json { ignoreUnknownKeys = true },
) : TelemetryRepository {

    private var webSocket: WebSocket? = null

    private val telemetryState = MutableStateFlow(TelemetryData())
    private val captureEventStream = MutableSharedFlow<CaptureControlResponse>(extraBufferCapacity = 32)

    override val telemetry: Flow<TelemetryData> = telemetryState.asStateFlow()
    override val captureEvents: Flow<CaptureControlResponse> = captureEventStream.asSharedFlow()

    override suspend fun connect() {
        if (webSocket != null) {
            return
        }

        val request = Request.Builder()
            .url("ws://$host:$port/ws")
            .build()

        webSocket = client.newWebSocket(
            request,
            object : WebSocketListener() {
                override fun onMessage(webSocket: WebSocket, text: String) {
                    handleInboundMessage(text)
                }

                override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                    webSocket.close(code, reason)
                }

                override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                    captureEventStream.tryEmit(
                        CaptureControlResponse(
                            type = "capture_control",
                            status = "error",
                            error = t.message ?: "websocket failure",
                        )
                    )
                }
            },
        )
    }

    override suspend fun disconnect() {
        webSocket?.close(1000, "client disconnect")
        webSocket = null
    }

    override suspend fun startCapture(label: String?) {
        sendCaptureCommand(CaptureCommand(command = "start", label = label))
    }

    override suspend fun stopCapture() {
        sendCaptureCommand(CaptureCommand(command = "stop"))
    }

    override suspend fun requestCaptureStatus() {
        sendCaptureCommand(CaptureCommand(command = "status"))
    }

    private fun sendCaptureCommand(command: CaptureCommand) {
        val payload = json.encodeToString(CaptureCommand.serializer(), command)
        webSocket?.send(payload)
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
}
