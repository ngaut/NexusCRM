package bootstrap

import (
	"context"
	"log"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
)

// InitializeSchema creates core system tables using SchemaManager
// This replaces raw SQL execution with declarative table definitions
func InitializeSchema(db *database.TiDBConnection) error {
	log.Println("🔧 Initializing core system schema...")

	// Create SchemaManager for centralized DDL execution
	schemaMgr := services.NewSchemaManager(db.DB())

	// Get all system table definitions
	tableDefs := GetSystemTableDefinitions()

	// CRITICAL: Ensure base metadata tables exist PHYSICALLY before any registration happens.
	// This breaks the circular dependency where registering _System_Table requires _System_Object to exist.
	coreTables := []string{constants.TableTable, constants.TableObject, constants.TableField}
	for _, coreName := range coreTables {
		for _, def := range tableDefs {
			if def.TableName == coreName {
				log.Printf("🧱 Pre-creating physical table: %s", coreName)
				if err := schemaMgr.CreatePhysicalTable(context.Background(), def); err != nil {
					log.Printf("⚠️  Warning pre-creating %s: %v", coreName, err)
					return err
				}
				break
			}
		}
	}

	// Phase 1: Create all tables in PARALLEL and BATCH register (DDL + _System_Table registry)
	// BatchCreatePhysicalTables handles parallel DDL and batch registration internally
	log.Println("⚡️ Creating tables and registering (Super Batch)...")
	if err := schemaMgr.BatchCreatePhysicalTables(context.Background(), tableDefs); err != nil {
		log.Printf("   ⚠️  Batch physical table creation failed: %v", err)
		return err
	}

	// Phase 2: BATCH register all object and field metadata
	log.Println("📋 Registering system object metadata (BATCH mode)...")

	// Collect all objects
	var allObjects []*models.ObjectMetadata
	var allFields []services.FieldWithContext

	for _, def := range tableDefs {
		objectID := services.GenerateObjectID(def.TableName)
		isCustom := constants.TableType(def.TableType) == constants.TableTypeCustomObject
		label := def.Description
		if label == "" {
			label = def.TableName
		}
		description := def.Description

		allObjects = append(allObjects, &models.ObjectMetadata{
			ID:           objectID,
			APIName:      def.TableName,
			Label:        label,
			PluralLabel:  def.TableName + "s",
			Description:  &description,
			IsCustom:     isCustom,
			SharingModel: models.SharingModel(constants.SharingModelPublicReadWrite),
		})

		// Collect fields for this table
		for _, col := range def.Columns {
			allFields = append(allFields, schemaMgr.PrepareFieldForBatch(def.TableName, col))
		}
	}

	// Batch insert objects (single statement)
	log.Printf("   📦 Batch inserting %d objects...", len(allObjects))
	if err := schemaMgr.BatchSaveObjectMetadata(allObjects, nil); err != nil {
		log.Printf("   ⚠️  Batch object insert failed: %v", err)
		return err
	}

	// Batch insert fields (batched in groups of 50)
	log.Printf("   📦 Batch inserting %d fields...", len(allFields))
	if err := schemaMgr.BatchSaveFieldMetadata(allFields, nil); err != nil {
		log.Printf("   ⚠️  Batch field insert failed: %v", err)
		return err
	}

	log.Println("✅ Core system schema initialized (optimized)")
	return nil
}
