package rest

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	appErrors "github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Schema Handlers ====================

// GetSchemas handles GET /api/metadata/objects
func (h *MetadataHandler) GetSchemas(c *gin.Context) {
	HandleGetEnvelope(c, "schemas", func() (interface{}, error) {
		return h.svc.GetSchemas(c.Request.Context()), nil
	})
}

// GetSchema handles GET /api/metadata/objects/:apiName
func (h *MetadataHandler) GetSchema(c *gin.Context) {
	user := GetUserFromContext(c)
	apiName := strings.ToLower(c.Param("apiName"))

	HandleGetEnvelope(c, "schema", func() (interface{}, error) {
		// Get effective schema (filtered by permissions)
		schema := h.svc.GetEffectiveSchema(c.Request.Context(), apiName, user)
		if schema == nil {
			return nil, appErrors.NewNotFoundError("Schema", apiName)
		}
		return schema, nil
	})
}

// CreateSchema handles POST /api/metadata/objects
func (h *MetadataHandler) CreateSchema(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	// Define request structure to include optional app additions
	var req struct {
		models.ObjectMetadata
	}

	HandleCreateEnvelope(c, "schema", "Schema created successfully", &req, func() error {
		schema := &req.ObjectMetadata

		// Validation inside generic handler action
		if schema.APIName == "" || schema.Label == "" {
			return appErrors.NewValidationError("api_name", "API Name and Label are required")
		}

		// Use CreateObjectInApp if AppID is present
		if schema.AppID != nil && *schema.AppID != "" {
			if err := h.svc.Metadata.CreateObjectInApp(c.Request.Context(), *schema.AppID, schema); err != nil {
				return err
			}
		} else {
			if err := h.svc.Metadata.CreateSchema(c.Request.Context(), schema); err != nil {
				return err
			}
		}

		// Perform post-creation updates (Permissions) in a transaction
		// App navigation is now handled in CreateObjectInApp for that flow,
		// but we still need to grant permissions.
		err := h.svc.TxManager.WithTransaction(func(tx *sql.Tx) error {
			adminProfileID := constants.ProfileSystemAdmin
			perm := models.SystemObjectPerms{
				ProfileID:     &adminProfileID,
				ObjectAPIName: schema.APIName,
				AllowRead:     true,
				AllowCreate:   true,
				AllowEdit:     true,
				AllowDelete:   true,
				ViewAll:       true,
				ModifyAll:     true,
			}

			// Use Transactional version
			if err := h.svc.Permissions.UpdateObjectPermissionTx(tx, perm); err != nil {
				return fmt.Errorf("failed to grant admin permissions: %w", err)
			}
			return nil
		})

		if err != nil {
			// Compensation: If post-updates fail, we should ideally delete the schema to maintain atomicity.
			// For now, we return the error, but the schema exists.
			log.Printf("CRITICAL: Schema created but post-updates failed: %v", err)
			return err
		}

		return nil
	})
}

// UpdateSchema handles PATCH /api/metadata/objects/:apiName
func (h *MetadataHandler) UpdateSchema(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	apiName := strings.ToLower(c.Param("apiName"))
	var updates models.ObjectMetadata
	HandleUpdateEnvelope(c, "schema", "Schema updated successfully", &updates, func() error {
		return h.svc.Metadata.UpdateSchema(c.Request.Context(), apiName, &updates)
	})
}

// DeleteSchema handles DELETE /api/metadata/objects/:apiName
func (h *MetadataHandler) DeleteSchema(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	apiName := strings.ToLower(c.Param("apiName"))

	HandleDeleteEnvelope(c, "Object deleted successfully", func() error {
		// Step 1: Drop table (DDL - auto-commits, cannot be in transaction)
		if err := h.svc.Metadata.DeleteSchema(c.Request.Context(), apiName); err != nil {
			return err
		}

		// Step 2: Cleanup metadata (DML - in transaction for atomicity)
		// Note: Permissions and layouts are auto-deleted via CASCADE foreign keys
		// We only need to manually clean up app navigation items
		err := h.svc.TxManager.WithTransaction(func(tx *sql.Tx) error {
			// Remove object from all app navigation items
			if err := h.svc.UIMetadata.RemoveObjectFromAllAppsTx(c.Request.Context(), tx, apiName); err != nil {
				log.Printf("⚠️  Warning: Failed to remove object '%s' from app navigation: %v", apiName, err)
				// Don't fail the entire deletion if this cleanup fails
				// The object is already deleted from DB, this is just UI cleanup
			}
			return nil
		})

		if err != nil {
			log.Printf("⚠️  Critical: Object '%s' deleted but post-deletion cleanup failed: %v", apiName, err)
		}

		return nil
	})
}

// ==================== Field Handlers ====================

// CreateField handles POST /api/metadata/objects/:apiName/fields
func (h *MetadataHandler) CreateField(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	objectAPIName := strings.ToLower(c.Param("apiName"))
	var field models.FieldMetadata

	HandleCreateEnvelope(c, "field", "Field created successfully", &field, func() error {
		if field.APIName == "" || field.Label == "" {
			return appErrors.NewValidationError("api_name", "API Name and Label are required")
		}

		// Validate field type specific requirements
		switch field.Type {
		case constants.FieldTypePicklist:
			if len(field.Options) == 0 {
				return appErrors.NewValidationError("options", "Picklist fields require at least one option")
			}
		case constants.FieldTypeLookup:
			if len(field.ReferenceTo) == 0 {
				return appErrors.NewValidationError("reference_to", "Lookup fields require a referenced object")
			}
		}

		if err := h.svc.Metadata.CreateField(c.Request.Context(), objectAPIName, &field); err != nil {
			return err
		}

		return nil
	})
}

// UpdateField handles PATCH /api/metadata/objects/:apiName/fields/:fieldApiName
func (h *MetadataHandler) UpdateField(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	objectAPIName := strings.ToLower(c.Param("apiName"))
	fieldAPIName := c.Param("fieldApiName")
	var updates models.FieldMetadata

	// No return key for UpdateField
	HandleUpdateEnvelope(c, "", "Field updated successfully", &updates, func() error {
		return h.svc.Metadata.UpdateField(c.Request.Context(), objectAPIName, fieldAPIName, &updates)
	})
}

// DeleteField handles DELETE /api/metadata/objects/:apiName/fields/:fieldApiName
func (h *MetadataHandler) DeleteField(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	objectAPIName := strings.ToLower(c.Param("apiName"))
	fieldAPIName := c.Param("fieldApiName")

	HandleDeleteEnvelope(c, "Field deleted successfully", func() error {
		return h.svc.Metadata.DeleteField(c.Request.Context(), objectAPIName, fieldAPIName)
	})
}

// ==================== Validation Rule Handlers ====================

// CreateValidationRule handles POST /api/metadata/validation-rules
func (h *MetadataHandler) CreateValidationRule(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	var rule models.ValidationRule
	HandleCreateEnvelope(c, "rule", "Validation rule created successfully", &rule, func() error {
		return h.svc.Metadata.CreateValidationRule(c.Request.Context(), &rule)
	})
}

// UpdateValidationRule handles PATCH /api/metadata/validation-rules/:id
func (h *MetadataHandler) UpdateValidationRule(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	id := c.Param(constants.FieldID)
	var updates models.ValidationRule
	// No return key for UpdateValidationRule? Check existing... "c.JSON(OK, ... message)". Yes.
	HandleUpdateEnvelope(c, "", "Validation rule updated successfully", &updates, func() error {
		return h.svc.Metadata.UpdateValidationRule(c.Request.Context(), id, &updates)
	})
}

// DeleteValidationRule handles DELETE /api/metadata/validation-rules/:id
func (h *MetadataHandler) DeleteValidationRule(c *gin.Context) {
	// requireSystemAdmin handled by middleware

	id := c.Param(constants.FieldID)
	HandleDeleteEnvelope(c, "Validation rule deleted successfully", func() error {
		return h.svc.Metadata.DeleteValidationRule(c.Request.Context(), id)
	})
}

// GetValidationRules handles GET /api/metadata/validation-rules
func (h *MetadataHandler) GetValidationRules(c *gin.Context) {
	objectAPIName := strings.ToLower(c.Query("objectApiName"))
	HandleGetEnvelope(c, "rules", func() (interface{}, error) {
		if objectAPIName == "" {
			return nil, appErrors.NewValidationError("objectApiName", "is required")
		}
		return h.svc.Metadata.GetValidationRules(c.Request.Context(), objectAPIName), nil
	})
}

// GetFieldTypes handles GET /api/metadata/fieldtypes
// Returns all available field types including custom plugin types
func (h *MetadataHandler) GetFieldTypes(c *gin.Context) {
	HandleGetEnvelope(c, "fieldTypes", func() (interface{}, error) {
		return GetAllFieldTypes(), nil
	})
}
