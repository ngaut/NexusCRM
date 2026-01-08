package persistence

import (
	"context"
	"database/sql"
	"fmt"

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
	query := fmt.Sprintf(`
		INSERT INTO %s (id, user_id, token, expires_at, ip_address, user_agent, is_revoked, last_activity, created_date, last_modified_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		constants.TableSession)

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
	query := fmt.Sprintf(`
		SELECT id, user_id, token, expires_at, ip_address, user_agent, is_revoked, last_activity, created_date, last_modified_date 
		FROM %s 
		WHERE id = ? LIMIT 1`,
		constants.TableSession)

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
	query := fmt.Sprintf("UPDATE %s SET is_revoked = 1, last_modified_date = NOW() WHERE id = ?", constants.TableSession)
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}

// UpdateLastActivity updates the last activity timestamp
func (r *SessionRepository) UpdateLastActivity(ctx context.Context, sessionID string) error {
	query := fmt.Sprintf("UPDATE %s SET last_activity = NOW() WHERE id = ?", constants.TableSession)
	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}
