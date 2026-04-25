package com.anodyne.iloadobd2.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuAnchorType
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.PrimaryScrollableTabRow
import androidx.compose.material3.Tab
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.anodyne.iloadobd2.data.TelemetryMode
import com.anodyne.iloadobd2.data.dtcDescriptions
import com.anodyne.iloadobd2.model.MapData
import com.anodyne.iloadobd2.viewmodel.DashboardViewModel
import kotlinx.coroutines.delay

data class BleDeviceOption(
    val name: String,
    val address: String,
    val rssi: Int? = null,
    val lastSeenMillis: Long? = null,
)

data class BleDiagnostics(
    val hasRequiredPermissions: Boolean,
    val adapterAvailable: Boolean,
    val adapterEnabled: Boolean,
    val selectedAddress: String?,
)

private enum class AppTab(val label: String) {
    DASHBOARD("Dashboard"),
    ECU("ECU Info"),
    MAPS("Engine Maps"),
    DTC("DTC"),
    SETTINGS("Settings"),
}

@Composable
fun AppScreen(
    viewModel: DashboardViewModel,
    currentHost: String,
    currentPort: Int,
    currentMode: TelemetryMode,
    currentBleDeviceAddress: String?,
    bondedBleDevices: List<BleDeviceOption>,
    discoveredBleDevices: List<BleDeviceOption>,
    isBleScanInProgress: Boolean,
    bleScanStartedAtMillis: Long?,
    bleDiagnostics: BleDiagnostics,
    settingsNotice: String?,
    onRefreshBleDevices: () -> Unit,
    onStartBleScan: () -> Unit,
    onStopBleScan: () -> Unit,
    onApplySettings: (String, Int, TelemetryMode, String?) -> Unit,
) {
    var selectedTab by rememberSaveable { mutableIntStateOf(AppTab.DASHBOARD.ordinal) }

    Column(modifier = Modifier.fillMaxSize()) {
        PrimaryScrollableTabRow(selectedTabIndex = selectedTab) {
            AppTab.entries.forEachIndexed { index, tab ->
                Tab(
                    selected = selectedTab == index,
                    onClick = { selectedTab = index },
                    text = { Text(tab.label) },
                )
            }
        }

        when (AppTab.entries[selectedTab]) {
            AppTab.DASHBOARD -> DashboardScreen(viewModel = viewModel)
            AppTab.ECU -> EcuInfoScreen(viewModel = viewModel)
            AppTab.MAPS -> EngineMapsScreen(viewModel = viewModel)
            AppTab.DTC -> DtcScreen(viewModel = viewModel)
            AppTab.SETTINGS -> SettingsScreen(
                viewModel = viewModel,
                currentHost = currentHost,
                currentPort = currentPort,
                currentMode = currentMode,
                currentBleDeviceAddress = currentBleDeviceAddress,
                bondedBleDevices = bondedBleDevices,
                discoveredBleDevices = discoveredBleDevices,
                isBleScanInProgress = isBleScanInProgress,
                bleScanStartedAtMillis = bleScanStartedAtMillis,
                bleDiagnostics = bleDiagnostics,
                settingsNotice = settingsNotice,
                onRefreshBleDevices = onRefreshBleDevices,
                onStartBleScan = onStartBleScan,
                onStopBleScan = onStopBleScan,
                onApplySettings = onApplySettings,
            )
        }
    }
}

@Composable
private fun EcuInfoScreen(viewModel: DashboardViewModel) {
    val telemetry by viewModel.telemetry.collectAsState()
    val info = telemetry.ecuInfo

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        item {
            Text("ECU Information", style = MaterialTheme.typography.headlineSmall)
        }

        if (info == null) {
            item {
                Text("No ECU data received yet.")
            }
        } else {
            item { InfoRow(label = "VIN", value = info.vin ?: "") }
            item { InfoRow(label = "Version", value = info.version ?: "") }
            item { InfoRow(label = "Hardware", value = info.hardware ?: "") }
            item { InfoRow(label = "Software", value = info.software ?: "") }
            item { InfoRow(label = "Calibration", value = info.calibration ?: "") }
            item { InfoRow(label = "Build Date", value = info.buildDate ?: "") }
            item { InfoRow(label = "Protocol", value = info.protocol ?: "") }
        }
    }
}

@Composable
private fun EngineMapsScreen(viewModel: DashboardViewModel) {
    val telemetry by viewModel.telemetry.collectAsState()
    val maps = telemetry.engineMaps

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        contentPadding = PaddingValues(bottom = 24.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        item {
            Text("Engine Maps", style = MaterialTheme.typography.headlineSmall)
        }

        if (maps == null || (maps.fuel == null && maps.timing == null)) {
            item {
                Text("No engine map data received yet.")
            }
        } else {
            maps.fuel?.let { fuel ->
                item {
                    MapCard(title = "Fuel Map", map = fuel)
                }
            }

            maps.timing?.let { timing ->
                item {
                    MapCard(title = "Timing Map", map = timing)
                }
            }
        }
    }
}

@Composable
private fun MapCard(title: String, map: MapData) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(
            modifier = Modifier.padding(12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(title, style = MaterialTheme.typography.titleMedium)
            Text(
                "${map.values.size}x${map.values.firstOrNull()?.size ?: 0} cells",
                style = MaterialTheme.typography.bodySmall,
            )

            if (map.values.isEmpty()) {
                Text("No values in map.")
            } else {
                val flatValues = map.values.flatten()
                val minValue = flatValues.minOrNull() ?: 0.0
                val maxValue = flatValues.maxOrNull() ?: 0.0

                map.values.forEach { row ->
                    Row(horizontalArrangement = Arrangement.spacedBy(2.dp)) {
                        row.forEach { value ->
                            Box(
                                modifier = Modifier
                                    .width(14.dp)
                                    .height(14.dp)
                                    .background(mapValueToColor(value, minValue, maxValue)),
                            )
                        }
                    }
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text("Min: %.2f".format(minValue), style = MaterialTheme.typography.bodySmall)
                    Text("Max: %.2f".format(maxValue), style = MaterialTheme.typography.bodySmall)
                }
            }
        }
    }
}

private fun mapValueToColor(value: Double, min: Double, max: Double): Color {
    if (max <= min) {
        return Color(0xFF6D8EA0)
    }

    val normalized = ((value - min) / (max - min)).toFloat().coerceIn(0f, 1f)
    val red = (40 + (200 * normalized)).toInt()
    val blue = (200 - (160 * normalized)).toInt()
    return Color(red = red, green = 90, blue = blue)
}

@Composable
private fun DtcScreen(viewModel: DashboardViewModel) {
    val telemetry by viewModel.telemetry.collectAsState()
    val dtcs = telemetry.dtcs

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        contentPadding = PaddingValues(bottom = 24.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp),
    ) {
        item {
            Text("Diagnostic Trouble Codes", style = MaterialTheme.typography.headlineSmall)
        }

        if (dtcs.isEmpty()) {
            item {
                Text("No DTCs found.")
            }
        } else {
            items(dtcs) { code ->
                Card(modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        Text(code, style = MaterialTheme.typography.titleMedium)
                        Text(
                            dtcDescriptions[code] ?: "Description not found in local dictionary.",
                            style = MaterialTheme.typography.bodyMedium,
                        )
                    }
                }
            }
        }
    }
}

@Composable
@OptIn(ExperimentalLayoutApi::class, ExperimentalMaterial3Api::class)
private fun SettingsScreen(
    viewModel: DashboardViewModel,
    currentHost: String,
    currentPort: Int,
    currentMode: TelemetryMode,
    currentBleDeviceAddress: String?,
    bondedBleDevices: List<BleDeviceOption>,
    discoveredBleDevices: List<BleDeviceOption>,
    isBleScanInProgress: Boolean,
    bleScanStartedAtMillis: Long?,
    bleDiagnostics: BleDiagnostics,
    settingsNotice: String?,
    onRefreshBleDevices: () -> Unit,
    onStartBleScan: () -> Unit,
    onStopBleScan: () -> Unit,
    onApplySettings: (String, Int, TelemetryMode, String?) -> Unit,
) {
    val profiles by viewModel.profiles.collectAsState()
    val profileSyncStatus by viewModel.profileSyncStatus.collectAsState()
    val profileSyncDiagnostics by viewModel.profileSyncDiagnostics.collectAsState()

    var hostText by rememberSaveable { mutableStateOf(currentHost) }
    var portText by rememberSaveable { mutableStateOf(currentPort.toString()) }
    var mode by rememberSaveable { mutableStateOf(currentMode) }
    var bleDeviceAddressText by rememberSaveable { mutableStateOf(currentBleDeviceAddress ?: "") }
    var modeMenuExpanded by remember { mutableStateOf(false) }
    var bleDeviceMenuExpanded by remember { mutableStateOf(false) }
    var profileMenuExpanded by remember { mutableStateOf(false) }
    var scanElapsedSeconds by remember { mutableLongStateOf(0L) }

    var selectedProfileId by rememberSaveable { mutableStateOf<String?>(null) }
    var profileDisplayName by rememberSaveable { mutableStateOf("") }
    var profileNotes by rememberSaveable { mutableStateOf("") }
    var profileTags by rememberSaveable { mutableStateOf("") }
    var profileMake by rememberSaveable { mutableStateOf("") }
    var profileModel by rememberSaveable { mutableStateOf("") }
    var profileYear by rememberSaveable { mutableStateOf("") }

    val normalizedBleAddress = normalizeBleAddress(bleDeviceAddressText)
    val isBleAddressValid = normalizedBleAddress.isBlank() || isValidBleAddress(normalizedBleAddress)

    LaunchedEffect(currentHost, currentPort, currentMode, currentBleDeviceAddress) {
        hostText = currentHost
        portText = currentPort.toString()
        mode = currentMode
        bleDeviceAddressText = currentBleDeviceAddress ?: ""
    }

    LaunchedEffect(profiles, selectedProfileId) {
        if (profiles.isEmpty()) {
            selectedProfileId = null
            profileDisplayName = ""
            profileNotes = ""
            profileTags = ""
            profileMake = ""
            profileModel = ""
            profileYear = ""
            return@LaunchedEffect
        }

        val selected = profiles.firstOrNull { it.id == selectedProfileId } ?: profiles.first()
        selectedProfileId = selected.id
        profileDisplayName = selected.displayName.orEmpty()
        profileNotes = selected.notes.orEmpty()
        profileTags = selected.tags.joinToString(",")
        profileMake = selected.make.orEmpty()
        profileModel = selected.model.orEmpty()
        profileYear = selected.year?.toString().orEmpty()
    }

    LaunchedEffect(isBleScanInProgress, bleScanStartedAtMillis) {
        if (!isBleScanInProgress || bleScanStartedAtMillis == null) {
            scanElapsedSeconds = 0L
            return@LaunchedEffect
        }

        while (isBleScanInProgress) {
            val elapsedMillis = System.currentTimeMillis() - bleScanStartedAtMillis
            scanElapsedSeconds = (elapsedMillis / 1000L).coerceAtLeast(0L)
            delay(1000L)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("Connection Settings", style = MaterialTheme.typography.headlineSmall)

        OutlinedTextField(
            value = hostText,
            onValueChange = { hostText = it },
            label = { Text("Host") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )

        OutlinedTextField(
            value = portText,
            onValueChange = { portText = it.filter { c -> c.isDigit() } },
            label = { Text("Port") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
        )

        ExposedDropdownMenuBox(
            expanded = modeMenuExpanded,
            onExpandedChange = { modeMenuExpanded = !modeMenuExpanded },
        ) {
            OutlinedTextField(
                value = mode.name,
                onValueChange = {},
                readOnly = true,
                label = { Text("Telemetry Mode") },
                trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = modeMenuExpanded) },
                modifier = Modifier
                    .menuAnchor(ExposedDropdownMenuAnchorType.PrimaryNotEditable)
                    .fillMaxWidth(),
            )

            ExposedDropdownMenu(
                expanded = modeMenuExpanded,
                onDismissRequest = { modeMenuExpanded = false },
            ) {
                TelemetryMode.entries.forEach { option ->
                    DropdownMenuItem(
                        text = { Text(option.name) },
                        onClick = {
                            mode = option
                            modeMenuExpanded = false
                        },
                    )
                }
            }
        }

        if (mode == TelemetryMode.BLE_DIRECT) {
            val allBleDevices = (bondedBleDevices + discoveredBleDevices)
                .distinctBy { it.address }
                .sortedWith(
                    compareByDescending<BleDeviceOption> { it.rssi ?: Int.MIN_VALUE }
                        .thenByDescending { it.lastSeenMillis ?: Long.MIN_VALUE }
                        .thenBy { it.name.lowercase() },
                )

            Card(modifier = Modifier.fillMaxWidth()) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(12.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    Text("BLE Diagnostics", style = MaterialTheme.typography.titleMedium)
                    Text("Permissions: ${if (bleDiagnostics.hasRequiredPermissions) "granted" else "missing"}", style = MaterialTheme.typography.bodySmall)
                    Text("Adapter: ${if (bleDiagnostics.adapterAvailable) "available" else "unavailable"}", style = MaterialTheme.typography.bodySmall)
                    Text("Bluetooth: ${if (bleDiagnostics.adapterEnabled) "enabled" else "disabled"}", style = MaterialTheme.typography.bodySmall)
                    Text(
                        "Selected: ${bleDiagnostics.selectedAddress ?: "none"}",
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
            }

            FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Button(onClick = onRefreshBleDevices) {
                    Text("Refresh Paired Devices")
                }

                Button(onClick = onStartBleScan) {
                    Text(if (isBleScanInProgress) "Scanning..." else "Scan Nearby")
                }

                if (isBleScanInProgress) {
                    Button(onClick = onStopBleScan) {
                        Text("Stop Scan")
                    }
                }
            }

            if (isBleScanInProgress) {
                Text(
                    text = "Scan in progress: ${scanElapsedSeconds}s",
                    style = MaterialTheme.typography.bodySmall,
                )
            }

            ExposedDropdownMenuBox(
                expanded = bleDeviceMenuExpanded,
                onExpandedChange = { bleDeviceMenuExpanded = !bleDeviceMenuExpanded },
            ) {
                OutlinedTextField(
                    value = bleDeviceAddressText,
                    onValueChange = { bleDeviceAddressText = it },
                    label = { Text("BLE Device Address") },
                    placeholder = { Text("AA:BB:CC:DD:EE:FF") },
                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = bleDeviceMenuExpanded) },
                    modifier = Modifier
                        .menuAnchor(ExposedDropdownMenuAnchorType.PrimaryEditable)
                        .fillMaxWidth(),
                    singleLine = true,
                )

                ExposedDropdownMenu(
                    expanded = bleDeviceMenuExpanded,
                    onDismissRequest = { bleDeviceMenuExpanded = false },
                ) {
                    allBleDevices.forEach { option ->
                        val signalBadge = option.rssi?.let { "[${rssiQualityLabel(it)}] " } ?: ""
                        val signalSuffix = option.rssi?.let { " | ${it} dBm" } ?: ""
                        val freshnessSuffix = option.lastSeenMillis?.let { " | seen ${relativeSeenLabel(it)}" } ?: ""
                        DropdownMenuItem(
                            text = { Text("$signalBadge${option.name} (${option.address})$signalSuffix$freshnessSuffix") },
                            onClick = {
                                bleDeviceAddressText = option.address
                                bleDeviceMenuExpanded = false
                            },
                        )
                    }
                }
            }

            Text(
                text = "Paired devices found: ${bondedBleDevices.size}",
                style = MaterialTheme.typography.bodySmall,
            )
            Text(
                text = "Discovered devices found: ${discoveredBleDevices.size}",
                style = MaterialTheme.typography.bodySmall,
            )

            if (!isBleAddressValid) {
                Text(
                    text = "Invalid BLE address format. Use AA:BB:CC:DD:EE:FF",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.error,
                )
            }
        }

        if (!settingsNotice.isNullOrBlank()) {
            Text(
                text = settingsNotice,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.error,
            )
        }

        Card(modifier = Modifier.fillMaxWidth()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                Text("Profile Sync Diagnostics", style = MaterialTheme.typography.titleMedium)
                Text("Total profiles: ${profileSyncDiagnostics.totalProfiles}", style = MaterialTheme.typography.bodySmall)
                Text(
                    "Pending sync: ${profileSyncDiagnostics.pendingCount}",
                    style = MaterialTheme.typography.bodySmall,
                    color = if (profileSyncDiagnostics.pendingCount > 0) {
                        MaterialTheme.colorScheme.primary
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    },
                )
                Text(
                    "Failed sync: ${profileSyncDiagnostics.failedCount}",
                    style = MaterialTheme.typography.bodySmall,
                    color = if (profileSyncDiagnostics.failedCount > 0) {
                        MaterialTheme.colorScheme.error
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    },
                )
                Text(
                    "Synced: ${profileSyncDiagnostics.syncedCount}",
                    style = MaterialTheme.typography.bodySmall,
                    color = if (profileSyncDiagnostics.syncedCount > 0) {
                        MaterialTheme.colorScheme.primary
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    },
                )
                Text(
                    text = "Last successful sync: ${formatSyncTime(profileSyncDiagnostics.lastSuccessfulSyncAtMillis)}",
                    style = MaterialTheme.typography.bodySmall,
                )

                FlowRow(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    Button(onClick = { viewModel.syncProfilesNow() }) {
                        Text("Sync Profiles Now")
                    }

                    Button(
                        enabled = profileSyncDiagnostics.failedCount > 0,
                        onClick = { viewModel.retryFailedProfilesNow() },
                    ) {
                        Text("Retry Failed Only")
                    }
                }

                if (!profileSyncStatus.isNullOrBlank()) {
                    Text(
                        text = profileSyncStatus.orEmpty(),
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
            }
        }

        Card(modifier = Modifier.fillMaxWidth()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text("Vehicle Profiles", style = MaterialTheme.typography.titleMedium)

                if (profiles.isEmpty()) {
                    Text(
                        "No profiles yet. Connect to a vehicle with VIN data to auto-create one.",
                        style = MaterialTheme.typography.bodySmall,
                    )
                } else {
                    val selectedProfile = profiles.firstOrNull { it.id == selectedProfileId } ?: profiles.first()

                    ExposedDropdownMenuBox(
                        expanded = profileMenuExpanded,
                        onExpandedChange = { profileMenuExpanded = !profileMenuExpanded },
                    ) {
                        OutlinedTextField(
                            value = "${selectedProfile.displayName ?: selectedProfile.vin} (${selectedProfile.syncState.name})",
                            onValueChange = {},
                            readOnly = true,
                            label = { Text("Active Profile") },
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = profileMenuExpanded) },
                            modifier = Modifier
                                .menuAnchor(ExposedDropdownMenuAnchorType.PrimaryNotEditable)
                                .fillMaxWidth(),
                            singleLine = true,
                        )

                        ExposedDropdownMenu(
                            expanded = profileMenuExpanded,
                            onDismissRequest = { profileMenuExpanded = false },
                        ) {
                            profiles.forEach { profile ->
                                DropdownMenuItem(
                                    text = {
                                        Text("${profile.displayName ?: profile.vin} (${profile.syncState.name})")
                                    },
                                    onClick = {
                                        selectedProfileId = profile.id
                                        profileMenuExpanded = false
                                    },
                                )
                            }
                        }
                    }

                    OutlinedTextField(
                        value = profileDisplayName,
                        onValueChange = { profileDisplayName = it },
                        label = { Text("Display Name") },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                    )

                    OutlinedTextField(
                        value = profileNotes,
                        onValueChange = { profileNotes = it },
                        label = { Text("Notes") },
                        modifier = Modifier.fillMaxWidth(),
                    )

                    OutlinedTextField(
                        value = profileTags,
                        onValueChange = { profileTags = it },
                        label = { Text("Tags (comma-separated)") },
                        modifier = Modifier.fillMaxWidth(),
                    )

                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        OutlinedTextField(
                            value = profileMake,
                            onValueChange = { profileMake = it },
                            label = { Text("Make") },
                            modifier = Modifier.weight(1f),
                            singleLine = true,
                        )

                        OutlinedTextField(
                            value = profileModel,
                            onValueChange = { profileModel = it },
                            label = { Text("Model") },
                            modifier = Modifier.weight(1f),
                            singleLine = true,
                        )
                    }

                    OutlinedTextField(
                        value = profileYear,
                        onValueChange = { profileYear = it.filter { c -> c.isDigit() } },
                        label = { Text("Year") },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                    )

                    Button(
                        onClick = {
                            val tags = profileTags.split(',')
                                .map { it.trim() }
                                .filter { it.isNotBlank() }
                            val year = profileYear.toIntOrNull()

                            viewModel.updateVehicleProfile(
                                profileId = selectedProfile.id,
                                displayName = profileDisplayName.ifBlank { null },
                                notes = profileNotes.ifBlank { null },
                                tags = tags,
                                make = profileMake.ifBlank { null },
                                model = profileModel.ifBlank { null },
                                year = year,
                            )
                        },
                    ) {
                        Text("Save Profile")
                    }

                }
            }
        }

        FlowRow(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Button(
                enabled = mode != TelemetryMode.BLE_DIRECT || isBleAddressValid,
                onClick = {
                    val parsedPort = portText.toIntOrNull() ?: currentPort
                    onApplySettings(
                        hostText.ifBlank { currentHost },
                        parsedPort,
                        mode,
                        normalizedBleAddress.ifBlank { null },
                    )
                },
            ) {
                Text("Apply and Reconnect")
            }
        }
    }
}

private fun normalizeBleAddress(value: String): String {
    return value
        .trim()
        .replace('-', ':')
        .uppercase()
}

private fun isValidBleAddress(value: String): Boolean {
    val pattern = Regex("^([0-9A-F]{2}:){5}[0-9A-F]{2}$")
    return pattern.matches(value)
}

private fun rssiQualityLabel(rssi: Int): String {
    return when {
        rssi >= -60 -> "Excellent"
        rssi >= -70 -> "Good"
        rssi >= -80 -> "Fair"
        else -> "Weak"
    }
}

private fun relativeSeenLabel(lastSeenMillis: Long): String {
    val deltaSeconds = ((System.currentTimeMillis() - lastSeenMillis) / 1000L).coerceAtLeast(0L)
    return "${deltaSeconds}s ago"
}

private fun formatSyncTime(timestampMillis: Long?): String {
    if (timestampMillis == null) {
        return "never"
    }

    val deltaSeconds = ((System.currentTimeMillis() - timestampMillis) / 1000L).coerceAtLeast(0L)
    return "${deltaSeconds}s ago"
}

@Composable
private fun InfoRow(label: String, value: String) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Text(label, style = MaterialTheme.typography.labelMedium)
            Text(value.ifBlank { "-" }, style = MaterialTheme.typography.bodyLarge)
        }
    }
}
