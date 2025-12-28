package utils

import (
	"log"

	"github.com/google/uuid"
)

// GenerateID generates a new UUID v4 string
func GenerateID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Failed to generate UUID: %v", err)
		return ""
	}
	return id.String()
}

// IsValidUUID checks if the string is a valid UUID
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
