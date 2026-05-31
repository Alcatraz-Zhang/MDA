package membership

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testStatus(minutes int, device string) *MembershipStatus {
	return &MembershipStatus{
		TierCode:            "orange_free",
		TierName:            "Orange Free",
		DailyRuntimeMinutes: minutes,
		DeviceCode: DeviceCodeV7{
			CPUHash: device,
		},
	}
}

func isolateQuotaState(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	path, err := quotaStatePath()
	if err != nil {
		t.Fatalf("quotaStatePath() failed: %v", err)
	}
	return path
}

func mustSaveQuotaState(t *testing.T, path string, state quotaState) {
	t.Helper()
	if err := saveQuotaState(path, state); err != nil {
		t.Fatalf("saveQuotaState() failed: %v", err)
	}
}

func TestNormalizeQuotaStateCarriesOneDayDebt(t *testing.T) {
	path := isolateQuotaState(t)
	status := testStatus(10, "device-a")
	device := deviceHash(status.DeviceCode)
	mustSaveQuotaState(t, path, quotaState{
		BusinessDate: "2026-05-28",
		DeviceHash:   device,
		TierCode:     "orange_free",
		LimitSeconds: 600,
		UsedSeconds:  725,
	})

	_, state, err := normalizeQuotaState(status, time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("normalizeQuotaState() failed: %v", err)
	}
	if state.UsedSeconds != 125 {
		t.Fatalf("UsedSeconds = %d, want 125", state.UsedSeconds)
	}
	if state.CarriedDebtSeconds != 125 {
		t.Fatalf("CarriedDebtSeconds = %d, want 125", state.CarriedDebtSeconds)
	}
	snapshot := snapshotFromState(status, state)
	if snapshot.RemainingSeconds != 475 {
		t.Fatalf("RemainingSeconds = %d, want 475", snapshot.RemainingSeconds)
	}
	if snapshot.CarriedDebtSeconds != 125 {
		t.Fatalf("snapshot.CarriedDebtSeconds = %d, want 125", snapshot.CarriedDebtSeconds)
	}
}

func TestNormalizeQuotaStateClearsWhenNoDebt(t *testing.T) {
	path := isolateQuotaState(t)
	status := testStatus(10, "device-a")
	device := deviceHash(status.DeviceCode)
	mustSaveQuotaState(t, path, quotaState{
		BusinessDate: "2026-05-28",
		DeviceHash:   device,
		TierCode:     "orange_free",
		LimitSeconds: 600,
		UsedSeconds:  500,
	})

	_, state, err := normalizeQuotaState(status, time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("normalizeQuotaState() failed: %v", err)
	}
	if state.UsedSeconds != 0 {
		t.Fatalf("UsedSeconds = %d, want 0", state.UsedSeconds)
	}
	if state.CarriedDebtSeconds != 0 {
		t.Fatalf("CarriedDebtSeconds = %d, want 0", state.CarriedDebtSeconds)
	}
}

func TestCarriedQuotaDebtDecaysAcrossMultipleDays(t *testing.T) {
	state := quotaState{
		BusinessDate: "2026-05-28",
		LimitSeconds: 600,
		UsedSeconds:  1900,
	}

	cases := map[string]int64{
		"2026-05-29": 1300,
		"2026-05-30": 700,
		"2026-06-01": 0,
	}
	for businessDate, want := range cases {
		if got := carriedQuotaDebt(state, businessDate, 600); got != want {
			t.Fatalf("carriedQuotaDebt(%s) = %d, want %d", businessDate, got, want)
		}
	}
}

func TestSnapshotPreservesSameDayOverage(t *testing.T) {
	status := testStatus(10, "device-a")
	snapshot := snapshotFromState(status, quotaState{
		BusinessDate: "2026-05-28",
		LimitSeconds: 600,
		UsedSeconds:  725,
	})

	if snapshot.UsedSeconds != 725 {
		t.Fatalf("UsedSeconds = %d, want 725", snapshot.UsedSeconds)
	}
	if snapshot.RemainingSeconds != 0 {
		t.Fatalf("RemainingSeconds = %d, want 0", snapshot.RemainingSeconds)
	}
}

func TestNormalizeQuotaStateResetsOnDeviceChange(t *testing.T) {
	path := isolateQuotaState(t)
	oldStatus := testStatus(10, "device-a")
	newStatus := testStatus(10, "device-b")
	mustSaveQuotaState(t, path, quotaState{
		BusinessDate: "2026-05-28",
		DeviceHash:   deviceHash(oldStatus.DeviceCode),
		TierCode:     "orange_free",
		LimitSeconds: 600,
		UsedSeconds:  725,
	})

	_, state, err := normalizeQuotaState(newStatus, time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("normalizeQuotaState() failed: %v", err)
	}
	if state.UsedSeconds != 0 {
		t.Fatalf("UsedSeconds = %d, want 0", state.UsedSeconds)
	}
	if state.DeviceHash != deviceHash(newStatus.DeviceCode) {
		t.Fatalf("DeviceHash was not updated")
	}
}

func TestLimitedMemberCarriesDebt(t *testing.T) {
	path := isolateQuotaState(t)
	status := testStatus(60, "device-a")
	status.IsMember = true
	status.TierCode = "orange_plus"
	status.TierName = "Orange Plus"
	device := deviceHash(status.DeviceCode)
	mustSaveQuotaState(t, path, quotaState{
		BusinessDate: "2026-05-28",
		DeviceHash:   device,
		TierCode:     "orange_plus",
		LimitSeconds: 3600,
		UsedSeconds:  3900,
	})

	_, state, err := normalizeQuotaState(status, time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("normalizeQuotaState() failed: %v", err)
	}
	if state.UsedSeconds != 300 {
		t.Fatalf("UsedSeconds = %d, want 300", state.UsedSeconds)
	}
	if state.CarriedDebtSeconds != 300 {
		t.Fatalf("CarriedDebtSeconds = %d, want 300", state.CarriedDebtSeconds)
	}
}

func TestUpgradeToLimitedMemberKeepsDebtWithNewLimit(t *testing.T) {
	path := isolateQuotaState(t)
	freeStatus := testStatus(10, "device-a")
	memberStatus := testStatus(60, "device-a")
	memberStatus.IsMember = true
	memberStatus.TierCode = "orange_plus"
	memberStatus.TierName = "Orange Plus"
	device := deviceHash(freeStatus.DeviceCode)
	mustSaveQuotaState(t, path, quotaState{
		BusinessDate: "2026-05-29",
		DeviceHash:   device,
		TierCode:     "orange_free",
		LimitSeconds: 600,
		UsedSeconds:  1800,
	})

	_, state, err := normalizeQuotaState(memberStatus, time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("normalizeQuotaState() failed: %v", err)
	}
	snapshot := snapshotFromState(memberStatus, state)
	if snapshot.UsedSeconds != 1800 {
		t.Fatalf("UsedSeconds = %d, want 1800", snapshot.UsedSeconds)
	}
	if snapshot.RemainingSeconds != 1800 {
		t.Fatalf("RemainingSeconds = %d, want 1800", snapshot.RemainingSeconds)
	}
}

func TestUnlimitedRuntimeClearsDebt(t *testing.T) {
	path := isolateQuotaState(t)
	status := testStatus(10, "device-a")
	device := deviceHash(status.DeviceCode)
	mustSaveQuotaState(t, path, quotaState{
		BusinessDate:       "2026-05-29",
		DeviceHash:         device,
		TierCode:           "orange_free",
		LimitSeconds:       600,
		UsedSeconds:        1800,
		CarriedDebtSeconds: 1200,
	})
	status.UnlimitedRuntime = true
	status.IsMember = true

	_, state, err := normalizeQuotaState(status, time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("normalizeQuotaState() failed: %v", err)
	}
	if state.UsedSeconds != 0 {
		t.Fatalf("UsedSeconds = %d, want 0", state.UsedSeconds)
	}
	if state.CarriedDebtSeconds != 0 {
		t.Fatalf("CarriedDebtSeconds = %d, want 0", state.CarriedDebtSeconds)
	}
}

func TestOldQuotaStateFallsBackToCurrentLimit(t *testing.T) {
	path := isolateQuotaState(t)
	status := testStatus(10, "device-a")
	device := deviceHash(status.DeviceCode)
	oldJSON := []byte(`{
  "business_date": "2026-05-28",
  "device_hash": "` + device + `",
  "tier_code": "orange_free",
  "used_seconds": 725,
  "updated_at": "2026-05-28T12:00:00+08:00"
}`)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(path, oldJSON, 0644); err != nil {
		t.Fatalf("WriteFile() failed: %v", err)
	}

	_, state, err := normalizeQuotaState(status, time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("normalizeQuotaState() failed: %v", err)
	}
	if state.UsedSeconds != 125 {
		t.Fatalf("UsedSeconds = %d, want 125", state.UsedSeconds)
	}
}

func TestAddQuotaUsageUsesBillableDuration(t *testing.T) {
	isolateQuotaState(t)
	status := testStatus(10, "device-a")

	snapshot, err := AddQuotaUsage(status, 2*time.Minute)
	if err != nil {
		t.Fatalf("AddQuotaUsage() failed: %v", err)
	}
	if snapshot.UsedSeconds != 120 {
		t.Fatalf("UsedSeconds = %d, want 120", snapshot.UsedSeconds)
	}
	if snapshot.RemainingSeconds != 480 {
		t.Fatalf("RemainingSeconds = %d, want 480", snapshot.RemainingSeconds)
	}
}

func TestNextQuotaTickInterval(t *testing.T) {
	cases := map[int64]time.Duration{
		0:   quotaTickMinInterval,
		3:   quotaTickMinInterval,
		30:  30 * time.Second,
		120: quotaTickMaxInterval,
	}
	for remainingSeconds, want := range cases {
		if got := nextQuotaTickInterval(remainingSeconds); got != want {
			t.Fatalf("nextQuotaTickInterval(%d) = %s, want %s", remainingSeconds, got, want)
		}
	}
}
