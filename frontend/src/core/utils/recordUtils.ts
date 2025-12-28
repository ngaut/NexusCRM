import { SObject, ObjectMetadata } from '../../types';

/**
 * Safely converts any value to a string.
 * Handles null/undefined by returning an empty string or default.
 */
export function getSafeString(value: unknown, fallback: string = ''): string {
    if (value === null || value === undefined) return fallback;
    return String(value);
}

/**
 * Heuristic to find the best display name for a record.
 * 1. Checks schema's 'is_name_field'
 * 2. Checks standard 'name' field
 * 3. Checks common alternatives (title, subject, case_number, etc.)
 * 4. Fallback to ID
 */
export function getRecordDisplayName(record: SObject, schema?: ObjectMetadata): string {
    if (!record) return '';

    // 1. Use Schema if available
    if (schema?.fields) {
        const nameField = schema.fields.find(f => f.is_name_field);
        if (nameField && record[nameField.api_name]) {
            return String(record[nameField.api_name]);
        }
    }

    // 2. Standard 'name'
    if (record.name) return String(record.name);

    // 3. Common alternatives
    if (record.title) return String(record.title);
    if (record.subject) return String(record.subject);
    if (record.case_number) return String(record.case_number);
    if (record.full_name) return String(record.full_name);
    if (record.project_name) return String(record.project_name);

    // 4. Fallback
    return record.id || 'Untitled Record';
}

/**
 * Identifies the field to use for the Path/Status visualizer.
 * Checks for standard status/stage fields.
 */
export function getPathField(schema: ObjectMetadata): import('../../types').FieldMetadata | undefined {
    return schema.fields.find(f =>
        f.api_name === 'stage_name' ||
        f.api_name === 'status' ||
        f.api_name === 'lifecycle_stage'
    );
}

/**
 * Identifies key fields to show in highlights panels or headers.
 * Filters out system fields but includes important ones like Owner, Amount, Status.
 */
export function getHighlightFields(schema: ObjectMetadata, count: number = 5): import('../../types').FieldMetadata[] {
    return schema.fields
        .filter(f => !f.is_system || ['owner_id', 'amount', 'stage_name', 'status'].includes(f.api_name))
        .slice(0, count);
}

