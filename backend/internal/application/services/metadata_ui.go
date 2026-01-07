package services

import (
	"context"

	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== UI Component Metadata Methods ====================

// UpsertUIComponent created or updates a UI component definition
func (ms *MetadataService) UpsertUIComponent(component *models.UIComponent) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.UpsertUIComponent(context.Background(), component)
}

// ==================== Setup Page Metadata Methods ====================

// UpsertSetupPage creates or updates a setup page definition
func (ms *MetadataService) UpsertSetupPage(page *models.SetupPage) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.repo.UpsertSetupPage(context.Background(), page)
}
