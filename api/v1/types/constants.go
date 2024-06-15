package v1

type StrictSchedulingStatus string

const StrictSchedulingStatusEnable StrictSchedulingStatus = "enable"
const StrictSchedulingStatusDisable StrictSchedulingStatus = "disable"

type MachineType string

const MachineTypeLowPriority MachineType = "low-priority"
const MachineTypeDefaultPriority MachineType = "default-priority"

const (
	PrivateModeNSLabelKey = "private_mode"
	PrivateObjectLabelKey = "private_object"
)
