/**
 * System Objects - Re-exported from generated constants
 * 
 * Table names are from system_tables.json via generated-schema.ts (SSOT)
 * Other constants from shared/generated/constants.ts
 */

// Table names from SSOT (system_tables.json)
import { SYSTEM_TABLE_NAMES, type SystemTableName } from '../../generated-schema';
export { SYSTEM_TABLE_NAMES, type SystemTableName };

// Object categorization re-exports
export {
    isCustomObject,
    OBJECT_CATEGORIES,
    type ObjectCategory,
    getObjectCategory,
} from './SchemaDefinitions';

// Helper function using SSOT table names
export function isSystemTable(objectApiName: string): boolean {
    return objectApiName.startsWith('_System_');
}
