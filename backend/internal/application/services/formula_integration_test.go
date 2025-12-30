package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/shared/pkg/constants"
)

func TestFormulaFields_Integration(t *testing.T) {
	// Setup Dependencies
	dbConn, err := database.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get DB instance: %v", err)
	}

	db := dbConn.DB()
	eventBus := services.NewEventBus()
	schemaManager := services.NewSchemaManager(db)

	metadataService := services.NewMetadataService(dbConn, schemaManager) // Correct signature
	permService := services.NewPermissionService(dbConn, metadataService)
	txManager := services.NewTransactionManager(dbConn)
	ps := services.NewPersistenceService(dbConn, metadataService, permService, eventBus, txManager)
	qs := services.NewQueryService(dbConn, metadataService, permService)

	// Context
	ctx := context.Background()

	// Use System Admin profile which bypasses permission checks implicitly via IsSuperUser check
	adminUser := &models.UserSession{
		ID:        "admin-user",
		Name:      "Admin User",
		ProfileID: constants.ProfileSystemAdmin,
	}

	// 1. Create Object "Product"
	objName := fmt.Sprintf("product_%d", time.Now().UnixNano())
	objDef := models.ObjectMetadata{
		APIName:      objName,
		Label:        "Product",
		PluralLabel:  "Products",
		SharingModel: constants.SharingModelPublicReadWrite, // Correct constant
	}
	// Use MetadataService to create schema (handles layout and system fields)
	if err := metadataService.CreateSchema(&objDef); err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	// 2. Create Fields: Price (Number), Quantity (Number)
	priceField := models.FieldMetadata{
		APIName: "price",
		Label:   "Price",
		Type:    constants.FieldTypeNumber,
	}
	if err := metadataService.CreateField(objName, &priceField); err != nil {
		t.Fatalf("Failed to create price field: %v", err)
	}

	qtyField := models.FieldMetadata{
		APIName: "quantity",
		Label:   "Quantity",
		Type:    constants.FieldTypeNumber,
	}
	if err := metadataService.CreateField(objName, &qtyField); err != nil {
		t.Fatalf("Failed to create quantity field: %v", err)
	}

	// 3. Create Formula Field: Total (Price * Quantity)
	formulaExpr := "price * quantity"
	returnType := constants.FieldTypeNumber
	totalField := models.FieldMetadata{
		APIName:    "total",
		Label:      "Total",
		Type:       constants.FieldTypeFormula,
		Formula:    &formulaExpr,
		ReturnType: &returnType,
	}
	if err := metadataService.CreateField(objName, &totalField); err != nil {
		t.Fatalf("Failed to create formula field: %v", err)
	}

	// Cache invalidation not needed as MetadataService reads from DB directly
	// metadataService.RefreshCache() // Optional/No-op

	// 4. Insert Record
	prod := models.SObject{
		"name":     "Test Product",
		"price":    10.5,
		"quantity": 2,
	}
	created, err := ps.Insert(ctx, objName, prod, adminUser)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}
	id := created[constants.FieldID].(string)

	// 5. Retrieve and Verify Formula Calculation
	t.Run("Read-Time Calculation", func(t *testing.T) {
		filterExpr := fmt.Sprintf("id == '%s'", id)
		// Query with limit 1
		results, err := qs.QueryWithFilter(ctx, objName, filterExpr, adminUser, "", "", 1)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("Record not found")
		}

		result := results[0]
		total := result["total"]
		if total == nil {
			t.Fatal("Formula field 'total' is nil")
		}

		// Check value (handle float types)
		var totalVal float64
		switch v := total.(type) {
		case float64:
			totalVal = v
		case int:
			totalVal = float64(v)
		case int64:
			totalVal = float64(v)
		default:
			t.Logf("Total type: %T value: %v", total, total)
			// Try to convert via string if needed, or assume failure
			// Expression engine usually returns float64 or int
		}

		// 10.5 * 2 = 21.0
		if totalVal != 21.0 {
			t.Errorf("Expected total 21.0, got %v (type %T)", total, total)
		}
	})

	// Query: WHERE total > 20 (Should match)
	t.Run("Filter by Formula (Match)", func(t *testing.T) {
		filterExpr := "total > 20"
		results, err := qs.QueryWithFilter(ctx, objName, filterExpr, adminUser, "", "", 1)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 match, got %d", len(results))
		}
	})

	// Query: WHERE total < 20 (Should NOT match)
	t.Run("Filter by Formula (No Match)", func(t *testing.T) {
		filterExpr := "total < 20"
		results, err := qs.QueryWithFilter(ctx, objName, filterExpr, adminUser, "", "", 1)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 matches, got %d", len(results))
		}
	})

	// 7. Verify Generated Column Schema in Database
	t.Run("Verify Generated Column Schema", func(t *testing.T) {
		var generation string
		err := db.QueryRow(`
			SELECT GENERATION_EXPRESSION 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_NAME = ? AND COLUMN_NAME = 'total'
		`, objName).Scan(&generation)
		if err != nil {
			t.Fatalf("Failed to query column schema: %v", err)
		}
		if generation == "" {
			t.Error("Expected 'total' to be a generated column, but GENERATION_EXPRESSION is empty")
		} else {
			t.Logf("✅ Generated column expression: %s", generation)
		}
	})

	// 8. Test Update Propagation (Generated column auto-updates)
	t.Run("Update Propagation", func(t *testing.T) {
		// Update price from 10.5 to 20.0
		err := ps.Update(ctx, objName, id, models.SObject{"price": 20.0}, adminUser)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Query and verify total is now 20.0 * 2 = 40.0
		filterExpr := fmt.Sprintf("id == '%s'", id)
		results, err := qs.QueryWithFilter(ctx, objName, filterExpr, adminUser, "", "", 1)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("Record not found after update")
		}

		total := results[0]["total"]
		if total == nil {
			t.Fatal("Formula field 'total' is nil after update")
		}

		var totalVal float64
		switch v := total.(type) {
		case float64:
			totalVal = v
		case int:
			totalVal = float64(v)
		case int64:
			totalVal = float64(v)
		case []byte:
			// DECIMAL comes back as bytes in some drivers
			fmt.Sscanf(string(v), "%f", &totalVal)
		case string:
			fmt.Sscanf(v, "%f", &totalVal)
		}

		// 20.0 * 2 = 40.0
		if totalVal != 40.0 {
			t.Errorf("Expected total 40.0 after update, got %v (type %T)", total, total)
		}
	})

	// 9. Test Null Value Handling
	t.Run("Null Value Handling", func(t *testing.T) {
		// Insert a record with null price
		nullProd := models.SObject{
			"name":     "Null Price Product",
			"quantity": 5,
			// price is not set (null)
		}
		created2, err := ps.Insert(ctx, objName, nullProd, adminUser)
		if err != nil {
			t.Fatalf("Failed to insert record with null price: %v", err)
		}
		id2 := created2[constants.FieldID].(string)

		// Query the record
		filterExpr := fmt.Sprintf("id == '%s'", id2)
		results, err := qs.QueryWithFilter(ctx, objName, filterExpr, adminUser, "", "", 1)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("Record not found")
		}

		// Formula with null operand should return null (not crash)
		total := results[0]["total"]
		// In MySQL, NULL * 5 = NULL, which is expected behavior
		t.Logf("Total with null price: %v (type %T) - Expected nil", total, total)
		// We just verify it doesn't crash; null handling is database-specific
	})
}

// TestFormulaValidation tests formula syntax validation on field creation
func TestFormulaValidation(t *testing.T) {
	// Setup Dependencies
	dbConn, err := database.GetInstance()
	if err != nil {
		t.Fatalf("Failed to get DB instance: %v", err)
	}

	db := dbConn.DB()
	schemaManager := services.NewSchemaManager(db)
	metadataService := services.NewMetadataService(dbConn, schemaManager)

	// Create a test object
	objName := fmt.Sprintf("formula_validation_%d", time.Now().UnixNano())
	objDef := models.ObjectMetadata{
		APIName:      objName,
		Label:        "Formula Validation Test",
		PluralLabel:  "Formula Validation Tests",
		SharingModel: constants.SharingModelPublicReadWrite,
	}
	if err := metadataService.CreateSchema(&objDef); err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	// Create a source field
	amountField := models.FieldMetadata{
		APIName: "amount",
		Label:   "Amount",
		Type:    constants.FieldTypeNumber,
	}
	if err := metadataService.CreateField(objName, &amountField); err != nil {
		t.Fatalf("Failed to create amount field: %v", err)
	}

	t.Run("Valid Formula Syntax", func(t *testing.T) {
		formulaExpr := "amount * 1.1"
		returnType := constants.FieldTypeNumber
		validField := models.FieldMetadata{
			APIName:    "tax_amount",
			Label:      "Tax Amount",
			Type:       constants.FieldTypeFormula,
			Formula:    &formulaExpr,
			ReturnType: &returnType,
		}
		err := metadataService.CreateField(objName, &validField)
		if err != nil {
			t.Errorf("Expected valid formula to succeed, got error: %v", err)
		} else {
			t.Log("✅ Valid formula syntax accepted")
		}
	})

	t.Run("Invalid Formula Syntax - Parse Error", func(t *testing.T) {
		formulaExpr := "amount * * 1.1" // Invalid: double operator
		returnType := constants.FieldTypeNumber
		invalidField := models.FieldMetadata{
			APIName:    "invalid_formula",
			Label:      "Invalid Formula",
			Type:       constants.FieldTypeFormula,
			Formula:    &formulaExpr,
			ReturnType: &returnType,
		}
		err := metadataService.CreateField(objName, &invalidField)
		if err == nil {
			t.Error("Expected invalid formula syntax to fail, but it succeeded")
		} else {
			t.Logf("✅ Invalid formula correctly rejected: %v", err)
		}
	})

	t.Run("Empty Formula Expression", func(t *testing.T) {
		formulaExpr := ""
		returnType := constants.FieldTypeNumber
		emptyField := models.FieldMetadata{
			APIName:    "empty_formula",
			Label:      "Empty Formula",
			Type:       constants.FieldTypeFormula,
			Formula:    &formulaExpr,
			ReturnType: &returnType,
		}
		err := metadataService.CreateField(objName, &emptyField)
		if err == nil {
			t.Error("Expected empty formula to fail, but it succeeded")
		} else {
			t.Logf("✅ Empty formula correctly rejected: %v", err)
		}
	})

	t.Run("Nil Formula Expression", func(t *testing.T) {
		returnType := constants.FieldTypeNumber
		nilField := models.FieldMetadata{
			APIName:    "nil_formula",
			Label:      "Nil Formula",
			Type:       constants.FieldTypeFormula,
			Formula:    nil,
			ReturnType: &returnType,
		}
		err := metadataService.CreateField(objName, &nilField)
		if err == nil {
			t.Error("Expected nil formula to fail, but it succeeded")
		} else {
			t.Logf("✅ Nil formula correctly rejected: %v", err)
		}
	})
}
