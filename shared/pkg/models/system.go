package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// SystemAuditLog represents an audit log entry
type SystemAuditLog struct {
	ID               string    `json:"id,omitempty"`
	ObjectAPIName    string    `json:"object_api_name"`
	RecordID         string    `json:"record_id"`
	FieldName        string    `json:"field_name"`
	OldValue         string    `json:"old_value"`
	NewValue         string    `json:"new_value"`
	ChangedByID      string    `json:"changed_by_id"`
	ChangedAt        time.Time `json:"changed_at"`
	CreatedDate      time.Time `json:"created_date,omitempty"`
	LastModifiedDate time.Time `json:"last_modified_date,omitempty"`
}

func (a *SystemAuditLog) ToSObject() SObject {
	b, _ := json.Marshal(a)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// SystemLog represents a system log entry
type SystemLog struct {
	ID        string  `json:"id"`
	Timestamp string  `json:"timestamp"`
	Level     string  `json:"level"` // INFO, WARN, ERROR
	Source    string  `json:"source"`
	Message   string  `json:"message"`
	Details   *string `json:"details,omitempty"`
}

func (l *SystemLog) ToSObject() SObject {
	b, _ := json.Marshal(l)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// SystemConfig represents a system configuration key-value pair
type SystemConfig struct {
	KeyName     string  `json:"key_name"`
	Value       string  `json:"value"`
	IsSecret    bool    `json:"is_secret"`
	Description *string `json:"description,omitempty"`
}

func (c *SystemConfig) ToSObject() SObject {
	b, _ := json.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// SystemTable represents a table registry record
type SystemTable struct {
	ID            string    `json:"id,omitempty"`
	TableName     string    `json:"table_name"`
	TableType     string    `json:"table_type"`
	Category      string    `json:"category,omitempty"`
	Description   string    `json:"description,omitempty"`
	IsManaged     bool      `json:"is_managed"`
	SchemaVersion string    `json:"schema_version"`
	CreatedDate   time.Time `json:"created_date,omitempty"`
}

func (t *SystemTable) ToSObject() SObject {
	b, _ := json.Marshal(t)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// SystemNotification represents a notification record
type SystemNotification struct {
	ID               string    `json:"id,omitempty"`
	RecipientID      string    `json:"recipient_id"`
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	Link             string    `json:"link,omitempty"`
	NotificationType string    `json:"notification_type"`
	IsRead           bool      `json:"is_read"`
	CreatedDate      time.Time `json:"created_date"`
}

func (n *SystemNotification) ToSObject() SObject {
	b, _ := json.Marshal(n)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// Helper functions for JSON marshaling with sql.Null* types
func NewNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func NullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func NewNullInt64(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(*i), Valid: true}
}

func NullInt64ToPtr(ni sql.NullInt64) *int {
	if !ni.Valid {
		return nil
	}
	val := int(ni.Int64)
	return &val
}

func NewNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func NullFloat64ToPtr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

func NewNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

func NullBoolToPtr(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	return &nb.Bool
}

func NewNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func NullTimeToPtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// Helper to parse JSON strings
func ParseJSON(data string, v interface{}) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), v)
}

// Helper to stringify JSON
func StringifyJSON(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
