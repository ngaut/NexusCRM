package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nexuscrm/backend/pkg/query"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// RecordRepository handles dynamic CRUD operations for any object.
// It abstracts the SQL generation and execution, allowing services to work with high-level SObjects.
type RecordRepository struct {
	db *sql.DB
}

// NewRecordRepository creates a new RecordRepository
func NewRecordRepository(db *sql.DB) *RecordRepository {
	return &RecordRepository{db: db}
}

// GetExecutor returns the transaction if present, or the DB connection
func (r *RecordRepository) GetExecutor(tx *sql.Tx) Executor {
	if tx != nil {
		return tx
	}
	return r.db
}

// Exists checks if a record exists by ID
func (r *RecordRepository) Exists(ctx context.Context, tx *sql.Tx, tableName string, id string) (bool, error) {
	queryP := query.From(tableName).
		Select([]string{constants.FieldID}).
		Where(constants.FieldID+" = ?", id).
		Limit(1).
		Build()

	exec := r.GetExecutor(tx)
	rows, err := exec.QueryContext(ctx, queryP.SQL, queryP.Params...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// GetLock retrieves a record by ID and locks it (SELECT ... FOR UPDATE)
func (r *RecordRepository) GetLock(ctx context.Context, tx *sql.Tx, tableName string, id string) (models.SObject, error) {
	// Must be in a transaction for locking to work effectively
	if tx == nil {
		return nil, fmt.Errorf("transaction required for locking record %s", id)
	}

	q := query.From(tableName).
		Select([]string{"*"}).
		Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
		Limit(1).
		Build()

	// Inject FOR UPDATE
	q.SQL += " FOR UPDATE"

	rows, err := tx.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := query.ScanRowsToSObjects(rows)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results[0], nil
}

// Insert executes an INSERT statement
func (r *RecordRepository) Insert(ctx context.Context, tx *sql.Tx, tableName string, record models.SObject) error {
	q := query.Insert(tableName, record).Build()

	exec := r.GetExecutor(tx)
	_, err := exec.ExecContext(ctx, q.SQL, q.Params...)
	return err
}

// BulkInsert executes a multi-row INSERT statement for batch insertions
// Records are inserted in batches of batchSize to avoid query size limits
func (r *RecordRepository) BulkInsert(ctx context.Context, tx *sql.Tx, tableName string, records []models.SObject, columns []string, batchSize int) error {
	if len(records) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	exec := r.GetExecutor(tx)

	// Process in batches
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		// Convert to []map[string]interface{} for query builder
		batchMaps := make([]map[string]interface{}, len(batch))
		for j, rec := range batch {
			batchMaps[j] = rec
		}

		sql, params := query.BulkInsertOrdered(tableName, columns, batchMaps)
		if sql == "" {
			continue
		}

		if _, err := exec.ExecContext(ctx, sql, params...); err != nil {
			return fmt.Errorf("bulk insert batch %d-%d failed: %w", i, end, err)
		}
	}

	return nil
}

// Update executes an UPDATE statement
func (r *RecordRepository) Update(ctx context.Context, tx *sql.Tx, tableName string, id string, updates models.SObject) error {
	builder := query.Update(tableName).
		Set(updates).
		Where(fmt.Sprintf("%s = ?", constants.FieldID), id)

	q := builder.Build()

	exec := r.GetExecutor(tx)
	_, err := exec.ExecContext(ctx, q.SQL, q.Params...)
	return err
}

// Delete executes a DELETE statement
func (r *RecordRepository) Delete(ctx context.Context, tx *sql.Tx, tableName string, id string) error {
	q := query.Delete(tableName).
		Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
		Build()

	exec := r.GetExecutor(tx)
	_, err := exec.ExecContext(ctx, q.SQL, q.Params...)
	return err
}

// GetChildren retrieves active child records for cascade delete
func (r *RecordRepository) GetChildren(ctx context.Context, tx *sql.Tx, childTable string, foreignKey string, parentID string) ([]models.SObject, error) {
	q := query.From(childTable).
		Select([]string{constants.FieldID}).
		Where(fmt.Sprintf("%s = ?", foreignKey), parentID).
		ExcludeDeleted(). // Generic "Active" check
		Build()

	exec := r.GetExecutor(tx)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return query.ScanRowsToSObjects(rows)
}

// ExistsByField checks if any record exists matching a specific field value (useful for Restrict delete rules)
func (r *RecordRepository) ExistsByField(ctx context.Context, tx *sql.Tx, tableName string, fieldName string, value interface{}) (bool, error) {
	q := query.From(tableName).
		Select([]string{constants.FieldID}).
		Where(fmt.Sprintf("%s = ?", fieldName), value).
		ExcludeDeleted().
		Limit(1).
		Build()

	exec := r.GetExecutor(tx)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// FindOne returns a single record by ID, excluding deleted
func (r *RecordRepository) FindOne(ctx context.Context, tx *sql.Tx, tableName string, id string) (models.SObject, error) {
	q := query.From(tableName).
		Select([]string{"*"}).
		Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
		ExcludeDeleted().
		Limit(1).
		Build()

	exec := r.GetExecutor(tx)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := query.ScanRowsToSObjects(rows)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// FindAny returns a record by ID, INCLUDING deleted ones
func (r *RecordRepository) FindAny(ctx context.Context, tx *sql.Tx, tableName string, id string) (models.SObject, error) {
	q := query.From(tableName).
		Select([]string{"*"}).
		Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
		Limit(1).
		Build()

	exec := r.GetExecutor(tx)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := query.ScanRowsToSObjects(rows)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// PhysicalDelete permanently removes a record
func (r *RecordRepository) PhysicalDelete(ctx context.Context, tx *sql.Tx, tableName string, id string) error {
	q := query.Delete(tableName).
		Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
		Build()

	exec := r.GetExecutor(tx)
	_, err := exec.ExecContext(ctx, q.SQL, q.Params...)
	return err
}

// DeleteByField deletes records matching a specific field value
func (r *RecordRepository) DeleteByField(ctx context.Context, tx *sql.Tx, tableName string, fieldName string, value interface{}) error {
	q := query.Delete(tableName).
		Where(fmt.Sprintf("%s = ?", fieldName), value).
		Build()

	exec := r.GetExecutor(tx)
	_, err := exec.ExecContext(ctx, q.SQL, q.Params...)
	return err
}

// CheckUniqueness checks if a value exists for a field, excluding a specific ID
func (r *RecordRepository) CheckUniqueness(ctx context.Context, tableName string, fieldName string, value interface{}, excludeID string) (bool, error) {
	builder := query.From(tableName).
		Select([]string{constants.FieldID}).
		Where(fmt.Sprintf("%s = ?", fieldName), value).
		Limit(1)

	if excludeID != "" {
		builder.Where(fmt.Sprintf("%s != ?", constants.FieldID), excludeID)
	}

	q := builder.Build()

	exec := r.GetExecutor(nil)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// FindRecycleBinItemsByUser retrieves recycle bin items deleted by a specific user
func (r *RecordRepository) FindRecycleBinItemsByUser(ctx context.Context, username string) ([]models.SObject, error) {
	q := query.From(constants.TableRecycleBin).
		Select([]string{"*"}).
		Where(constants.FieldDeletedBy+" = ?", username).
		OrderBy(constants.FieldSysRecycleBin_DeletedDate, constants.SortDESC).
		Build()

	exec := r.GetExecutor(nil)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return query.ScanRowsToSObjects(rows)
}

// FindAllRecycleBinItems retrieves all recycle bin items (Admin only typically)
func (r *RecordRepository) FindAllRecycleBinItems(ctx context.Context) ([]models.SObject, error) {
	q := query.From(constants.TableRecycleBin).
		Select([]string{"*"}).
		OrderBy(constants.FieldSysRecycleBin_DeletedDate, constants.SortDESC).
		Build()

	exec := r.GetExecutor(nil)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return query.ScanRowsToSObjects(rows)
}

// FindRecycleBinEntry finds a recycle bin entry by original Record ID
func (r *RecordRepository) FindRecycleBinEntry(ctx context.Context, recordID string) (models.SObject, error) {
	q := query.From(constants.TableRecycleBin).
		Select([]string{"*"}).
		Where(constants.FieldRecordID+" = ?", recordID).
		Limit(1).
		Build()

	exec := r.GetExecutor(nil)
	rows, err := exec.QueryContext(ctx, q.SQL, q.Params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := query.ScanRowsToSObjects(rows)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}
