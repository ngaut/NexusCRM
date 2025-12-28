package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/pkg/constants"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/utils"
)

type DataHandler struct {
	svc *services.ServiceManager
}

func NewDataHandler(svc *services.ServiceManager) *DataHandler {
	return &DataHandler{svc: svc}
}

// QueryRequest represents a query request
type QueryRequest struct {
	ObjectApiName string `json:"object_api_name" binding:"required"`
	FilterExpr    string `json:"filter_expr"` // Formula expression for filtering
	SortField     string `json:"sort_field"`
	SortDirection string `json:"sort_direction"`
	Limit         int    `json:"limit"`
}

// Query handles POST /api/data/query
func (h *DataHandler) Query(c *gin.Context) {
	user := GetUserFromContext(c)
	var req QueryRequest
	if !BindJSON(c, &req) {
		return
	}

	// Normalize object API name from JSON body
	req.ObjectApiName = strings.ToLower(req.ObjectApiName)

	HandleGetEnvelope(c, "records", func() (interface{}, error) {
		return h.svc.QuerySvc.QueryWithFilter(
			c.Request.Context(),
			req.ObjectApiName,
			req.FilterExpr,
			user,
			req.SortField,
			req.SortDirection,
			req.Limit,
		)
	})
}

// Search handles POST /api/data/search
func (h *DataHandler) Search(c *gin.Context) {
	user := GetUserFromContext(c)
	var req struct {
		Term string `json:"term" binding:"required"`
	}

	if !BindJSON(c, &req) {
		return
	}

	HandleGetEnvelope(c, "results", func() (interface{}, error) {
		return h.svc.QuerySvc.GlobalSearch(c.Request.Context(), req.Term, user)
	})
}

// SearchSingleObject handles searching within a single object
func (h *DataHandler) SearchSingleObject(c *gin.Context) {
	user := GetUserFromContext(c)
	objectName := strings.ToLower(c.Param("objectApiName"))
	term := c.Query("term")

	HandleGetEnvelope(c, "records", func() (interface{}, error) {
		if term == "" {
			return nil, errors.NewValidationError("term", "Search term is required")
		}
		return h.svc.QuerySvc.SearchSingleObject(c.Request.Context(), objectName, term, user)
	})
}

// GetRecycleBinItems handles GET /api/data/recyclebin/items
func (h *DataHandler) GetRecycleBinItems(c *gin.Context) {
	user := GetUserFromContext(c)
	scope := c.Query("scope") // "mine" or "all"
	items, err := h.svc.Persistence.GetRecycleBinItems(c.Request.Context(), user, scope)
	if err != nil {
		RespondError(c, errors.GetHTTPStatus(err), err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// RestoreFromRecycleBin handles POST /api/data/recyclebin/restore/:id
func (h *DataHandler) RestoreFromRecycleBin(c *gin.Context) {
	user := GetUserFromContext(c)
	id := c.Param("id")
	if err := h.svc.Persistence.Restore(c.Request.Context(), id, user); err != nil {
		RespondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to restore record: %v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Record restored successfully"})
}

// PurgeFromRecycleBin handles DELETE /api/data/recyclebin/:id
func (h *DataHandler) PurgeFromRecycleBin(c *gin.Context) {
	user := GetUserFromContext(c)
	recordId := c.Param("id")

	HandleDeleteEnvelope(c, "Record purged successfully", func() error {
		return h.svc.Persistence.Purge(c.Request.Context(), recordId, user)
	})
}

// GetRecord handles GET /api/data/:objectApiName/:id
func (h *DataHandler) GetRecord(c *gin.Context) {
	user := GetUserFromContext(c)
	objectApiName := strings.ToLower(c.Param("objectApiName"))
	id := c.Param("id")

	HandleGetEnvelope(c, "record", func() (interface{}, error) {
		if !utils.IsValidUUID(id) {
			return nil, errors.NewValidationError("id", "Invalid ID format")
		}
		// Use formula expression for ID lookup
		filterExpr := fmt.Sprintf("id == '%s'", id)
		records, err := h.svc.QuerySvc.QueryWithFilter(
			c.Request.Context(),
			objectApiName,
			filterExpr,
			user,
			constants.FieldCreatedDate,
			"DESC",
			1,
		)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			// Ensure we return a NotFoundError so it maps to 404
			return nil, errors.NewNotFoundError(objectApiName, id)
		}

		record := records[0]

		// Secure Read: Check single record access (Ownership/Sharing)
		schema := h.svc.Metadata.GetSchema(objectApiName)
		if schema != nil {
			if !h.svc.Permissions.CheckRecordAccess(schema, record, "read", user) {
				// Return PermissionError (403) or NotFound (404) depending on security policy.
				// For strict security, 404 prevents enumerating IDs.
				// But we'll use PermissionError for clarity for now, or match persistence service.
				return nil, errors.NewPermissionError("read", objectApiName+"/"+id)
			}
		}

		return record, nil
	})
}

// CreateRecord handles POST /api/data/:objectApiName
func (h *DataHandler) CreateRecord(c *gin.Context) {
	user := GetUserFromContext(c)
	objectApiName := strings.ToLower(c.Param("objectApiName"))

	var data models.SObject
	// Use manual binding here to preserve original map structure before envelope
	// HandleCreateEnvelope will bind to it.
	data = make(models.SObject)

	// We need to capture the created record to return it
	HandleCreateEnvelope(c, "record", "Record created successfully", &data, func() error {
		// Data is already bound by HandleCreateEnvelope
		record, err := h.svc.Persistence.Insert(c.Request.Context(), objectApiName, data, user)
		if err != nil {
			return err
		}
		// Update data with returned record (which includes ID and systems fields)
		data = record
		return nil
	})
}

// UpdateRecord handles PATCH /api/data/:objectApiName/:id
func (h *DataHandler) UpdateRecord(c *gin.Context) {
	user := GetUserFromContext(c)
	objectApiName := strings.ToLower(c.Param("objectApiName"))
	id := c.Param("id")

	var updates models.SObject
	updates = make(models.SObject)

	HandleUpdateEnvelope(c, "", "Record updated successfully", &updates, func() error {
		return h.svc.Persistence.Update(c.Request.Context(), objectApiName, id, updates, user)
	})
}

// DeleteRecord handles DELETE /api/data/:objectApiName/:id
func (h *DataHandler) DeleteRecord(c *gin.Context) {
	user := GetUserFromContext(c)
	objectApiName := strings.ToLower(c.Param("objectApiName"))
	id := c.Param("id")

	HandleDeleteEnvelope(c, "Record deleted successfully", func() error {
		return h.svc.Persistence.Delete(c.Request.Context(), objectApiName, id, user)
	})
}

// RunAnalytics handles POST /api/data/analytics
func (h *DataHandler) RunAnalytics(c *gin.Context) {
	user := GetUserFromContext(c)
	var query models.AnalyticsQuery

	if !BindJSON(c, &query) {
		return
	}

	// Normalize object API name from JSON body
	query.ObjectAPIName = strings.ToLower(query.ObjectAPIName)

	HandleGetEnvelope(c, "result", func() (interface{}, error) {
		return h.svc.QuerySvc.RunAnalytics(c.Request.Context(), query, user)
	})
}

// Calculate handles POST /api/data/:objectApiName/calculate
func (h *DataHandler) Calculate(c *gin.Context) {
	user := GetUserFromContext(c)
	objectApiName := strings.ToLower(c.Param("objectApiName"))

	var record models.SObject
	if !BindJSON(c, &record) {
		return
	}

	HandleGetEnvelope(c, "record", func() (interface{}, error) {
		return h.svc.QuerySvc.Calculate(objectApiName, record, user)
	})
}
