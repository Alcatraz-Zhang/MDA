package membership

import maa "github.com/MaaXYZ/maa-framework-go/v4"

// Register registers membership quota checks and runtime tracking.
func Register() {
	maa.AgentServerRegisterCustomAction("MembershipCheck", &MembershipCheckAction{})
	maa.AgentServerRegisterCustomAction("RuntimeQuotaCheck", &RuntimeQuotaCheckAction{})
	maa.AgentServerAddTaskerSink(&RuntimeTracker{})
}
