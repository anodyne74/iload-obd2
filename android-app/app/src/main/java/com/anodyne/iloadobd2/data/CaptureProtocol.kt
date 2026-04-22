package com.anodyne.iloadobd2.data

import kotlinx.serialization.Serializable

@Serializable
data class CaptureCommand(
    val command: String,
    val label: String? = null,
    val metadata: Map<String, String>? = null,
)

@Serializable
data class CaptureControlResponse(
    val type: String,
    val status: String? = null,
    val message: String? = null,
    val error: String? = null,
    val filePath: String? = null,
    val frameCount: Int? = null,
    val durationSeconds: Long? = null,
)
