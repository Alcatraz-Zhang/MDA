package membership

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type quotaState struct {
	BusinessDate string `json:"business_date"`
	DeviceHash   string `json:"device_hash"`
	TierCode     string `json:"tier_code"`
	UsedSeconds  int64  `json:"used_seconds"`
	UpdatedAt    string `json:"updated_at"`
}

type QuotaSnapshot struct {
	TierName         string
	TierCode         string
	LimitSeconds     int64
	UsedSeconds      int64
	RemainingSeconds int64
	BusinessDate     string
	SponsorURL       string
	UnlimitedRuntime bool
}

var quotaMu sync.Mutex

func quotaBusinessDate(now time.Time) string {
	return now.Add(-4 * time.Hour).Format("2006-01-02")
}

func quotaStatePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	path := filepath.Join(dir, "MDA", "go-service")
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	return filepath.Join(path, "membership-quota.json"), nil
}

func deviceHash(device DeviceCodeV7) string {
	sum := sha256.Sum256([]byte(device.CPUHash + device.UUIDHash + device.BIOSHash + device.BoardHash + device.DiskHash + device.GUIDHash))
	return hex.EncodeToString(sum[:])
}

func loadQuotaState(path string) quotaState {
	data, err := os.ReadFile(path)
	if err != nil {
		return quotaState{}
	}
	var state quotaState
	if err := json.Unmarshal(data, &state); err != nil {
		return quotaState{}
	}
	return state
}

func saveQuotaState(path string, state quotaState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func normalizeQuotaState(status *MembershipStatus, now time.Time) (string, quotaState, error) {
	path, err := quotaStatePath()
	if err != nil {
		return "", quotaState{}, err
	}
	state := loadQuotaState(path)
	businessDate := quotaBusinessDate(now)
	device := deviceHash(status.DeviceCode)
	if state.BusinessDate != businessDate || state.DeviceHash != device {
		state = quotaState{
			BusinessDate: businessDate,
			DeviceHash:   device,
			UsedSeconds:  0,
		}
	}
	state.TierCode = status.TierCode
	if state.TierCode == "" {
		state.TierCode = "orange_free"
	}
	state.UpdatedAt = now.Format(time.RFC3339)
	return path, state, nil
}

func quotaSnapshotLocked(status *MembershipStatus, now time.Time) (QuotaSnapshot, error) {
	path, state, err := normalizeQuotaState(status, now)
	if err != nil {
		return QuotaSnapshot{}, err
	}
	if err := saveQuotaState(path, state); err != nil {
		return QuotaSnapshot{}, err
	}
	return snapshotFromState(status, state), nil
}

func snapshotFromState(status *MembershipStatus, state quotaState) QuotaSnapshot {
	if status.UnlimitedRuntime {
		return QuotaSnapshot{
			TierName:         status.TierName,
			TierCode:         status.TierCode,
			LimitSeconds:     0,
			UsedSeconds:      0,
			RemainingSeconds: 0,
			BusinessDate:     state.BusinessDate,
			UnlimitedRuntime: true,
		}
	}

	limit := int64(status.DailyRuntimeMinutes) * 60
	if limit <= 0 {
		limit = 10 * 60
	}
	used := state.UsedSeconds
	if used < 0 {
		used = 0
	}
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	return QuotaSnapshot{
		TierName:         status.TierName,
		TierCode:         status.TierCode,
		LimitSeconds:     limit,
		UsedSeconds:      used,
		RemainingSeconds: remaining,
		BusinessDate:     state.BusinessDate,
		SponsorURL:       SponsorURL(status),
		UnlimitedRuntime: false,
	}
}

func GetQuotaSnapshot(status *MembershipStatus) (QuotaSnapshot, error) {
	if status.UnlimitedRuntime {
		return snapshotFromState(status, quotaState{BusinessDate: quotaBusinessDate(time.Now())}), nil
	}
	quotaMu.Lock()
	defer quotaMu.Unlock()
	return quotaSnapshotLocked(status, time.Now())
}

func AddQuotaUsage(status *MembershipStatus, delta time.Duration) (QuotaSnapshot, error) {
	if status.UnlimitedRuntime {
		return snapshotFromState(status, quotaState{BusinessDate: quotaBusinessDate(time.Now())}), nil
	}
	if delta <= 0 {
		return GetQuotaSnapshot(status)
	}
	quotaMu.Lock()
	defer quotaMu.Unlock()
	now := time.Now()
	path, state, err := normalizeQuotaState(status, now)
	if err != nil {
		return QuotaSnapshot{}, err
	}
	seconds := int64(delta.Round(time.Second) / time.Second)
	if seconds <= 0 {
		seconds = 1
	}
	state.UsedSeconds += seconds
	state.UpdatedAt = now.Format(time.RFC3339)
	if err := saveQuotaState(path, state); err != nil {
		return QuotaSnapshot{}, err
	}
	return snapshotFromState(status, state), nil
}

func EnsureQuotaAvailable(status *MembershipStatus) (QuotaSnapshot, bool, error) {
	snapshot, err := GetQuotaSnapshot(status)
	if err != nil {
		fallback := snapshotFromState(status, quotaState{BusinessDate: quotaBusinessDate(time.Now())})
		return fallback, true, err
	}
	if snapshot.UnlimitedRuntime {
		return snapshot, true, nil
	}
	return snapshot, snapshot.RemainingSeconds > 0, nil
}

func FormatMinutes(seconds int64) int64 {
	if seconds <= 0 {
		return 0
	}
	return (seconds + 59) / 60
}
