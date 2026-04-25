package com.anodyne.iloadobd2

import android.annotation.SuppressLint
import android.Manifest
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothManager
import android.content.BroadcastReceiver
import android.content.Intent
import android.content.IntentFilter
import android.content.pm.PackageManager
import android.os.Bundle
import android.os.Build
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.core.content.edit
import androidx.core.content.ContextCompat
import com.anodyne.iloadobd2.data.TelemetryMode
import com.anodyne.iloadobd2.ui.AppScreen
import com.anodyne.iloadobd2.ui.BleDiagnostics
import com.anodyne.iloadobd2.ui.BleDeviceOption
import kotlinx.coroutines.delay

private data class PendingConnectionSettings(
    val host: String,
    val port: Int,
    val mode: TelemetryMode,
    val bleDeviceAddress: String?,
)

private const val DISCOVERED_DEVICE_STALE_AFTER_MILLIS = 45_000L
private const val DISCOVERED_DEVICE_PRUNE_INTERVAL_MILLIS = 5_000L

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
        val initialBleDeviceAddress = prefs.getString("ble_device_address", null)

        setContent {
            var host by remember { mutableStateOf(initialHost) }
            var port by remember { mutableIntStateOf(initialPort) }
            var mode by remember { mutableStateOf(initialMode) }
            var bleDeviceAddress by remember { mutableStateOf(initialBleDeviceAddress) }
            var pendingSettings by remember { mutableStateOf<PendingConnectionSettings?>(null) }
            var bondedBleDevices by remember { mutableStateOf(readBondedBleDevices()) }
            var discoveredBleDevices by remember { mutableStateOf<List<BleDeviceOption>>(emptyList()) }
            var isBleScanInProgress by remember { mutableStateOf(false) }
            var bleScanStartedAtMillis by remember { mutableStateOf<Long?>(null) }
            var settingsNotice by remember { mutableStateOf<String?>(null) }
            var viewModel by remember {
                mutableStateOf(
                    DashboardViewModelFactory.create(
                        context = applicationContext,
                        host = host,
                        port = port,
                        mode = mode,
                        bleDeviceAddress = bleDeviceAddress,
                    ),
                )
            }

            DisposableEffect(Unit) {
                val receiver = object : BroadcastReceiver() {
                    override fun onReceive(context: android.content.Context?, intent: Intent?) {
                        when (intent?.action) {
                            BluetoothDevice.ACTION_FOUND -> {
                                val device = extractBluetoothDevice(intent)
                                if (device != null && hasBlePermissions()) {
                                    val rssi = extractRssi(intent)
                                    val option = BleDeviceOption(
                                        name = bluetoothDeviceName(device),
                                        address = device.address,
                                        rssi = rssi,
                                        lastSeenMillis = System.currentTimeMillis(),
                                    )
                                    discoveredBleDevices = (discoveredBleDevices.filterNot { it.address == option.address } + option)
                                        .sortedWith(
                                            compareByDescending<BleDeviceOption> { it.rssi ?: Int.MIN_VALUE }
                                                .thenByDescending { it.lastSeenMillis ?: Long.MIN_VALUE }
                                                .thenBy { it.name.lowercase() },
                                        )
                                }
                            }

                            BluetoothAdapter.ACTION_DISCOVERY_FINISHED -> {
                                isBleScanInProgress = false
                                bleScanStartedAtMillis = null
                                if (discoveredBleDevices.isEmpty()) {
                                    settingsNotice = "Scan complete: no nearby Bluetooth devices found."
                                }
                            }
                        }
                    }
                }

                val filter = IntentFilter().apply {
                    addAction(BluetoothDevice.ACTION_FOUND)
                    addAction(BluetoothAdapter.ACTION_DISCOVERY_FINISHED)
                }

                registerReceiver(receiver, filter)

                onDispose {
                    runCatching { unregisterReceiver(receiver) }
                    cancelBluetoothDiscovery()
                }
            }

            LaunchedEffect(discoveredBleDevices.isNotEmpty()) {
                if (discoveredBleDevices.isEmpty()) {
                    return@LaunchedEffect
                }

                while (true) {
                    delay(DISCOVERED_DEVICE_PRUNE_INTERVAL_MILLIS)
                    val now = System.currentTimeMillis()
                    discoveredBleDevices = discoveredBleDevices.filter { option ->
                        val lastSeenMillis = option.lastSeenMillis ?: return@filter true
                        now - lastSeenMillis <= DISCOVERED_DEVICE_STALE_AFTER_MILLIS
                    }

                    if (discoveredBleDevices.isEmpty()) {
                        break
                    }
                }
            }

            fun startBleScan() {
                if (!hasBlePermissions()) {
                    settingsNotice = "Bluetooth permission is required before scanning."
                    return
                }

                val adapter = bluetoothAdapter()
                if (adapter == null) {
                    settingsNotice = "Bluetooth adapter is unavailable on this device."
                    return
                }

                if (!adapter.isEnabled) {
                    settingsNotice = "Bluetooth is disabled. Enable it to scan for devices."
                    return
                }

                settingsNotice = null
                discoveredBleDevices = emptyList()
                isBleScanInProgress = restartBluetoothDiscovery(adapter)
                if (!isBleScanInProgress) {
                    settingsNotice = "Unable to start Bluetooth scan."
                    bleScanStartedAtMillis = null
                } else {
                    bleScanStartedAtMillis = System.currentTimeMillis()
                }
            }

            fun stopBleScan() {
                cancelBluetoothDiscovery()
                isBleScanInProgress = false
                bleScanStartedAtMillis = null
                settingsNotice = "Bluetooth scan stopped."
            }

            fun applySettings(nextHost: String, nextPort: Int, nextMode: TelemetryMode, nextBleDeviceAddress: String?) {
                host = nextHost
                port = nextPort
                mode = nextMode
                bleDeviceAddress = nextBleDeviceAddress
                settingsNotice = null
                prefs.edit {
                    putString("host", host)
                    putInt("port", port)
                    putString("mode", mode.name)
                    putString("ble_device_address", bleDeviceAddress)
                }
                viewModel = DashboardViewModelFactory.create(
                    context = applicationContext,
                    host = host,
                    port = port,
                    mode = mode,
                    bleDeviceAddress = bleDeviceAddress,
                )
            }

            val permissionLauncher = rememberLauncherForActivityResult(
                contract = ActivityResultContracts.RequestMultiplePermissions(),
            ) { result ->
                val granted = result.values.all { it }
                val pending = pendingSettings
                pendingSettings = null
                bondedBleDevices = readBondedBleDevices()

                if (granted && pending != null) {
                    applySettings(
                        nextHost = pending.host,
                        nextPort = pending.port,
                        nextMode = pending.mode,
                        nextBleDeviceAddress = pending.bleDeviceAddress,
                    )
                } else if (!granted) {
                    settingsNotice = "Bluetooth permission denied. BLE mode requires BLUETOOTH_CONNECT and BLUETOOTH_SCAN."
                }
            }

            MaterialTheme {
                Surface {
                    AppScreen(
                        viewModel = viewModel,
                        currentHost = host,
                        currentPort = port,
                        currentMode = mode,
                        currentBleDeviceAddress = bleDeviceAddress,
                        bondedBleDevices = bondedBleDevices,
                        discoveredBleDevices = discoveredBleDevices,
                                                isBleScanInProgress = isBleScanInProgress,
                                                bleScanStartedAtMillis = bleScanStartedAtMillis,
                        bleDiagnostics = readBleDiagnostics(selectedAddress = bleDeviceAddress),
                        settingsNotice = settingsNotice,
                        onRefreshBleDevices = {
                            bondedBleDevices = readBondedBleDevices()
                            settingsNotice = if (bondedBleDevices.isEmpty()) {
                                "No paired Bluetooth devices found. Pair your adapter in system Bluetooth settings first."
                            } else {
                                null
                            }
                        },
                                                onStartBleScan = { startBleScan() },
                                                onStopBleScan = { stopBleScan() },
                        onApplySettings = { nextHost, nextPort, nextMode, nextBleDeviceAddress ->
                            if (nextMode == TelemetryMode.BLE_DIRECT && !hasBlePermissions()) {
                                pendingSettings = PendingConnectionSettings(
                                    host = nextHost,
                                    port = nextPort,
                                    mode = nextMode,
                                    bleDeviceAddress = nextBleDeviceAddress,
                                )
                                permissionLauncher.launch(requiredBlePermissions())
                            } else {
                                applySettings(nextHost, nextPort, nextMode, nextBleDeviceAddress)
                            }
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

    private fun hasBlePermissions(): Boolean {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.S) {
            return true
        }

        return requiredBlePermissions().all { permission ->
            ContextCompat.checkSelfPermission(this, permission) == PackageManager.PERMISSION_GRANTED
        }
    }

    private fun requiredBlePermissions(): Array<String> {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.S) {
            return emptyArray()
        }

        return arrayOf(
            Manifest.permission.BLUETOOTH_CONNECT,
            Manifest.permission.BLUETOOTH_SCAN,
        )
    }

    private fun readBondedBleDevices(): List<BleDeviceOption> {
        if (!hasBlePermissions()) {
            return emptyList()
        }

        val adapter = bluetoothAdapter() ?: return emptyList()
        return bondedDevices(adapter)
            .map { device ->
                BleDeviceOption(
                    name = bluetoothDeviceName(device),
                    address = device.address,
                    rssi = null,
                    lastSeenMillis = null,
                )
            }
            .sortedBy { option -> option.name.lowercase() }
    }

    private fun readBleDiagnostics(selectedAddress: String?): BleDiagnostics {
        val adapter = bluetoothAdapter()
        return BleDiagnostics(
            hasRequiredPermissions = hasBlePermissions(),
            adapterAvailable = adapter != null,
            adapterEnabled = adapter?.isEnabled == true,
            selectedAddress = selectedAddress,
        )
    }

    private fun bluetoothAdapter(): BluetoothAdapter? {
        return getSystemService(BluetoothManager::class.java)?.adapter
    }

    @SuppressLint("MissingPermission")
    private fun cancelBluetoothDiscovery() {
        bluetoothAdapter()?.cancelDiscovery()
    }

    @SuppressLint("MissingPermission")
    private fun restartBluetoothDiscovery(adapter: BluetoothAdapter): Boolean {
        adapter.cancelDiscovery()
        return adapter.startDiscovery()
    }

    @SuppressLint("MissingPermission")
    private fun bondedDevices(adapter: BluetoothAdapter): Set<BluetoothDevice> {
        return adapter.bondedDevices
    }

    @SuppressLint("MissingPermission")
    private fun bluetoothDeviceName(device: BluetoothDevice): String {
        return device.name ?: "Unknown Device"
    }

    @Suppress("DEPRECATION")
    private fun extractBluetoothDevice(intent: Intent): BluetoothDevice? {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            intent.getParcelableExtra(BluetoothDevice.EXTRA_DEVICE, BluetoothDevice::class.java)
        } else {
            intent.getParcelableExtra(BluetoothDevice.EXTRA_DEVICE)
        }
    }

    private fun extractRssi(intent: Intent): Int? {
        val raw = intent.getShortExtra(BluetoothDevice.EXTRA_RSSI, Short.MIN_VALUE)
        return if (raw == Short.MIN_VALUE) {
            null
        } else {
            raw.toInt()
        }
    }
}
