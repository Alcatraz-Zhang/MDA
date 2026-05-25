package membership

// membershipLevels maps tier names to their numeric user level.
var membershipLevels = map[string]int{
	"普通用户":    0,
	"金Doro会员": 3,
}

const MemberStatusURL = "https://doropay.top/api/members/v7/device-status"
