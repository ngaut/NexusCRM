package models

import (
	"database/sql"
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
)

// SObject represents a generic record
type SObject map[string]interface{}

// Helper methods for SObject
func (s SObject) GetString(key string) string {
	if val, ok := s[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (s SObject) GetBool(key string) bool {
	if val, ok := s[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (s SObject) GetTime(key string) time.Time {
	if val, ok := s[key]; ok {
		if t, ok := val.(time.Time); ok {
			return t
		}
		if tStr, ok := val.(string); ok {
			parsed, _ := time.Parse(time.RFC3339, tStr)
			return parsed
		}
	}
	return time.Time{}
}

func (s SObject) Get(key string) interface{} {
	return s[key]
}

// SearchResult represents global search results
type SearchResult struct {
	ObjectLabel   string    `json:"object_label"`
	ObjectAPIName string    `json:"object_api_name"`
	Icon          string    `json:"icon"`
	Matches       []SObject `json:"matches"`
}

// AnalyticsQuery represents an analytics query
type AnalyticsQuery struct {
	ObjectAPIName string  `json:"object_api_name"`
	Operation     string  `json:"operation"` // count, sum, avg, group_by
	Field         *string `json:"field"`
	GroupBy       *string `json:"group_by"`
	FilterExpr    string  `json:"filter_expr,omitempty"` // Formula expression for filtering
}

// QueryCriterion represents a single query filter criterion
type QueryCriterion struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Val   interface{} `json:"val"`
}

// QueryRequest represents a generic query request
type QueryRequest struct {
	ObjectAPIName string           `json:"object_api_name" binding:"required"`
	Criteria      []QueryCriterion `json:"criteria,omitempty"`
	FilterExpr    string           `json:"filter_expr,omitempty"` // Formula expression for filtering
	SortField     string           `json:"sort_field,omitempty"`
	SortDirection string           `json:"sort_direction,omitempty"`
	Limit         int              `json:"limit,omitempty"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Term string `json:"term" binding:"required"`
}

// RecycleBinItem represents an item in the recycle bin
type RecycleBinItem struct {
	ID            string `json:"id"`
	RecordID      string `json:"record_id"`
	ObjectAPIName string `json:"object_api_name"`
	RecordName    string `json:"record_name"`
	DeletedBy     string `json:"deleted_by"`
	DeletedDate   string `json:"deleted_date"`
}

// SystemFile represents a file record
type SystemFile struct {
	ID             string    `json:"id,omitempty"`
	TargetID       string    `json:"target_id"`
	TargetObject   string    `json:"target_object"`
	OrganizationID string    `json:"organization_id"`
	Filename       string    `json:"filename"`
	Path           string    `json:"path"`
	MimeType       string    `json:"mime_type"`
	Size           int64     `json:"size"`
	UploadedBy     string    `json:"uploaded_by"`
	CreatedDate    time.Time `json:"created_date"`
}

func (f *SystemFile) ToSObject() SObject {
	return SObject{
		constants.FieldID:               f.ID,
		constants.FieldSysFile_ParentID: f.TargetID,
		// TargetObject and OrganizationID are not in _System_File table definition
		constants.FieldSysFile_Name:        f.Filename,
		constants.FieldSysFile_StoragePath: f.Path,
		constants.FieldSysFile_MimeType:    f.MimeType,
		constants.FieldSysFile_SizeBytes:   f.Size,
		constants.FieldCreatedByID:         f.UploadedBy,
		constants.FieldCreatedDate:         f.CreatedDate,
	}
}

// SystemComment represents a comment record
type SystemComment struct {
	ID              string    `json:"id,omitempty"`
	Body            string    `json:"body"`
	ObjectAPIName   string    `json:"object_api_name"`
	RecordID        string    `json:"record_id"`
	ParentCommentID *string   `json:"parent_comment_id,omitempty"`
	IsResolved      bool      `json:"is_resolved"`
	CreatedBy       string    `json:"created_by"`
	CreatedDate     time.Time `json:"created_date"`
}

func (c *SystemComment) ToSObject() SObject {
	m := SObject{
		constants.FieldSysComment_Body:          c.Body,
		constants.FieldSysComment_ObjectAPIName: c.ObjectAPIName,
		constants.FieldSysComment_RecordID:      c.RecordID,
		constants.FieldSysComment_IsResolved:    c.IsResolved,
		constants.FieldCreatedByID:              c.CreatedBy,
		constants.FieldCreatedDate:              c.CreatedDate,
	}
	if c.ID != "" {
		m[constants.FieldID] = c.ID
	}
	if c.ParentCommentID != nil {
		m[constants.FieldSysComment_ParentCommentID] = *c.ParentCommentID
	}
	return m
}

// RecentItem represents a recently viewed item
type RecentItem struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	ObjectAPIName string `json:"object_api_name"`
	RecordID      string `json:"record_id"`
	RecordName    string `json:"record_name"`
	Timestamp     string `json:"timestamp"`
}

func (r *RecentItem) ToSObject() SObject {
	return SObject{
		constants.FieldID:                      r.ID,
		constants.FieldSysRecent_UserID:        r.UserID,
		constants.FieldSysRecent_ObjectAPIName: r.ObjectAPIName,
		constants.FieldSysRecent_RecordID:      r.RecordID,
		constants.FieldSysRecent_RecordName:    r.RecordName,
		constants.FieldSysRecent_Timestamp:     r.Timestamp,
	}
}

// Transaction represents a database transaction
type Transaction interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Commit() error
	Rollback() error
}
