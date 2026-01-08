package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/utils"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Recycle Bin Operations ====================

// GetRecycleBinItems returns items from the recycle bin
func (ps *PersistenceService) GetRecycleBinItems(ctx context.Context, currentUser *models.UserSession, scope string) ([]models.SObject, error) {
	// Permission check for "all" scope
	if scope == "all" && !currentUser.IsSuperUser() {
		return nil, errors.NewPermissionError("view_all_recycle_bin", "recycle_bin")
	}

	// Determine which repository method to call based on scope
	if scope == "all" {
		// Logic: Admin check already done
		return ps.repo.FindAllRecycleBinItems(ctx)
	}

	// Default to "mine"
	return ps.repo.FindRecycleBinItemsByUser(ctx, currentUser.Name)
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
	if !ps.permissions.CheckObjectPermissionWithUser(ctx, objectName, constants.PermDelete, currentUser) {
		return fmt.Errorf("insufficient permissions to restore %s", objectName)
	}

	// Verify the record exists and is deleted
	if err := ps.verifyDeletedRecord(ctx, objectName, recordId); err != nil {
		return err
	}

	// Execute restore within transaction
	err = ps.txManager.WithRetry(func(tx *sql.Tx) error {
		// 1. Undelete record (IsDeleted = false)
		updates := models.SObject{constants.FieldIsDeleted: constants.IsDeletedFalse}
		if err := ps.repo.Update(ctx, tx, objectName, recordId, updates); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}

		// 2. Remove from recycle bin
		if err := ps.repo.DeleteByField(ctx, tx, constants.TableRecycleBin, constants.FieldRecordID, recordId); err != nil {
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
	if !ps.permissions.CheckObjectPermissionWithUser(ctx, objectName, constants.PermDelete, currentUser) {
		return fmt.Errorf("insufficient permissions to purge %s", objectName)
	}

	// Verify the record exists and is deleted
	if err := ps.verifyDeletedRecord(ctx, objectName, recordId); err != nil {
		return err
	}

	// Execute purge within transaction
	err = ps.txManager.WithRetry(func(tx *sql.Tx) error {
		// Permanently delete the record
		if err := ps.repo.PhysicalDelete(ctx, tx, objectName, recordId); err != nil {
			return fmt.Errorf("purge failed: %w", err)
		}

		// Remove from recycle bin
		if err := ps.repo.DeleteByField(ctx, tx, constants.TableRecycleBin, constants.FieldRecordID, recordId); err != nil {
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
	binRecord, err := ps.repo.FindRecycleBinEntry(ctx, recordId)
	if err != nil {
		return "", fmt.Errorf("failed to find recycle bin record: %w", err)
	}
	if binRecord == nil {
		return "", fmt.Errorf("record not found in recycle bin")
	}

	objectName, ok := binRecord[constants.FieldSysRecycleBin_ObjectAPIName].(string)
	if !ok {
		return "", fmt.Errorf("invalid object_api_name in recycle bin")
	}
	return objectName, nil
}

func (ps *PersistenceService) verifyDeletedRecord(ctx context.Context, objectName, recordId string) error {
	record, err := ps.repo.FindAny(ctx, nil, objectName, recordId)
	if err != nil {
		return fmt.Errorf("failed to verify record: %w", err)
	}

	if record == nil {
		return fmt.Errorf("record not found")
	}

	// Check IsDeleted flag
	isDeleted, ok := record[constants.FieldIsDeleted]
	if !ok {
		// If no IsDeleted field, it can't be deleted
		return fmt.Errorf("record does not support deletion status")
	}

	// Use standardized utility for type coercion
	if utils.ToBool(isDeleted) {
		return nil
	}

	return fmt.Errorf("record is not deleted")
}
