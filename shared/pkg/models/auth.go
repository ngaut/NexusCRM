package models

// UserSession represents an authenticated user session
type UserSession struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Email         *string `json:"email,omitempty"`
	ProfileID     string  `json:"profile_id"`
	RoleID        *string `json:"role_id,omitempty"`
	IsSystemAdmin bool    `json:"is_system_admin"`
}

// SystemPermissionSetAssignment - use generated
// type PermissionSetAssignment struct { ... }

// SystemGroupMember - use generated
// type GroupMember struct { ... }
