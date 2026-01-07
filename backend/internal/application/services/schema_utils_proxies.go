package services

import (
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/models"
)

// =========================================================================================
// Proxies for Schema Helper Functions moved to persistence (Backward Compatibility)
// =========================================================================================

// NormalizeSObject keys to match schema (case-insensitive) and coerce values to correct type.
func NormalizeSObject(schema *models.ObjectMetadata, input models.SObject) models.SObject {
	return persistence.NormalizeSObject(schema, input)
}

// ToStorageRecord prepares an SObject for database insertion/update.
func ToStorageRecord(schema *models.ObjectMetadata, data models.SObject) models.SObject {
	return persistence.ToStorageRecord(schema, data)
}

// FromStorageRecord hydrates an SObject from a database row.
func FromStorageRecord(schema *models.ObjectMetadata, record models.SObject, visibleFields []string) models.SObject {
	return persistence.FromStorageRecord(schema, record, visibleFields)
}

// FindField finds a field in the schema by name (case-insensitive)
func FindField(schema *models.ObjectMetadata, fieldName string) *models.FieldMetadata {
	return persistence.FindField(schema, fieldName)
}

// GetPolymorphicTypeColumnName returns the standardized column name for the type discriminator
func GetPolymorphicTypeColumnName(fieldAPIName string) string {
	return persistence.GetPolymorphicTypeColumnName(fieldAPIName)
}

// GenerateObjectID generates a standardized ID for an object based on its API Name
func GenerateObjectID(apiName string) string {
	return persistence.GenerateObjectID(apiName)
}

// GenerateFieldID generates a standardized ID for a field
func GenerateFieldID(objectAPIName, fieldAPIName string) string {
	return persistence.GenerateFieldID(objectAPIName, fieldAPIName)
}

// GenerateTableID generates a standardized ID for a table
func GenerateTableID(tableName string) string {
	return persistence.GenerateTableID(tableName)
}

// GenerateAutoNumberID generates a standardized ID for an auto-number sequence
func GenerateAutoNumberID(objectAPIName, fieldAPIName string) string {
	return persistence.GenerateAutoNumberID(objectAPIName, fieldAPIName)
}

// GenerateAppID generates a standardized ID for an app
func GenerateAppID(apiName string) string {
	return persistence.GenerateAppID(apiName)
}
