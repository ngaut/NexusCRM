package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/models"
)

// MetadataService manages CRM metadata
type MetadataService struct {
	schemaMgr *SchemaManager
	repo      *persistence.MetadataRepository
	mu        sync.RWMutex

	// Cache - Schemas
	schemas   []*models.ObjectMetadata
	schemaMap map[string]*models.ObjectMetadata
	fieldMap  map[string][]models.FieldMetadata // key: ObjectAPIName (lowercase)

	// Cache - Flows
	flows   []*models.Flow
	flowMap map[string]*models.Flow // key: flow ID

	// Cache - Per-Object Metadata
	validationRulesMap map[string][]*models.ValidationRule // key: objectAPIName (lowercase)
	autoNumbersMap     map[string][]*models.AutoNumber     // key: objectAPIName (lowercase)

	// Dependencies
	validationSvc *ValidationService
}

// NewMetadataService creates a new MetadataService
func NewMetadataService(repo *persistence.MetadataRepository, schemaMgr *SchemaManager) *MetadataService {
	return &MetadataService{
		schemaMgr: schemaMgr,
		repo:      repo,
	}
}

// SetValidationService sets the validation service dependency
func (ms *MetadataService) SetValidationService(vs *ValidationService) {
	ms.validationSvc = vs
}

// RefreshCache reloads all metadata from the database
func (ms *MetadataService) RefreshCache() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.refreshCacheLocked()
}

// refreshCacheLocked reloads metadata assuming the write lock is already held
func (ms *MetadataService) refreshCacheLocked() error {
	log.Println("🔄 Refreshing metadata cache...")
	ctx := context.Background()

	// 1. Load all schemas
	// Critical: If this fails, we cannot proceed
	schemas, err := ms.repo.GetAllSchemas(ctx)
	if err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	// 2. Build schema maps
	schemaMap := make(map[string]*models.ObjectMetadata)
	fieldMap := make(map[string][]models.FieldMetadata)

	for _, schema := range schemas {
		key := strings.ToLower(schema.APIName)
		schemaMap[key] = schema
		fieldMap[key] = schema.Fields
	}

	// 3. Load all flows
	// Critical: If this fails, returning empty flows would silently disable all automation.
	// We must fail the refresh instead.
	flows, err := ms.repo.GetAllFlows(ctx)
	if err != nil {
		return fmt.Errorf("failed to load flows: %w", err)
	}

	flowMap := make(map[string]*models.Flow)
	for _, flow := range flows {
		flowMap[flow.ID] = flow
	}

	// 4. Load validation rules and auto-numbers per object
	// Non-Critical: If these fail for specific objects, we can log and continue (partial degrade),
	// or strict fail. Given strict review, let's prefer partial degrade only for specific objects,
	// but generally try to load everything.
	validationRulesMap := make(map[string][]*models.ValidationRule)
	autoNumbersMap := make(map[string][]*models.AutoNumber)

	for _, schema := range schemas {
		key := strings.ToLower(schema.APIName)

		// Validation Rules (skip system tables)
		if !isSystemTableForCaching(schema.APIName) {
			rules, err := ms.repo.GetValidationRules(ctx, schema.APIName)
			if err != nil {
				// Log but don't fail entire cache?
				// If we fail here, one bad table breaks entire system.
				// Better to log error and treat as empty rules for that table.
				log.Printf("⚠️ Failed to load validation rules for %s: %v", schema.APIName, err)
			} else {
				validationRulesMap[key] = rules
			}
		}

		// Auto-Numbers
		ans, err := ms.repo.GetAutoNumbers(ctx, schema.APIName)
		if err != nil {
			log.Printf("⚠️ Failed to load auto-numbers for %s: %v", schema.APIName, err)
		} else {
			autoNumbersMap[key] = ans
		}
	}

	// 5. ATOMIC SWAP
	// Only update state after all critical data is loaded successfully
	ms.schemas = schemas
	ms.schemaMap = schemaMap
	ms.fieldMap = fieldMap
	ms.flows = flows
	ms.flowMap = flowMap
	ms.validationRulesMap = validationRulesMap
	ms.autoNumbersMap = autoNumbersMap

	log.Printf("✅ Metadata cache refreshed: %d objects, %d flows loaded", len(schemas), len(flows))
	return nil
}

// isSystemTableForCaching checks if a table is a system table (for caching optimization)
func isSystemTableForCaching(apiName string) bool {
	return strings.HasPrefix(apiName, "_System_")
}

// ensureCacheInitialized ensures that metadata is loaded (Double-Checked Locking)
func (ms *MetadataService) ensureCacheInitialized() error {
	// 1. Fast path: Read Lock
	ms.mu.RLock()
	loaded := ms.schemas != nil
	ms.mu.RUnlock()

	if loaded {
		return nil
	}

	// 2. Slow path: Write Lock
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Double check
	if ms.schemas != nil {
		return nil
	}

	return ms.refreshCacheLocked()
}

// Getter methods

func (ms *MetadataService) GetSchema(ctx context.Context, apiName string) *models.ObjectMetadata {
	// Ensure cache is loaded
	if err := ms.ensureCacheInitialized(); err != nil {
		log.Printf("⚠️ Failed to initialize cache in GetSchema: %v", err)
		return nil
	}

	normalizedName := strings.ToLower(apiName)

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.schemaMap != nil {
		if schema, ok := ms.schemaMap[normalizedName]; ok {
			return schema
		}
	}

	return nil
}

// InvalidateCache clears the cache, forcing a refresh on next read (Thread-safe)
func (ms *MetadataService) InvalidateCache() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.invalidateCacheLocked()
}

// invalidateCacheLocked clears the cache (Caller must hold lock)
func (ms *MetadataService) invalidateCacheLocked() {
	ms.schemas = nil
	ms.schemaMap = nil
	ms.fieldMap = nil
	ms.flows = nil
	ms.flowMap = nil
	ms.validationRulesMap = nil
	ms.autoNumbersMap = nil
	log.Println("🗑️ Metadata cache invalidated")
}

// GetSchemaOrError returns the schema or a NotFoundError if not found
func (ms *MetadataService) GetSchemaOrError(ctx context.Context, apiName string) (*models.ObjectMetadata, error) {
	schema := ms.GetSchema(ctx, apiName)
	if schema == nil {
		return nil, errors.NewNotFoundError("Object Metadata", apiName)
	}
	return schema, nil
}

func (ms *MetadataService) GetField(objectAPIName, fieldAPIName string) *models.FieldMetadata {
	// Ensure cache is loaded
	if err := ms.ensureCacheInitialized(); err != nil {
		log.Printf("⚠️ Failed to initialize cache in GetField: %v", err)
		return nil
	}

	normalizedObj := strings.ToLower(objectAPIName)

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.fieldMap != nil {
		if fields, ok := ms.fieldMap[normalizedObj]; ok {
			for _, field := range fields {
				if field.APIName == fieldAPIName {
					result := field // Copy to be safe
					return &result
				}
			}
		}
	}

	return nil
}

func (ms *MetadataService) GetSchemas(ctx context.Context) []*models.ObjectMetadata {
	// Ensure cache is loaded
	if err := ms.ensureCacheInitialized(); err != nil {
		log.Printf("⚠️ Failed to initialize cache in GetSchemas: %v", err)
		return []*models.ObjectMetadata{}
	}

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.schemas != nil {
		return ms.schemas
	}

	return []*models.ObjectMetadata{}
}
