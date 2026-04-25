package com.anodyne.iloadobd2.data

import com.anodyne.iloadobd2.model.VehicleProfile
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody

class HttpProfileSyncService(
    private val baseUrl: String,
    private val client: OkHttpClient,
    private val json: Json = Json { ignoreUnknownKeys = true },
) : ProfileSyncService {
    @Serializable
    private data class UpsertRequest(
        val profile: VehicleProfile,
    )

    @Serializable
    private data class UpsertResponse(
        val cloudVersion: String? = null,
    )

    override suspend fun upsertProfile(profile: VehicleProfile): ProfileUpsertResult = withContext(Dispatchers.IO) {
        val endpoint = "${baseUrl.trimEnd('/')}/api/v1/vehicle-profiles/upsert"
        val payload = json.encodeToString(
            UpsertRequest.serializer(),
            UpsertRequest(profile = profile),
        )

        val request = Request.Builder()
            .url(endpoint)
            .post(payload.toRequestBody("application/json".toMediaType()))
            .build()

        client.newCall(request).execute().use { response ->
            if (!response.isSuccessful) {
                throw IllegalStateException("profile upsert failed: HTTP ${response.code}")
            }

            val body = response.body.string()
            if (body.isBlank()) {
                return@withContext ProfileUpsertResult(cloudVersion = null)
            }

            val parsed = runCatching {
                json.decodeFromString(UpsertResponse.serializer(), body)
            }.getOrNull()

            ProfileUpsertResult(cloudVersion = parsed?.cloudVersion)
        }
    }
}
