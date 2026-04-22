package com.anodyne.iloadobd2.data

import android.content.Context
import com.anodyne.iloadobd2.model.EcuInfo
import com.anodyne.iloadobd2.model.ProfileSyncState
import com.anodyne.iloadobd2.model.VehicleProfile
import com.anodyne.iloadobd2.model.VehicleProfileCatalog
import com.anodyne.iloadobd2.model.VehicleProfileUpdate
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import java.util.UUID

class VehicleProfileRepository(
    context: Context,
    private val json: Json = Json {
        ignoreUnknownKeys = true
        prettyPrint = true
        encodeDefaults = true
    },
) {
    private val mutex = Mutex()
    private val catalogFile = context.applicationContext.filesDir.resolve("vehicle_profiles.json")

    suspend fun ensureProfileForDetectedVehicle(
        ecuInfo: EcuInfo,
        source: String = "telemetry_detected",
    ): VehicleProfile? = mutex.withLock {
        val normalizedVin = ecuInfo.vin?.trim()?.uppercase().orEmpty()
        if (normalizedVin.isEmpty()) {
            return null
        }

        val now = System.currentTimeMillis()
        val catalog = readCatalog()
        val existing = catalog.profiles.firstOrNull { it.vin == normalizedVin }
        if (existing != null) {
            val nextProfile = existing.copy(
                make = existing.make ?: ecuInfo.hardware,
                model = existing.model ?: ecuInfo.software,
                displayName = existing.displayName ?: normalizedVin,
                updatedAtMillis = now,
                lastDetectedAtMillis = now,
            )

            writeCatalog(
                catalog.copy(
                    profiles = catalog.profiles.map {
                        if (it.id == nextProfile.id) nextProfile else it
                    },
                ),
            )

            return nextProfile
        }

        val profile = VehicleProfile(
            id = UUID.randomUUID().toString(),
            version = 1,
            vin = normalizedVin,
            displayName = normalizedVin,
            make = ecuInfo.hardware,
            model = ecuInfo.software,
            creationSource = source,
            createdAtMillis = now,
            updatedAtMillis = now,
            lastDetectedAtMillis = now,
            syncState = ProfileSyncState.PENDING_CREATE,
        )

        writeCatalog(catalog.copy(profiles = catalog.profiles + profile))
        profile
    }

    suspend fun updateProfile(
        profileId: String,
        update: VehicleProfileUpdate,
    ): VehicleProfile? = mutex.withLock {
        val catalog = readCatalog()
        val current = catalog.profiles.firstOrNull { it.id == profileId } ?: return null
        val now = System.currentTimeMillis()

        val nextSyncState = when (current.syncState) {
            ProfileSyncState.PENDING_CREATE -> ProfileSyncState.PENDING_CREATE
            else -> ProfileSyncState.PENDING_UPDATE
        }

        val updated = current.copy(
            version = current.version + 1,
            displayName = update.displayName ?: current.displayName,
            notes = update.notes ?: current.notes,
            tags = update.tags ?: current.tags,
            make = update.make ?: current.make,
            model = update.model ?: current.model,
            year = update.year ?: current.year,
            updatedAtMillis = now,
            syncState = nextSyncState,
        )

        writeCatalog(
            catalog.copy(
                profiles = catalog.profiles.map {
                    if (it.id == profileId) updated else it
                },
            ),
        )

        updated
    }

    suspend fun listProfiles(): List<VehicleProfile> = mutex.withLock {
        readCatalog().profiles.sortedByDescending { it.updatedAtMillis }
    }

    suspend fun getProfileByVin(vin: String): VehicleProfile? = mutex.withLock {
        val normalizedVin = vin.trim().uppercase()
        if (normalizedVin.isBlank()) {
            return null
        }

        readCatalog().profiles.firstOrNull { it.vin == normalizedVin }
    }

    suspend fun listPendingSyncProfiles(limit: Int = 100): List<VehicleProfile> = mutex.withLock {
        readCatalog().profiles
            .filter {
                it.syncState == ProfileSyncState.PENDING_CREATE ||
                    it.syncState == ProfileSyncState.PENDING_UPDATE ||
                    it.syncState == ProfileSyncState.FAILED
            }
            .sortedBy { it.updatedAtMillis }
            .take(limit)
    }

    suspend fun listFailedSyncProfiles(limit: Int = 100): List<VehicleProfile> = mutex.withLock {
        readCatalog().profiles
            .filter { it.syncState == ProfileSyncState.FAILED }
            .sortedBy { it.updatedAtMillis }
            .take(limit)
    }

    suspend fun syncPendingProfiles(
        syncService: ProfileSyncService,
        limit: Int = 100,
    ) {
        val pending = listPendingSyncProfiles(limit = limit)
        pending.forEach { profile ->
            runCatching {
                syncService.upsertProfile(profile)
            }.onSuccess { result ->
                markProfileSynced(profileId = profile.id, cloudVersion = result.cloudVersion)
            }.onFailure {
                markProfileSyncFailed(profileId = profile.id)
            }
        }
    }

    suspend fun syncFailedProfiles(
        syncService: ProfileSyncService,
        limit: Int = 100,
    ) {
        val failed = listFailedSyncProfiles(limit = limit)
        failed.forEach { profile ->
            runCatching {
                syncService.upsertProfile(profile)
            }.onSuccess { result ->
                markProfileSynced(profileId = profile.id, cloudVersion = result.cloudVersion)
            }.onFailure {
                markProfileSyncFailed(profileId = profile.id)
            }
        }
    }

    suspend fun markProfileSynced(profileId: String, cloudVersion: String?): VehicleProfile? = mutex.withLock {
        val catalog = readCatalog()
        val current = catalog.profiles.firstOrNull { it.id == profileId } ?: return null
        val now = System.currentTimeMillis()

        val updated = current.copy(
            syncState = ProfileSyncState.SYNCED,
            cloudVersion = cloudVersion,
            syncUpdatedAtMillis = now,
            updatedAtMillis = now,
        )

        writeCatalog(
            catalog.copy(
                profiles = catalog.profiles.map {
                    if (it.id == profileId) updated else it
                },
            ),
        )

        updated
    }

    suspend fun markProfileSyncFailed(profileId: String): VehicleProfile? = mutex.withLock {
        val catalog = readCatalog()
        val current = catalog.profiles.firstOrNull { it.id == profileId } ?: return null
        val now = System.currentTimeMillis()

        val updated = current.copy(
            syncState = ProfileSyncState.FAILED,
            updatedAtMillis = now,
        )

        writeCatalog(
            catalog.copy(
                profiles = catalog.profiles.map {
                    if (it.id == profileId) updated else it
                },
            ),
        )

        updated
    }

    private suspend fun readCatalog(): VehicleProfileCatalog = withContext(Dispatchers.IO) {
        if (!catalogFile.exists()) {
            return@withContext VehicleProfileCatalog()
        }

        val raw = catalogFile.readText()
        if (raw.isBlank()) {
            return@withContext VehicleProfileCatalog()
        }

        runCatching {
            json.decodeFromString(VehicleProfileCatalog.serializer(), raw)
        }.getOrDefault(VehicleProfileCatalog())
    }

    private suspend fun writeCatalog(catalog: VehicleProfileCatalog) = withContext(Dispatchers.IO) {
        if (!catalogFile.parentFile.exists()) {
            catalogFile.parentFile?.mkdirs()
        }
        catalogFile.writeText(json.encodeToString(VehicleProfileCatalog.serializer(), catalog))
    }
}
