package services

import (
	"context"
	"fmt"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
)

type NotificationService struct {
	persistence *PersistenceService
	query       *QueryService
}

func NewNotificationService(persistence *PersistenceService, query *QueryService) *NotificationService {
	return &NotificationService{
		persistence: persistence,
		query:       query,
	}
}

// GetMyNotifications returns unread notifications for the user
func (s *NotificationService) GetMyNotifications(ctx context.Context, user *models.UserSession) ([]models.SystemNotification, error) {
	// Query _System_Notification where recipient_id = user.ID
	// Using formula expression for filtering
	filterExpr := fmt.Sprintf("%s == '%s'", constants.FieldSysNotification_RecipientID, user.ID)
	results, err := s.query.QueryWithFilter(
		ctx,
		constants.TableNotification,
		filterExpr,
		user,
		constants.FieldCreatedDate,
		"DESC",
		20,
	)

	if err != nil {
		return nil, err
	}

	notifications := make([]models.SystemNotification, len(results))
	for i, record := range results {
		notifications[i] = *s.mapToNotification(record)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id string, user *models.UserSession) error {
	updates := map[string]interface{}{
		constants.FieldSysNotification_IsRead: true,
	}
	return s.persistence.Update(ctx, constants.TableNotification, id, updates, user)
}

// CreateNotification creates a notification (System internal use usually, but exposed for testing/admin)
func (s *NotificationService) CreateNotification(ctx context.Context, notification models.SystemNotification, user *models.UserSession) error {
	data := notification.ToSObject()
	// Ensure is_read is false default
	data[constants.FieldSysNotification_IsRead] = false

	_, err := s.persistence.Insert(ctx, constants.TableNotification, data, user)
	return err
}

func (s *NotificationService) mapToNotification(record models.SObject) *models.SystemNotification {
	return &models.SystemNotification{
		ID:               record.GetString(constants.FieldID),
		RecipientID:      record.GetString(constants.FieldSysNotification_RecipientID),
		Title:            record.GetString(constants.FieldSysNotification_Title),
		Body:             record.GetString(constants.FieldSysNotification_Body),
		Link:             record.GetString(constants.FieldSysNotification_Link),
		NotificationType: record.GetString(constants.FieldSysNotification_NotificationType),
		IsRead:           record.GetBool(constants.FieldSysNotification_IsRead),
		CreatedDate:      record.GetTime(constants.FieldCreatedDate),
	}
}
