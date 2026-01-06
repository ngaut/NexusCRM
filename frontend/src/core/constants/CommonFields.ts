/**
 * Common Field Constants - Re-exported from generated constants
 * 
 * Provides project-wide access to common field names (id, name, owner_id, etc.)
 * derived from the single source of truth (system_tables.json).
 */

export {
    COMMON_FIELDS,
} from '../../generated-schema';

/**
 * Check if a field name is a standard system field
 */
/**
 * Standard system field names
 */
export const SYSTEM_FIELDS = [
    'id',
    'owner_id',
    'created_date',
    'created_by_id',
    'last_modified_date',
    'last_modified_by_id',
    'is_deleted'
];

/**
 * Check if a field name is a standard system field
 */
export function isSystemField(fieldName: string): boolean {
    return SYSTEM_FIELDS.includes(fieldName);
}
