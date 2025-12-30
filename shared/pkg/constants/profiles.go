package constants

// Profile IDs - sourced from shared/constants/system.json
// Run `./scripts/typegen.sh` to regenerate after modifying system.json
const (
	ProfileSystemAdmin  = "system_admin"
	ProfileStandardUser = "standard_user"
)

// IsSystemAdmin checks if a profile ID is the system admin profile
func IsSystemAdmin(profileID string) bool {
	return profileID == ProfileSystemAdmin
}

// IsStandardUser checks if a profile ID is the standard user profile
func IsStandardUser(profileID string) bool {
	return profileID == ProfileStandardUser
}

// IsSuperUser checks if a profile ID has super user privileges
func IsSuperUser(profileID string) bool {
	return profileID == ProfileSystemAdmin
}
