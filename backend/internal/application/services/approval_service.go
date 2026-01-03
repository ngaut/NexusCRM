package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ApprovalService handles business logic for approval processes
type ApprovalService struct {
	persistence     *PersistenceService
	query           *QueryService
	permissions     *PermissionService
	flowExecutor    *FlowExecutor
	flowInstanceSvc *FlowInstanceService
}

// NewApprovalService creates a new ApprovalService
func NewApprovalService(
	p *PersistenceService,
	q *QueryService,
	perm *PermissionService,
	fe *FlowExecutor,
	fis *FlowInstanceService,
) *ApprovalService {
	return &ApprovalService{
		persistence:     p,
		query:           q,
		permissions:     perm,
		flowExecutor:    fe,
		flowInstanceSvc: fis,
	}
}

// CheckProcess checks if there is an active approval process for the object
func (s *ApprovalService) CheckProcess(ctx context.Context, objectAPIName string, user *models.UserSession) (models.SObject, error) {
	return s.findActiveProcess(ctx, objectAPIName, user)
}

// SubmitRequest represents the input for submitting a record
type SubmitRequest struct {
	ObjectAPIName string
	RecordID      string
	Comments      string
}

// Submit submits a record for approval
func (s *ApprovalService) Submit(ctx context.Context, req SubmitRequest, user *models.UserSession) (models.SObject, error) {
	// Find active approval process for this object
	process, err := s.findActiveProcess(ctx, req.ObjectAPIName, user)
	if err != nil || process == nil {
		return nil, errors.New("no active approval process found for this object")
	}

	// Security check: verify user has read access to the record
	if !s.permissions.CheckObjectPermissionWithUser(req.ObjectAPIName, constants.PermRead, user) {
		return nil, errors.New("you don't have permission to submit this record for approval")
	}

	// Check if already pending
	hasPending, err := s.hasPendingApproval(ctx, req.ObjectAPIName, req.RecordID, user)
	if err != nil {
		return nil, fmt.Errorf("failed to check for pending approvals: %w", err)
	}
	if hasPending {
		return nil, errors.New("record already has a pending approval")
	}

	// Determine approver (Business Logic)
	var approverID interface{}
	if process[constants.FieldSysApprovalProcess_ApproverType] == "Self" {
		approverID = user.ID
	} else {
		approverID = process[constants.FieldSysApprovalProcess_ApproverID]
	}

	// Create work item
	workItem := models.SObject{
		constants.FieldSysApprovalWorkItem_ProcessID:     process[constants.FieldID],
		constants.FieldSysApprovalWorkItem_ObjectAPIName: req.ObjectAPIName,
		constants.FieldSysApprovalWorkItem_RecordID:      req.RecordID,
		constants.FieldSysApprovalWorkItem_Status:        constants.ApprovalStatusPending,
		constants.FieldSysApprovalWorkItem_SubmittedByID: user.ID,
		constants.FieldSysApprovalWorkItem_ApproverID:    approverID,
		constants.FieldSysApprovalWorkItem_Comments:      req.Comments,
	}

	return s.persistence.Insert(ctx, constants.TableApprovalWorkItem, workItem, user)
}

// Approve approves a pending work item
func (s *ApprovalService) Approve(ctx context.Context, workItemID, comments string, user *models.UserSession) error {
	return s.processAction(ctx, workItemID, constants.ApprovalStatusApproved, comments, user)
}

// Reject rejects a pending work item
func (s *ApprovalService) Reject(ctx context.Context, workItemID, comments string, user *models.UserSession) error {
	return s.processAction(ctx, workItemID, constants.ApprovalStatusRejected, comments, user)
}

// GetPending returns pending approvals for the current user
func (s *ApprovalService) GetPending(ctx context.Context, user *models.UserSession) ([]models.SObject, error) {
	filterExpr := fmt.Sprintf("%s == '%s' && %s == '%s'", constants.FieldSysApprovalWorkItem_ApproverID, user.ID, constants.FieldSysApprovalWorkItem_Status, constants.ApprovalStatusPending)
	return s.query.QueryWithFilter(
		ctx,
		constants.TableApprovalWorkItem,
		filterExpr,
		user,
		constants.FieldCreatedDate, constants.SortDESC,
		100,
	)
}

// GetHistory returns history for a record
func (s *ApprovalService) GetHistory(ctx context.Context, objectAPIName, recordID string, user *models.UserSession) ([]models.SObject, error) {
	filterExpr := fmt.Sprintf("%s == '%s' && %s == '%s'", constants.FieldSysApprovalWorkItem_ObjectAPIName, objectAPIName, constants.FieldSysApprovalWorkItem_RecordID, recordID)
	return s.query.QueryWithFilter(
		ctx,
		constants.TableApprovalWorkItem,
		filterExpr,
		user,
		constants.FieldCreatedDate, constants.SortDESC,
		50,
	)
}

// GetFlowProgress returns the progress of a flow instance
func (s *ApprovalService) GetFlowProgress(ctx context.Context, instanceID string, user *models.UserSession) (*FlowInstanceProgress, error) {
	return s.flowInstanceSvc.GetInstanceProgress(ctx, instanceID, user)
}

// Private helpers

func (s *ApprovalService) processAction(ctx context.Context, workItemID, newStatus, comments string, user *models.UserSession) error {
	// Execute in transaction to ensure atomicity of update and flow resumption
	return s.persistence.RunInTransaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		// Fetch and validate work item
		item, err := s.getWorkItem(txCtx, workItemID, user)
		if err != nil {
			return errors.New("approval work item not found")
		}

		if item[constants.FieldSysApprovalWorkItem_Status] != constants.ApprovalStatusPending {
			return errors.New("work item is not pending")
		}

		// Verify user is authorized to act on this item
		if !s.isAuthorizedApprover(item, user) {
			action := "approve"
			if newStatus == constants.ApprovalStatusRejected {
				action = "reject"
			}
			return fmt.Errorf("you are not authorized to %s this item", action)
		}

		// Update work item
		updates := models.SObject{
			constants.FieldSysApprovalWorkItem_Status:       newStatus,
			constants.FieldSysApprovalWorkItem_ApprovedByID: user.ID,
			constants.FieldSysApprovalWorkItem_ApprovedDate: time.Now().UTC(),
			constants.FieldSysApprovalWorkItem_Comments:     comments,
		}

		// Update using txCtx which carries the transaction
		if err := s.persistence.Update(txCtx, constants.TableApprovalWorkItem, workItemID, updates, user); err != nil {
			return fmt.Errorf("failed to update approval: %w", err)
		}

		// Resume flow if needed - passed txCtx will ensure it participates in transaction
		s.resumeFlowIfNeeded(txCtx, item, newStatus == constants.ApprovalStatusApproved, user)

		return nil
	})
}

func (s *ApprovalService) findActiveProcess(ctx context.Context, objectAPIName string, user *models.UserSession) (models.SObject, error) {
	filterExpr := fmt.Sprintf("%s == '%s' && %s == true", constants.FieldSysApprovalProcess_ObjectAPIName, objectAPIName, constants.FieldSysApprovalProcess_IsActive)
	processes, err := s.query.QueryWithFilter(
		ctx,
		constants.TableApprovalProcess,
		filterExpr,
		user,
		"", "",
		1,
	)
	if err != nil || len(processes) == 0 {
		return nil, err
	}
	return processes[0], nil
}

func (s *ApprovalService) hasPendingApproval(ctx context.Context, objectAPIName, recordID string, user *models.UserSession) (bool, error) {
	filterExpr := fmt.Sprintf("%s == '%s' && %s == '%s' && %s == '%s'", constants.FieldSysApprovalWorkItem_ObjectAPIName, objectAPIName, constants.FieldSysApprovalWorkItem_RecordID, recordID, constants.FieldSysApprovalWorkItem_Status, constants.ApprovalStatusPending)
	existing, err := s.query.QueryWithFilter(
		ctx,
		constants.TableApprovalWorkItem,
		filterExpr,
		user,
		"", "",
		1,
	)
	if err != nil {
		return false, err
	}
	return len(existing) > 0, nil
}

func (s *ApprovalService) getWorkItem(ctx context.Context, workItemID string, user *models.UserSession) (models.SObject, error) {
	filterExpr := fmt.Sprintf("%s == '%s'", constants.FieldID, workItemID)
	items, err := s.query.QueryWithFilter(
		ctx,
		constants.TableApprovalWorkItem,
		filterExpr,
		user,
		"", "",
		1,
	)
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return items[0], nil
}

func (s *ApprovalService) isAuthorizedApprover(item models.SObject, user *models.UserSession) bool {
	approverID, ok := item[constants.FieldSysApprovalWorkItem_ApproverID].(string)
	if !ok || approverID == "" {
		return true // No specific approver set - anyone can approve
	}
	return approverID == user.ID
}

func (s *ApprovalService) resumeFlowIfNeeded(ctx context.Context, workItem models.SObject, approved bool, user *models.UserSession) {
	flowInstanceID, ok := workItem[constants.FieldSysApprovalWorkItem_FlowInstanceID].(string)
	if !ok || flowInstanceID == "" {
		return
	}

	flowStepID, ok := workItem[constants.FieldSysApprovalWorkItem_FlowStepID].(string)
	if !ok || flowStepID == "" {
		return
	}

	stepExecutor := func(ctx context.Context, instance *models.FlowInstance, step *models.FlowStep, record models.SObject, user *models.UserSession) error {
		payload := RecordEventPayload{
			ObjectAPIName: instance.ObjectAPIName,
			Record:        record,
			CurrentUser:   user,
		}
		return s.flowExecutor.ExecuteInstanceStep(ctx, instance, step, payload)
	}

	if err := s.flowInstanceSvc.ResumeAfterApproval(ctx, flowInstanceID, flowStepID, approved, stepExecutor, user); err != nil {
		log.Printf("⚠️ Failed to resume flow after approval: %v", err)
	}
}
