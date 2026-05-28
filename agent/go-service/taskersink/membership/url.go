package membership

import "fmt"

func SponsorURL(status *MembershipStatus) string {
	return fmt.Sprintf(
		"https://doropay.top?cpu=%s&uuid=%s&bios=%s&board=%s&disk=%s&guid=%s",
		status.DeviceCode.CPUHash,
		status.DeviceCode.UUIDHash,
		status.DeviceCode.BIOSHash,
		status.DeviceCode.BoardHash,
		status.DeviceCode.DiskHash,
		status.DeviceCode.GUIDHash,
	)
}
