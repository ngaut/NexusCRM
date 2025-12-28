package events

// EventType defines the type of event in the system
type EventType string

const (
	// Record Events
	RecordBeforeCreate EventType = "record.beforeCreate"
	RecordCreated      EventType = "record.created"
	RecordAfterCreate  EventType = RecordCreated // Alias for record.created
	RecordBeforeUpdate EventType = "record.beforeUpdate"
	RecordUpdated      EventType = "record.updated"
	RecordAfterUpdate  EventType = RecordUpdated // Alias for record.updated
	RecordBeforeDelete EventType = "record.beforeDelete"
	RecordDeleted      EventType = "record.deleted"
	RecordAfterDelete  EventType = RecordDeleted // Alias for record.deleted

	// Schema Events
	ObjectCreated EventType = "schema.object_created"
	FieldCreated  EventType = "schema.field_created"

	// System Events
	SystemStartup EventType = "system.startup"
)

// String returns the string representation of the event type
func (e EventType) String() string {
	return string(e)
}
