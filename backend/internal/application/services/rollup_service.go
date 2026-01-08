package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/models"
)

// RollupService handles rollup summary field calculations
// RollupService handles rollup summary field calculations
type RollupService struct {
	repo      *persistence.RollupRepository
	metadata  *MetadataService
	txManager *persistence.TransactionManager
}

// NewRollupService creates a new RollupService
func NewRollupService(repo *persistence.RollupRepository, metadata *MetadataService, txManager *persistence.TransactionManager) *RollupService {
	return &RollupService{
		repo:      repo,
		metadata:  metadata,
		txManager: txManager,
	}
}

// ProcessRollups orchestration method.
// The tx parameter ensures rollup updates participate in the caller's transaction for ACID compliance.
// Pass nil for tx to execute outside a transaction (not recommended for data consistency).
func (rs *RollupService) ProcessRollups(ctx context.Context, tx *sql.Tx, childObjName string, childRecord models.SObject) error {
	affected, err := rs.FindAffectedRollups(ctx, childObjName, childRecord)
	if err != nil {
		return err
	}

	for _, item := range affected {
		newVal, err := rs.CalculateRollup(ctx, item, tx)
		if err != nil {
			return fmt.Errorf("failed to calculate rollup %s.%s: %w", item.ParentObjName, item.RollupField.APIName, err)
		}

		// Direct Update of Parent via Repository
		log.Printf("🔄 Updating Rollup %s.%s on %s = %v", item.ParentObjName, item.RollupField.APIName, item.ParentID, newVal)

		if err := rs.repo.UpdateParentRollup(ctx, tx, item.ParentObjName, item.ParentID, item.RollupField.APIName, newVal); err != nil {
			return fmt.Errorf("failed to update parent rollup %s: %w", item.ParentID, err)
		}
	}
	return nil
}

// AffectedRollup represents a rollup calculation that needs to be performed
type AffectedRollup struct {
	ParentObjName string
	ParentID      string
	RollupField   models.FieldMetadata
}

// FindAffectedRollups identifies parent records that need rollup recalculation
// based on a change to a child record.
func (rs *RollupService) FindAffectedRollups(ctx context.Context, childObjName string, childRecord models.SObject) ([]AffectedRollup, error) {
	var affected []AffectedRollup

	// iterate over all schemas to find objects that have a rollup summary pointing to this child object
	schemas := rs.metadata.GetSchemas(ctx)
	for _, parentSchema := range schemas {
		for _, field := range parentSchema.Fields {
			// Check if it's a Rollup Summary field
			if field.RollupConfig == nil {
				continue
			}

			// Check if it matches this child object
			if field.RollupConfig.SummaryObject != childObjName {
				continue
			}

			// Find the relationship field value on the child record
			relField := field.RollupConfig.RelationshipField
			if relField == "" {
				// Fallback: If there is only one lookup to Parent, use it?
				// For now, strict: must have relationship_field.
				continue
			}

			// Get Parent ID from Child Record
			parentIDVal, ok := childRecord[relField]
			if !ok || parentIDVal == nil {
				continue // No parent linked, nothing to roll up
			}

			parentID, ok := parentIDVal.(string)
			if !ok || parentID == "" {
				continue
			}

			affected = append(affected, AffectedRollup{
				ParentObjName: parentSchema.APIName,
				ParentID:      parentID,
				RollupField:   field,
			})
		}
	}

	return affected, nil
}

// CalculateRollup performs the aggregation query and returns the value
func (rs *RollupService) CalculateRollup(ctx context.Context, rollup AffectedRollup, tx *sql.Tx) (interface{}, error) {
	config := rollup.RollupField.RollupConfig

	// Default value for no rows is handled by Repo unless date/time special case
	// But actually Repo returns 0 or null.
	// We might need to handle empty result logic here or in Repo.
	// Repo handles SUM/AVG -> 0, MIN/MAX -> null.
	// For date fields, 0 is invalid, so let's defer to Repo's return.

	// Validate Filter Logic
	if config.Filter != nil && *config.Filter != "" {
		filter := *config.Filter
		// Validate filter contains only safe SQL expression characters
		if !isValidSQLFilter(filter) {
			return nil, fmt.Errorf("invalid filter expression: contains potentially unsafe characters")
		}
	}

	return rs.repo.CalculateRollup(
		ctx,
		tx,
		config.CalcType,
		config.SummaryObject,
		config.SummaryField,
		config.RelationshipField,
		rollup.ParentID,
		config.Filter,
	)
}

// isValidSQLFilter validates that a SQL filter expression contains only safe characters.
// It rejects expressions that could be used for SQL injection attacks.
// Allowed: field names, operators, literals, and common SQL keywords.
func isValidSQLFilter(filter string) bool {
	// Reject empty filters
	if filter == "" {
		return false
	}

	// Reject dangerous SQL keywords/patterns (case-insensitive check)
	lowerFilter := strings.ToLower(filter)
	dangerousPatterns := []string{
		";",                  // Statement terminator
		"--",                 // Comment
		"/*",                 // Block comment
		"drop ",              // DDL
		"delete ",            // DML (not in WHERE context)
		"insert ",            // DML
		"update ",            // DML
		"truncate ",          // DDL
		"alter ",             // DDL
		"create ",            // DDL
		"exec ",              // Execute
		"execute ",           // Execute
		"union ",             // UNION attacks
		"information_schema", // Schema probing
		"sleep(",             // Time-based attacks
		"benchmark(",         // Time-based attacks
		"load_file(",         // File access
		"into outfile",       // File write
		"into dumpfile",      // File write
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerFilter, pattern) {
			return false
		}
	}

	return true
}
