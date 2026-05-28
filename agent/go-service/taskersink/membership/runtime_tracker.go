package membership

import (
	"fmt"
	"sync"
	"time"

	"github.com/1204244136/MDA/agent/go-service/pkg/i18n"
	"github.com/1204244136/MDA/agent/go-service/pkg/maafocus"
	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type RuntimeTracker struct {
	mu      sync.Mutex
	active  bool
	taskID  uint64
	entry   string
	last    time.Time
	stopCh  chan struct{}
	stopped bool
}

var _ maa.TaskerEventSink = &RuntimeTracker{}

const quotaTickInterval = 15 * time.Second

func (t *RuntimeTracker) OnTaskerTask(tasker *maa.Tasker, event maa.EventStatus, detail maa.TaskerTaskDetail) {
	if detail.Entry == "MaaTaskerPostStop" {
		return
	}

	switch event {
	case maa.EventStatusStarting:
		t.start(tasker, detail)
	case maa.EventStatusSucceeded, maa.EventStatusFailed:
		t.finish()
	}
}

func (t *RuntimeTracker) start(tasker *maa.Tasker, detail maa.TaskerTaskDetail) {
	t.finish()

	status := GetMembershipStatus()
	snapshot, ok, err := EnsureQuotaAvailable(status)
	if err != nil {
		log.Warn().Err(err).Msg("RuntimeTracker: failed to check quota at task start")
	}
	if !ok {
		printQuotaExhausted(snapshot)
		tasker.PostStop()
		return
	}

	t.mu.Lock()
	t.active = true
	t.taskID = detail.TaskID
	t.entry = detail.Entry
	t.last = time.Now()
	t.stopCh = make(chan struct{})
	t.stopped = false
	stopCh := t.stopCh
	t.mu.Unlock()

	log.Info().
		Uint64("task_id", detail.TaskID).
		Str("entry", detail.Entry).
		Int64("remaining_seconds", snapshot.RemainingSeconds).
		Bool("unlimited_runtime", snapshot.UnlimitedRuntime).
		Msg("RuntimeTracker: started quota tracking")

	if snapshot.UnlimitedRuntime {
		return
	}

	go t.tick(tasker, status, stopCh)
}

func (t *RuntimeTracker) finish() {
	t.mu.Lock()
	if !t.active {
		t.mu.Unlock()
		return
	}
	last := t.last
	stopCh := t.stopCh
	t.active = false
	t.stopCh = nil
	close(stopCh)
	t.mu.Unlock()

	status := GetMembershipStatus()
	if _, err := AddQuotaUsage(status, time.Since(last)); err != nil {
		log.Warn().Err(err).Msg("RuntimeTracker: failed to flush final quota usage")
	}
}

func (t *RuntimeTracker) tick(tasker *maa.Tasker, status *MembershipStatus, stopCh <-chan struct{}) {
	ticker := time.NewTicker(quotaTickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if t.consumeTick(tasker, status) {
				return
			}
		case <-stopCh:
			return
		}
	}
}

func (t *RuntimeTracker) consumeTick(tasker *maa.Tasker, status *MembershipStatus) bool {
	now := time.Now()
	t.mu.Lock()
	if !t.active {
		t.mu.Unlock()
		return true
	}
	delta := now.Sub(t.last)
	t.last = now
	taskID := t.taskID
	entry := t.entry
	alreadyStopped := t.stopped
	t.mu.Unlock()

	snapshot, err := AddQuotaUsage(status, delta)
	if err != nil {
		log.Warn().Err(err).Msg("RuntimeTracker: failed to record quota usage")
		return false
	}

	log.Debug().
		Uint64("task_id", taskID).
		Str("entry", entry).
		Int64("used_seconds", snapshot.UsedSeconds).
		Int64("remaining_seconds", snapshot.RemainingSeconds).
		Msg("RuntimeTracker: quota usage recorded")

	if snapshot.RemainingSeconds > 0 || alreadyStopped {
		return false
	}

	t.mu.Lock()
	t.stopped = true
	t.mu.Unlock()
	printQuotaExhausted(snapshot)
	tasker.PostStop()
	return false
}

func printQuotaExhausted(snapshot QuotaSnapshot) {
	maafocus.PrintLargeContentTrimNewline(fmt.Sprintf(
		i18n.T("tasker.membership_check.denied"),
		snapshot.TierName,
		FormatMinutes(snapshot.LimitSeconds),
		snapshot.SponsorURL,
	))
}
