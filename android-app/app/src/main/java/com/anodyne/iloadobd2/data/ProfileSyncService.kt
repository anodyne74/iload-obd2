package com.anodyne.iloadobd2.data

import com.anodyne.iloadobd2.model.VehicleProfile

data class ProfileUpsertResult(
    val cloudVersion: String? = null,
)

interface ProfileSyncService {
    suspend fun upsertProfile(profile: VehicleProfile): ProfileUpsertResult
}
