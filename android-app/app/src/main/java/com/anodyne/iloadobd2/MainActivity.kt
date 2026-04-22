package com.anodyne.iloadobd2

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.anodyne.iloadobd2.data.TelemetryMode
import com.anodyne.iloadobd2.ui.AppScreen

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        val prefs = getSharedPreferences("iload_settings", MODE_PRIVATE)
        val initialHost = prefs.getString("host", "192.168.1.100") ?: "192.168.1.100"
        val initialPort = prefs.getInt("port", 8080)
        val initialMode = runCatching {
            TelemetryMode.valueOf(prefs.getString("mode", TelemetryMode.WEBSOCKET.name) ?: TelemetryMode.WEBSOCKET.name)
        }.getOrDefault(TelemetryMode.WEBSOCKET)

        setContent {
            var host by remember { mutableStateOf(initialHost) }
            var port by remember { mutableIntStateOf(initialPort) }
            var mode by remember { mutableStateOf(initialMode) }
            var viewModel by remember {
                mutableStateOf(
                    DashboardViewModelFactory.create(host = host, port = port, mode = mode),
                )
            }

            MaterialTheme {
                Surface {
                    AppScreen(
                        viewModel = viewModel,
                        currentHost = host,
                        currentPort = port,
                        currentMode = mode,
                        onApplySettings = { nextHost, nextPort, nextMode ->
                            host = nextHost
                            port = nextPort
                            mode = nextMode
                            prefs.edit()
                                .putString("host", host)
                                .putInt("port", port)
                                .putString("mode", mode.name)
                                .apply()
                            viewModel = DashboardViewModelFactory.create(
                                host = host,
                                port = port,
                                mode = mode,
                            )
                        },
                    )
                }
            }
        }
    }

    override fun onDestroy() {
        DashboardViewModelFactory.dispose()
        super.onDestroy()
    }
}
