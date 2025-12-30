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
