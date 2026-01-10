package bootstrap

import (
	"context"
	"log"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// InitializeSchema creates core system tables using SchemaManager
// This replaces raw SQL execution with declarative table definitions
func InitializeSchema(db *database.TiDBConnection) error {
	log.Println("🔧 Initializing core system schema...")

	// Create SchemaManager for centralized DDL execution
	repo := persistence.NewSchemaRepository(db.DB())
	schemaMgr := services.NewSchemaManager(repo)

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
	// Filter out core tables as they are already created
	var remainingDefs []schema.TableDefinition
	for _, def := range tableDefs {
		isCore := false
		for _, core := range coreTables {
			if def.TableName == core {
				isCore = true
				break
			}
		}
		if !isCore {
			remainingDefs = append(remainingDefs, def)
		}
	}

	// BatchCreatePhysicalTables handles parallel DDL and batch registration internally
	log.Println("⚡️ Creating tables and registering (Super Batch)...")
	if err := schemaMgr.BatchCreatePhysicalTables(context.Background(), remainingDefs); err != nil {
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
			fc := schemaMgr.PrepareFieldForBatch(def.TableName, col)
			fc.ObjectID = objectID // Enforce linkage to the actual object ID we generated above
			allFields = append(allFields, fc)
		}
	}

	// Batch insert objects (single statement) within a Transaction
	tx, err := db.DB().Begin()
	if err != nil {
		log.Printf("   ⚠️  Failed to begin transaction: %v", err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	log.Printf("   📦 Batch inserting %d objects...", len(allObjects))
	if err := schemaMgr.BatchSaveObjectMetadata(allObjects, tx); err != nil {
		log.Printf("   ⚠️  Batch object insert failed: %v", err)
		return err
	}

	// Batch insert fields (batched in groups of 50)
	log.Printf("   📦 Batch inserting %d fields...", len(allFields))
	if err := schemaMgr.BatchSaveFieldMetadata(allFields, tx); err != nil {
		log.Printf("   ⚠️  Batch field insert failed: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("   ⚠️  Failed to commit transaction: %v", err)
		return err
	}

	log.Println("✅ Core system schema initialized (optimized)")
	return nil
}
