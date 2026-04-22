package com.anodyne.iloadobd2

import android.content.Context
import com.anodyne.iloadobd2.data.WebSocketTelemetryRepository
import com.anodyne.iloadobd2.data.BleTelemetryRepository
import com.anodyne.iloadobd2.data.TelemetryMode
import com.anodyne.iloadobd2.viewmodel.DashboardViewModel
import okhttp3.OkHttpClient
import java.util.concurrent.TimeUnit

object DashboardViewModelFactory {
    private data class ConnectionConfig(
        val host: String,
        val port: Int,
        val mode: TelemetryMode,
        val bleDeviceAddress: String?,
    )

    private val client: OkHttpClient by lazy {
        OkHttpClient.Builder()
            .connectTimeout(5, TimeUnit.SECONDS)
            .readTimeout(0, TimeUnit.MILLISECONDS)
            .build()
    }

    private var current: DashboardViewModel? = null
    private var currentConfig: ConnectionConfig? = null

    fun create(
        context: Context,
        host: String = "192.168.1.100",
        port: Int = 8080,
        mode: TelemetryMode = TelemetryMode.WEBSOCKET,
        bleDeviceAddress: String? = null,
    ): DashboardViewModel {
        val requested = ConnectionConfig(
            host = host,
            port = port,
            mode = mode,
            bleDeviceAddress = bleDeviceAddress,
        )
        if (current == null || currentConfig != requested) {
            current?.clear()

            val repository = when (mode) {
                TelemetryMode.WEBSOCKET -> WebSocketTelemetryRepository(
                    host = host,
                    port = port,
                    client = client,
                )

                TelemetryMode.BLE_DIRECT -> BleTelemetryRepository(
                    context = context.applicationContext,
                    preferredDeviceAddress = bleDeviceAddress,
                )
            }

            current = DashboardViewModel(repository = repository)
            currentConfig = requested
        }

        return requireNotNull(current)
    }

    fun dispose() {
        current?.clear()
        current = null
        currentConfig = null
    }
}
