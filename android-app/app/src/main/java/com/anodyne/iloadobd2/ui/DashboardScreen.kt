package com.anodyne.iloadobd2.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.anodyne.iloadobd2.viewmodel.DashboardViewModel
import java.util.Locale

@Composable
fun DashboardScreen(viewModel: DashboardViewModel) {
    val telemetry by viewModel.telemetry.collectAsState()
    val captureResponse by viewModel.captureResponse.collectAsState()
    val connectionState by viewModel.connectionState.collectAsState()

    var captureLabel by remember { mutableStateOf("") }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text(
            text = "iLoad OBD2 Android",
            style = MaterialTheme.typography.headlineSmall,
        )
        Text(
            text = "Connection: $connectionState",
            style = MaterialTheme.typography.bodyMedium,
        )

        Column(
            modifier = Modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            MetricCard(label = "RPM", value = telemetry.rpm.toInt().toString())
            MetricCard(label = "Speed", value = String.format(Locale.US, "%.1f km/h", telemetry.speed))
            MetricCard(label = "Temp", value = String.format(Locale.US, "%.1f C", telemetry.temp))
            MetricCard(label = "DTC Count", value = telemetry.dtcs.size.toString())
            MetricCard(label = "Capture", value = telemetry.capture?.status ?: "idle")
            MetricCard(label = "Frames", value = (telemetry.capture?.frameCount ?: 0).toString())
        }

        Card(modifier = Modifier.fillMaxWidth()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Text("Capture Control", style = MaterialTheme.typography.titleMedium)
                OutlinedTextField(
                    value = captureLabel,
                    onValueChange = { captureLabel = it },
                    label = { Text("Trip label") },
                    modifier = Modifier.fillMaxWidth(),
                )

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(onClick = { viewModel.startCapture(captureLabel.ifBlank { null }) }) {
                        Text("Start")
                    }
                    Button(onClick = { viewModel.stopCapture() }) {
                        Text("Stop")
                    }
                    Button(onClick = { viewModel.requestCaptureStatus() }) {
                        Text("Status")
                    }
                }

                if (captureResponse != null) {
                    Text(
                        text = "Last response: ${captureResponse?.status ?: ""} ${captureResponse?.error ?: captureResponse?.message ?: ""}",
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
            }
        }
    }
}

@Composable
private fun MetricCard(label: String, value: String) {
    Card(
        modifier = Modifier
            .widthIn(min = 150.dp)
            .padding(0.dp),
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(label, style = MaterialTheme.typography.labelMedium)
            Text(value, style = MaterialTheme.typography.titleLarge)
        }
    }
}
