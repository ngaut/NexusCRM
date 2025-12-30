package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRollupSummary_Sum(t *testing.T) {
	// Setup Dependencies
	conn, err := database.GetInstance()
	if err != nil {
		t.Skip("Skipping integration test: failed to connect to DB: " + err.Error())
	}

	db := conn.DB()

	// Initialize Services
	eventBus := NewEventBus()
	txManager := NewTransactionManager(conn)
	schemaMgr := NewSchemaManager(conn.DB())
	metadataSvc := NewMetadataService(conn, schemaMgr)
	permSvc := NewPermissionService(conn, metadataSvc)

	var ps *PersistenceService

	ctx := context.Background()
	adminUser := &models.UserSession{
		ID:        "test-admin",
		Name:      "Test Admin",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// 1. Define Parent Object (Invoice)
	parentName := fmt.Sprintf("invoice_%d", time.Now().UnixNano())
	parentCols := schemaMgr.GetStandardSystemColumns()
	parentCols = append(parentCols, schema.ColumnDefinition{
		Name: "total_amount", Type: "DECIMAL(18,2)", Nullable: true,
	})

	parentDef := schema.TableDefinition{
		TableName:   parentName,
		Description: "Master Invoice",
		TableType:   string(constants.TableTypeCustomObject),
		Columns:     parentCols,
	}
	require.NoError(t, schemaMgr.CreateTableFromDefinition(context.Background(), parentDef))
	defer schemaMgr.DropTable(parentName)

	// 2. Define Child Object (LineItem)
	childName := fmt.Sprintf("line_item_%d", time.Now().UnixNano())
	childCols := schemaMgr.GetStandardSystemColumns()
	childCols = append(childCols,
		schema.ColumnDefinition{Name: "amount", Type: "DECIMAL(18,2)", Nullable: false},
		schema.ColumnDefinition{Name: "invoice_id", Type: "VARCHAR(255)", Nullable: true},
	)

	childDef := schema.TableDefinition{
		TableName:   childName,
		Description: "Invoice Line Item",
		TableType:   "custom_object",
		Columns:     childCols,
	}
	require.NoError(t, schemaMgr.CreateTableFromDefinition(context.Background(), childDef))
	defer schemaMgr.DropTable(childName)

	// 3. Register Metadata & Rollup
	desc := "Test Object"
	require.NoError(t, schemaMgr.BatchSaveObjectMetadata([]*models.ObjectMetadata{
		{APIName: parentName, Label: "Invoice", Description: &desc},
		{APIName: childName, Label: "Line Item", Description: &desc},
	}, nil))

	// Register Fields for Parent
	rollupConfig := &models.RollupConfig{
		SummaryObject:     childName,
		SummaryField:      "amount",
		RelationshipField: "invoice_id",
		CalcType:          "SUM",
	}

	var parentObjID string
	err = db.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE api_name = ?", constants.TableObject), parentName).Scan(&parentObjID)
	require.NoError(t, err)

	var childObjID string
	err = db.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE api_name = ?", constants.TableObject), childName).Scan(&childObjID)
	require.NoError(t, err)

	parentFields := []models.FieldMetadata{
		{APIName: "total_amount", Label: "Total Amount", Type: constants.FieldTypeCurrency, RollupConfig: rollupConfig},
	}

	// Convert to FieldWithContext
	parentBatch := make([]FieldWithContext, len(parentFields))
	for i, f := range parentFields {
		f.IsSystem = false
		parentBatch[i] = FieldWithContext{
			ObjectID: parentObjID,
			FieldID:  fmt.Sprintf("fld_%s_%s", parentName, f.APIName),
			Field:    &f,
		}
	}
	require.NoError(t, schemaMgr.BatchSaveFieldMetadata(parentBatch, nil))

	// Register Fields for Child
	childFields := []models.FieldMetadata{
		{APIName: "amount", Label: "Amount", Type: constants.FieldTypeCurrency},
		{APIName: "invoice_id", Label: "Invoice", Type: constants.FieldTypeLookup, ReferenceTo: []string{parentName}},
	}
	childBatch := make([]FieldWithContext, len(childFields))
	for i, f := range childFields {
		f.IsSystem = false
		childBatch[i] = FieldWithContext{
			ObjectID: childObjID,
			FieldID:  fmt.Sprintf("fld_%s_%s", childName, f.APIName),
			Field:    &f,
		}
	}
	require.NoError(t, schemaMgr.BatchSaveFieldMetadata(childBatch, nil))

	// 4. Create Data
	// refresh metadata service to pick up new schemas
	metadataSvc = NewMetadataService(conn, schemaMgr)
	ps = NewPersistenceService(conn, metadataSvc, permSvc, eventBus, txManager)

	invoiceData := models.SObject{"total_amount": 0, "name": "Invoice 001"}
	invoice, err := ps.Insert(ctx, parentName, invoiceData, adminUser)
	require.NoError(t, err)
	invoiceID := invoice["id"].(string)

	checkInvoice := func(expected float64) {
		// Wait for async? No, rollup is synchronous in current implementation if transactional.
		// If we use EventBus async listener, we need to wait.
		// But ps.rollup.ProcessRollups is called synchronously in hooks.
		// However, does ps.Update commit? Yes.

		var val float64
		// We query total_amount
		err := db.QueryRow(fmt.Sprintf("SELECT total_amount FROM %s WHERE id = ?", parentName), invoiceID).Scan(&val)
		require.NoError(t, err)
		assert.Equal(t, expected, val)
	}
	checkInvoice(0)

	// Create Child Line Item 1 ($100)
	item1 := models.SObject{"amount": 100, "invoice_id": invoiceID, "name": "Item 1"}
	_, err = ps.Insert(ctx, childName, item1, adminUser)
	require.NoError(t, err)
	checkInvoice(100)

	// Create Child Line Item 2 ($50)
	item2 := models.SObject{"amount": 50, "invoice_id": invoiceID, "name": "Item 2"}
	_, err = ps.Insert(ctx, childName, item2, adminUser)
	require.NoError(t, err)
	checkInvoice(150)

	// Update Item 1 ($100 -> $200) - need ID
	item3, err := ps.Insert(ctx, childName, models.SObject{"amount": 25, "invoice_id": invoiceID, "name": "Item 3"}, adminUser)
	require.NoError(t, err)
	item3ID := item3["id"].(string)
	checkInvoice(175)

	// Update Item 3 ($25 -> $125)
	err = ps.Update(ctx, childName, item3ID, models.SObject{"amount": 125}, adminUser)
	require.NoError(t, err)
	checkInvoice(275)

	// Delete Item 3
	err = ps.Delete(ctx, childName, item3ID, adminUser)
	require.NoError(t, err)
	checkInvoice(150)

	// ================= REPARENTING TEST =================
	// Create a Second Invoice
	invoice2Data := models.SObject{"total_amount": 0, "name": "Invoice 002"}
	invoice2, err := ps.Insert(ctx, parentName, invoice2Data, adminUser)
	require.NoError(t, err)
	invoice2ID := invoice2["id"].(string)

	checkInvoice2 := func(expected float64) {
		var val float64
		err := db.QueryRow(fmt.Sprintf("SELECT total_amount FROM %s WHERE id = ?", parentName), invoice2ID).Scan(&val)
		require.NoError(t, err)
		assert.Equal(t, expected, val)
	}
	checkInvoice2(0)

	// Move Item 2 ($50) from Invoice 1 to Invoice 2
	// Item 2 ID is not captured above, let's capture it during creation or query it
	// We need to refactor creation or query.
	// For simplicity, let's create a NEW Item 4 ($75) on Invoice 1
	item4, err := ps.Insert(ctx, childName, models.SObject{"amount": 75, "invoice_id": invoiceID, "name": "Item 4"}, adminUser)
	require.NoError(t, err)
	item4ID := item4["id"].(string)

	// Invoice 1 should be 150 + 75 = 225
	checkInvoice(225) // FAILS IF REPARENTING BROKEN? No, this is insert.

	// Now Reparent Item 4 to Invoice 2
	err = ps.Update(ctx, childName, item4ID, models.SObject{"invoice_id": invoice2ID}, adminUser)
	require.NoError(t, err)

	// Verify Invoice 2 (Should be 75) - This works because it's the "new" parent
	checkInvoice2(75)

	// Verify Invoice 1 (Should be 150) - This FAILS if old parent isn't recalculated
	checkInvoice(150)
}
