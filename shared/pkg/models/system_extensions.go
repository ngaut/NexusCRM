package models

import (
	"encoding/json"

	"github.com/nexuscrm/shared/pkg/constants"
)

// ToSObject converts SystemAuditLog to SObject
func (a *SystemAuditLog) ToSObject() SObject {
	b, _ := json.Marshal(a)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemLog to SObject
func (l *SystemLog) ToSObject() SObject {
	b, _ := json.Marshal(l)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemConfig to SObject
func (c *SystemConfig) ToSObject() SObject {
	b, _ := json.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemTable to SObject
func (t *SystemTable) ToSObject() SObject {
	b, _ := json.Marshal(t)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemNotification to SObject
func (n *SystemNotification) ToSObject() SObject {
	b, _ := json.Marshal(n)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemUser to SObject
func (u *SystemUser) ToSObject() SObject {
	b, _ := json.Marshal(u)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	// Manually add password as it is hidden from JSON
	if u.Password != "" {
		m["password"] = u.Password
	}
	return m
}

// ToSObject converts SystemSession to SObject
func (s *SystemSession) ToSObject() SObject {
	b, _ := json.Marshal(s)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemComment to SObject
func (c *SystemComment) ToSObject() SObject {
	b, _ := json.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemFile to SObject
func (f *SystemFile) ToSObject() SObject {
	b, _ := json.Marshal(f)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemProfile to SObject
func (p *SystemProfile) ToSObject() SObject {
	b, _ := json.Marshal(p)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemRecent to SObject
func (r *SystemRecent) ToSObject() SObject {
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemRole to SObject
func (r *SystemRole) ToSObject() SObject {
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemGroup to SObject
func (g *SystemGroup) ToSObject() SObject {
	b, _ := json.Marshal(g)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemPermissionSet to SObject
func (p *SystemPermissionSet) ToSObject() SObject {
	b, _ := json.Marshal(p)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemSharingRule to SObject
func (s *SystemSharingRule) ToSObject() SObject {
	b, _ := json.Marshal(s)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemObjectPerms to SObject
func (p *SystemObjectPerms) ToSObject() SObject {
	b, _ := json.Marshal(p)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// ToSObject converts SystemFieldPerms to SObject
func (p *SystemFieldPerms) ToSObject() SObject {
	b, _ := json.Marshal(p)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	return m
}

// UserSession Helper Methods

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
