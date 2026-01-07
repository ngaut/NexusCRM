package services

import (
	"context"

	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== UI Component Metadata Methods ====================

// UpsertUIComponent created or updates a UI component definition
func (ms *MetadataService) UpsertUIComponent(ctx context.Context, component *models.UIComponent) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.UpsertUIComponent(ctx, component)
}

// ==================== Setup Page Metadata Methods ====================

// UpsertSetupPage creates or updates a setup page definition
func (ms *MetadataService) UpsertSetupPage(ctx context.Context, page *models.SetupPage) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.UpsertSetupPage(ctx, page)
}

// GetSetupPages returns all setup pages
func (ms *MetadataService) GetSetupPages(ctx context.Context) ([]models.SetupPage, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.repo.GetSetupPages(ctx)
}
