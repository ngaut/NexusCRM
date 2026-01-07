package constants

// FieldType represents the type of a field
type FieldType string

const (
	FieldTypeText          FieldType = "Text"
	FieldTypeNumber        FieldType = "Number"
	FieldTypeCurrency      FieldType = "Currency"
	FieldTypeDate          FieldType = "Date"
	FieldTypeDateTime      FieldType = "DateTime"
	FieldTypePicklist      FieldType = "Picklist"
	FieldTypeEmail         FieldType = "Email"
	FieldTypePhone         FieldType = "Phone"
	FieldTypeTextArea      FieldType = "TextArea"
	FieldTypeLookup        FieldType = "Lookup"
	FieldTypeURL           FieldType = "Url"
	FieldTypeBoolean       FieldType = "Boolean"
	FieldTypeFormula       FieldType = "Formula"
	FieldTypePercent       FieldType = "Percent"
	FieldTypeRollupSummary FieldType = "RollupSummary"
	FieldTypeJSON          FieldType = "JSON"
	FieldTypeLongTextArea  FieldType = "LongTextArea"
	FieldTypeRichText      FieldType = "RichText"
	FieldTypePassword      FieldType = "Password"
	FieldTypeAutoNumber    FieldType = "AutoNumber"
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
