package constants

// Action type constants - single source of truth from registry
const (
	ActionTypeCreateRecord = "CreateRecord"
	ActionTypeUpdateRecord = "UpdateRecord"
	ActionTypeDeleteRecord = "DeleteRecord"
	ActionTypeSendEmail    = "SendEmail"
	ActionTypeCallWebhook  = "CallWebhook"
	ActionTypeComposite    = "Composite"
)

// Flow trigger type constants
const (
	TriggerBeforeCreate = "beforeCreate"
	TriggerAfterCreate  = "afterCreate"
	TriggerBeforeUpdate = "beforeUpdate"
	TriggerAfterUpdate  = "afterUpdate"
	TriggerBeforeDelete = "beforeDelete"
	TriggerAfterDelete  = "afterDelete"
)
