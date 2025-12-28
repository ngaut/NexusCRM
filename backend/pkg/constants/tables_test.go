package constants

import (
	"testing"
)

func TestIsSystemTable(t *testing.T) {
	tests := []struct {
		tableName string
		want      bool
	}{
		{"_System_Object", true},
		{"_System_Field", true},
		{"_System_User", true},
		{"_System_Custom", true},
		{"Custom_Object", false},
		{"User", false},
		{"system_lower", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tableName, func(t *testing.T) {
			if got := IsSystemTable(tt.tableName); got != tt.want {
				t.Errorf("IsSystemTable(%q) = %v, want %v", tt.tableName, got, tt.want)
			}
		})
	}
}
