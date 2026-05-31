package membership

import (
	"testing"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

func TestMultiplierForEntry(t *testing.T) {
	cases := map[string]int64{
		"SmallEventMain":   2000,
		"LargeEventMain":   3000,
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

func TestOnNodePipelineNodeAppliesDailyLoginMultiplier(t *testing.T) {
	tracker := &RuntimeTracker{
		active: true,
		taskID: 7,
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: multiplierScale,
			Reason:        "default",
		},
	}

	tracker.OnNodePipelineNode(nil, maa.EventStatusStarting, maa.NodePipelineNodeDetail{TaskID: 7, Name: "DailyRewardsDailyLogin"})

	if got := tracker.multiplier.ExtraPermille; got != 1500 {
		t.Fatalf("ExtraPermille = %d, want 1500", got)
	}
}

func TestOnNodePipelineNodeAppliesLateDailyLoginMultiplier(t *testing.T) {
	tracker := &RuntimeTracker{
		active: true,
		taskID: 7,
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: multiplierScale,
			Reason:        "default",
		},
	}
	if got := tracker.consumeBillableSeconds(20*time.Second, false); got != 20 {
		t.Fatalf("initial consumeBillableSeconds() = %d, want 20", got)
	}

	tracker.OnNodePipelineNode(nil, maa.EventStatusStarting, maa.NodePipelineNodeDetail{TaskID: 7, Name: "DailyRewardsDailyLogin"})

	if got := tracker.multiplier.ExtraPermille; got != 1500 {
		t.Fatalf("ExtraPermille = %d, want 1500", got)
	}
	if got := tracker.consumeBillableSeconds(0, true); got != 10 {
		t.Fatalf("backfill consumeBillableSeconds() = %d, want 10", got)
	}
}

func TestApplyExtraMultiplier(t *testing.T) {
	tracker := &RuntimeTracker{
		active: true,
		taskID: 7,
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: multiplierScale,
			Reason:        "default",
		},
	}

	tracker.applyExtraMultiplier(7, 1500, "daily_login_enabled")

	if got := tracker.multiplier.ExtraPermille; got != 1500 {
		t.Fatalf("ExtraPermille = %d, want 1500", got)
	}
	if got := tracker.multiplier.totalPermille(); got != 1500 {
		t.Fatalf("totalPermille() = %d, want 1500", got)
	}
	if tracker.multiplier.Reason != "daily_login_enabled" {
		t.Fatalf("Reason = %s, want daily_login_enabled", tracker.multiplier.Reason)
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

func TestApplyExtraMultiplierIgnoresOtherTask(t *testing.T) {
	tracker := &RuntimeTracker{
		active: true,
		taskID: 7,
		multiplier: quotaMultiplier{
			BasePermille:  multiplierScale,
			ExtraPermille: multiplierScale,
		},
	}

	tracker.applyExtraMultiplier(8, 1500, "daily_login_enabled")

	if got := tracker.multiplier.ExtraPermille; got != multiplierScale {
		t.Fatalf("ExtraPermille = %d, want %d", got, multiplierScale)
	}
}
