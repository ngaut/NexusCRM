package services

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/backend/internal/domain/ports"
	"github.com/nexuscrm/backend/pkg/auth"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetadataService for testing
type MockMetadataService struct {
	flows []*models.Flow
}

func (m *MockMetadataService) GetFlows() []*models.Flow {
	return m.flows
}

func (m *MockMetadataService) GetFlow(id string) *models.Flow {
	for _, f := range m.flows {
		if f.ID == id {
			return f
		}
	}
	return nil
}

func (m *MockMetadataService) GetSupportedEvents() []string {
	return []string{
		string(events.RecordBeforeCreate),
		string(events.RecordBeforeUpdate),
		string(events.RecordAfterCreate), // Added support
	}
}

// MockActionService
type MockActionService struct {
	mock.Mock
}

func (m *MockActionService) ExecuteAction(ctx context.Context, actionID string, contextRecord models.SObject, user *models.UserSession) error {
	args := m.Called(ctx, actionID, contextRecord, user)
	return args.Error(0)
}

func (m *MockActionService) ExecuteActionDirect(ctx context.Context, action *models.ActionMetadata, record models.SObject, user *models.UserSession) error {
	args := m.Called(ctx, action, record, user)
	return args.Error(0)
}

func (m *MockActionService) GetAllActions() ([]*models.ActionMetadata, error) {
	return nil, nil
}
func (m *MockActionService) GetAction(id string) (*models.ActionMetadata, error) {
	return nil, nil
}
func (m *MockActionService) CreateAction(action *models.ActionMetadata) error {
	return nil
}
func (m *MockActionService) UpdateAction(action *models.ActionMetadata) error {
	return nil
}
func (m *MockActionService) DeleteAction(id string) error {
	return nil
}
func (m *MockActionService) GetActionsForObject(object string) ([]*models.ActionMetadata, error) {
	return nil, nil
}

// MockEventBus
type MockEventBus struct {
	handlers map[events.EventType][]ports.EventHandler
	mu       sync.Mutex
}

func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		handlers: make(map[events.EventType][]ports.EventHandler),
	}
}

func (m *MockEventBus) Subscribe(eventType events.EventType, handler ports.EventHandler) func() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[eventType] = append(m.handlers[eventType], handler)
	return func() {}
}

func (m *MockEventBus) Publish(ctx context.Context, eventType events.EventType, payload interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handlers, ok := m.handlers[eventType]; ok {
		for _, h := range handlers {
			if err := h(ctx, payload); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MockEventBus) PublishAsync(eventType events.EventType, payload interface{}) {
	_ = m.Publish(context.Background(), eventType, payload)
}

// ----------------------------------------------------------------------------
// Tests
// ----------------------------------------------------------------------------

func TestFlowExecutor_PasswordHashing(t *testing.T) {
	mockMetadata := &MockMetadataService{
		flows: []*models.Flow{
			{
				ID:            "flow-test-hash",
				Name:          "Test_Password_Hash",
				Status:        constants.FlowStatusActive,
				TriggerObject: constants.TableUser,
				TriggerType:   "beforeCreate",
				ActionType:    "UpdateRecord",
				ActionConfig: map[string]interface{}{
					"field_mappings": map[string]interface{}{
						"password": "=BCRYPT(password)",
					},
				},
			},
		},
	}

	eventBus := NewMockEventBus()
	executor := NewFlowExecutor(mockMetadata, nil, eventBus, nil, nil)
	executor.RegisterFlowHandlers()

	t.Run("Should hash password on beforeCreate", func(t *testing.T) {
		rawPassword := "SecurePass123!"
		userRecord := models.SObject{
			"id":       "user-1",
			"name":     "Test User",
			"email":    "test@example.com",
			"password": rawPassword,
		}

		payload := RecordEventPayload{
			ObjectAPIName: constants.TableUser,
			Record:        userRecord,
			CurrentUser:   &models.UserSession{ID: "admin"},
		}

		err := eventBus.Publish(context.Background(), events.RecordBeforeCreate, payload)
		assert.NoError(t, err)

		newPassword := userRecord["password"].(string)
		assert.NotEqual(t, rawPassword, newPassword)
		assert.True(t, auth.VerifyPassword(rawPassword, newPassword))
		log.Printf("Verified: %s -> %s", rawPassword, newPassword)
	})
}

func TestFlowExecutor_HighPriorityAlert(t *testing.T) {
	mockMetadata := &MockMetadataService{
		flows: []*models.Flow{
			{
				ID:               "flow-alert",
				Name:             "High Priority Alert",
				Status:           constants.FlowStatusActive,
				TriggerObject:    "Ticket",
				TriggerType:      "afterCreate",
				TriggerCondition: `priority == "High"`,
				ActionType:       "createRecord",
				ActionConfig: map[string]interface{}{
					constants.ConfigTargetObject: "Ticket",
					constants.ConfigFieldMappings: map[string]interface{}{
						"name":     "URGENT: Alert",
						"priority": "Medium",
						"status":   "New",
					},
				},
			},
		},
	}

	mockActionSvc := new(MockActionService)
	// Match ActionMetadata
	mockActionSvc.On("ExecuteActionDirect",
		mock.Anything,
		mock.MatchedBy(func(action *models.ActionMetadata) bool {
			return action.Type == constants.ActionTypeCreateRecord &&
				*action.TargetObject == "Ticket"
		}),
		mock.Anything,
		mock.Anything,
	).Return(nil)

	eventBus := NewMockEventBus()
	executor := NewFlowExecutor(mockMetadata, mockActionSvc, eventBus, nil, nil)
	executor.RegisterFlowHandlers()

	// 1. Test with High Priority (Should Trigger)
	t.Run("Should Trigger Alert on High Priority", func(t *testing.T) {
		payload := RecordEventPayload{
			ObjectAPIName: "Ticket",
			Record: models.SObject{
				"id":       "ticket-1",
				"priority": "High",
				"name":     "Crash",
			},
			CurrentUser: &models.UserSession{ID: "admin"},
		}

		err := eventBus.Publish(context.Background(), events.RecordAfterCreate, payload)
		assert.NoError(t, err)

		mockActionSvc.AssertNumberOfCalls(t, "ExecuteActionDirect", 1)
	})

	// 2. Test with Low Priority (Should NOT Trigger)
	t.Run("Should NOT Trigger Alert on Low Priority", func(t *testing.T) {
		mockActionSvc.Calls = nil // Reset calls

		payload := RecordEventPayload{
			ObjectAPIName: "Ticket",
			Record: models.SObject{
				"id":       "ticket-2",
				"priority": "Low",
				"name":     "Minor bug",
			},
			CurrentUser: &models.UserSession{ID: "admin"},
		}

		err := eventBus.Publish(context.Background(), events.RecordAfterCreate, payload)
		assert.NoError(t, err)

		// Should NOT call action
		mockActionSvc.AssertNumberOfCalls(t, "ExecuteActionDirect", 0)
	})
}
