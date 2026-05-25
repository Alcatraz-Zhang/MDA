package membership

import (
	"fmt"
	"sync"

	"github.com/1204244136/MDA/agent/go-service/pkg/i18n"
	"github.com/1204244136/MDA/agent/go-service/pkg/maafocus"
	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

// MembershipCheckAction is a custom action that checks membership before executing member-only tasks.
// It runs synchronously in the pipeline, blocking execution for non-members.
type MembershipCheckAction struct{}

var _ maa.CustomActionRunner = &MembershipCheckAction{}

var notifyOnce sync.Once

func (a *MembershipCheckAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	status := GetMembershipStatus()

	// 构建赞助链接（无论是否会员都显示）
	sponsorURL := fmt.Sprintf(
		"https://doropay.top?cpu=%s&uuid=%s&bios=%s&board=%s&disk=%s&guid=%s",
		status.DeviceCode.CPUHash,
		status.DeviceCode.UUIDHash,
		status.DeviceCode.BIOSHash,
		status.DeviceCode.BoardHash,
		status.DeviceCode.DiskHash,
		status.DeviceCode.GUIDHash,
	)

	if status.IsMember {
		log.Info().
			Str("tier", status.MembershipType).
			Int("level", status.UserLevel).
			Str("expiry", status.VirtualExpiry).
			Msg("MembershipCheck: member verified, allowing")

		// 会员提示只在首次启动时显示
		notifyOnce.Do(func() {
			maafocus.Print(ctx, fmt.Sprintf(
				i18n.T("tasker.membership_check.verified"),
				status.MembershipType, status.VirtualExpiry,
			))
			maafocus.Print(ctx, fmt.Sprintf(
				i18n.T("tasker.membership_check.sponsor"),
				sponsorURL,
			))
		})
		return true
	}

	// 非会员每次都显示提示
	maafocus.Print(ctx, fmt.Sprintf(
		i18n.T("tasker.membership_check.denied"),
		sponsorURL,
	))

	return false
}
