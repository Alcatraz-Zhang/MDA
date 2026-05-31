package membership

import maa "github.com/MaaXYZ/maa-framework-go/v4"

var runtimeTracker = &RuntimeTracker{}

// Register registers membership quota checks and runtime tracking.
func Register() {
	maa.AgentServerRegisterCustomAction("RuntimeQuotaCheck", &RuntimeQuotaCheckAction{})
	maa.AgentServerAddTaskerSink(runtimeTracker)
	maa.AgentServerAddContextSink(runtimeTracker)
}
