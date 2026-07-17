package membership

import (
	"testing"
	"time"
)

func TestMultiplierForEntry(t *testing.T) {
	cases := map[string]int64{
		"SmallEventMain":   1000,
		"LargeEventMain":   1000,
		"MapPushingFlow":   5000,
		"DailyRewardsMain": 1000,
	}
	for entry, want := range cases {
		if got := multiplierForEntry(entry).BasePermille; got != want {
			t.Fatalf("multiplierForEntry(%s).BasePermille = %d, want %d", entry, got, want)
		}
	}
}

func TestBillableDuration(t *testing.T) {
	multiplier := quotaMultiplier{BasePermille: 3000, ExtraPermille: 1500}
	if got := multiplier.billableDuration(time.Minute); got != 270*time.Second {
		t.Fatalf("billableDuration() = %s, want 4m30s", got)
	}
}

func TestConsumeBillableSecondsKeepsFractionUntilFlush(t *testing.T) {
	tracker := &RuntimeTracker{
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: 1500,
		},
	}

	if got := tracker.consumeBillableSeconds(500*time.Millisecond, false); got != 0 {
		t.Fatalf("first consumeBillableSeconds() = %d, want 0", got)
	}
	if got := tracker.consumeBillableSeconds(500*time.Millisecond, false); got != 1 {
		t.Fatalf("second consumeBillableSeconds() = %d, want 1", got)
	}
	if got := tracker.consumeBillableSeconds(0, true); got != 1 {
		t.Fatalf("flush consumeBillableSeconds() = %d, want 1", got)
	}
}

func TestConsumeBillableSecondsCeilsOnFlush(t *testing.T) {
	tracker := &RuntimeTracker{
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: 1500,
		},
	}

	if got := tracker.consumeBillableSeconds(500*time.Millisecond, true); got != 1 {
		t.Fatalf("flush consumeBillableSeconds() = %d, want 1", got)
	}
}

func TestConsumeTickIgnoresStaleGeneration(t *testing.T) {
	isolateQuotaState(t)
	status := testStatus(10, "device-a")
	tracker := &RuntimeTracker{
		active:     true,
		generation: 2,
		last:       time.Now().Add(-time.Minute),
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: multiplierScale,
		},
	}

	if _, done := tracker.consumeTick(nil, status, quotaRouteRegular, 1); !done {
		t.Fatalf("consumeTick() with stale generation should stop")
	}

	snapshot, err := GetQuotaSnapshot(status, quotaPoolRegularDaily)
	if err != nil {
		t.Fatalf("GetQuotaSnapshot() failed: %v", err)
	}
	if snapshot.UsedSeconds != 0 {
		t.Fatalf("UsedSeconds = %d, want 0", snapshot.UsedSeconds)
	}
}

func TestFinishFlushesOnlyUnchargedTail(t *testing.T) {
	isolateQuotaState(t)
	status := testStatus(10, "device-a")
	tracker := &RuntimeTracker{
		active: true,
		route:  quotaRouteRegular,
		status: status,
		last:   time.Now().Add(-500 * time.Millisecond),
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: multiplierScale,
		},
		realNs:         int64(time.Minute),
		chargedSeconds: 60,
		stopCh:         make(chan struct{}),
	}

	tracker.finish()

	snapshot, err := GetQuotaSnapshot(status, quotaPoolRegularDaily)
	if err != nil {
		t.Fatalf("GetQuotaSnapshot() failed: %v", err)
	}
	if snapshot.UsedSeconds != 1 {
		t.Fatalf("UsedSeconds = %d, want 1", snapshot.UsedSeconds)
	}
}
