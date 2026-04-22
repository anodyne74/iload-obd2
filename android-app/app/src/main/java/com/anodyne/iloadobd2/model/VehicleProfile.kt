package com.anodyne.iloadobd2.model

import kotlinx.serialization.Serializable

@Serializable
enum class ProfileSyncState {
    PENDING_CREATE,
    PENDING_UPDATE,
    SYNCED,
    FAILED,
}

@Serializable
data class VehicleProfile(
    val id: String,
    val version: Int,
    val vin: String,
    val displayName: String? = null,
    val make: String? = null,
    val model: String? = null,
    val year: Int? = null,
    val notes: String? = null,
    val tags: List<String> = emptyList(),
    val creationSource: String,
    val createdAtMillis: Long,
    val updatedAtMillis: Long,
    val lastDetectedAtMillis: Long,
    val syncState: ProfileSyncState,
    val cloudVersion: String? = null,
    val syncUpdatedAtMillis: Long? = null,
)

@Serializable
data class VehicleProfileCatalog(
    val schemaVersion: Int = 1,
    val profiles: List<VehicleProfile> = emptyList(),
)

data class VehicleProfileUpdate(
    val displayName: String? = null,
    val notes: String? = null,
    val tags: List<String>? = null,
    val make: String? = null,
    val model: String? = null,
    val year: Int? = null,
)
