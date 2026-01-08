package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// SchedulerService manages scheduled flow execution
type SchedulerService struct {
	repo         *persistence.SchedulerRepository
	metadata     *MetadataService
	flowExecutor *FlowExecutor
	stopChan     chan struct{}
	wg           sync.WaitGroup
	mu           sync.Mutex
	running      bool
	stopped      bool // Prevents double-close of stopChan
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(repo *persistence.SchedulerRepository, metadata *MetadataService, flowExecutor *FlowExecutor) *SchedulerService {
	return &SchedulerService{
		repo:         repo,
		metadata:     metadata,
		flowExecutor: flowExecutor,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the scheduler background loop
func (s *SchedulerService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	log.Println("⏰ Scheduler service starting...")

	ticker := time.NewTicker(time.Duration(constants.ScheduleCheckInterval) * time.Second)
	defer ticker.Stop()

	// Run immediately on start
	s.runPendingJobs()

	for {
		select {
		case <-ticker.C:
			s.runPendingJobs()
		case <-s.stopChan:
			log.Println("⏰ Scheduler service stopping...")
			s.wg.Wait() // Wait for running jobs to complete
			log.Println("⏰ Scheduler service stopped")
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	if !s.running || s.stopped {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.stopped = true
	s.mu.Unlock()

	close(s.stopChan)
}

// runPendingJobs finds and executes all due scheduled flows
func (s *SchedulerService) runPendingJobs() {
	flows := s.metadata.GetScheduledFlows(context.Background())

	now := time.Now().UTC()
	for _, flow := range flows {
		// Skip if not active
		if flow.Status != constants.FlowStatusActive {
			continue
		}

		// Skip if no schedule defined
		if flow.Schedule == nil || *flow.Schedule == "" {
			continue
		}

		// Check if flow is due
		if !s.isFlowDue(&flow, now) {
			continue
		}

		// Execute in goroutine
		s.wg.Add(1)
		go func(f models.Flow) {
			defer s.wg.Done()
			s.executeScheduledFlow(&f)
		}(flow)
	}
}

// isFlowDue checks if a scheduled flow should run now
func (s *SchedulerService) isFlowDue(flow *models.Flow, now time.Time) bool {
	// If NextRunAt is set and is in the past or equal to now, it's due
	if flow.NextRunAt != nil && !now.Before(*flow.NextRunAt) {
		return true
	}

	// If no NextRunAt and has never run (LastRunAt is nil), this is a new schedule - run immediately
	// This ensures we don't keep running on every check when NextRunAt hasn't been set yet
	if flow.NextRunAt == nil && flow.LastRunAt == nil && flow.Schedule != nil {
		return true
	}

	return false
}

// executeScheduledFlow runs a single scheduled flow with safety guards
func (s *SchedulerService) executeScheduledFlow(flow *models.Flow) {
	flowID := flow.ID
	log.Printf("⏰ Starting scheduled flow: %s (%s)", flow.Name, flowID)

	// 1. Atomically acquire execution lock
	acquired, err := s.repo.AcquireExecutionLock(flowID)
	if err != nil {
		log.Printf("⚠️ Failed to acquire lock for flow %s: %v", flowID, err)
		return
	}
	if !acquired {
		log.Printf("⏭️ Flow %s is already running, skipping", flow.Name)
		return
	}

	// 2. Ensure cleanup on exit (panic recovery)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🔥 Panic in scheduled flow %s: %v", flow.Name, r)
		}
		_ = s.repo.ReleaseExecutionLock(flowID)
	}()

	// 3. Create timeout context
	timeout := time.Duration(constants.ScheduleMaxRuntimeMins) * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 4. Execute the flow
	startTime := time.Now()
	execErr := s.executeFlowLogic(ctx, flow)
	duration := time.Since(startTime)

	// 5. Update execution status
	if execErr != nil {
		log.Printf("❌ Scheduled flow %s failed after %v: %v", flow.Name, duration, execErr)
		// Logging failure with error message, keeping LastRunAt update
		_ = s.repo.UpdateFlowRunStatus(flowID)
		s.logFlowExecution(flowID, false, execErr.Error())
	} else {
		log.Printf("✅ Scheduled flow %s completed in %v", flow.Name, duration)
		_ = s.repo.UpdateFlowRunStatus(flowID)
	}

	// 6. Calculate and set next run time
	s.scheduleNextRun(flow)
}

// executeFlowLogic performs the actual flow execution
func (s *SchedulerService) executeFlowLogic(ctx context.Context, flow *models.Flow) error {
	// For scheduled flows, we execute the action directly with a system context
	// Since scheduled flows typically don't have a triggering record, we pass an empty record

	// Create a system user session for scheduled execution
	// Using a fixed system user ID for scheduled tasks
	systemUser := &models.UserSession{
		ID:        "00000000-0000-0000-0000-000000000000", // System scheduler user
		Name:      "System Scheduler",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// Handle multistep flows differently - they need FlowExecutor
	if flow.FlowType == constants.FlowTypeMultistep {
		// For multistep flows, we need to create a flow instance and execute via FlowExecutor
		// This is a simplified execution - multistep scheduled flows may need more context
		log.Printf("⚠️ Multistep scheduled flows not fully supported yet, using simple execution for flow %s", flow.Name)
	}

	// For simple flows, execute the action directly via ActionService
	action := &models.ActionMetadata{
		ID:            flow.ID,
		ObjectAPIName: flow.TriggerObject,
		Type:          flow.ActionType,
		Config:        flow.ActionConfig,
	}

	return s.flowExecutor.actionSvc.ExecuteActionDirect(ctx, action, models.SObject{}, systemUser)
}

// scheduleNextRun calculates and sets the next run time
func (s *SchedulerService) scheduleNextRun(flow *models.Flow) {
	if flow.Schedule == nil || *flow.Schedule == "" {
		return
	}

	nextRun, err := s.calculateNextRun(*flow.Schedule, flow.ScheduleTimezone)
	if err != nil {
		log.Printf("⚠️ Failed to calculate next run for flow %s: %v", flow.Name, err)
		return
	}

	if err := s.repo.UpdateNextRunAt(flow.ID, nextRun); err != nil {
		log.Printf("⚠️ Failed to update next_run_at for flow %s: %v", flow.Name, err)
	}
}

// calculateNextRun parses cron expression and returns next execution time
func (s *SchedulerService) calculateNextRun(cronExpr string, timezone *string) (time.Time, error) {
	// Determine location
	loc := time.UTC
	if timezone != nil && *timezone != "" && *timezone != constants.ScheduleDefaultTimezone {
		var err error
		loc, err = time.LoadLocation(*timezone)
		if err != nil {
			log.Printf("⚠️ Invalid timezone %s, falling back to UTC", *timezone)
			loc = time.UTC
		}
	}

	// Parse cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Calculate next run time
	now := time.Now().In(loc)
	nextRun := schedule.Next(now)

	return nextRun.UTC(), nil
}

// logFlowExecution logs a flow execution event (stub - can be expanded)
func (s *SchedulerService) logFlowExecution(flowID string, success bool, errMsg string) {
	// This could write to _System_Flow_Log table
	// For now, just log to stdout
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	log.Printf("📝 Flow Execution Log: flowID=%s status=%s error=%s", flowID, status, errMsg)
}
