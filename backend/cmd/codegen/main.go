// Package main provides a code generator that reads system_tables.json
// and generates Go constants, structs, and TypeScript types.
//
// This ensures a single source of truth for schema definitions.
// If system_tables.json changes, regenerate to get compile-time errors
// for any code using removed/renamed fields.
//
// Usage: go run ./cmd/codegen
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

// TableDefinition matches the JSON structure in system_tables.json
type TableDefinition struct {
	TableName   string       `json:"tableName"`
	TableType   string       `json:"tableType"`
	Category    string       `json:"category"`
	Description string       `json:"description"`
	Columns     []ColumnDef  `json:"columns"`
	Indices     []IndexDef   `json:"indices,omitempty"`
	ForeignKeys []ForeignKey `json:"foreignKeys,omitempty"`
}

type ColumnDef struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	PrimaryKey    bool     `json:"primaryKey,omitempty"`
	Nullable      bool     `json:"nullable,omitempty"`
	Unique        bool     `json:"unique,omitempty"`
	Default       string   `json:"default,omitempty"`
	AutoIncrement bool     `json:"autoIncrement,omitempty"`
	LogicalType   string   `json:"logicalType,omitempty"`
	ReferenceTo   []string `json:"referenceTo,omitempty"`
	IsNameField   bool     `json:"isNameField,omitempty"`
}

type IndexDef struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique,omitempty"`
}

type ForeignKey struct {
	Column     string `json:"column"`
	References string `json:"references"`
	OnDelete   string `json:"onDelete,omitempty"`
	OnUpdate   string `json:"onUpdate,omitempty"`
}

// Generation context
type genContext struct {
	tables    []TableDefinition
	timestamp string
}

var commonFieldNames = []string{
	// Primary/Core fields
	"__sys_gen_id",
	"name",
	"api_name",
	"label",
	"plural_label",
	"description",

	// Audit fields
	"__sys_gen_created_date",
	"__sys_gen_created_by_id",
	"__sys_gen_last_modified_date",
	"__sys_gen_last_modified_by_id",
	"__sys_gen_owner_id",
	"__sys_gen_is_deleted",
	"created_by",

	// User fields
	"email",
	"username",
	"password",
	"first_name",
	"last_name",
	"profile_id",
	"permission_set_id",
	"role_id",
	"is_active",
	"last_login_date",

	// Object/Field Metadata
	"type",
	"object_id",
	"reference_to",
	"subject",
	"is_custom",
	"is_system",
	"required",
	"unique",

	// Flow/Action fields
	"status",
	"sort_order",
	"condition",
	"config",

	// Approval fields
	"process_id",
	"work_item_id",
	"submitted_date",
	"approver_id",
	"approved_by_id",
	"approved_date",
	"comments",
	"entry_criteria",
	"approver_type",

	// Flow/Step fields
	"flow_id",
	"flow_instance_id",
	"flow_step_id",
	"current_step_id",
	"step_name",
	"step_order",
	"step_type",
	"action_type",
	"trigger_type",
	"trigger_object",
	"on_success_step",
	"on_failure_step",

	// Recycle Bin fields
	"record_id",
	"object_api_name",
	"record_name",
	"deleted_by",
	"deleted_date",

	// Recent Items fields
	"user_id",
	"timestamp",

	// Config fields
	"key_name",
	"value",
	"is_secret",

	// Log fields
	"level",
	"source",
	"message",
	"details",

	// Session fields
	"token",
	"expires_at",
	"last_activity",
	"ip_address",
	"user_agent",
	"is_revoked",

	// Metadata fields
	"table_name",
	"table_type",
	"category",
	"filters",
	"is_managed",
	"schema_version",
}

func main() {
	// Find project root by looking for system_tables.json
	jsonPath := findSystemTablesJSON()
	if jsonPath == "" {
		log.Fatal("❌ Could not find backend/internal/bootstrap/system_tables.json")
	}
	fmt.Printf("📖 Reading: %s\n", jsonPath)

	// Determine project root from json path
	// jsonPath is like: .../backend/internal/bootstrap/system_tables.json
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(jsonPath))))

	// Read and parse JSON
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("❌ Failed to read file: %v", err)
	}

	var tables []TableDefinition
	if err := json.Unmarshal(content, &tables); err != nil {
		log.Fatalf("❌ Failed to parse JSON: %v", err)
	}

	fmt.Printf("📊 Total Definitions: %d\n", len(tables))

	ctx := &genContext{
		tables:    tables,
		timestamp: time.Now().Format(time.RFC3339),
	}

	// Generate all outputs
	if err := generateGoTableConstants(ctx, projectRoot); err != nil {
		log.Fatalf("❌ Failed to generate table constants: %v", err)
	}

	if err := generateGoFieldConstants(ctx, projectRoot); err != nil {
		log.Fatalf("❌ Failed to generate field constants: %v", err)
	}

	if err := generateGoStructs(ctx, projectRoot); err != nil {
		log.Fatalf("❌ Failed to generate structs: %v", err)
	}

	if err := generateTypeScript(ctx, projectRoot); err != nil {
		log.Fatalf("❌ Failed to generate TypeScript: %v", err)
	}

	if err := generateSchemaDefinitions(ctx, projectRoot); err != nil {
		log.Fatalf("❌ Failed to generate SchemaDefinitions: %v", err)
	}

	if err := generateMCPTypes(ctx, projectRoot); err != nil {
		log.Fatalf("❌ Failed to generate MCP types: %v", err)
	}

	fmt.Println("\n🎉 Code generation complete!")
}

func findSystemTablesJSON() string {
	paths := []string{
		"backend/internal/bootstrap/system_tables.json",
		"internal/bootstrap/system_tables.json",
		"../internal/bootstrap/system_tables.json",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return ""
}

// ============================================================================
// Go Table Constants Generation
// ============================================================================

func generateGoTableConstants(ctx *genContext, projectRoot string) error {
	var sb strings.Builder

	sb.WriteString("// Code generated by cmd/codegen. DO NOT EDIT.\n")
	sb.WriteString("// Source: internal/bootstrap/system_tables.json\n")
	sb.WriteString("// Generated at: " + ctx.timestamp + "\n\n")
	sb.WriteString("package constants\n\n")

	sb.WriteString("// System Table Names\n")
	sb.WriteString("const (\n")

	// Sort tables for deterministic output
	sortedTables := make([]TableDefinition, len(ctx.tables))
	copy(sortedTables, ctx.tables)
	sort.Slice(sortedTables, func(i, j int) bool {
		return sortedTables[i].TableName < sortedTables[j].TableName
	})

	for _, t := range sortedTables {
		constName := tableNameToConstant(t.TableName)
		sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constName, t.TableName))
	}

	sb.WriteString(")\n\n")

	// Generate slice of all tables
	sb.WriteString("// AllSystemTableNames returns all system table names for validation\n")
	sb.WriteString("var AllSystemTableNames = []string{\n")
	for _, t := range sortedTables {
		constName := tableNameToConstant(t.TableName)
		sb.WriteString(fmt.Sprintf("\t%s,\n", constName))
	}
	sb.WriteString("}\n")

	outPath := filepath.Join(projectRoot, "shared", "pkg", "constants", "z_generated_tables.go")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	fmt.Printf("✅ Generated: %s (%d bytes)\n", outPath, sb.Len())
	return nil
}

// ============================================================================
// Go Field Constants Generation
// ============================================================================

func generateGoFieldConstants(ctx *genContext, projectRoot string) error {
	var sb strings.Builder

	sb.WriteString("// Code generated by cmd/codegen. DO NOT EDIT.\n")
	sb.WriteString("// Source: internal/bootstrap/system_tables.json\n")
	sb.WriteString("// Generated at: " + ctx.timestamp + "\n\n")
	sb.WriteString("package constants\n\n")

	// Generate common fields first (backward compatible)
	sb.WriteString("// Common field names (backward compatible)\n")
	sb.WriteString("const (\n")

	// Sort common fields for deterministic output (slice is already sorted in code if we keep it sorted, but let's sort to be safe)
	sort.Strings(commonFieldNames)

	for _, field := range commonFieldNames {
		// ALGORITHMIC NAMING: Field + PascalCase
		pascalName := snakeToPascal(field)
		constName := "Field" + pascalName

		sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constName, field))
	}
	sb.WriteString(")\n\n")

	// Generate table-specific field constants
	sortedTables := make([]TableDefinition, len(ctx.tables))
	copy(sortedTables, ctx.tables)
	sort.Slice(sortedTables, func(i, j int) bool {
		return sortedTables[i].TableName < sortedTables[j].TableName
	})

	for _, t := range sortedTables {
		tablePrefix := tableNameToPrefix(t.TableName)
		sb.WriteString(fmt.Sprintf("// %s fields\n", t.TableName))
		sb.WriteString("const (\n")

		// Sort columns for deterministic output
		sortedCols := make([]ColumnDef, len(t.Columns))
		copy(sortedCols, t.Columns)
		sort.Slice(sortedCols, func(i, j int) bool {
			return sortedCols[i].Name < sortedCols[j].Name
		})

		for _, col := range sortedCols {
			constName := fmt.Sprintf("Field%s_%s", tablePrefix, snakeToPascal(col.Name))
			sb.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constName, col.Name))
		}
		sb.WriteString(")\n\n")
	}

	outPath := filepath.Join(projectRoot, "shared", "pkg", "constants", "z_generated_fields.go")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	fmt.Printf("✅ Generated: %s (%d bytes)\n", outPath, sb.Len())
	return nil
}

// ============================================================================
// Go Struct Generation
// ============================================================================

func generateGoStructs(ctx *genContext, projectRoot string) error {
	var sb strings.Builder

	sb.WriteString("// Code generated by cmd/codegen. DO NOT EDIT.\n")
	sb.WriteString("// Source: internal/bootstrap/system_tables.json\n")
	sb.WriteString("// Generated at: " + ctx.timestamp + "\n\n")
	sb.WriteString("//go:generate go run ../../../cmd/codegen\n\n")
	sb.WriteString("package models\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"encoding/json\"\n")
	sb.WriteString("\t\"time\"\n")
	sb.WriteString(")\n\n")

	// Suppress unused import warning
	sb.WriteString("// Ensure imports are used\n")
	sb.WriteString("var (\n")
	sb.WriteString("\t_ json.RawMessage\n")
	sb.WriteString("\t_ time.Time\n")
	sb.WriteString(")\n\n")

	// Sort tables for deterministic output
	sortedTables := make([]TableDefinition, len(ctx.tables))
	copy(sortedTables, ctx.tables)
	sort.Slice(sortedTables, func(i, j int) bool {
		return sortedTables[i].TableName < sortedTables[j].TableName
	})

	for _, t := range sortedTables {
		structName := tableNameToStructName(t.TableName) // Removed Gen prefix for direct usage

		// Add description as comment
		if t.Description != "" {
			sb.WriteString(fmt.Sprintf("// %s represents the %s table (generated).\n", structName, t.TableName))
			sb.WriteString(fmt.Sprintf("// %s\n", t.Description))
		} else {
			sb.WriteString(fmt.Sprintf("// %s represents the %s table (generated).\n", structName, t.TableName))
		}

		sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

		for _, col := range t.Columns {
			fieldName := snakeToPascal(col.Name)
			goType := sqlTypeToGoType(col.Type, col.Nullable, col.LogicalType)
			jsonTag := col.Name
			omitEmpty := col.Nullable && !col.PrimaryKey

			// Special handling for password fields - exclude from JSON
			if col.LogicalType == "Password" {
				sb.WriteString(fmt.Sprintf("\t%s %s `json:\"-\"`\n", fieldName, goType))
			} else if omitEmpty {
				sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s,omitempty\"`\n", fieldName, goType, jsonTag))
			} else {
				sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, goType, jsonTag))
			}
		}

		sb.WriteString("}\n\n")

		// Generate GetTableName method (using Get prefix to avoid conflict with TableName field)
		sb.WriteString(fmt.Sprintf("// GetTableName returns the database table name for %s.\n", structName))
		sb.WriteString(fmt.Sprintf("func (%s) GetTableName() string {\n", structName))
		sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", t.TableName))
		sb.WriteString("}\n\n")
	}

	outPath := filepath.Join(projectRoot, "shared", "pkg", "models", "z_generated.go")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	fmt.Printf("✅ Generated: %s (%d bytes)\n", outPath, sb.Len())
	return nil
}

// ============================================================================
// TypeScript Generation
// ============================================================================

func generateTypeScript(ctx *genContext, projectRoot string) error {
	var sb strings.Builder

	sb.WriteString("// Code generated by cmd/codegen. DO NOT EDIT.\n")
	sb.WriteString("// Source: backend/internal/bootstrap/system_tables.json\n")
	sb.WriteString("// Generated at: " + ctx.timestamp + "\n\n")

	// Generate table name constants
	sb.WriteString("// ==================== System Table Names ====================\n\n")
	sb.WriteString("export const SYSTEM_TABLE_NAMES = {\n")

	sortedTables := make([]TableDefinition, len(ctx.tables))
	copy(sortedTables, ctx.tables)
	sort.Slice(sortedTables, func(i, j int) bool {
		return sortedTables[i].TableName < sortedTables[j].TableName
	})

	for _, t := range sortedTables {
		constKey := tableNameToTSKey(t.TableName)
		sb.WriteString(fmt.Sprintf("    %s: '%s',\n", constKey, t.TableName))
	}
	sb.WriteString("} as const;\n\n")
	sb.WriteString("export type SystemTableName = typeof SYSTEM_TABLE_NAMES[keyof typeof SYSTEM_TABLE_NAMES];\n\n")

	// Generate common fields first (TypeScript version)
	sb.WriteString("// ==================== Common Fields ====================\n\n")
	sb.WriteString("export const COMMON_FIELDS = {\n")

	sort.Strings(commonFieldNames)

	for _, field := range commonFieldNames {
		// Convert "FieldID" to "ID" for cleaner TS constants
		// Algorithm: snakeToPascal -> upper snake case?
		// e.g. "created_date" -> "CreatedDate" -> "CREATED_DATE"

		// Use helper
		pascalName := snakeToPascal(field) // "CreatedDate"

		// Handle legacy overrides

		// Convert Pascal to Snake for TS Constant Key: "CreatedDate" -> "CREATED_DATE"
		// Using pascalToSnake helper then Upper
		snakeName := pascalToSnake(pascalName)
		constName := strings.ToUpper(snakeName)

		// Special cases from original code:
		// if constName == "APIName" -> "API_NAME" (handled by pascalToSnake usually?)
		// pascalToSnake("APIName") -> "api_name"?
		// Let's check pascalToSnake implementation.
		// "API" -> "Api" -> "api".
		// So "APIName" -> "ApiName" -> "api_name".
		// Upper -> "API_NAME". Correct.

		sb.WriteString(fmt.Sprintf("    %s: '%s',\n", constName, field))
	}
	sb.WriteString("} as const;\n\n")

	// Generate field constants per table
	sb.WriteString("// ==================== Field Constants ====================\n\n")

	for _, t := range sortedTables {
		constName := fmt.Sprintf("FIELDS_%s", tableNameToTSKey(t.TableName))
		sb.WriteString(fmt.Sprintf("export const %s = {\n", constName))

		sortedCols := make([]ColumnDef, len(t.Columns))
		copy(sortedCols, t.Columns)
		sort.Slice(sortedCols, func(i, j int) bool {
			return sortedCols[i].Name < sortedCols[j].Name
		})

		for _, col := range sortedCols {
			// Generate clean key: __sys_gen_id -> ID
			pascal := snakeToPascal(col.Name) // ID
			snake := pascalToSnake(pascal)    // id
			key := strings.ToUpper(snake)     // ID

			sb.WriteString(fmt.Sprintf("    %s: '%s',\n", key, col.Name))
		}
		sb.WriteString("} as const;\n\n")
	}

	// Generate TypeScript interfaces
	sb.WriteString("// ==================== TypeScript Interfaces ====================\n\n")

	for _, t := range sortedTables {
		interfaceName := tableNameToStructName(t.TableName)

		if t.Description != "" {
			sb.WriteString(fmt.Sprintf("/** %s - %s */\n", t.TableName, t.Description))
		}

		sb.WriteString(fmt.Sprintf("export interface %s {\n", interfaceName))

		for _, col := range t.Columns {
			tsType := sqlTypeToTSType(col.Type, col.LogicalType)
			optional := col.Nullable && !col.PrimaryKey

			// Don't expose password fields
			if col.LogicalType == "Password" {
				continue
			}

			if optional {
				sb.WriteString(fmt.Sprintf("    %s?: %s;\n", col.Name, tsType))
			} else {
				sb.WriteString(fmt.Sprintf("    %s: %s;\n", col.Name, tsType))
			}

			// Generate friendly aliases for __sys_gen_* fields
			// This allows frontend code to use .id, .created_date, etc.
			if strings.HasPrefix(col.Name, "__sys_gen_") {
				alias := strings.TrimPrefix(col.Name, "__sys_gen_")
				// Always make aliases optional to avoid conflicts
				sb.WriteString(fmt.Sprintf("    %s?: %s; // Alias for %s\n", alias, tsType, col.Name))
			}
		}
		sb.WriteString("}\n\n")
	}

	outPath := filepath.Join(projectRoot, "frontend", "src", "generated-schema.ts")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	fmt.Printf("✅ Generated: %s (%d bytes)\n", outPath, sb.Len())
	return nil
}

// ============================================================================
// SchemaDefinitions.ts Generation (from shared/constants/*.json)
// ============================================================================

// SystemConfig matches shared/constants/system.json
type SystemConfig struct {
	Profiles         map[string]ProfileDef    `json:"profiles"`
	SystemFields     map[string]SystemField   `json:"systemFields"`
	SystemTables     map[string]string        `json:"systemTables"`
	Permissions      map[string]PermissionDef `json:"permissions"`
	LogLevels        map[string]string        `json:"logLevels"`
	ObjectCategories map[string]string        `json:"objectCategories"`
	Defaults         map[string]interface{}   `json:"defaults"`
}

type ProfileDef struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	IsSuperUser bool   `json:"is_super_user"`
}

type SystemField struct {
	APIName   string `json:"apiName"`
	Label     string `json:"label"`
	ProtoName string `json:"protoName"`
}

type PermissionDef struct {
	Value     string `json:"value"`
	ProtoName string `json:"protoName"`
}

// FieldTypeDef matches shared/constants/fieldTypes.json entries
type FieldTypeDef struct {
	SQLType           *string  `json:"sqlType"`
	Icon              string   `json:"icon"`
	Label             string   `json:"label"`
	Description       string   `json:"description"`
	IsSearchable      bool     `json:"isSearchable"`
	IsGroupable       bool     `json:"isGroupable"`
	IsSummable        bool     `json:"isSummable"`
	ValidationPattern *string  `json:"validationPattern,omitempty"`
	ValidationMessage *string  `json:"validationMessage,omitempty"`
	IsFK              *bool    `json:"isFK,omitempty"`
	IsVirtual         *bool    `json:"isVirtual,omitempty"`
	IsSystemOnly      *bool    `json:"isSystemOnly,omitempty"`
	Operators         []string `json:"operators"`
}

// OperatorDef matches shared/constants/operators.json entries
type OperatorDef struct {
	Label         string `json:"label"`
	Symbol        string `json:"symbol"`
	SQLOperator   string `json:"sqlOperator"`
	SQLPattern    string `json:"sqlPattern,omitempty"`
	RequiresRange bool   `json:"requiresRange,omitempty"`
	RequiresList  bool   `json:"requiresList,omitempty"`
	NoValue       bool   `json:"noValue,omitempty"`
}

func generateSchemaDefinitions(ctx *genContext, projectRoot string) error {
	var sb strings.Builder

	// Read system.json
	systemPath := filepath.Join(projectRoot, "shared", "constants", "system.json")
	systemData, err := os.ReadFile(systemPath)
	if err != nil {
		return fmt.Errorf("read system.json: %w", err)
	}
	var sysConfig SystemConfig
	if err := json.Unmarshal(systemData, &sysConfig); err != nil {
		return fmt.Errorf("parse system.json: %w", err)
	}

	// Read fieldTypes.json
	fieldTypesPath := filepath.Join(projectRoot, "shared", "constants", "fieldTypes.json")
	fieldTypesData, err := os.ReadFile(fieldTypesPath)
	if err != nil {
		return fmt.Errorf("read fieldTypes.json: %w", err)
	}
	var fieldTypes map[string]FieldTypeDef
	if err := json.Unmarshal(fieldTypesData, &fieldTypes); err != nil {
		return fmt.Errorf("parse fieldTypes.json: %w", err)
	}

	// Read operators.json
	operatorsPath := filepath.Join(projectRoot, "shared", "constants", "operators.json")
	operatorsData, err := os.ReadFile(operatorsPath)
	if err != nil {
		return fmt.Errorf("read operators.json: %w", err)
	}
	var operators map[string]OperatorDef
	if err := json.Unmarshal(operatorsData, &operators); err != nil {
		return fmt.Errorf("parse operators.json: %w", err)
	}

	// Header
	sb.WriteString("// Code generated by cmd/codegen. DO NOT EDIT.\n")
	sb.WriteString("// Source: shared/constants/*.json\n")
	sb.WriteString("// Generated at: " + ctx.timestamp + "\n\n")

	// ==================== Profiles ====================
	sb.WriteString("// ==================== Profiles ====================\n\n")

	// PROFILE_IDS
	sb.WriteString("export const PROFILE_IDS = {\n")
	var profileKeys []string
	for k := range sysConfig.Profiles {
		profileKeys = append(profileKeys, k)
	}
	sort.Strings(profileKeys)
	for _, k := range profileKeys {
		sb.WriteString(fmt.Sprintf("    %s: '%s',\n", k, sysConfig.Profiles[k].ID))
	}
	sb.WriteString("} as const;\n\n")
	sb.WriteString("export type ProfileId = typeof PROFILE_IDS[keyof typeof PROFILE_IDS];\n\n")

	// ProfileMetadata interface
	sb.WriteString("export interface ProfileMetadata {\n")
	sb.WriteString("    id: ProfileId;\n")
	sb.WriteString("    label: string;\n")
	sb.WriteString("    description: string;\n")
	sb.WriteString("    is_system: boolean;\n")
	sb.WriteString("    is_super_user: boolean;\n")
	sb.WriteString("}\n\n")

	// SYSTEM_PROFILES
	sb.WriteString("export const SYSTEM_PROFILES: Record<string, ProfileMetadata> = {\n")
	for _, k := range profileKeys {
		p := sysConfig.Profiles[k]
		sb.WriteString(fmt.Sprintf("    \"%s\": {\n", k))
		sb.WriteString(fmt.Sprintf("        \"id\": \"%s\",\n", p.ID))
		sb.WriteString(fmt.Sprintf("        \"label\": \"%s\",\n", p.Label))
		sb.WriteString(fmt.Sprintf("        \"description\": \"%s\",\n", p.Description))
		sb.WriteString(fmt.Sprintf("        \"is_system\": %t,\n", p.IsSystem))
		sb.WriteString(fmt.Sprintf("        \"is_super_user\": %t\n", p.IsSuperUser))
		sb.WriteString("    },\n")
	}
	sb.WriteString("};\n\n")

	// Helper functions for profiles
	sb.WriteString("export function isSuperUserProfile(profileId: string | undefined): boolean {\n")
	sb.WriteString("    if (!profileId) return false;\n")
	sb.WriteString("    const profile = SYSTEM_PROFILES[profileId];\n")
	sb.WriteString("    return profile?.is_super_user ?? false;\n")
	sb.WriteString("}\n\n")

	sb.WriteString("export function isSystemProfile(profileId: string | undefined): boolean {\n")
	sb.WriteString("    if (!profileId) return false;\n")
	sb.WriteString("    const profile = SYSTEM_PROFILES[profileId];\n")
	sb.WriteString("    return profile?.is_system ?? false;\n")
	sb.WriteString("}\n\n")

	// ==================== System Tables ====================
	sb.WriteString("// ==================== System Tables ====================\n")
	sb.WriteString("// Re-exported from generated-schema.ts for backward compatibility\n")
	sb.WriteString("// SYSTEM_TABLE_NAMES is the single source of truth (generated from backend)\n\n")

	sb.WriteString("import { SYSTEM_TABLE_NAMES, type SystemTableName } from '../../generated-schema';\n")
	sb.WriteString("export { SYSTEM_TABLE_NAMES, type SystemTableName };\n\n")

	sb.WriteString("export function isSystemTable(objectApiName: string): boolean {\n")
	sb.WriteString("    return objectApiName.startsWith('_System_');\n")
	sb.WriteString("}\n\n")

	// ==================== Custom Objects ====================
	sb.WriteString("// ==================== Custom Objects ====================\n\n")
	sb.WriteString("export function isCustomObject(objectApiName: string): boolean {\n")
	sb.WriteString("    return objectApiName.endsWith('__c');\n")
	sb.WriteString("}\n\n")

	// ==================== Object Categories ====================
	sb.WriteString("// ==================== Object Categories ====================\n\n")
	sb.WriteString("export const OBJECT_CATEGORIES = {\n")
	var catKeys []string
	for k := range sysConfig.ObjectCategories {
		catKeys = append(catKeys, k)
	}
	sort.Strings(catKeys)
	for _, k := range catKeys {
		sb.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", k, sysConfig.ObjectCategories[k]))
	}
	sb.WriteString("} as const;\n\n")
	sb.WriteString("export type ObjectCategory = typeof OBJECT_CATEGORIES[keyof typeof OBJECT_CATEGORIES];\n\n")

	// getObjectCategory function
	sb.WriteString("export function getObjectCategory(objectApiName: string): ObjectCategory {\n")
	sb.WriteString("    if (isSystemTable(objectApiName)) {\n")
	sb.WriteString("        const securityTables: SystemTableName[] = [\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_PROFILE,\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_ROLE,\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_OBJECTPERMS,\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_FIELDPERMS,\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_SHARINGRULE,\n")
	sb.WriteString("        ];\n\n")
	sb.WriteString("        const utilityTables: SystemTableName[] = [\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_RECYCLEBIN,\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_LOG,\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_RECENT,\n")
	sb.WriteString("            SYSTEM_TABLE_NAMES.SYSTEM_CONFIG,\n")
	sb.WriteString("        ];\n\n")
	sb.WriteString("        if (securityTables.includes(objectApiName as SystemTableName)) {\n")
	sb.WriteString("            return OBJECT_CATEGORIES.SECURITY;\n")
	sb.WriteString("        } else if (utilityTables.includes(objectApiName as SystemTableName)) {\n")
	sb.WriteString("            return OBJECT_CATEGORIES.UTILITY;\n")
	sb.WriteString("        }\n")
	sb.WriteString("        return OBJECT_CATEGORIES.SYSTEM_METADATA;\n")
	sb.WriteString("    }\n\n")
	sb.WriteString("    return OBJECT_CATEGORIES.BUSINESS_CUSTOM;\n")
	sb.WriteString("}\n\n")

	// ==================== Defaults ====================
	sb.WriteString("// ==================== Defaults ====================\n\n")
	sb.WriteString("export const DEFAULTS = {\n")
	var defKeys []string
	for k := range sysConfig.Defaults {
		defKeys = append(defKeys, k)
	}
	sort.Strings(defKeys)
	for _, k := range defKeys {
		v := sysConfig.Defaults[k]
		switch vt := v.(type) {
		case string:
			sb.WriteString(fmt.Sprintf("    \"%s\": \"%s\",\n", k, vt))
		case float64:
			sb.WriteString(fmt.Sprintf("    \"%s\": %v,\n", k, vt))
		default:
			sb.WriteString(fmt.Sprintf("    \"%s\": %v,\n", k, vt))
		}
	}
	sb.WriteString("} as const;\n\n")

	// ==================== Field Types ====================
	sb.WriteString("// ==================== Field Types ====================\n\n")

	// Generate FieldType union type
	var ftKeys []string
	for k := range fieldTypes {
		ftKeys = append(ftKeys, k)
	}
	sort.Strings(ftKeys)

	sb.WriteString("export type FieldType = ")
	for i, k := range ftKeys {
		if i > 0 {
			sb.WriteString(" | ")
		}
		sb.WriteString(fmt.Sprintf("'%s'", k))
	}
	sb.WriteString(";\n\n")

	// FieldTypeDefinition interface
	sb.WriteString("export interface FieldTypeDefinition {\n")
	sb.WriteString("    sqlType: string | null;\n")
	sb.WriteString("    icon: string;\n")
	sb.WriteString("    label: string;\n")
	sb.WriteString("    description: string;\n")
	sb.WriteString("    isSearchable: boolean;\n")
	sb.WriteString("    isGroupable: boolean;\n")
	sb.WriteString("    isSummable: boolean;\n")
	sb.WriteString("    validationPattern?: string;\n")
	sb.WriteString("    validationMessage?: string;\n")
	sb.WriteString("    isFK?: boolean;\n")
	sb.WriteString("    isVirtual?: boolean;\n")
	sb.WriteString("    isSystemOnly?: boolean;\n")
	sb.WriteString("    operators: string[];\n")
	sb.WriteString("}\n\n")

	// FIELD_TYPES
	sb.WriteString("export const FIELD_TYPES: Record<FieldType, FieldTypeDefinition> = {\n")
	for _, k := range ftKeys {
		ft := fieldTypes[k]
		sb.WriteString(fmt.Sprintf("    \"%s\": {\n", k))
		if ft.SQLType != nil {
			sb.WriteString(fmt.Sprintf("        \"sqlType\": \"%s\",\n", *ft.SQLType))
		} else {
			sb.WriteString("        \"sqlType\": null,\n")
		}
		sb.WriteString(fmt.Sprintf("        \"icon\": \"%s\",\n", ft.Icon))
		sb.WriteString(fmt.Sprintf("        \"label\": \"%s\",\n", ft.Label))
		sb.WriteString(fmt.Sprintf("        \"description\": \"%s\",\n", ft.Description))
		sb.WriteString(fmt.Sprintf("        \"isSearchable\": %t,\n", ft.IsSearchable))
		sb.WriteString(fmt.Sprintf("        \"isGroupable\": %t,\n", ft.IsGroupable))
		sb.WriteString(fmt.Sprintf("        \"isSummable\": %t,\n", ft.IsSummable))
		if ft.ValidationPattern != nil {
			// Escape backslashes for JSON
			escaped := strings.ReplaceAll(*ft.ValidationPattern, "\\", "\\\\")
			sb.WriteString(fmt.Sprintf("        \"validationPattern\": \"%s\",\n", escaped))
		}
		if ft.ValidationMessage != nil {
			sb.WriteString(fmt.Sprintf("        \"validationMessage\": \"%s\",\n", *ft.ValidationMessage))
		}
		if ft.IsFK != nil && *ft.IsFK {
			sb.WriteString("        \"isFK\": true,\n")
		}
		if ft.IsVirtual != nil && *ft.IsVirtual {
			sb.WriteString("        \"isVirtual\": true,\n")
		}
		if ft.IsSystemOnly != nil && *ft.IsSystemOnly {
			sb.WriteString("        \"isSystemOnly\": true,\n")
		}
		sb.WriteString("        \"operators\": [\n")
		for _, op := range ft.Operators {
			sb.WriteString(fmt.Sprintf("            \"%s\",\n", op))
		}
		sb.WriteString("        ]\n")
		sb.WriteString("    },\n")
	}
	sb.WriteString("};\n\n")

	// Helper functions for field types
	sb.WriteString("export function getSqlType(type: FieldType): string | null {\n")
	sb.WriteString("    return FIELD_TYPES[type]?.sqlType ?? null;\n")
	sb.WriteString("}\n\n")

	sb.WriteString("export function isSearchableType(type: string): boolean {\n")
	sb.WriteString("    return FIELD_TYPES[type as FieldType]?.isSearchable ?? false;\n")
	sb.WriteString("}\n\n")

	sb.WriteString("export function isGroupableType(type: string): boolean {\n")
	sb.WriteString("    return FIELD_TYPES[type as FieldType]?.isGroupable ?? false;\n")
	sb.WriteString("}\n\n")

	sb.WriteString("export function isSummableType(type: string): boolean {\n")
	sb.WriteString("    return FIELD_TYPES[type as FieldType]?.isSummable ?? false;\n")
	sb.WriteString("}\n\n")

	// ==================== Operators ====================
	sb.WriteString("// ==================== Operators ====================\n\n")

	sb.WriteString("export interface OperatorDefinition {\n")
	sb.WriteString("    label: string;\n")
	sb.WriteString("    symbol: string;\n")
	sb.WriteString("    sqlOperator: string;\n")
	sb.WriteString("    sqlPattern?: string;\n")
	sb.WriteString("    requiresRange?: boolean;\n")
	sb.WriteString("    requiresList?: boolean;\n")
	sb.WriteString("    noValue?: boolean;\n")
	sb.WriteString("}\n\n")

	sb.WriteString("export const OPERATORS: Record<string, OperatorDefinition> = {\n")
	var opKeys []string
	for k := range operators {
		opKeys = append(opKeys, k)
	}
	sort.Strings(opKeys)
	for _, k := range opKeys {
		op := operators[k]
		sb.WriteString(fmt.Sprintf("    \"%s\": {\n", k))
		sb.WriteString(fmt.Sprintf("        \"label\": \"%s\",\n", op.Label))
		sb.WriteString(fmt.Sprintf("        \"symbol\": \"%s\",\n", op.Symbol))
		sb.WriteString(fmt.Sprintf("        \"sqlOperator\": \"%s\",\n", op.SQLOperator))
		if op.SQLPattern != "" {
			sb.WriteString(fmt.Sprintf("        \"sqlPattern\": \"%s\",\n", op.SQLPattern))
		}
		if op.RequiresRange {
			sb.WriteString("        \"requiresRange\": true,\n")
		}
		if op.RequiresList {
			sb.WriteString("        \"requiresList\": true,\n")
		}
		if op.NoValue {
			sb.WriteString("        \"noValue\": true,\n")
		}
		sb.WriteString("    },\n")
	}
	sb.WriteString("};\n\n")

	sb.WriteString("export function getOperatorsForType(type: FieldType): string[] {\n")
	sb.WriteString("    return FIELD_TYPES[type]?.operators ?? [];\n")
	sb.WriteString("}\n")

	// Write file
	outPath := filepath.Join(projectRoot, "frontend", "src", "core", "constants", "SchemaDefinitions.ts")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write SchemaDefinitions.ts: %w", err)
	}
	fmt.Printf("✅ Generated: %s (%d bytes)\n", outPath, sb.Len())
	return nil
}

// ============================================================================
// MCP Types Generation
// ============================================================================

func generateMCPTypes(ctx *genContext, projectRoot string) error {
	var sb strings.Builder

	sb.WriteString("// Code generated by cmd/codegen. DO NOT EDIT.\n")
	sb.WriteString("// Source: backend/internal/bootstrap/system_tables.json\n")
	sb.WriteString("// Generated at: " + ctx.timestamp + "\n\n")
	sb.WriteString("package models\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"encoding/json\"\n")
	sb.WriteString("\t\"time\"\n")
	sb.WriteString(")\n\n")

	// Suppress unused import warning
	sb.WriteString("// Ensure imports are used\n")
	sb.WriteString("var (\n")
	sb.WriteString("\t_ json.RawMessage\n")
	sb.WriteString("\t_ time.Time\n")
	sb.WriteString(")\n\n")

	// For MCP, we generate a subset of commonly used types
	// Focus on types that MCP tools actually interact with
	mcpRelevantTables := map[string]bool{
		"_System_User":       true,
		"_System_Object":     true,
		"_System_Field":      true,
		"_System_Dashboard":  true,
		"_System_Profile":    true,
		"_System_Role":       true,
		"_System_Validation": true,
		"_System_Flow":       true,
		"_System_App":        true,
		"_System_Layout":     true,
		"_System_ListView":   true,
	}

	sortedTables := make([]TableDefinition, len(ctx.tables))
	copy(sortedTables, ctx.tables)
	sort.Slice(sortedTables, func(i, j int) bool {
		return sortedTables[i].TableName < sortedTables[j].TableName
	})

	for _, t := range sortedTables {
		if !mcpRelevantTables[t.TableName] {
			continue
		}

		structName := tableNameToStructName(t.TableName) // Removed Gen prefix

		if t.Description != "" {
			sb.WriteString(fmt.Sprintf("// %s represents the %s table (generated).\n", structName, t.TableName))
			sb.WriteString(fmt.Sprintf("// %s\n", t.Description))
		} else {
			sb.WriteString(fmt.Sprintf("// %s represents the %s table (generated).\n", structName, t.TableName))
		}

		sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

		for _, col := range t.Columns {
			// Skip password fields for MCP
			if col.LogicalType == "Password" {
				continue
			}

			fieldName := snakeToPascal(col.Name)
			goType := sqlTypeToGoType(col.Type, col.Nullable, col.LogicalType)
			jsonTag := col.Name
			omitEmpty := col.Nullable && !col.PrimaryKey

			if omitEmpty {
				sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s,omitempty\"`\n", fieldName, goType, jsonTag))
			} else {
				sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, goType, jsonTag))
			}
		}

		sb.WriteString("}\n\n")
	}

	outPath := filepath.Join(projectRoot, "mcp", "pkg", "models", "z_generated.go")
	if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	fmt.Printf("✅ Generated: %s (%d bytes)\n", outPath, sb.Len())
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// tableNameToConstant converts "_System_ApprovalWorkItem" to "TableApprovalWorkItem"
// and "my_table" to "TableMyTable"
func tableNameToConstant(name string) string {
	clean := strings.TrimPrefix(name, "_System_")
	clean = strings.TrimPrefix(clean, "_")
	// Table names are already PascalCase (e.g., ApprovalWorkItem), just prefix with Table
	return "Table" + snakeToPascal(clean)
}

// tableNameToPrefix converts "_System_User" to "SysUser" and "account" to "Account"
func tableNameToPrefix(name string) string {
	if strings.HasPrefix(name, "_System_") {
		clean := strings.TrimPrefix(name, "_System_")
		return "Sys" + snakeToPascal(clean)
	}
	// For business objects (custom or legacy), just use PascalCase name
	return snakeToPascal(name)
}

// tableNameToStructName converts "_System_User" to "SystemUser" and "account" to "Account"
func tableNameToStructName(name string) string {
	if strings.HasPrefix(name, "_System_") {
		clean := strings.TrimPrefix(name, "_")
		clean = strings.ReplaceAll(clean, "_", "")
		// Handle special cases
		clean = strings.ReplaceAll(clean, "System", "System")
		return clean
	}
	return snakeToPascal(name)
}

// tableNameToTSKey converts "_System_User" to "SYSTEM_USER" and "account" to "ACCOUNT"
func tableNameToTSKey(name string) string {
	clean := strings.TrimPrefix(name, "_")
	clean = strings.ReplaceAll(clean, "_", "_")
	return strings.ToUpper(clean)
}

// snakeToPascal converts "created_date" to "CreatedDate"
func snakeToPascal(s string) string {
	// Strip system prefix for clean Go/TS identifiers
	s = strings.TrimPrefix(s, "__sys_gen_")

	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	result := strings.Join(parts, "")

	// Handle common abbreviations
	result = strings.ReplaceAll(result, "Id", "ID")
	result = strings.ReplaceAll(result, "Api", "API")
	result = strings.ReplaceAll(result, "Ui", "UI")
	result = strings.ReplaceAll(result, "Url", "URL")
	result = strings.ReplaceAll(result, "Html", "HTML")
	result = strings.ReplaceAll(result, "Json", "JSON")
	result = strings.ReplaceAll(result, "Sql", "SQL")
	result = strings.ReplaceAll(result, "Ip", "IP")

	return result
}

// pascalToSnake converts "CreatedDate" to "created_date" and "ID" to "id"
func pascalToSnake(s string) string {
	// Handle special acronyms
	s = strings.ReplaceAll(s, "ID", "Id")
	s = strings.ReplaceAll(s, "API", "Api")
	s = strings.ReplaceAll(s, "UI", "Ui")
	s = strings.ReplaceAll(s, "URL", "Url")
	s = strings.ReplaceAll(s, "IP", "Ip")
	s = strings.ReplaceAll(s, "JSON", "Json")
	s = strings.ReplaceAll(s, "HTML", "Html")

	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// sqlTypeToGoType converts SQL types to Go types
func sqlTypeToGoType(sqlType string, nullable bool, logicalType string) string {
	sqlType = strings.ToUpper(sqlType)

	// Handle ENUM types
	if strings.HasPrefix(sqlType, "ENUM") {
		if nullable {
			return "*string"
		}
		return "string"
	}

	// Extract base type (e.g., VARCHAR(255) -> VARCHAR)
	baseType := regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(sqlType, "")
	baseType = strings.TrimSpace(baseType)

	var goType string
	switch baseType {
	case "VARCHAR", "TEXT", "CHAR", "LONGTEXT", "MEDIUMTEXT":
		goType = "string"
	case "INT", "INTEGER", "SMALLINT", "MEDIUMINT":
		goType = "int"
	case "BIGINT":
		goType = "int64"
	case "TINYINT":
		// TINYINT(1) is typically boolean
		if strings.Contains(sqlType, "(1)") {
			goType = "bool"
		} else {
			goType = "int8"
		}
	case "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE":
		goType = "float64"
	case "DATETIME", "TIMESTAMP", "DATE", "TIME":
		goType = "time.Time"
	case "BOOLEAN", "BOOL":
		goType = "bool"
	case "JSON":
		goType = "json.RawMessage"
	case "BLOB", "LONGBLOB", "MEDIUMBLOB":
		goType = "[]byte"
	default:
		goType = "string" // fallback
	}

	// Handle nullable types (use pointers)
	if nullable && goType != "bool" && goType != "json.RawMessage" && goType != "[]byte" {
		return "*" + goType
	}

	return goType
}

// sqlTypeToTSType converts SQL types to TypeScript types
func sqlTypeToTSType(sqlType string, logicalType string) string {
	sqlType = strings.ToUpper(sqlType)

	// Handle ENUM types
	if strings.HasPrefix(sqlType, "ENUM") {
		return "string"
	}

	baseType := regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(sqlType, "")
	baseType = strings.TrimSpace(baseType)

	switch baseType {
	case "VARCHAR", "TEXT", "CHAR", "LONGTEXT", "MEDIUMTEXT":
		return "string"
	case "INT", "INTEGER", "SMALLINT", "MEDIUMINT", "BIGINT", "TINYINT":
		// TINYINT(1) is typically boolean
		if strings.Contains(sqlType, "(1)") && baseType == "TINYINT" {
			return "boolean"
		}
		return "number"
	case "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE":
		return "number"
	case "DATETIME", "TIMESTAMP", "DATE", "TIME":
		return "string" // ISO date strings in JSON
	case "BOOLEAN", "BOOL":
		return "boolean"
	case "JSON":
		return "Record<string, unknown>"
	default:
		return "unknown"
	}
}
