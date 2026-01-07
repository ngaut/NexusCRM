package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Insert Operations ====================

// Insert creates a new record with ACID transaction guarantees
func (ps *PersistenceService) Insert(
	ctx context.Context,
	objectName string,
	data models.SObject,
	currentUser *models.UserSession,
) (models.SObject, error) {
	schema, err := ps.prepareOperation(ctx, objectName, constants.PermCreate, currentUser)
	if err != nil {
		return nil, err
	}

	// Apply defaults
	data = ps.applyDefaults(data, schema, currentUser)

	// Validate Polymorphic Lookups (Database Check) & Resolve Types
	resolvedTypes, err := ps.validatePolymorphicLookups(ctx, data, schema)
	if err != nil {
		return nil, err
	}

	// Inject resolved types into data for persistence
	for fieldName, objType := range resolvedTypes {
		data[GetPolymorphicTypeColumnName(fieldName)] = objType
	}

	// Validate Static Rules
	validationRules := ps.metadata.GetValidationRules(ctx, objectName)
	if err := ps.validator.ValidateRecord(data, schema, validationRules, nil); err != nil {
		return nil, err
	}

	// Check uniqueness
	if err := ps.checkUniqueness(ctx, objectName, data, schema, ""); err != nil {
		return nil, err
	}

	// Generate system fields dynamically from metadata
	systemFields := ps.generateSystemFields(objectName, data, currentUser, true)

	// Merge data with system fields
	for k, v := range systemFields {
		data[k] = v
	}

	// CRITICAL: Ensure _System_User passwords are hashed if not already.
	// This provides a fallback if the system Flow is missing or fails.
	if strings.EqualFold(objectName, constants.TableUser) {
		if pwd, ok := data[constants.FieldPassword].(string); ok && pwd != "" {
			// Check if already hashed (Bcrypt starts with $2a$, $2b$, or $2y$)
			if !strings.HasPrefix(pwd, "$2") {
				hashed, err := auth.HashPassword(pwd)
				if err != nil {
					return nil, fmt.Errorf("failed to hash password: %w", err)
				}
				data[constants.FieldPassword] = string(hashed)
			}
		}
	}

	// Execute Transactional Work
	err = ps.RunInTransaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		// Generate AutoNumbers WITHIN transaction to ensure atomicity
		if err := ps.generateAutoNumbers(txCtx, tx, objectName, data); err != nil {
			return fmt.Errorf("failed to generate auto-numbers: %w", err)
		}

		// Publish beforeCreate event (synchronous, can fail transaction)
		if err := ps.publishRecordEvent(txCtx, events.RecordBeforeCreate, objectName, data, nil, currentUser); err != nil {
			return err
		}

		// Extract physical fields only
		physicalData := ToStorageRecord(schema, data)

		// Build and execute insert within transaction
		builder := query.Insert(objectName, physicalData)
		q := builder.Build()

		if _, err := tx.Exec(q.SQL, q.Params...); err != nil {
			return fmt.Errorf("insert failed: %w", err)
		}

		// Process Mentions for Comments
		if objectName == constants.TableComment {
			if err := ps.processMentions(tx, data, currentUser); err != nil {
				// Log error but generally don't fail the insert
				log.Printf("Failed to process mentions: %v", err)
			}
		}

		// Hook: Rollup Summary (inside transaction for ACID compliance)
		if err := ps.rollup.ProcessRollups(txCtx, tx, objectName, data); err != nil {
			return fmt.Errorf("failed to process rollups: %w", err)
		}

		// Enqueue afterCreate event to outbox (inside transaction for guaranteed delivery)
		if ps.outbox != nil {
			if err := ps.outbox.EnqueueEventTx(txCtx, tx, events.RecordCreated, RecordEventPayload{
				ObjectAPIName: objectName,
				Record:        data,
				CurrentUser:   currentUser,
			}); err != nil {
				return fmt.Errorf("failed to enqueue record created event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	id, _ := data[constants.FieldID].(string)
	log.Printf("✨ Created record %s in %s (User: %s)", id, objectName, getUserID(currentUser))

	return data, nil
}

// processMentions parses the comment body for mentions and creates notifications
func (ps *PersistenceService) processMentions(tx *sql.Tx, data models.SObject, sender *models.UserSession) error {
	body, ok := data[constants.FieldSysComment_Body].(string)
	if !ok {
		return nil
	}

	// Find data-id="USER_ID" from TipTap mention elements
	search := `data-id="`
	startIndex := 0
	for {
		idx := strings.Index(body[startIndex:], search)
		if idx == -1 {
			break
		}
		idx += startIndex
		quoteStart := idx + len(search)
		quoteEnd := strings.Index(body[quoteStart:], `"`)
		if quoteEnd == -1 {
			break
		}
		userID := body[quoteStart : quoteStart+quoteEnd]

		// Create Notification
		notifID := GenerateID()

		log.Printf("📧 sending email to %s: You were mentioned by %s", userID, sender.Name)

		notifStruct := models.SystemNotification{
			ID:               notifID,
			RecipientID:      userID,
			Title:            fmt.Sprintf("New mention by %s", sender.Name),
			Body:             "You were mentioned in a comment.",
			Link:             fmt.Sprintf("/object/%s/%s", data[constants.FieldSysComment_ObjectAPIName], data[constants.FieldSysComment_RecordID]),
			NotificationType: "mention",
			IsRead:           false,
			CreatedDate:      time.Now(),
		}
		notif := notifStruct.ToSObject()

		q := query.Insert(constants.TableNotification, notif).Build()
		if _, err := tx.Exec(q.SQL, q.Params...); err != nil {
			log.Printf("Failed to create notification for %s: %v", userID, err)
		}

		startIndex = quoteStart + quoteEnd + 1
	}
	return nil
}
