package com.anodyne.iloadobd2.ui

import androidx.compose.foundation.background
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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.ScrollableTabRow
import androidx.compose.material3.Tab
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.collectAsState
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
    onApplySettings: (String, Int, TelemetryMode) -> Unit,
) {
    var selectedTab by rememberSaveable { mutableStateOf(AppTab.DASHBOARD.ordinal) }

    Column(modifier = Modifier.fillMaxSize()) {
        ScrollableTabRow(selectedTabIndex = selectedTab) {
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
                currentHost = currentHost,
                currentPort = currentPort,
                currentMode = currentMode,
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
@OptIn(ExperimentalMaterial3Api::class)
private fun SettingsScreen(
    currentHost: String,
    currentPort: Int,
    currentMode: TelemetryMode,
    onApplySettings: (String, Int, TelemetryMode) -> Unit,
) {
    var hostText by rememberSaveable { mutableStateOf(currentHost) }
    var portText by rememberSaveable { mutableStateOf(currentPort.toString()) }
    var mode by rememberSaveable { mutableStateOf(currentMode) }
    var modeMenuExpanded by remember { mutableStateOf(false) }

    LaunchedEffect(currentHost, currentPort, currentMode) {
        hostText = currentHost
        portText = currentPort.toString()
        mode = currentMode
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
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
                    .menuAnchor()
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

        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Button(
                onClick = {
                    val parsedPort = portText.toIntOrNull() ?: currentPort
                    onApplySettings(hostText.ifBlank { currentHost }, parsedPort, mode)
                },
            ) {
                Text("Apply and Reconnect")
            }
        }
    }
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
