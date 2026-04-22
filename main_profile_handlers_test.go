package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetProfileStoreForTest() {
	profileStoreMu.Lock()
	defer profileStoreMu.Unlock()
	profileStore = make(map[string]storedProfile)
}

func TestProfileUpsertHandlerRejectsMissingProfileID(t *testing.T) {
	resetProfileStoreForTest()

	body := []byte(`{"profile":{"vin":"KMH123"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicle-profiles/upsert", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	profileUpsertHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestProfileUpsertAndListHandlers(t *testing.T) {
	resetProfileStoreForTest()

	upsertBody := []byte(`{"profile":{"id":"profile-1","vin":"KMH123","displayName":"Van"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicle-profiles/upsert", bytes.NewReader(upsertBody))
	rr := httptest.NewRecorder()

	profileUpsertHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 on first upsert, got %d", rr.Code)
	}

	var upsertResp ProfileUpsertResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &upsertResp); err != nil {
		t.Fatalf("failed to parse upsert response: %v", err)
	}
	if upsertResp.CloudVersion != "1" {
		t.Fatalf("expected cloudVersion 1, got %q", upsertResp.CloudVersion)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/vehicle-profiles/upsert", bytes.NewReader(upsertBody))
	rr2 := httptest.NewRecorder()
	profileUpsertHandler(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected status 200 on second upsert, got %d", rr2.Code)
	}

	var upsertResp2 ProfileUpsertResponse
	if err := json.Unmarshal(rr2.Body.Bytes(), &upsertResp2); err != nil {
		t.Fatalf("failed to parse second upsert response: %v", err)
	}
	if upsertResp2.CloudVersion != "2" {
		t.Fatalf("expected cloudVersion 2, got %q", upsertResp2.CloudVersion)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/vehicle-profiles", nil)
	listRR := httptest.NewRecorder()
	profileListHandler(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 on list, got %d", listRR.Code)
	}

	var profiles []StoredProfileView
	if err := json.Unmarshal(listRR.Body.Bytes(), &profiles); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("expected 1 stored profile, got %d", len(profiles))
	}

	if profiles[0].ID != "profile-1" {
		t.Fatalf("expected profile id profile-1, got %q", profiles[0].ID)
	}

	if profiles[0].CloudVersion != "2" {
		t.Fatalf("expected listed cloudVersion 2, got %q", profiles[0].CloudVersion)
	}
}
