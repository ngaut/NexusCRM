package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
)

func main() {
	// Test ID Generation
	userTable := "_System_User"
	actionTable := "_System_Action"

	// Simulate what SchemaRepository calls
	// It calls GenerateObjectID matching persistence/schema_utils.go
	// But we need to use the one from persistence package which we imported

	uid := persistence.GenerateObjectID(userTable)
	aid := persistence.GenerateObjectID(actionTable)

	fmt.Printf("User Table ID:   %s\n", uid)
	fmt.Printf("Action Table ID: %s\n", aid)

	if uid == aid {
		fmt.Println("❌ COLLISION DETECTED!")
	} else {
		fmt.Println("✅ IDs are distinct.")
	}

	// Test JSON Parsing
	content, err := os.ReadFile("internal/bootstrap/system_tables.json")
	if err != nil {
		panic(err)
	}

	var tables []schema.TableDefinition
	if err := json.Unmarshal(content, &tables); err != nil {
		panic(err)
	}

	fmt.Printf("Parsed %d tables.\n", len(tables))

	var userDef schema.TableDefinition
	found := false
	for _, t := range tables {
		if t.TableName == "_System_User" {
			userDef = t
			found = true
			break
		}
	}

	if !found {
		fmt.Println("❌ _System_User NOT FOUND in JSON.")
		return
	}

	fmt.Println("--- _System_User Columns ---")
	hasRole := false
	hasConfig := false
	for _, c := range userDef.Columns {
		fmt.Printf("- %s (%s)\n", c.Name, c.Type)
		if c.Name == "role_id" {
			hasRole = true
		}
		if c.Name == "config" {
			hasConfig = true
		}
	}

	if hasRole {
		fmt.Println("✅ role_id found in parsed definition.")
	} else {
		fmt.Println("❌ role_id NOT found in parsing.")
	}
	if hasConfig {
		fmt.Println("❌ config field wrongly found in _System_User!")
	} else {
		fmt.Println("✅ config field correctly usage (not present).")
	}
}
