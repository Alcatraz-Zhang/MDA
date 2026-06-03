package membership

import (
	"fmt"
	"sync"

	"github.com/1204244136/MDA/agent/go-service/pkg/i18n"
	"github.com/1204244136/MDA/agent/go-service/pkg/maafocus"
	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type RuntimeQuotaCheckAction struct{}

var _ maa.CustomActionRunner = &RuntimeQuotaCheckAction{}

var notifyOnce sync.Once

func (a *RuntimeQuotaCheckAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	return runRuntimeQuotaCheck(ctx)
}

func runRuntimeQuotaCheck(ctx *maa.Context) bool {
	if isDebugEnvironment() {
		return true
	}

	status := GetMembershipStatus()
	if status.UpdateRequired {
		if status.UpdateMessage != "" {
			maafocus.Print(ctx, status.UpdateMessage)
		} else {
			maafocus.Print(ctx, fmt.Sprintf(
				i18n.T("tasker.membership_check.update_required"),
				status.MinimumSupportedVersion,
			))
		}
		return false
	}

	snapshot, ok, err := EnsureQuotaAvailable(status)
	if err != nil {
		log.Warn().Err(err).Msg("RuntimeQuotaCheck: failed to read local quota state")
	}

	log.Info().
		Str("tier_code", snapshot.TierCode).
		Str("tier_name", snapshot.TierName).
		Int64("limit_seconds", snapshot.LimitSeconds).
		Int64("used_seconds", snapshot.UsedSeconds).
		Int64("remaining_seconds", snapshot.RemainingSeconds).
		Int64("carried_debt_seconds", snapshot.CarriedDebtSeconds).
		Bool("unlimited_runtime", snapshot.UnlimitedRuntime).
		Str("business_date", snapshot.BusinessDate).
		Msg("RuntimeQuotaCheck: quota evaluated")

	if ok {
		notifyOnce.Do(func() {
			if snapshot.UnlimitedRuntime {
				maafocus.Print(ctx, i18n.T("tasker.membership_check.debug_unlimited"))
				return
			}
			maafocus.Print(ctx, formatQuotaVerifiedMessage(snapshot))
			if snapshot.CarriedDebtSeconds > 0 {
				maafocus.Print(ctx, fmt.Sprintf(
					i18n.T("tasker.membership_check.debt"),
					FormatMinutes(snapshot.CarriedDebtSeconds),
				))
			}
			maafocus.Print(ctx, fmt.Sprintf(
				i18n.T("tasker.membership_check.sponsor"),
				snapshot.SponsorURL,
			))
		})
		return true
	}

	maafocus.Print(ctx, formatQuotaDeniedMessage(snapshot))
	return false
}

func formatQuotaVerifiedMessage(snapshot QuotaSnapshot) string {
	return fmt.Sprintf(
		i18n.T("tasker.membership_check.verified"),
		snapshot.TierName,
		FormatMinutes(snapshot.UsedSeconds),
		FormatMinutes(snapshot.LimitSeconds),
	)
}

func formatQuotaDeniedMessage(snapshot QuotaSnapshot) string {
	messageKey := "tasker.membership_check.denied"
	if snapshot.CarriedDebtSeconds > 0 {
		messageKey = "tasker.membership_check.denied_debt"
	}
	return fmt.Sprintf(
		i18n.T(messageKey),
		snapshot.TierName,
		FormatMinutes(snapshot.LimitSeconds),
		snapshot.SponsorURL,
	)
}
