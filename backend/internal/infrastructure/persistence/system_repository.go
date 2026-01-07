package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

type SystemRepository struct {
	db *sql.DB
}

func NewSystemRepository(db *sql.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

// GetLogs retrieves system logs
func (r *SystemRepository) GetLogs(ctx context.Context, limit int) ([]*models.SystemLog, error) {
	if limit <= 0 {
		limit = 100
	}

	q := query.From(constants.TableLog).
		Select([]string{constants.FieldID, constants.FieldTimestamp, constants.FieldLevel, constants.FieldSource, constants.FieldMessage, constants.FieldDetails}).
		OrderBy(constants.FieldTimestamp, constants.SortDESC).
		Limit(limit).
		Build()

	rows, err := r.db.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]*models.SystemLog, 0)
	for rows.Next() {
		var log models.SystemLog
		var details *string

		if err := rows.Scan(&log.ID, &log.Timestamp, &log.Level, &log.Source, &log.Message, &details); err != nil {
			continue
		}

		log.Details = details
		logs = append(logs, &log)
	}

	return logs, nil
}

// GetRecentItems retrieves recently viewed items for a user
func (r *SystemRepository) GetRecentItems(ctx context.Context, userID string, limit int) ([]*models.RecentItem, error) {
	if userID == "" {
		return []*models.RecentItem{}, nil
	}

	if limit <= 0 {
		limit = 10
	}

	q := query.From(constants.TableRecent).
		Select([]string{constants.FieldID, constants.FieldUserID, constants.FieldObjectAPIName, constants.FieldRecordID, constants.FieldRecordName, constants.FieldTimestamp}).
		Where(constants.FieldUserID+" = ?", userID).
		OrderBy(constants.FieldTimestamp, constants.SortDESC).
		Limit(limit).
		Build()

	rows, err := r.db.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*models.RecentItem, 0)
	for rows.Next() {
		var item models.RecentItem

		if err := rows.Scan(&item.ID, &item.UserID, &item.ObjectAPIName, &item.RecordID, &item.RecordName, &item.Timestamp); err != nil {
			continue
		}

		items = append(items, &item)
	}

	return items, nil
}

// GetConfig retrieves a system configuration value
func (r *SystemRepository) GetConfig(ctx context.Context, key string) (*string, error) {
	q := query.From(constants.TableConfig).
		Select([]string{constants.FieldValue}).
		Where(constants.FieldKeyName+" = ?", key).
		Limit(1).
		Build()

	rows, err := r.db.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		return &value, nil
	}

	return nil, nil
}

// GetAllConfigs retrieves all system configurations
func (r *SystemRepository) GetAllConfigs(ctx context.Context) ([]*models.SystemConfig, error) {
	q := query.From(constants.TableConfig).
		Select([]string{constants.FieldKeyName, constants.FieldValue, constants.FieldIsSecret, constants.FieldDescription}).
		Build()

	rows, err := r.db.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make([]*models.SystemConfig, 0)
	for rows.Next() {
		var config models.SystemConfig
		var isSecret int // TiDB/MySQL boolean is tinyint
		var description *string

		if err := rows.Scan(&config.KeyName, &config.Value, &isSecret, &description); err != nil {
			continue
		}

		config.IsSecret = isSecret != 0
		config.Description = description

		configs = append(configs, &config)
	}

	return configs, nil
}

// CheckConfigExists checks if a config key exists and returns the key if found
func (r *SystemRepository) CheckConfigExists(ctx context.Context, key string) (bool, error) {
	var exists int
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", constants.TableConfig, constants.FieldKeyName)
	err := r.db.QueryRowContext(ctx, query, key).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// CheckProfileExistsByName checks existence by name and returns ID if found
func (r *SystemRepository) GetProfileIDByName(ctx context.Context, name string) (string, error) {
	var id string
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", constants.FieldID, constants.TableProfile, constants.FieldName)
	err := r.db.QueryRowContext(ctx, query, name).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil // Not found
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetUserIDByEmail checks existence by email and returns ID if found
func (r *SystemRepository) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
	var id string
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", constants.FieldID, constants.TableUser, constants.FieldEmail)
	err := r.db.QueryRowContext(ctx, query, email).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return id, nil
}

// BatchUpsertProfiles inserts multiple profiles using direct SQL
func (r *SystemRepository) BatchUpsertProfiles(ctx context.Context, profiles []models.Profile) error {
	if len(profiles) == 0 {
		return nil
	}

	values := []string{}
	args := []interface{}{}

	for _, p := range profiles {
		values = append(values, "(?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)")
		args = append(args, p.ID, p.Name, p.Description, p.IsActive, p.IsSystem)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (id, name, description, is_active, is_system, created_date, last_modified_date) VALUES %s ON DUPLICATE KEY UPDATE description=VALUES(description), is_active=VALUES(is_active), is_system=VALUES(is_system), last_modified_date=CURRENT_TIMESTAMP",
		constants.TableProfile,
		strings.Join(values, ","),
	)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("batch upsert profiles failed: %w", err)
	}
	return nil
}

// BatchUpsertUsers inserts multiple users using direct SQL
func (r *SystemRepository) BatchUpsertUsers(ctx context.Context, users []models.SystemUser) error {
	if len(users) == 0 {
		return nil
	}

	values := []string{}
	args := []interface{}{}

	for _, u := range users {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)")
		// ID, Username, Email, Password, FirstName, LastName, ProfileID, IsActive
		args = append(args, u.ID, u.Email, u.Email, u.Password, u.FirstName, u.LastName, u.ProfileID, true)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (id, username, email, password, first_name, last_name, profile_id, is_active, created_date, last_modified_date) 
		VALUES %s 
		ON DUPLICATE KEY UPDATE 
		username=VALUES(username), password=VALUES(password), first_name=VALUES(first_name), 
		last_name=VALUES(last_name), profile_id=VALUES(profile_id), is_active=VALUES(is_active), last_modified_date=CURRENT_TIMESTAMP`,
		constants.TableUser,
		strings.Join(values, ","),
	)

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("batch upsert users failed: %w", err)
	}
	return nil
}
