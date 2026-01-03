package constants

// Flow status constants
const (
	FlowStatusActive   = "Active"
	FlowStatusInactive = "Inactive"
	FlowStatusDraft    = "Draft"
)

// Flow type constants
const (
	FlowTypeSimple    = "simple"
	FlowTypeMultistep = "multistep"
)

// Flow step type constants
const (
	FlowStepTypeApproval = "approval"
	FlowStepTypeAction   = "action"
	FlowStepTypeDecision = "decision"
)

// Flow instance status constants
const (
	FlowInstanceStatusRunning   = "Running"
	FlowInstanceStatusPaused    = "Paused"
	FlowInstanceStatusCompleted = "Completed"
	FlowInstanceStatusFailed    = "Failed"
)

// Approval work item status constants
const (
	ApprovalStatusPending  = "Pending"
	ApprovalStatusApproved = "Approved"
	ApprovalStatusRejected = "Rejected"
)

// Flow trigger types
const (
	TriggerTypeRecordCreated = "record_created"
	TriggerTypeRecordUpdated = "record_updated"
	TriggerTypeRecordDeleted = "record_deleted"
	TriggerTypeSchedule      = "schedule"
)

// Schedule-related constants
const (
	ScheduleDefaultTimezone = "UTC"
	ScheduleCheckInterval   = 60 // Seconds between scheduler checks
	ScheduleMaxRuntimeMins  = 30 // Maximum execution time before timeout (minutes)
	ScheduleMinIntervalMins = 1  // Minimum interval between runs (minutes)
)
