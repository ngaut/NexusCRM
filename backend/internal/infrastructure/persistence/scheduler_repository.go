package persistence

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
)

// SchedulerRepository handles direct database operations for the SchedulerService
// specifically targeting the _System_Flow table for execution locking and status updates.
type SchedulerRepository struct {
	db *sql.DB
}

// NewSchedulerRepository creates a new SchedulerRepository
func NewSchedulerRepository(db *sql.DB) *SchedulerRepository {
	return &SchedulerRepository{
		db: db,
	}
}

// AcquireExecutionLock atomically sets is_running = true if not already running
func (r *SchedulerRepository) AcquireExecutionLock(flowID string) (bool, error) {
	query := fmt.Sprintf(`
		%s %s 
		%s %s = %s 
		%s %s = ? %s (%s = %s %s %s IS %s)
	`, KeywordUpdate, constants.TableFlow,
		KeywordSet, constants.FieldSysFlow_IsRunning, KeywordTrue,
		KeywordWhere, constants.FieldID, KeywordAnd, constants.FieldSysFlow_IsRunning, KeywordFalse, KeywordOr, constants.FieldSysFlow_IsRunning, KeywordNull)

	result, err := r.db.Exec(query, flowID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// ReleaseExecutionLock sets is_running = false
func (r *SchedulerRepository) ReleaseExecutionLock(flowID string) error {
	query := fmt.Sprintf(`%s %s %s %s = %s %s %s = ?`,
		KeywordUpdate, constants.TableFlow, KeywordSet, constants.FieldSysFlow_IsRunning, KeywordFalse, KeywordWhere, constants.FieldID)
	_, err := r.db.Exec(query, flowID)
	return err
}

// UpdateFlowRunStatus updates last_run_at
func (r *SchedulerRepository) UpdateFlowRunStatus(flowID string) error {
	now := time.Now().UTC()
	query := fmt.Sprintf(`%s %s %s %s = ? %s %s = ?`,
		KeywordUpdate, constants.TableFlow, KeywordSet, constants.FieldSysFlow_LastRunAt, KeywordWhere, constants.FieldID)
	_, err := r.db.Exec(query, now, flowID)
	return err
}

// UpdateNextRunAt updates next_run_at
func (r *SchedulerRepository) UpdateNextRunAt(flowID string, nextRun time.Time) error {
	query := fmt.Sprintf(`%s %s %s %s = ? %s %s = ?`,
		KeywordUpdate, constants.TableFlow, KeywordSet, constants.FieldSysFlow_NextRunAt, KeywordWhere, constants.FieldID)
	_, err := r.db.Exec(query, nextRun, flowID)
	return err
}
