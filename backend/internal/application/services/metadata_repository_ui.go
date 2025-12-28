package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
)

// scanApp scans a row into an AppConfig struct
func (ms *MetadataService) scanApp(row Scannable) (*models.AppConfig, error) {
	var app models.AppConfig
	var name string
	// fix: scan color as well
	var description, icon, color, navItems sql.NullString

	if err := row.Scan(&app.ID, &name, &app.Label, &description, &icon, &color, &navItems); err != nil {
		return nil, err
	}

	app.Description = ScanNullStringValue(description)
	app.Icon = ScanNullStringValue(icon)
	app.Color = ScanNullStringValue(color)
	UnmarshalJSONField(navItems, &app.NavigationItems)

	return &app, nil
}

// queryAllApps queries all apps
func (ms *MetadataService) queryAllApps() ([]*models.AppConfig, error) {
	// fix: include color in SELECT
	query := fmt.Sprintf("SELECT id, name, label, description, icon, color, navigation_items FROM %s", constants.TableApp)
	rows, err := ms.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := make([]*models.AppConfig, 0)
	for rows.Next() {
		app, err := ms.scanApp(rows)
		if err != nil {
			continue
		}
		apps = append(apps, app)
	}

	return apps, nil
}

// queryApp queries a single app by ID
func (ms *MetadataService) queryApp(id string) (*models.AppConfig, error) {
	// fix: include color in SELECT
	query := fmt.Sprintf("SELECT id, name, label, description, icon, color, navigation_items FROM %s WHERE id = ?", constants.TableApp)

	app, err := ms.scanApp(ms.db.QueryRow(query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}

	return app, nil
}

// scanLayout scans a row into a PageLayout struct
func (ms *MetadataService) scanLayout(row Scannable) (*models.PageLayout, error) {
	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		return nil, err
	}

	var layout models.PageLayout
	if err := json.Unmarshal([]byte(configJSON), &layout); err != nil {
		return nil, err
	}
	return &layout, nil
}

// queryLayouts queries all layouts for an object
func (ms *MetadataService) queryLayouts(objectAPIName string) ([]*models.PageLayout, error) {
	rows, err := ms.db.Query(fmt.Sprintf("SELECT config FROM %s WHERE object_api_name = ?", constants.TableLayout), objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	layouts := make([]*models.PageLayout, 0)
	for rows.Next() {
		layout, err := ms.scanLayout(rows)
		if err != nil {
			continue
		}
		layouts = append(layouts, layout)
	}

	return layouts, nil
}

// queryLayout queries a single layout by ID
func (ms *MetadataService) queryLayout(layoutID string) (*models.PageLayout, error) {
	layout, err := ms.scanLayout(ms.db.QueryRow(fmt.Sprintf("SELECT config FROM %s WHERE id = ?", constants.TableLayout), layoutID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return layout, nil
}

// scanDashboard scans a row into a DashboardConfig struct
func (ms *MetadataService) scanDashboard(row Scannable) (*models.DashboardConfig, error) {
	var db models.DashboardConfig
	var description, widgetsJSON sql.NullString

	if err := row.Scan(&db.ID, &db.Label, &description, &db.Layout, &widgetsJSON); err != nil {
		return nil, err
	}

	db.Description = ScanNullString(description)
	UnmarshalJSONField(widgetsJSON, &db.Widgets)
	return &db, nil
}

// queryDashboards queries all dashboards
func (ms *MetadataService) queryDashboards() ([]*models.DashboardConfig, error) {
	rows, err := ms.db.Query(fmt.Sprintf("SELECT id, name, description, layout, widgets FROM %s", constants.TableDashboard))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dashboards := make([]*models.DashboardConfig, 0)
	for rows.Next() {
		db, err := ms.scanDashboard(rows)
		if err != nil {
			continue
		}
		dashboards = append(dashboards, db)
	}
	return dashboards, nil
}

// queryDashboard queries a single dashboard
func (ms *MetadataService) queryDashboard(id string) (*models.DashboardConfig, error) {
	db, err := ms.scanDashboard(ms.db.QueryRow(
		fmt.Sprintf("SELECT id, name, description, layout, widgets FROM %s WHERE id = ?", constants.TableDashboard),
		id,
	))

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return db, nil
}

// scanListView scans a row into a ListView struct
func (ms *MetadataService) scanListView(row Scannable) (*models.ListView, error) {
	var view models.ListView
	var filtersJSON, fieldsJSON sql.NullString
	if err := row.Scan(&view.ID, &view.ObjectAPIName, &view.Label, &filtersJSON, &fieldsJSON); err != nil {
		return nil, err
	}
	UnmarshalJSONField(filtersJSON, &view.Filters)
	UnmarshalJSONField(fieldsJSON, &view.Fields)
	return &view, nil
}

// queryListViews queries list views for an object
func (ms *MetadataService) queryListViews(objectAPIName string) ([]*models.ListView, error) {
	query := fmt.Sprintf("SELECT id, object_api_name, label, filters, fields FROM %s WHERE object_api_name = ?", constants.TableListView)
	rows, err := ms.db.Query(query, objectAPIName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	views := make([]*models.ListView, 0)
	for rows.Next() {
		view, err := ms.scanListView(rows)
		if err != nil {
			continue
		}
		views = append(views, view)
	}
	return views, nil
}

// UpsertUIComponent inserts or updates a UI component definition
func (ms *MetadataService) UpsertUIComponent(component *models.UIComponent) error {
	// Check if exists by Name
	var existingID string
	err := ms.db.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE name = ?", constants.TableUIComponent), component.Name).Scan(&existingID)

	if err == nil {
		// Found, update it
		component.ID = existingID
		query := fmt.Sprintf(`
			UPDATE %s SET 
				description = ?, 
				type = ?,
				is_embeddable = ?,
				component_path = ?,
				last_modified_date = CURRENT_TIMESTAMP
			WHERE id = ?`, constants.TableUIComponent)

		_, err = ms.db.Exec(query, component.Description, component.Type, component.IsEmbeddable, component.ComponentPath, component.ID)
		return err
	}

	if err != sql.ErrNoRows {
		return err // Real error
	}

	// Not found, Insert
	if component.ID == "" {
		component.ID = GenerateID()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, type, is_embeddable, description, component_path, created_date, last_modified_date)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, constants.TableUIComponent)

	_, err = ms.db.Exec(query, component.ID, component.Name, component.Type, component.IsEmbeddable, component.Description, component.ComponentPath)
	return err
}

// UpsertSetupPage inserts or updates a setup page definition
func (ms *MetadataService) UpsertSetupPage(page *models.SetupPage) error {
	// Check if exists by ID
	var existingID string
	err := ms.db.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE id = ?", constants.TableSetupPage), page.ID).Scan(&existingID)

	if err == nil {
		// Found, update it
		query := fmt.Sprintf(`
			UPDATE %s SET 
				label = ?, 
				icon = ?,
				component_name = ?,
				category = ?,
				page_order = ?,
				permission_required = ?,
				is_enabled = ?,
				description = ?,
				last_modified_date = CURRENT_TIMESTAMP
			WHERE id = ?`, constants.TableSetupPage)

		_, err = ms.db.Exec(query, page.Label, page.Icon, page.ComponentName, page.Category, page.PageOrder, page.PermissionRequired, page.IsEnabled, page.Description, page.ID)
		return err
	}

	if err != sql.ErrNoRows {
		return err // Real error
	}

	// Not found, Insert
	query := fmt.Sprintf(`
		INSERT INTO %s (id, label, icon, component_name, category, page_order, permission_required, is_enabled, description, created_date, last_modified_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, constants.TableSetupPage)

	_, err = ms.db.Exec(query, page.ID, page.Label, page.Icon, page.ComponentName, page.Category, page.PageOrder, page.PermissionRequired, page.IsEnabled, page.Description)
	return err
}
