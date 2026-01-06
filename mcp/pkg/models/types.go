package models

import (
	shared "github.com/nexuscrm/shared/pkg/models"
)

// SObject represents a generic CRM object record (map of field names to values)
type SObject = shared.SObject

// FieldType alias
type FieldType = shared.FieldType

// ObjectMetadata describes the schema of a CRM object
type ObjectMetadata = shared.ObjectMetadata

// FieldMetadata describes a single field on an object
type FieldMetadata = shared.FieldMetadata

// UserSession represents the authenticated user context
type UserSession = shared.UserSession

// AnalyticsQuery represents an analytics query
type AnalyticsQuery = shared.AnalyticsQuery

// NavigationItem represents a navigation item in an app
type NavigationItem = shared.NavigationItem

// QueryRequest represents a generic query request
type QueryRequest = shared.QueryRequest

// DashboardWidget represents a widget configuration for dashboards
// Alias to shared.WidgetConfig
type DashboardWidget = shared.WidgetConfig

// DashboardCreate is the request payload for creating a dashboard
// Alias to shared.DashboardConfig
type DashboardCreate = shared.DashboardConfig

// AppConfig represents an application configuration (navigation group)
// Alias to shared.AppConfig
type AppConfig = shared.AppConfig

// DashboardConfig represents a dashboard configuration
type DashboardConfig = shared.DashboardConfig

// RecycleBinItem represents an item in the recycle bin
type RecycleBinItem = shared.RecycleBinItem

// ValidationRule represents a validation rule
type ValidationRule = shared.ValidationRule
