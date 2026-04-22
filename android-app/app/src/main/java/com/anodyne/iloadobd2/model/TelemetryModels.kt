package com.anodyne.iloadobd2.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class TelemetryData(
    val rpm: Double = 0.0,
    val speed: Double = 0.0,
    val temp: Double = 0.0,
    val dtcs: List<String> = emptyList(),
    val ecuInfo: EcuInfo? = null,
    val engineMaps: EngineMaps? = null,
    val canFrames: List<CanFrame> = emptyList(),
    val capture: CaptureState? = null,
)

@Serializable
data class EcuInfo(
    val version: String? = null,
    val hardware: String? = null,
    val software: String? = null,
    val calibration: String? = null,
    val vin: String? = null,
    val buildDate: String? = null,
    val protocol: String? = null,
)

@Serializable
data class EngineMaps(
    val fuel: MapData? = null,
    val timing: MapData? = null,
)

@Serializable
data class MapData(
    val values: List<List<Double>> = emptyList(),
    val xAxis: List<Double> = emptyList(),
    val yAxis: List<Double> = emptyList(),
)

@Serializable
data class CanFrame(
    val id: Long,
    @SerialName("data") val payload: String = "",
    val timestamp: String,
)

@Serializable
data class CaptureState(
    val status: String,
    val frameCount: Int = 0,
    val durationSeconds: Long = 0,
)
