package com.anodyne.iloadobd2.viewmodel

import com.anodyne.iloadobd2.data.CaptureControlResponse
import com.anodyne.iloadobd2.data.ProfileSyncService
import com.anodyne.iloadobd2.data.TelemetryRepository
import com.anodyne.iloadobd2.data.VehicleProfileRepository
import com.anodyne.iloadobd2.model.ProfileSyncState
import com.anodyne.iloadobd2.model.TelemetryData
import com.anodyne.iloadobd2.model.VehicleProfile
import com.anodyne.iloadobd2.model.VehicleProfileUpdate
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

data class ProfileSyncDiagnostics(
    val totalProfiles: Int = 0,
    val pendingCount: Int = 0,
    val failedCount: Int = 0,
    val syncedCount: Int = 0,
    val lastSuccessfulSyncAtMillis: Long? = null,
)

class DashboardViewModel(
    private val repository: TelemetryRepository,
    private val vehicleProfiles: VehicleProfileRepository? = null,
    private val profileSync: ProfileSyncService? = null,
) {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)

    private val _telemetry = MutableStateFlow(TelemetryData())
    val telemetry: StateFlow<TelemetryData> = _telemetry.asStateFlow()

    private val _captureResponse = MutableStateFlow<CaptureControlResponse?>(null)
    val captureResponse: StateFlow<CaptureControlResponse?> = _captureResponse.asStateFlow()

    private val _connectionState = MutableStateFlow("connecting")
    val connectionState: StateFlow<String> = _connectionState.asStateFlow()

    private val _profiles = MutableStateFlow<List<VehicleProfile>>(emptyList())
    val profiles: StateFlow<List<VehicleProfile>> = _profiles.asStateFlow()

    private val _profileSyncStatus = MutableStateFlow<String?>(null)
    val profileSyncStatus: StateFlow<String?> = _profileSyncStatus.asStateFlow()

    private val _profileSyncDiagnostics = MutableStateFlow(ProfileSyncDiagnostics())
    val profileSyncDiagnostics: StateFlow<ProfileSyncDiagnostics> = _profileSyncDiagnostics.asStateFlow()

    init {
        scope.launch {
            runCatching {
                repository.connect()
            }.onSuccess {
                _connectionState.value = "connected"
            }.onFailure {
                _connectionState.value = "error"
                _captureResponse.value = CaptureControlResponse(
                    type = "capture_control",
                    status = "error",
                    error = it.message ?: "connection failed",
                )
            }
        }

        scope.launch {
            refreshProfilesSnapshot()
        }

        scope.launch {
            repository.telemetry.collect { data ->
                _telemetry.value = data
                val ecuInfo = data.ecuInfo
                if (ecuInfo != null) {
                    val ensured = vehicleProfiles?.ensureProfileForDetectedVehicle(ecuInfo)
                    if (ensured != null) {
                        refreshProfilesSnapshot()
                    }
                    if (ensured != null && ensured.syncState != ProfileSyncState.SYNCED) {
                        syncProfilesOnce()
                    }
                }
            }
        }

        scope.launch {
            repository.captureEvents.collect { event ->
                _captureResponse.value = event
            }
        }

        scope.launch {
            while (true) {
                syncProfilesOnce()
                kotlinx.coroutines.delay(PROFILE_SYNC_INTERVAL_MS)
            }
        }
    }

    fun startCapture(label: String?) {
        scope.launch {
            val metadata = buildCaptureMetadata()
            repository.startCapture(label = label, metadata = metadata)
        }
    }

    fun stopCapture() {
        scope.launch { repository.stopCapture() }
    }

    fun requestCaptureStatus() {
        scope.launch { repository.requestCaptureStatus() }
    }

    fun updateVehicleProfile(
        profileId: String,
        displayName: String? = null,
        notes: String? = null,
        tags: List<String>? = null,
        make: String? = null,
        model: String? = null,
        year: Int? = null,
    ) {
        val updates = VehicleProfileUpdate(
            displayName = displayName,
            notes = notes,
            tags = tags,
            make = make,
            model = model,
            year = year,
        )
        scope.launch {
            vehicleProfiles?.updateProfile(profileId = profileId, update = updates)
            refreshProfilesSnapshot()
            syncProfilesOnce()
        }
    }

    fun syncProfilesNow() {
        scope.launch {
            syncProfilesOnce()
        }
    }

    fun retryFailedProfilesNow() {
        scope.launch {
            val profiles = vehicleProfiles ?: return@launch
            val sync = profileSync ?: return@launch

            val failedCount = profiles.listFailedSyncProfiles().size
            if (failedCount == 0) {
                _profileSyncStatus.value = "No failed profiles to retry."
                refreshProfilesSnapshot()
                return@launch
            }

            _profileSyncStatus.value = "Retrying $failedCount failed profile(s)..."
            profiles.syncFailedProfiles(syncService = sync)
            refreshProfilesSnapshot()

            val remainingFailed = profiles.listFailedSyncProfiles().size
            _profileSyncStatus.value = if (remainingFailed == 0) {
                "Failed profiles successfully retried."
            } else {
                "Retry finished: $remainingFailed failed profile(s) remain."
            }
        }
    }

    private suspend fun syncProfilesOnce() {
        val profiles = vehicleProfiles ?: return
        val sync = profileSync ?: return
        val pendingCount = profiles.listPendingSyncProfiles().size

        if (pendingCount == 0) {
            _profileSyncStatus.value = "Profiles already synchronized."
            refreshProfilesSnapshot()
            return
        }

        _profileSyncStatus.value = "Syncing $pendingCount profile(s)..."
        profiles.syncPendingProfiles(syncService = sync)
        refreshProfilesSnapshot()

        val remaining = profiles.listPendingSyncProfiles().size
        _profileSyncStatus.value = if (remaining == 0) {
            "Profile sync completed."
        } else {
            "Profile sync pending: $remaining remaining."
        }
    }

    private suspend fun refreshProfilesSnapshot() {
        val profiles = vehicleProfiles ?: return
        val snapshot = profiles.listProfiles()
        _profiles.value = snapshot

        val pendingCount = snapshot.count {
            it.syncState == ProfileSyncState.PENDING_CREATE || it.syncState == ProfileSyncState.PENDING_UPDATE
        }
        val failedCount = snapshot.count { it.syncState == ProfileSyncState.FAILED }
        val synced = snapshot.filter { it.syncState == ProfileSyncState.SYNCED }
        val lastSuccessfulSyncAtMillis = synced.maxOfOrNull { it.syncUpdatedAtMillis ?: 0L }
            ?.takeIf { it > 0L }

        _profileSyncDiagnostics.value = ProfileSyncDiagnostics(
            totalProfiles = snapshot.size,
            pendingCount = pendingCount,
            failedCount = failedCount,
            syncedCount = synced.size,
            lastSuccessfulSyncAtMillis = lastSuccessfulSyncAtMillis,
        )
    }

    private suspend fun buildCaptureMetadata(): Map<String, String>? {
        val telemetry = _telemetry.value
        val ecuInfo = telemetry.ecuInfo ?: return null
        val vin = ecuInfo.vin?.trim()?.uppercase().orEmpty()
        if (vin.isBlank()) {
            return null
        }

        val profile = vehicleProfiles?.getProfileByVin(vin)
        val metadata = linkedMapOf<String, String>()
        metadata["vin"] = vin

        if (profile != null) {
            metadata["profile_id"] = profile.id
            metadata["profile_version"] = profile.version.toString()
            metadata["profile_sync_state"] = profile.syncState.name.lowercase()
            profile.make?.let { metadata["make"] = it }
            profile.model?.let { metadata["model"] = it }
            profile.year?.let { metadata["year"] = it.toString() }
        }

        return if (metadata.isEmpty()) null else metadata
    }

    fun clear() {
        scope.launch {
            repository.disconnect()
            _connectionState.value = "disconnected"
            scope.cancel()
        }
    }

    companion object {
        private const val PROFILE_SYNC_INTERVAL_MS = 30_000L
    }
}
