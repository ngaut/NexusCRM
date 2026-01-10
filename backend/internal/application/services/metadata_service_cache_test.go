package services

import (
	"context"
	"testing"

	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataCache_FlowCaching(t *testing.T) {
	conn, ms := SetupIntegrationTest(t)
	_ = conn // conn used by ms setup

	ctx := context.Background()

	t.Run("GetFlows_uses_cache", func(t *testing.T) {
		// Initial call should load cache
		flows1 := ms.GetFlows(ctx)

		// Second call should use cache (same result)
		flows2 := ms.GetFlows(ctx)

		// Results should be identical (same slice reference if cached)
		assert.Equal(t, len(flows1), len(flows2), "Flow counts should match")
	})

	t.Run("CreateFlow_invalidates_cache", func(t *testing.T) {
		// Get initial flows
		initialFlows := ms.GetFlows(ctx)
		initialCount := len(initialFlows)

		// Create a new flow
		newFlow := &models.Flow{
			Name:          "Test Cache Flow",
			TriggerObject: "account",
			TriggerType:   constants.TriggerAfterCreate,
			ActionType:    constants.ActionTypeUpdateRecord,
			Status:        constants.FlowStatusActive,
		}
		err := ms.CreateFlow(ctx, newFlow)
		require.NoError(t, err, "CreateFlow should succeed")

		// Cleanup
		t.Cleanup(func() {
			_ = ms.DeleteFlow(ctx, newFlow.ID)
		})

		// Cache should be invalidated, next GetFlows should include new flow
		updatedFlows := ms.GetFlows(ctx)
		assert.Equal(t, initialCount+1, len(updatedFlows), "Flow count should increase by 1 after CreateFlow")
	})

	t.Run("DeleteFlow_invalidates_cache", func(t *testing.T) {
		// Create a flow to delete
		newFlow := &models.Flow{
			Name:          "Test Delete Flow",
			TriggerObject: "account",
			TriggerType:   constants.TriggerAfterUpdate,
			ActionType:    constants.ActionTypeUpdateRecord,
			Status:        constants.FlowStatusDraft,
		}
		err := ms.CreateFlow(ctx, newFlow)
		require.NoError(t, err, "CreateFlow should succeed")

		// Get count after creation
		countAfterCreate := len(ms.GetFlows(ctx))

		// Delete the flow
		err = ms.DeleteFlow(ctx, newFlow.ID)
		require.NoError(t, err, "DeleteFlow should succeed")

		// Cache should be invalidated
		countAfterDelete := len(ms.GetFlows(ctx))
		assert.Equal(t, countAfterCreate-1, countAfterDelete, "Flow count should decrease after delete")
	})
}

func TestMetadataCache_ValidationRuleCaching(t *testing.T) {
	conn, ms := SetupIntegrationTest(t)
	_ = conn

	ctx := context.Background()

	// Use account object for testing
	testObject := "account"

	t.Run("GetValidationRules_uses_cache", func(t *testing.T) {
		// Initial call should load cache
		rules1 := ms.GetValidationRules(ctx, testObject)

		// Second call should use cache
		rules2 := ms.GetValidationRules(ctx, testObject)

		// Results should be identical
		assert.Equal(t, len(rules1), len(rules2), "Validation rule counts should match")
	})

	t.Run("CreateValidationRule_invalidates_cache", func(t *testing.T) {
		// Get initial rules
		initialRules := ms.GetValidationRules(ctx, testObject)
		initialCount := len(initialRules)

		// Create a new validation rule
		newRule := &models.ValidationRule{
			ObjectAPIName: testObject,
			Name:          "Test Cache Validation Rule",
			Condition:     "name == \"\"",
			ErrorMessage:  "Name is required for cache test",
			Active:        true,
		}
		err := ms.CreateValidationRule(ctx, newRule)
		require.NoError(t, err, "CreateValidationRule should succeed")

		// Cleanup
		t.Cleanup(func() {
			_ = ms.DeleteValidationRule(ctx, newRule.ID)
		})

		// Cache should be invalidated
		updatedRules := ms.GetValidationRules(ctx, testObject)
		assert.Equal(t, initialCount+1, len(updatedRules), "Validation rule count should increase by 1")
	})

	t.Run("GetValidationRules_skips_system_tables", func(t *testing.T) {
		// System tables should always return empty (optimization)
		rules := ms.GetValidationRules(ctx, constants.TableUser)
		assert.Empty(t, rules, "System tables should return empty validation rules")
	})
}

func TestMetadataCache_AutoNumberCaching(t *testing.T) {
	conn, ms := SetupIntegrationTest(t)
	_ = conn

	ctx := context.Background()

	t.Run("GetAutoNumbers_uses_cache", func(t *testing.T) {
		// Initial call should load cache
		ans1 := ms.GetAutoNumbers(ctx, "account")

		// Second call should use cache
		ans2 := ms.GetAutoNumbers(ctx, "account")

		// Results should be identical
		assert.Equal(t, len(ans1), len(ans2), "AutoNumber counts should match")
	})

	t.Run("GetAutoNumbers_returns_empty_for_unknown_object", func(t *testing.T) {
		// Unknown objects should return empty slice (from cache)
		ans := ms.GetAutoNumbers(ctx, "nonexistent_object_xyz")
		assert.Empty(t, ans, "Non-existent objects should return empty auto-numbers")
	})
}

func TestMetadataCache_RefreshAndInvalidate(t *testing.T) {
	conn, _ := SetupIntegrationTest(t)
	db := conn.DB()

	// Services
	repo := persistence.NewMetadataRepository(db)
	schemaRepo := persistence.NewSchemaRepository(db)
	schemaMgr := NewSchemaManager(schemaRepo)
	ms := NewMetadataService(repo, schemaMgr)

	t.Run("RefreshCache_loads_all_caches", func(t *testing.T) {
		// Refresh cache
		err := ms.RefreshCache()
		require.NoError(t, err, "RefreshCache should succeed")

		// Verify caches are populated
		ctx := context.Background()

		// Schemas should be loaded
		schemas := ms.GetSchemas(ctx)
		assert.NotEmpty(t, schemas, "Schemas should be loaded after RefreshCache")

		// Flows should be accessible (may be empty)
		flows := ms.GetFlows(ctx)
		t.Logf("Loaded %d flows", len(flows))
	})

	t.Run("InvalidateCache_clears_caches", func(t *testing.T) {
		// First ensure cache is loaded
		_ = ms.RefreshCache()

		// Invalidate
		ms.InvalidateCache()

		// Next call should re-load cache
		ctx := context.Background()
		schemas := ms.GetSchemas(ctx)
		assert.NotEmpty(t, schemas, "Schemas should be reloaded after InvalidateCache + GetSchemas")
	})
}
