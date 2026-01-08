package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/interfaces/rest"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockApprovalService is a mock implementation of the ApprovalService
type MockApprovalService struct {
	mock.Mock
}

func (m *MockApprovalService) Submit(ctx context.Context, req services.SubmitRequest, user *models.UserSession) (models.SObject, error) {
	args := m.Called(ctx, req, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.SObject), args.Error(1)
}

func (m *MockApprovalService) Approve(ctx context.Context, workItemID, comments string, user *models.UserSession) error {
	args := m.Called(ctx, workItemID, comments, user)
	return args.Error(0)
}

func (m *MockApprovalService) Reject(ctx context.Context, workItemID, comments string, user *models.UserSession) error {
	args := m.Called(ctx, workItemID, comments, user)
	return args.Error(0)
}

func (m *MockApprovalService) GetPending(ctx context.Context, user *models.UserSession) ([]models.SObject, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SObject), args.Error(1)
}

func (m *MockApprovalService) GetHistory(ctx context.Context, objectAPIName, recordID string, user *models.UserSession) ([]models.SObject, error) {
	args := m.Called(ctx, objectAPIName, recordID, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SObject), args.Error(1)
}

func (m *MockApprovalService) CheckProcess(ctx context.Context, objectAPIName string, user *models.UserSession) (models.SObject, error) {
	args := m.Called(ctx, objectAPIName, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.SObject), args.Error(1)
}

func (m *MockApprovalService) GetFlowProgress(ctx context.Context, instanceID string, user *models.UserSession) (*services.FlowInstanceProgress, error) {
	args := m.Called(ctx, instanceID, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.FlowInstanceProgress), args.Error(1)
}
func TestApprovalHandler_Submit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockApprovalService)
	handler := rest.NewApprovalHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Mock context user (must be auth.UserSession)
		authUser := auth.UserSession{
			ID:        "user123",
			Name:      "Test User",
			Email:     "test@example.com",
			ProfileId: "Standard User",
		}
		c.Set(constants.ContextKeyUser, authUser)

		// Expected model user (converted by helper)
		modelUser := &models.UserSession{
			ID:            "user123",
			Name:          "Test User",
			Email:         services.StringPtr("test@example.com"),
			ProfileID:     "Standard User",
			IsSystemAdmin: false,
		}

		reqBody := rest.SubmitRequest{
			ObjectAPIName: "Opportunity",
			RecordID:      "rec123",
			Comments:      "Please approve",
		}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/approvals/submit", bytes.NewBuffer(jsonBytes))

		serviceReq := services.SubmitRequest{
			ObjectAPIName: "Opportunity",
			RecordID:      "rec123",
			Comments:      "Please approve",
		}
		expectedResp := models.SObject{constants.FieldID: "req1"}

		// Use mock.Anything for context as gin context wraps request context
		// Expect modelUser (converted from authUser)
		mockService.On("Submit", mock.Anything, serviceReq, modelUser).Return(expectedResp, nil).Once()

		handler.Submit(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Permission Denied", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		authUser := auth.UserSession{ID: "user1", ProfileId: "Standard User"}
		c.Set(constants.ContextKeyUser, authUser)

		reqBody := rest.SubmitRequest{ObjectAPIName: "Account", RecordID: "rec1"}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/approvals/submit", bytes.NewBuffer(jsonBytes))

		serviceReq := services.SubmitRequest{ObjectAPIName: "Account", RecordID: "rec1"}

		// Simulate permission error
		mockService.On("Submit", mock.Anything, serviceReq, mock.Anything).Return(nil, errors.New("you don't have permission to submit this record for approval")).Once()

		handler.Submit(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("No Active Process", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		authUser := auth.UserSession{ID: "user1"}
		c.Set(constants.ContextKeyUser, authUser)

		reqBody := rest.SubmitRequest{ObjectAPIName: "Custom", RecordID: "rec1"}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/approvals/submit", bytes.NewBuffer(jsonBytes))

		serviceReq := services.SubmitRequest{ObjectAPIName: "Custom", RecordID: "rec1"}

		mockService.On("Submit", mock.Anything, serviceReq, mock.Anything).Return(nil, errors.New("no active approval process found for this object")).Once()

		handler.Submit(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Already Pending", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		authUser := auth.UserSession{ID: "user1"}
		c.Set(constants.ContextKeyUser, authUser)

		reqBody := rest.SubmitRequest{ObjectAPIName: "Ticket", RecordID: "rec1"}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/approvals/submit", bytes.NewBuffer(jsonBytes))

		serviceReq := services.SubmitRequest{ObjectAPIName: "Ticket", RecordID: "rec1"}

		mockService.On("Submit", mock.Anything, serviceReq, mock.Anything).Return(nil, errors.New("record already has a pending approval")).Once()

		handler.Submit(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Generic Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		authUser := auth.UserSession{ID: "user1"}
		c.Set(constants.ContextKeyUser, authUser)

		reqBody := rest.SubmitRequest{ObjectAPIName: "Ticket", RecordID: "rec1"}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/approvals/submit", bytes.NewBuffer(jsonBytes))

		serviceReq := services.SubmitRequest{ObjectAPIName: "Ticket", RecordID: "rec1"}

		mockService.On("Submit", mock.Anything, serviceReq, mock.Anything).Return(nil, errors.New("db disconnect")).Once()

		handler.Submit(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
