package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// SessionRepository handles database operations for user sessions
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// InsertSession creates a new session in the database
func (r *SessionRepository) InsertSession(ctx context.Context, session *models.SystemSession) error {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysSession_UserID, constants.FieldSysSession_Token,
		constants.FieldSysSession_ExpiresAt, constants.FieldSysSession_IPAddress, constants.FieldSysSession_UserAgent,
		constants.FieldSysSession_IsRevoked, constants.FieldSysSession_LastActivity,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`, constants.TableSession, cols)

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.Token,
		session.ExpiresAt,
		session.IPAddress,
		session.UserAgent,
		session.IsRevoked,
		session.LastActivity,
	)
	return err
}

// GetSession retrieves a session by its ID (from JWT claim)
func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (*models.SystemSession, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysSession_UserID, constants.FieldSysSession_Token,
		constants.FieldSysSession_ExpiresAt, constants.FieldSysSession_IPAddress, constants.FieldSysSession_UserAgent,
		constants.FieldSysSession_IsRevoked, constants.FieldSysSession_LastActivity,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s FROM %s 
		WHERE %s = ? LIMIT 1`,
		cols, constants.TableSession, constants.FieldID)

	var s models.SystemSession
	var createdDateRaw, lastModifiedDateRaw, lastActivityRaw []byte

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&s.ID,
		&s.UserID,
		&s.Token,
		&s.ExpiresAt,
		&s.IPAddress,
		&s.UserAgent,
		&s.IsRevoked,
		&lastActivityRaw,
		&createdDateRaw,
		&lastModifiedDateRaw,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse timestamps
	s.LastActivity = parseTime(lastActivityRaw)
	s.CreatedDate = parseTime(createdDateRaw)
	s.LastModifiedDate = parseTime(lastModifiedDateRaw)

	return &s, nil
}

// RevokeSession marks a session as revoked
func (r *SessionRepository) RevokeSession(ctx context.Context, sessionID string) error {
	query := fmt.Sprintf("UPDATE %s SET %s = 1, %s = NOW() WHERE %s = ?",
		constants.TableSession, constants.FieldSysSession_IsRevoked, constants.FieldLastModifiedDate, constants.FieldID)
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}

// UpdateLastActivity updates the last activity timestamp
func (r *SessionRepository) UpdateLastActivity(ctx context.Context, sessionID string) error {
	query := fmt.Sprintf("UPDATE %s SET %s = NOW() WHERE %s = ?",
		constants.TableSession, constants.FieldSysSession_LastActivity, constants.FieldID)
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}
