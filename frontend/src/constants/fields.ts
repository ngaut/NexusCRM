/**
 * System Field Constants
 * 
 * Imports from shared JSON - single source of truth matching protobuf enums.
 * @see shared/constants/system.json
 * @see shared/proto/constants.proto (SystemField enum)
 */

import systemConstants from '@shared/constants/system.json';

// Extract system fields from shared JSON
const { systemFields, recordStates, defaults } = systemConstants;

// Field API names - derived from shared JSON
export const FieldID = systemFields.ID.apiName;
export const FieldName = systemFields.NAME.apiName;
export const FieldOwnerID = systemFields.OWNER_ID.apiName;
export const FieldCreatedDate = systemFields.CREATED_DATE.apiName;
export const FieldCreatedByID = systemFields.CREATED_BY_ID.apiName;
export const FieldLastModifiedDate = systemFields.LAST_MODIFIED_DATE.apiName;
export const FieldLastModifiedByID = systemFields.LAST_MODIFIED_BY_ID.apiName;
export const FieldIsDeleted = systemFields.IS_DELETED.apiName;



// Standard system fields array
export const SYSTEM_FIELDS = [
    FieldID,
    FieldOwnerID,
    FieldCreatedDate,
    FieldCreatedByID,
    FieldLastModifiedDate,
    FieldLastModifiedByID,
    FieldIsDeleted,
] as const;

// Audit fields array
export const AUDIT_FIELDS = [
    FieldCreatedDate,
    FieldCreatedByID,
    FieldLastModifiedDate,
    FieldLastModifiedByID,
] as const;

/**
 * Check if a field is a standard system field
 */
export function isSystemField(fieldName: string): boolean {
    return SYSTEM_FIELDS.includes(fieldName as typeof SYSTEM_FIELDS[number]);
}

/**
 * Check if a field is an audit field
 */
export function isAuditField(fieldName: string): boolean {
    return AUDIT_FIELDS.includes(fieldName as typeof AUDIT_FIELDS[number]);
}


