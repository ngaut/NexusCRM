package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/query"
)

// ==================== Recycle Bin Operations ====================

// GetRecycleBinItems returns items from the recycle bin
func (ps *PersistenceService) GetRecycleBinItems(ctx context.Context, currentUser *models.UserSession, scope string) ([]models.SObject, error) {
	// Query the recycle bin table
	builder := query.From(constants.TableRecycleBin).
		Select([]string{"*"})

	// Apply scope filtering
	if scope == "mine" {
		// Filter by user's Name (as stored in Delete)
		// Ideally should use ID, but verified Delete stores Name.
		builder.Where(constants.FieldDeletedBy+" = ?", currentUser.Name)
	} else if scope == "all" {
		// Only admins can see all
		if !currentUser.IsSuperUser() {
			return nil, errors.NewPermissionError("view_all_recycle_bin", "recycle_bin")
		}
	} else {
		// Default to mine if invalid scope, or error? Let's default to mine for safety.
		builder.Where(constants.FieldDeletedBy+" = ?", currentUser.Name)
	}

	binQ := builder.OrderBy(constants.FieldSysRecycleBin_DeletedDate, constants.SortDESC).Build()

	records, err := ExecuteQuery(ctx, ps.db, binQ)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recycle bin items: %w", err)
	}

	return records, nil
}

// Restore undeletes a record from the recycle bin
func (ps *PersistenceService) Restore(
	ctx context.Context,
	recordId string,
	currentUser *models.UserSession,
) error {
	// First, find the record in the recycle bin to get the objectApiName
	objectName, err := ps.getRecycleBinRecord(ctx, recordId)
	if err != nil {
		return err
	}

	// Check permissions
	if !ps.permissions.CheckObjectPermissionWithUser(objectName, constants.PermDelete, currentUser) {
		return fmt.Errorf("insufficient permissions to restore %s", objectName)
	}

	// Verify the record exists and is deleted
	if err := ps.verifyDeletedRecord(ctx, objectName, recordId); err != nil {
		return err
	}

	// Execute restore within transaction
	err = ps.txManager.WithRetry(func(tx *sql.Tx) error {
		restoreQ := query.Update(objectName).
			Set(models.SObject{constants.FieldIsDeleted: constants.IsDeletedFalse}).
			Where(constants.FieldID+" = ?", recordId).
			Build()

		if _, err := tx.Exec(restoreQ.SQL, restoreQ.Params...); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}

		// Remove from recycle bin
		deleteBinQ := query.Delete(constants.TableRecycleBin).
			Where(constants.FieldRecordID+" = ?", recordId).
			Build()

		if _, err := tx.Exec(deleteBinQ.SQL, deleteBinQ.Params...); err != nil {
			return fmt.Errorf("failed to remove from recycle bin: %w", err)
		}

		return nil
	}, 3)

	if err != nil {
		return err
	}

	// Publish restore event (after successful commit)
	ps.eventBus.PublishAsync(events.RecordUpdated, RecordEventPayload{
		ObjectAPIName: objectName,
		Record:        models.SObject{constants.FieldID: recordId},
		CurrentUser:   currentUser,
	})

	return nil
}

// Purge permanently deletes a record from the database
func (ps *PersistenceService) Purge(
	ctx context.Context,
	recordId string,
	currentUser *models.UserSession,
) error {
	// First, find the record in the recycle bin to get the objectApiName
	objectName, err := ps.getRecycleBinRecord(ctx, recordId)
	if err != nil {
		return err
	}

	// Check permissions (purge requires delete permission)
	if !ps.permissions.CheckObjectPermissionWithUser(objectName, constants.PermDelete, currentUser) {
		return fmt.Errorf("insufficient permissions to purge %s", objectName)
	}

	// Verify the record exists and is deleted
	if err := ps.verifyDeletedRecord(ctx, objectName, recordId); err != nil {
		return err
	}

	// Execute purge within transaction
	err = ps.txManager.WithRetry(func(tx *sql.Tx) error {
		// Permanently delete the record
		purgeQ := query.Delete(objectName).
			Where(constants.FieldID+" = ?", recordId).
			Build()

		if _, err := tx.Exec(purgeQ.SQL, purgeQ.Params...); err != nil {
			return fmt.Errorf("purge failed: %w", err)
		}

		// Remove from recycle bin
		deleteBinQ := query.Delete(constants.TableRecycleBin).
			Where(constants.FieldRecordID+" = ?", recordId).
			Build()

		if _, err := tx.Exec(deleteBinQ.SQL, deleteBinQ.Params...); err != nil {
			return fmt.Errorf("failed to remove from recycle bin: %w", err)
		}

		return nil
	}, 3)

	if err != nil {
		return err
	}

	// Publish purge event (after successful commit)
	ps.eventBus.PublishAsync(events.RecordDeleted, RecordEventPayload{
		ObjectAPIName: objectName,
		Record:        models.SObject{constants.FieldID: recordId},
		CurrentUser:   currentUser,
	})

	return nil
}

// ==================== Recycle Bin Helpers ====================

func (ps *PersistenceService) getRecycleBinRecord(ctx context.Context, recordId string) (string, error) {
	binQ := query.From(constants.TableRecycleBin).
		Select([]string{constants.FieldSysRecycleBin_ObjectAPIName}).
		Where(constants.FieldRecordID+" = ?", recordId).
		Limit(1).
		Build()

	binRecords, err := ExecuteQuery(ctx, ps.db, binQ)
	if err != nil {
		return "", fmt.Errorf("failed to scan recycle bin: %w", err)
	}

	if len(binRecords) == 0 {
		return "", fmt.Errorf("record not found in recycle bin")
	}

	objectName, ok := binRecords[0][constants.FieldSysRecycleBin_ObjectAPIName].(string)
	if !ok {
		return "", fmt.Errorf("invalid object_api_name in recycle bin")
	}
	return objectName, nil
}

func (ps *PersistenceService) verifyDeletedRecord(ctx context.Context, objectName, recordId string) error {
	checkQ := query.From(objectName).
		Select([]string{constants.FieldID}).
		Where(constants.FieldID+" = ?", recordId).
		Where(fmt.Sprintf("%s = %d", constants.FieldIsDeleted, constants.IsDeletedTrue)).
		Limit(1).
		Build()

	checkRecords, err := ExecuteQuery(ctx, ps.db, checkQ)
	if err != nil {
		return fmt.Errorf("failed to verify record: %w", err)
	}

	if len(checkRecords) == 0 {
		return fmt.Errorf("record not found or not deleted")
	}
	return nil
}
