/**
 * Generate TypeScript constants from shared JSON files
 * Run: node scripts/generate-ts-constants.js
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const SHARED_DIR = path.join(__dirname, '..', 'shared', 'constants');
const OUTPUT_DIR = path.join(__dirname, '..', 'shared', 'generated');

// Ensure output directory exists
if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

// Load JSON files
const system = JSON.parse(fs.readFileSync(path.join(SHARED_DIR, 'system.json'), 'utf8'));
const fieldTypes = JSON.parse(fs.readFileSync(path.join(SHARED_DIR, 'fieldTypes.json'), 'utf8'));
const operators = JSON.parse(fs.readFileSync(path.join(SHARED_DIR, 'operators.json'), 'utf8'));

// Generate TypeScript content
let tsContent = `/**
 * AUTO-GENERATED FILE - DO NOT EDIT
 * Generated from shared/constants/*.json
 * Run: node scripts/generate-ts-constants.js
 */

// ==================== Profiles ====================

export const PROFILE_IDS = {
${Object.entries(system.profiles).map(([key, val]) => `    ${key}: '${val.id}'`).join(',\n')}
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
${Object.entries(system.profiles).map(([key, val]) => `    [PROFILE_IDS.${key}]: ${JSON.stringify(val, null, 8).replace(/\n/g, '\n    ')}`).join(',\n')}
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
${Object.entries(system.systemTables).map(([key, val]) => `    ${key}: '${val}'`).join(',\n')}
} as const;

export type SystemTableName = typeof SYSTEM_TABLES[keyof typeof SYSTEM_TABLES];

export function isSystemTable(objectApiName: string): boolean {
    return objectApiName.startsWith('_System_');
}

// ==================== Standard Objects ====================

export const STANDARD_OBJECTS = {
${Object.entries(system.standardObjects).map(([key, val]) => `    ${key}: '${val}'`).join(',\n')}
} as const;

export type StandardObjectName = typeof STANDARD_OBJECTS[keyof typeof STANDARD_OBJECTS];

export function isStandardObject(objectApiName: string): boolean {
    return Object.values(STANDARD_OBJECTS).includes(objectApiName as any);
}

export function isCustomObject(objectApiName: string): boolean {
    return objectApiName.endsWith('__c');
}

// ==================== Object Categories ====================

export const OBJECT_CATEGORIES = {
${Object.entries(system.objectCategories).map(([key, val]) => `    ${key}: '${val}'`).join(',\n')}
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

    if (isStandardObject(objectApiName)) {
        return OBJECT_CATEGORIES.BUSINESS_STANDARD;
    }

    return OBJECT_CATEGORIES.BUSINESS_CUSTOM;
}

// ==================== Defaults ====================

export const DEFAULTS = {
    userName: '${system.defaults.userName}',
    userEmail: '${system.defaults.userEmail}',
    pageSize: ${system.defaults.pageSize},
    maxPageSize: ${system.defaults.maxPageSize}
} as const;

// ==================== Field Types ====================

export type FieldType = ${Object.keys(fieldTypes).map(t => `'${t}'`).join(' | ')};

export interface FieldTypeDefinition {
    sqlType: string | null;
    icon: string;
    label: string;
    description: string;
    isSearchable: boolean;
    isGroupable: boolean;
    isSummable: boolean;
    isFK?: boolean;
    isVirtual?: boolean;
    isSystemOnly?: boolean;
    operators: string[];
}

export const FIELD_TYPES: Record<FieldType, FieldTypeDefinition> = ${JSON.stringify(fieldTypes, null, 4)};

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

export const OPERATORS: Record<string, OperatorDefinition> = ${JSON.stringify(operators, null, 4)};

export function getOperatorsForType(type: FieldType): string[] {
    return FIELD_TYPES[type]?.operators ?? [];
}
`;

// Write TypeScript file
fs.writeFileSync(path.join(OUTPUT_DIR, 'constants.ts'), tsContent);
console.log('✅ Generated: shared/generated/constants.ts');

// Also create an index.ts for easy imports
const indexContent = `/**
 * AUTO-GENERATED FILE - DO NOT EDIT
 * Re-exports all generated constants
 */
export * from './constants';
`;

fs.writeFileSync(path.join(OUTPUT_DIR, 'index.ts'), indexContent);
console.log('✅ Generated: shared/generated/index.ts');

console.log('\\n🎉 TypeScript constants generation complete!');
