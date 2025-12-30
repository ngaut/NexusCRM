package models

import (
	"encoding/json"
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
)

// UserSession represents an authenticated user session
type UserSession struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Email         *string `json:"email,omitempty"`
	ProfileID     string  `json:"profile_id"`
	RoleID        *string `json:"role_id,omitempty"`
	IsSystemAdmin bool    `json:"is_system_admin"`
}

// ToMap converts UserSession to a map for formula context
func (u *UserSession) ToMap() map[string]interface{} {
	return map[string]interface{}{
		constants.FieldID:        u.ID,
		constants.FieldName:      u.Name,
		constants.FieldEmail:     u.Email,
		constants.FieldProfileID: u.ProfileID,
		constants.FieldRoleID:    u.RoleID,
	}
}

// IsSuperUser checks if the user has super user privileges
func (u *UserSession) IsSuperUser() bool {
	return constants.IsSuperUser(u.ProfileID)
}

// Profile represents a user profile
type Profile struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
	IsSystem    bool    `json:"is_system,omitempty"`
	IsSuperUser bool    `json:"is_super_user,omitempty"` // Super users bypass permission checks
}

func (p *Profile) ToSObject() SObject {
	b, _ := json.Marshal(p)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

type PermissionSet struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Label       string  `json:"label"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
}

type PermissionSetAssignment struct {
	ID              string `json:"id"`
	AssigneeID      string `json:"assignee_id"`
	PermissionSetID string `json:"permission_set_id"`
}

// Group represents a queue or user group
type Group struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Label string  `json:"label"`
	Type  string  `json:"type"` // Queue, Regular
	Email *string `json:"email,omitempty"`
}

// GroupMember represents membership in a group
type GroupMember struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
	UserID  string `json:"user_id"`
}

// Role represents a role in the hierarchy
type Role struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	ParentRoleID *string `json:"parent_role_id,omitempty"`
}

// ObjectPermission represents object-level permissions
type ObjectPermission struct {
	ProfileID       *string `json:"profile_id,omitempty"`
	PermissionSetID *string `json:"permission_set_id,omitempty"`
	ObjectAPIName   string  `json:"object_api_name"`
	AllowRead       bool    `json:"allow_read"`
	AllowCreate     bool    `json:"allow_create"`
	AllowEdit       bool    `json:"allow_edit"`
	AllowDelete     bool    `json:"allow_delete"`
	ViewAll         bool    `json:"view_all"`
	ModifyAll       bool    `json:"modify_all"`
}

// FieldPermission represents field-level permissions
type FieldPermission struct {
	ProfileID       *string `json:"profile_id,omitempty"`
	PermissionSetID *string `json:"permission_set_id,omitempty"`
	ObjectAPIName   string  `json:"object_api_name"`
	FieldAPIName    string  `json:"field_api_name"`
	Readable        bool    `json:"readable"`
	Editable        bool    `json:"editable"`
}

// SharingRule represents a sharing rule
type SharingRule struct {
	ID               string  `json:"id"`
	ObjectAPIName    string  `json:"object_api_name"`
	Name             string  `json:"name"`
	Criteria         string  `json:"criteria"`
	AccessLevel      string  `json:"access_level"` // Read, Edit
	ShareWithRoleID  *string `json:"share_with_role_id,omitempty"`
	ShareWithGroupID *string `json:"share_with_group_id,omitempty"`
}

// SystemUser represents a user record
type SystemUser struct {
	ID             string     `json:"id,omitempty"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Password       string     `json:"password,omitempty"`
	FirstName      string     `json:"first_name,omitempty"`
	LastName       string     `json:"last_name,omitempty"`
	ProfileID      string     `json:"profile_id"`
	RoleID         *string    `json:"role_id,omitempty"`
	IsActive       bool       `json:"is_active"`
	LastLoginDate  *time.Time `json:"last_login_date,omitempty"`
	OwnerID        string     `json:"owner_id,omitempty"`
	CreatedByID    string     `json:"created_by_id,omitempty"`
	LastModifiedID string     `json:"last_modified_by_id,omitempty"`
	CreatedDate    time.Time  `json:"created_date,omitempty"`
}

func (u *SystemUser) ToSObject() SObject {
	b, _ := json.Marshal(u)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// SystemSession represents a session record
type SystemSession struct {
	ID           string    `json:"id,omitempty"`
	UserID       string    `json:"user_id"`
	Token        string    `json:"token"`
	ExpiresAt    time.Time `json:"expires_at"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	IsRevoked    bool      `json:"is_revoked"`
	LastActivity time.Time `json:"last_activity"`
}

func (s *SystemSession) ToSObject() SObject {
	b, _ := json.Marshal(s)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}
