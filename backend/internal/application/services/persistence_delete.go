package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// Delete soft-deletes a record with ACID transaction guarantees
func (ps *PersistenceService) Delete(
	ctx context.Context,
	objectName string,
	id string,
	currentUser *models.UserSession,
) error {
	schema, err := ps.prepareOperation(objectName, constants.PermDelete, currentUser)
	if err != nil {
		return err
	}

	// Load record to check permissions and child relationships
	q := query.From(objectName).
		Select([]string{"*"}).
		Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
		ExcludeDeleted().
		Limit(1).
		Build()

	records, err := ExecuteQuery(ctx, ps.db, q)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil // Already deleted
	}

	record := records[0]

	// Check record-level access
	if schema != nil && !ps.permissions.CheckRecordAccess(schema, record, constants.PermDelete, currentUser) {
		return errors.NewPermissionError(constants.PermDelete, objectName+"/"+id)
	}

	// Check child relationships
	if schema != nil {
		children := ps.metadata.GetChildRelationships(objectName)
		for _, childSchema := range children {
			for _, field := range childSchema.Fields {
				isLookup := strings.EqualFold(string(field.Type), string(constants.FieldTypeLookup))
				if isLookup && ContainsStringIgnoreCase(field.ReferenceTo, objectName) {
					deleteRule := constants.DeleteRuleRestrict
					if field.DeleteRule != nil {
						deleteRule = *field.DeleteRule
					}

					checkQ := query.From(childSchema.APIName).
						Select([]string{constants.FieldID}).
						Where(fmt.Sprintf("`%s` = ?", field.APIName), id).
						ExcludeDeleted().
						Limit(1).
						Build()

					checkRecords, err := ExecuteQuery(ctx, ps.db, checkQ)
					if err != nil {
						log.Printf("Warning: failed to check child records for %s.%s: %v", childSchema.APIName, field.APIName, err)
						continue
					}

					if len(checkRecords) > 0 {
						if strings.EqualFold(string(deleteRule), string(constants.DeleteRuleRestrict)) {
							return errors.NewConflictError(objectName, "referenced by", childSchema.PluralLabel)
						}
					}
				}
			}
		}
	}

	// Execute Transactional Work
	err = ps.RunInTransaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		// Publish beforeDelete event
		if err := ps.publishRecordEvent(txCtx, events.RecordBeforeDelete, objectName, record, nil, currentUser); err != nil {
			return err
		}

		// Cascade Delete Children (Application-Level)
		if err := ps.cascadeDeleteChildren(txCtx, currentUser, objectName, id); err != nil {
			return fmt.Errorf("failed to cascade delete children: %w", err)
		}

		// Soft delete within transaction
		updateQ := query.Update(objectName).
			Set(models.SObject{constants.FieldIsDeleted: 1}).
			Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
			Build()

		if _, err := tx.Exec(updateQ.SQL, updateQ.Params...); err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}

		// Add to recycle bin within transaction
		recordName := ps.getRecordName(record, schema)
		deletedBy := constants.SystemUserName
		if currentUser != nil {
			deletedBy = currentUser.Name
		}

		binID := fmt.Sprintf("%s-%d", id, time.Now().UnixNano())
		binQ := query.Insert(constants.TableRecycleBin, models.SObject{
			constants.FieldID:               binID,
			constants.FieldRecordID:         id,
			constants.FieldObjectAPIName:    objectName,
			constants.FieldRecordName:       recordName,
			constants.FieldDeletedBy:        deletedBy,
			constants.FieldDeletedDate:      NowTimestamp(),
			constants.FieldCreatedDate:      NowTimestamp(),
			constants.FieldLastModifiedDate: NowTimestamp(),
		}).Build()

		if _, err := tx.Exec(binQ.SQL, binQ.Params...); err != nil {
			return fmt.Errorf("failed to add to recycle bin: %w", err)
		}

		// Hook: Rollup Summary (inside transaction for ACID compliance)
		if err := ps.rollup.ProcessRollups(txCtx, tx, objectName, record); err != nil {
			log.Printf("⚠️ Failed to process rollups for deleted record %s/%s: %v", objectName, id, err)
			return fmt.Errorf("failed to process rollups: %w", err)
		}

		// Enqueue afterDelete event to outbox (inside transaction for guaranteed delivery)
		if ps.outbox != nil {
			if err := ps.outbox.EnqueueEventTx(txCtx, tx, events.RecordDeleted, RecordEventPayload{
				ObjectAPIName: objectName,
				Record:        record,
				CurrentUser:   currentUser,
			}); err != nil {
				return fmt.Errorf("failed to enqueue record deleted event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("🗑️ Deleted record %s in %s (User: %s)", id, objectName, getUserID(currentUser))

	return nil
}

// cascadeDeleteChildren finds and deletes child records that have DeleteRuleCascade
// DESIGN ASSUMPTION: DeleteRule values may be stored in mixed case (e.g., "CASCADE", "Cascade").
// Always use strings.EqualFold for DeleteRule comparisons to ensure case-insensitive matching.
// DESIGN ASSUMPTION: Object API names are case-insensitive (see ContainsStringIgnoreCase).
func (ps *PersistenceService) cascadeDeleteChildren(ctx context.Context, user *models.UserSession, parentObjName, parentID string) error {
	schemas := ps.metadata.GetSchemas()

	for _, schema := range schemas {
		for _, field := range schema.Fields {
			// Check if field references our parent object with Cascade delete
			// IMPORTANT: Use case-insensitive checks for both object names and delete rules
			if field.Type == constants.FieldTypeLookup &&
				ContainsStringIgnoreCase(field.ReferenceTo, parentObjName) &&
				field.DeleteRule != nil &&
				strings.EqualFold(string(*field.DeleteRule), string(constants.DeleteRuleCascade)) {

				childObjName := schema.APIName
				lookupFieldName := field.APIName

				// Find IDs of children (excluding already deleted)
				childQuery := fmt.Sprintf("SELECT %s FROM %s WHERE `%s` = ? AND %s = false",
					constants.FieldID, childObjName, lookupFieldName, constants.FieldIsDeleted)

				var rows *sql.Rows
				var err error

				// ACID: Use existing transaction if available
				tx := ps.txManager.ExtractTx(ctx)
				if tx != nil {
					rows, err = tx.QueryContext(ctx, childQuery, parentID)
				} else {
					rows, err = ps.db.DB().QueryContext(ctx, childQuery, parentID)
				}

				if err != nil {
					// Ignore table not found errors (race conditions or init issues)
					if strings.Contains(err.Error(), "doesn't exist") {
						continue
					}
					return fmt.Errorf("failed to query children of %s via %s: %w", childObjName, lookupFieldName, err)
				}

				var childIDs []string
				for rows.Next() {
					var childID string
					if err := rows.Scan(&childID); err != nil {
						return err
					}
					childIDs = append(childIDs, childID)
				}
				rows.Close()

				// Recursively delete children
				for _, childID := range childIDs {
					// Log for safety/debugging
					log.Printf("Cascade deleting child %s/%s because parent %s/%s was deleted", childObjName, childID, parentObjName, parentID)
					if err := ps.Delete(ctx, childObjName, childID, user); err != nil {
						// Don't fail if already deleted (idempotency)
						if !errors.IsNotFound(err) {
							return fmt.Errorf("failed to cascade delete child %s (%s): %w", childObjName, childID, err)
						}
					}
				}
			}
		}
	}
	return nil
}
