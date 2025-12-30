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
 * 1. Checks schema.path_field
 * 2. Fallback to common status fields
 */
export function getPathField(schema: ObjectMetadata): import('../../types').FieldMetadata | undefined {
    // 1. Explicit metadata
    if (schema.path_field) {
        const field = schema.fields.find(f => f.api_name === schema.path_field);
        if (field) return field;
    }

    // 2. Heuristic fallback
    return schema.fields.find(f =>
        f.api_name === 'status' ||
        f.api_name === 'stage_name' ||
        f.api_name === 'lifecycle_stage'
    );
}

/**
 * Identifies key fields to show in highlights panels or headers.
 * 1. Uses compact_layout if defined in metadata
 * 2. Filters out system fields but includes important business ones
 */
export function getHighlightFields(schema: ObjectMetadata, count: number = 5): import('../../types').FieldMetadata[] {
    // Note: compact_layout is on PageLayout, but we often want a quick heuristic on the schema too.
    // We'll stick to a smarter heuristic here.
    return schema.fields
        .filter(f => !f.is_system || f.is_name_field)
        .slice(0, count);
}

