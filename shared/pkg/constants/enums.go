package constants

// SchemaFieldType represents the type of a field
type SchemaFieldType string

const (
	FieldTypeText            SchemaFieldType = "Text"
	FieldTypeNumber          SchemaFieldType = "Number"
	FieldTypeCurrency        SchemaFieldType = "Currency"
	FieldTypeDate            SchemaFieldType = "Date"
	FieldTypeDateTime        SchemaFieldType = "DateTime"
	FieldTypePicklist        SchemaFieldType = "Picklist"
	FieldTypeEmail           SchemaFieldType = "Email"
	FieldTypePhone           SchemaFieldType = "Phone"
	FieldTypeTextArea        SchemaFieldType = "TextArea"
	FieldTypeLookup          SchemaFieldType = "Lookup"
	FieldTypeURL             SchemaFieldType = "Url"
	FieldTypeBoolean         SchemaFieldType = "Boolean"
	FieldTypeFormula         SchemaFieldType = "Formula"
	FieldTypePercent         SchemaFieldType = "Percent"
	FieldTypeRollupSummary   SchemaFieldType = "RollupSummary"
	FieldTypeJSON            SchemaFieldType = "JSON"
	FieldTypeLongTextArea    SchemaFieldType = "LongTextArea"
	FieldTypeRichText        SchemaFieldType = "RichText"
	FieldTypePassword        SchemaFieldType = "Password"
	FieldTypeAutoNumber      SchemaFieldType = "AutoNumber"
	FieldTypeMultiPicklist   SchemaFieldType = "MultiPicklist"
	FieldTypeMasterDetail    SchemaFieldType = "MasterDetail"
	FieldTypeEncryptedString SchemaFieldType = "EncryptedString"
)

// GetAllFieldTypes returns all valid field types as a slice of strings
func GetAllFieldTypes() []string {
	return []string{
		string(FieldTypeText),
		string(FieldTypeNumber),
		string(FieldTypeCurrency),
		string(FieldTypeDate),
		string(FieldTypeDateTime),
		string(FieldTypePicklist),
		string(FieldTypeEmail),
		string(FieldTypePhone),
		string(FieldTypeTextArea),
		string(FieldTypeLookup),
		string(FieldTypeURL),
		string(FieldTypeBoolean),
		string(FieldTypeFormula),
		string(FieldTypePercent),
		string(FieldTypeRollupSummary),
		string(FieldTypeJSON),
		string(FieldTypeLongTextArea),
		string(FieldTypeRichText),
		string(FieldTypePassword),
		string(FieldTypeAutoNumber),
		string(FieldTypeMultiPicklist),
		string(FieldTypeMasterDetail),
		string(FieldTypeEncryptedString),
	}
}

// SharingModel represents object-level sharing model
type SharingModel string

const (
	SharingModelPrivate         SharingModel = "Private"
	SharingModelPublicRead      SharingModel = "PublicRead"
	SharingModelPublicReadWrite SharingModel = "PublicReadWrite"
)

// DeleteRule represents referential integrity rules
type DeleteRule string

const (
	DeleteRuleRestrict DeleteRule = "Restrict"
	DeleteRuleCascade  DeleteRule = "Cascade"
	DeleteRuleSetNull  DeleteRule = "SetNull"
)

// ObjectCategory defines the category of an object
type ObjectCategory string

const (
	ObjectCategoryCustom ObjectCategory = "Custom"
	ObjectCategorySystem ObjectCategory = "System"
)

// OutboxEventStatus represents the processing status of outbox events
type OutboxEventStatus string

const (
	OutboxStatusPending   OutboxEventStatus = "pending"
	OutboxStatusProcessed OutboxEventStatus = "processed"
	OutboxStatusFailed    OutboxEventStatus = "failed"
)
