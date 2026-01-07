package services

import (
	"context"
	"fmt"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

type FeedService struct {
	persistence *PersistenceService
	query       *QueryService
}

func NewFeedService(persistence *PersistenceService, query *QueryService) *FeedService {
	return &FeedService{
		persistence: persistence,
		query:       query,
	}
}

// CreateComment creates a new comment
func (s *FeedService) CreateComment(ctx context.Context, comment models.SystemComment, user *models.UserSession) (*models.SystemComment, error) {
	data := comment.ToSObject()

	// Create using Persistence Service
	record, err := s.persistence.Insert(ctx, constants.TableComment, data, user)
	if err != nil {
		return nil, err
	}

	// Map back to SystemComment
	return s.mapToComment(record), nil
}

// GetComments gets comments for a record
func (s *FeedService) GetComments(ctx context.Context, recordID string, user *models.UserSession) ([]models.SystemComment, error) {
	// Use formula expression for filtering
	filterExpr := fmt.Sprintf("%s == '%s'", constants.FieldSysComment_RecordID, recordID)
	results, err := s.query.QueryWithFilter(
		ctx,
		constants.TableComment,
		filterExpr,
		user,
		constants.FieldCreatedDate,
		constants.SortDESC,
		50,
	)

	if err != nil {
		return nil, err
	}

	comments := make([]models.SystemComment, len(results))
	for i, record := range results {
		comments[i] = *s.mapToComment(record)
	}

	return comments, nil
}

func (s *FeedService) mapToComment(record models.SObject) *models.SystemComment {
	return &models.SystemComment{
		ID:            record.GetString(constants.FieldID),
		Body:          record.GetString(constants.FieldSysComment_Body),
		RecordID:      record.GetString(constants.FieldSysComment_RecordID),
		ObjectAPIName: record.GetString(constants.FieldSysComment_ObjectAPIName),
		ParentCommentID: func() *string {
			if v := record.GetString(constants.FieldSysComment_ParentCommentID); v != "" {
				return &v
			}
			return nil
		}(),
		IsResolved:  record.GetBool(constants.FieldSysComment_IsResolved),
		CreatedBy:   record.GetString(constants.FieldCreatedByID),
		CreatedDate: record.GetTime(constants.FieldCreatedDate),
	}
}
