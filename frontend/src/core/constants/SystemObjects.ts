/**
 * System Objects - Re-exported from generated constants
 * 
 * Table names are from system_tables.json via generated-schema.ts (SSOT)
 * Other constants from shared/generated/constants.ts
 */

// Table names from SSOT (system_tables.json)
import { SYSTEM_TABLE_NAMES, type SystemTableName } from '../../generated-schema';
export { SYSTEM_TABLE_NAMES, type SystemTableName };

// Backward-compatible alias (prefer SYSTEM_TABLE_NAMES for new code)
export const SYSTEM_TABLES = SYSTEM_TABLE_NAMES;

// Object categorization from shared constants
export {
    isCustomObject,
    OBJECT_CATEGORIES,
    type ObjectCategory,
    getObjectCategory,
} from '../../../../shared/generated/constants';

// Helper function using SSOT table names
export function isSystemTable(objectApiName: string): boolean {
    return objectApiName.startsWith('_System_');
}

