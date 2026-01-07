package services

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/models"
)

// MetadataService manages CRM metadata
type MetadataService struct {
	db        *database.TiDBConnection
	schemaMgr *SchemaManager
	repo      *persistence.MetadataRepository
	mu        sync.RWMutex

	// Cache
	schemas   []*models.ObjectMetadata
	schemaMap map[string]*models.ObjectMetadata
	fieldMap  map[string][]models.FieldMetadata // key: ObjectAPIName

	// Dependencies
	permissionSvc *PermissionService
	validationSvc *ValidationService
}

// NewMetadataService creates a new MetadataService
func NewMetadataService(db *database.TiDBConnection, schemaMgr *SchemaManager) *MetadataService {
	return &MetadataService{
		db:        db,
		schemaMgr: schemaMgr,
		repo:      persistence.NewMetadataRepository(db.DB()),
	}
}

// SetPermissionService sets the permission service dependency (break circular dependency)
func (ms *MetadataService) SetPermissionService(ps *PermissionService) {
	ms.permissionSvc = ps
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

	// 1. Load all schemas
	schemas, err := ms.repo.GetAllSchemas(context.Background())
	if err != nil {
		return err
	}

	// 2. Build maps
	schemaMap := make(map[string]*models.ObjectMetadata)
	fieldMap := make(map[string][]models.FieldMetadata)

	for _, schema := range schemas {
		// Normalize key
		key := strings.ToLower(schema.APIName)
		schemaMap[key] = schema

		// Map fields
		fieldMap[key] = schema.Fields
	}

	ms.schemas = schemas
	ms.schemaMap = schemaMap
	ms.fieldMap = fieldMap

	log.Printf("✅ Metadata cache refreshed: %d objects loaded", len(schemas))
	return nil
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
