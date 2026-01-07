package models

import (
	"database/sql"
	"time"
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

// Transaction represents a database transaction
type Transaction interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Commit() error
	Rollback() error
}
