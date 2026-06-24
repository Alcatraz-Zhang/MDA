package membership

import (
	"strings"
	"testing"

	"github.com/1204244136/MDA/agent/go-service/pkg/i18n"
)

func initTestI18n() {
	i18n.Init()
}

func TestFormatQuotaVerifiedMessageUsesUsedRuntime(t *testing.T) {
	initTestI18n()
	message := formatQuotaVerifiedMessage(QuotaSnapshot{
		Route:        quotaRouteRegular,
		TierName:     "Orange Plus",
		LimitSeconds: 3600,
		UsedSeconds:  600,
	})

	if !strings.Contains(message, "10/60") {
		t.Fatalf("message does not show used runtime: %s", message)
	}
}

func TestFormatQuotaDeniedMessageUsesDebtText(t *testing.T) {
	initTestI18n()
	message := formatQuotaDeniedMessage(QuotaSnapshot{
		TierName:           "Orange Free",
		LimitSeconds:       600,
		CarriedDebtSeconds: 1500,
		SponsorURL:         "https://example.test",
	})

	if !strings.Contains(message, "此前超额运行") && !strings.Contains(message, "previous over-quota runtime") {
		t.Fatalf("message does not mention carried debt: %s", message)
	}
	if !strings.Contains(message, "15") {
		t.Fatalf("message does not show remaining carried debt: %s", message)
	}
}

func TestFormatQuotaDeniedMessageUsesSpecialDebtText(t *testing.T) {
	initTestI18n()
	message := formatQuotaDeniedMessage(QuotaSnapshot{
		Route:               quotaRouteSpecialThenRegular,
		TierName:            "Orange Plus",
		SpecialLimitSeconds: 36000,
		RegularLimitSeconds: 1800,
		LimitSeconds:        1800,
		CarriedDebtSeconds:  2700,
		SponsorURL:          "https://example.test",
	})

	if !strings.Contains(message, "专项额度") && !strings.Contains(message, "special quota") {
		t.Fatalf("message does not mention special quota: %s", message)
	}
	if !strings.Contains(message, "剩余待抵扣") && !strings.Contains(message, "remaining previous over-quota") {
		t.Fatalf("message does not show remaining carried debt: %s", message)
	}
}

func TestFormatCarriedDebtMessageShowsDeductedAndRemainingQuota(t *testing.T) {
	initTestI18n()
	message := formatCarriedDebtMessage(QuotaSnapshot{
		LimitSeconds:       1800,
		RemainingSeconds:   900,
		CarriedDebtSeconds: 900,
	})

	if !strings.Contains(message, "已抵扣") && !strings.Contains(message, "deducted") {
		t.Fatalf("message does not mention deducted carried debt: %s", message)
	}
	if strings.Count(message, "15") < 3 {
		t.Fatalf("message does not show carried, deducted, and remaining minutes: %s", message)
	}
}

func TestFormatQuotaDeniedMessageUsesNormalText(t *testing.T) {
	initTestI18n()
	message := formatQuotaDeniedMessage(QuotaSnapshot{
		Route:        quotaRouteRegular,
		TierName:     "Orange Free",
		LimitSeconds: 600,
		SponsorURL:   "https://example.test",
	})

	if strings.Contains(message, "此前超额运行") || strings.Contains(message, "previous over-quota runtime") {
		t.Fatalf("message unexpectedly mentions carried debt: %s", message)
	}
}

func TestFormatQuotaVerifiedMessageUsesSpecialRoute(t *testing.T) {
	initTestI18n()
	message := formatQuotaVerifiedMessage(QuotaSnapshot{
		Route:               quotaRouteSpecialThenRegular,
		TierName:            "Orange Plus",
		SpecialLimitSeconds: 36000,
		SpecialUsedSeconds:  600,
	})

	if !strings.Contains(message, "专项额度") && !strings.Contains(message, "special quota") {
		t.Fatalf("message does not mention special quota: %s", message)
	}
}

func TestQuotaRouteForRuntimeQuotaCheckEntry(t *testing.T) {
	if got := quotaRouteForEntry("MapPushingFlow"); got != quotaRouteSpecialThenRegular {
		t.Fatalf("quotaRouteForEntry(MapPushingFlow) = %s, want %s", got, quotaRouteSpecialThenRegular)
	}
}

func TestFormatMembershipVerificationUnavailableMessage(t *testing.T) {
	initTestI18n()
	message := formatMembershipVerificationUnavailableMessage()

	if !strings.Contains(message, "会员校验服务暂不可用") && !strings.Contains(message, "temporarily unavailable") {
		t.Fatalf("message does not mention membership verification service unavailable: %s", message)
	}
}
