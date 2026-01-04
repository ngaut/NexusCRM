/**
 * AUTO-GENERATED FILE - DO NOT EDIT
 * Generated from shared/constants/*.json
 * Run: node scripts/generate-ts-constants.js
 */

// ==================== Profiles ====================

export const PROFILE_IDS = {
    SYSTEM_ADMIN: 'system_admin',
    STANDARD_USER: 'standard_user'
} as const;

export type ProfileId = typeof PROFILE_IDS[keyof typeof PROFILE_IDS];

export interface ProfileMetadata {
    id: ProfileId;
    label: string;
    description: string;
    is_system: boolean;
    is_super_user: boolean;
}

export const SYSTEM_PROFILES: Record<string, ProfileMetadata> = {
    [PROFILE_IDS.SYSTEM_ADMIN]: {
        "id": "system_admin",
        "label": "System Administrator",
        "description": "Full access to all platform features and data. Can modify system configuration.",
        "is_system": true,
        "is_super_user": true
    },
    [PROFILE_IDS.STANDARD_USER]: {
        "id": "standard_user",
        "label": "Standard User",
        "description": "Default user profile with standard permissions as defined by administrators.",
        "is_system": true,
        "is_super_user": false
    }
};

export function isSuperUserProfile(profileId: string | undefined): boolean {
    if (!profileId) return false;
    const profile = SYSTEM_PROFILES[profileId];
    return profile?.is_super_user ?? false;
}

export function isSystemProfile(profileId: string | undefined): boolean {
    if (!profileId) return false;
    const profile = SYSTEM_PROFILES[profileId];
    return profile?.is_system ?? false;
}

// ==================== System Tables ====================

export const SYSTEM_TABLES = {
    OBJECT: '_System_Object',
    FIELD: '_System_Field',
    RELATIONSHIP: '_System_Relationship',
    RECORD_TYPE: '_System_RecordType',
    APP: '_System_App',
    TAB: '_System_Tab',
    LAYOUT: '_System_Layout',
    PROFILE_LAYOUT: '_System_ProfileLayout',
    DASHBOARD: '_System_Dashboard',
    LIST_VIEW: '_System_ListView',
    SETUP_PAGE: '_System_SetupPage',
    UI_THEME: '_System_UITheme',
    UI_COMPONENT: '_System_UIComponent',
    FIELD_RENDERING: '_System_FieldRendering',
    NAVIGATION_MENU: '_System_NavigationMenu',
    FLOW: '_System_Flow',
    ACTION_HANDLER: '_System_ActionHandler',
    FORMULA_FUNCTION: '_System_FormulaFunction',
    WEBHOOK: '_System_Webhook',
    EMAIL_TEMPLATE: '_System_EmailTemplate',
    API_ENDPOINT: '_System_ApiEndpoint',
    VALIDATION: '_System_Validation',
    TRANSFORMATION: '_System_Transformation',
    ACTION: '_System_Action',
    PROFILE: '_System_Profile',
    ROLE: '_System_Role',
    USER: '_System_User',
    OBJECT_PERMS: '_System_ObjectPerms',
    FIELD_PERMS: '_System_FieldPerms',
    SHARING_RULE: '_System_SharingRule',
    RECYCLE_BIN: '_System_RecycleBin',
    LOG: '_System_Log',
    RECENT: '_System_Recent',
    CONFIG: '_System_Config',
    LIMIT: '_System_Limit',
    PROMPT: '_System_Prompt'
} as const;

export type SystemTableName = typeof SYSTEM_TABLES[keyof typeof SYSTEM_TABLES];

export function isSystemTable(objectApiName: string): boolean {
    return objectApiName.startsWith('_System_');
}

// ==================== Custom Objects ====================

// They are created dynamically via the Metadata API as part of the pure meta-driven architecture.

export function isCustomObject(objectApiName: string): boolean {
    return objectApiName.endsWith('__c');
}

// ==================== Object Categories ====================

export const OBJECT_CATEGORIES = {
    SYSTEM_METADATA: 'system_metadata',
    BUSINESS_CUSTOM: 'business_custom',
    SECURITY: 'security',
    UTILITY: 'utility'
} as const;

export type ObjectCategory = typeof OBJECT_CATEGORIES[keyof typeof OBJECT_CATEGORIES];

export function getObjectCategory(objectApiName: string): ObjectCategory {
    if (isSystemTable(objectApiName)) {
        const securityTables = [
            SYSTEM_TABLES.PROFILE,
            SYSTEM_TABLES.ROLE,
            SYSTEM_TABLES.OBJECT_PERMS,
            SYSTEM_TABLES.FIELD_PERMS,
            SYSTEM_TABLES.SHARING_RULE,
        ];

        const utilityTables = [
            SYSTEM_TABLES.RECYCLE_BIN,
            SYSTEM_TABLES.LOG,
            SYSTEM_TABLES.RECENT,
            SYSTEM_TABLES.CONFIG,
        ];

        if (securityTables.includes(objectApiName as any)) {
            return OBJECT_CATEGORIES.SECURITY;
        } else if (utilityTables.includes(objectApiName as any)) {
            return OBJECT_CATEGORIES.UTILITY;
        }
        return OBJECT_CATEGORIES.SYSTEM_METADATA;
    }

    // All non-system objects are now treated as custom business objects
    return OBJECT_CATEGORIES.BUSINESS_CUSTOM;
}

// ==================== Defaults ====================

export const DEFAULTS = {
    userName: 'Unknown',
    userEmail: 'unknown@example.com',
    pageSize: 25,
    maxPageSize: 1000
} as const;

// ==================== Field Types ====================

export type FieldType = 'Text' | 'TextArea' | 'LongTextArea' | 'RichText' | 'Number' | 'Currency' | 'Percent' | 'Date' | 'DateTime' | 'Boolean' | 'Picklist' | 'Email' | 'Phone' | 'Url' | 'Lookup' | 'Formula' | 'RollupSummary' | 'JSON' | 'Password';

export interface FieldTypeDefinition {
    sqlType: string | null;
    icon: string;
    label: string;
    description: string;
    isSearchable: boolean;
    isGroupable: boolean;
    isSummable: boolean;
    validationPattern?: string;
    validationMessage?: string;
    isFK?: boolean;
    isVirtual?: boolean;
    isSystemOnly?: boolean;
    operators: string[];
}

export const FIELD_TYPES: Record<FieldType, FieldTypeDefinition> = {
    "Text": {
        "sqlType": "VARCHAR(255)",
        "icon": "Type",
        "label": "Text",
        "description": "Text field with maximum 255 characters",
        "isSearchable": true,
        "isGroupable": false,
        "isSummable": false,
        "operators": [
            "equals",
            "not_equals",
            "contains",
            "not_contains",
            "starts_with",
            "ends_with",
            "is_null",
            "is_not_null"
        ]
    },
    "TextArea": {
        "sqlType": "TEXT",
        "icon": "AlignLeft",
        "label": "Text Area",
        "description": "Multi-line text field",
        "isSearchable": true,
        "isGroupable": false,
        "isSummable": false,
        "operators": [
            "contains",
            "not_contains",
            "is_null",
            "is_not_null"
        ]
    },
    "LongTextArea": {
        "sqlType": "LONGTEXT",
        "icon": "FileText",
        "label": "Long Text Area",
        "description": "Large text field for extensive content",
        "isSearchable": false,
        "isGroupable": false,
        "isSummable": false,
        "operators": [
            "contains",
            "not_contains",
            "is_null",
            "is_not_null"
        ]
    },
    "RichText": {
        "sqlType": "LONGTEXT",
        "icon": "Edit3",
        "label": "Rich Text",
        "description": "HTML-formatted text content",
        "isSearchable": false,
        "isGroupable": false,
        "isSummable": false,
        "operators": [
            "contains",
            "not_contains",
            "is_null",
            "is_not_null"
        ]
    },
    "Number": {
        "sqlType": "DECIMAL(18,2)",
        "icon": "Hash",
        "label": "Number",
        "description": "Numeric field with decimal support",
        "isSearchable": false,
        "isGroupable": true,
        "isSummable": true,
        "operators": [
            "equals",
            "not_equals",
            "greater_than",
            "greater_or_equal",
            "less_than",
            "less_or_equal",
            "between",
            "is_null",
            "is_not_null"
        ]
    },
    "Currency": {
        "sqlType": "DECIMAL(18,2)",
        "icon": "DollarSign",
        "label": "Currency",
        "description": "Currency amount with 2 decimal places",
        "isSearchable": false,
        "isGroupable": true,
        "isSummable": true,
        "operators": [
            "equals",
            "not_equals",
            "greater_than",
            "greater_or_equal",
            "less_than",
            "less_or_equal",
            "between",
            "is_null",
            "is_not_null"
        ]
    },
    "Percent": {
        "sqlType": "DECIMAL(5,2)",
        "icon": "Percent",
        "label": "Percent",
        "description": "Percentage value",
        "isSearchable": false,
        "isGroupable": true,
        "isSummable": true,
        "operators": [
            "equals",
            "not_equals",
            "greater_than",
            "greater_or_equal",
            "less_than",
            "less_or_equal",
            "between",
            "is_null",
            "is_not_null"
        ]
    },
    "Date": {
        "sqlType": "DATE",
        "icon": "Calendar",
        "label": "Date",
        "description": "Date without time component",
        "isSearchable": false,
        "isGroupable": true,
        "isSummable": false,
        "operators": [
            "equals",
            "not_equals",
            "greater_than",
            "greater_or_equal",
            "less_than",
            "less_or_equal",
            "between",
            "is_null",
            "is_not_null"
        ]
    },
    "DateTime": {
        "sqlType": "DATETIME",
        "icon": "Clock",
        "label": "Date/Time",
        "description": "Date with time component",
        "isSearchable": false,
        "isGroupable": false,
        "isSummable": false,
        "operators": [
            "equals",
            "not_equals",
            "greater_than",
            "greater_or_equal",
            "less_than",
            "less_or_equal",
            "between",
            "is_null",
            "is_not_null"
        ]
    },
    "Boolean": {
        "sqlType": "BOOLEAN",
        "icon": "ToggleLeft",
        "label": "Checkbox",
        "description": "True/false toggle",
        "isSearchable": false,
        "isGroupable": true,
        "isSummable": false,
        "operators": [
            "equals",
            "not_equals"
        ]
    },
    "Picklist": {
        "sqlType": "VARCHAR(255)",
        "icon": "List",
        "label": "Picklist",
        "description": "Single selection from predefined options",
        "isSearchable": true,
        "isGroupable": true,
        "isSummable": false,
        "operators": [
            "equals",
            "not_equals",
            "in",
            "not_in",
            "is_null",
            "is_not_null"
        ]
    },
    "Email": {
        "sqlType": "VARCHAR(255)",
        "icon": "Mail",
        "label": "Email",
        "description": "Email address with validation",
        "isSearchable": true,
        "isGroupable": false,
        "isSummable": false,
        "validationPattern": "^[a-zA-Z0-9._%+\\-]+@[a-zA-Z0-9.\\-]+\\.[a-zA-Z]{2,}$",
        "validationMessage": "invalid email format",
        "operators": [
            "equals",
            "not_equals",
            "contains",
            "starts_with",
            "ends_with",
            "is_null",
            "is_not_null"
        ]
    },
    "Phone": {
        "sqlType": "VARCHAR(40)",
        "icon": "Phone",
        "label": "Phone",
        "description": "Phone number",
        "isSearchable": true,
        "isGroupable": false,
        "isSummable": false,
        "validationPattern": "^\\+?[1-9]\\d{1,14}$",
        "validationMessage": "invalid phone format",
        "operators": [
            "equals",
            "not_equals",
            "contains",
            "starts_with",
            "is_null",
            "is_not_null"
        ]
    },
    "Url": {
        "sqlType": "VARCHAR(1024)",
        "icon": "Link",
        "label": "URL",
        "description": "Web address",
        "isSearchable": true,
        "isGroupable": false,
        "isSummable": false,
        "validationPattern": "^https?://[^\\s/$.?#].[^\\s]*$",
        "validationMessage": "invalid URL format",
        "operators": [
            "equals",
            "not_equals",
            "contains",
            "starts_with",
            "is_null",
            "is_not_null"
        ]
    },
    "Lookup": {
        "sqlType": "VARCHAR(36)",
        "icon": "ExternalLink",
        "label": "Lookup",
        "description": "Reference to another object record",
        "isSearchable": true,
        "isGroupable": true,
        "isSummable": false,
        "isFK": true,
        "operators": [
            "equals",
            "not_equals",
            "is_null",
            "is_not_null"
        ]
    },
    "Formula": {
        "sqlType": null,
        "icon": "Zap",
        "label": "Formula",
        "description": "Calculated field based on expression",
        "isSearchable": false,
        "isGroupable": false,
        "isSummable": false,
        "isVirtual": true,
        "isSystemOnly": true,
        "operators": []
    },
    "RollupSummary": {
        "sqlType": null,
        "icon": "Sigma",
        "label": "Rollup Summary",
        "description": "Aggregation of child records",
        "isSearchable": false,
        "isGroupable": false,
        "isSummable": false,
        "isVirtual": true,
        "isSystemOnly": true,
        "operators": []
    },
    "JSON": {
        "sqlType": "JSON",
        "icon": "Code",
        "label": "JSON",
        "description": "Structured JSON data",
        "isSearchable": false,
        "isGroupable": false,
        "isSummable": false,
        "operators": [
            "is_null",
            "is_not_null"
        ]
    },
    "Password": {
        "sqlType": "VARCHAR(255)",
        "icon": "Lock",
        "label": "Password",
        "description": "Encrypted password field",
        "isSearchable": false,
        "isGroupable": false,
        "isSummable": false,
        "isSystemOnly": true,
        "operators": []
    }
};

export function getSqlType(type: FieldType): string | null {
    return FIELD_TYPES[type]?.sqlType ?? null;
}

export function isSearchableType(type: string): boolean {
    return FIELD_TYPES[type as FieldType]?.isSearchable ?? false;
}

export function isGroupableType(type: string): boolean {
    return FIELD_TYPES[type as FieldType]?.isGroupable ?? false;
}

export function isSummableType(type: string): boolean {
    return FIELD_TYPES[type as FieldType]?.isSummable ?? false;
}

// ==================== Operators ====================

export interface OperatorDefinition {
    label: string;
    symbol: string;
    sqlOperator: string;
    sqlPattern?: string;
    requiresRange?: boolean;
    requiresList?: boolean;
    noValue?: boolean;
}

export const OPERATORS: Record<string, OperatorDefinition> = {
    "equals": {
        "label": "equals",
        "symbol": "=",
        "sqlOperator": "="
    },
    "not_equals": {
        "label": "not equal to",
        "symbol": "≠",
        "sqlOperator": "!="
    },
    "contains": {
        "label": "contains",
        "symbol": "∋",
        "sqlOperator": "LIKE",
        "sqlPattern": "%{value}%"
    },
    "not_contains": {
        "label": "does not contain",
        "symbol": "∌",
        "sqlOperator": "NOT LIKE",
        "sqlPattern": "%{value}%"
    },
    "starts_with": {
        "label": "starts with",
        "symbol": "^",
        "sqlOperator": "LIKE",
        "sqlPattern": "{value}%"
    },
    "ends_with": {
        "label": "ends with",
        "symbol": "$",
        "sqlOperator": "LIKE",
        "sqlPattern": "%{value}"
    },
    "greater_than": {
        "label": "greater than",
        "symbol": ">",
        "sqlOperator": ">"
    },
    "greater_or_equal": {
        "label": "greater or equal",
        "symbol": "≥",
        "sqlOperator": ">="
    },
    "less_than": {
        "label": "less than",
        "symbol": "<",
        "sqlOperator": "<"
    },
    "less_or_equal": {
        "label": "less or equal",
        "symbol": "≤",
        "sqlOperator": "<="
    },
    "between": {
        "label": "between",
        "symbol": "↔",
        "sqlOperator": "BETWEEN",
        "requiresRange": true
    },
    "in": {
        "label": "in",
        "symbol": "∈",
        "sqlOperator": "IN",
        "requiresList": true
    },
    "not_in": {
        "label": "not in",
        "symbol": "∉",
        "sqlOperator": "NOT IN",
        "requiresList": true
    },
    "is_null": {
        "label": "is empty",
        "symbol": "∅",
        "sqlOperator": "IS NULL",
        "noValue": true
    },
    "is_not_null": {
        "label": "is not empty",
        "symbol": "≠∅",
        "sqlOperator": "IS NOT NULL",
        "noValue": true
    }
};

export function getOperatorsForType(type: FieldType): string[] {
    return FIELD_TYPES[type]?.operators ?? [];
}
