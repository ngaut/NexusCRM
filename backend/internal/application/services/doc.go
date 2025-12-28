// Package services provides the business logic layer for NexusCRM.
//
// This package contains all service implementations that handle:
//   - CRUD operations with validation and events (PersistenceService)
//   - Metadata management for objects, fields, and relationships (MetadataService)
//   - Permission checks and security  (PermissionService)
//   - UI metadata for apps, tabs, layouts, dashboards (UIMetadataService)
//   - Database schema management (SchemaManager)
//   - Transaction handling with retry logic (TransactionManager)
//   - Event publishing and subscription (EventBus)
//   - Formula field calculations (integrated with persistence)
//   - Query execution with filtering and sorting (QueryService)
//   - System operations like logging and configuration (SystemManager)
//
// All services follow clean architecture principles with dependency injection
// and are designed to be testable and maintainable.
package services
