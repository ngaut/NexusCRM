/**
 * Field Types Constants
 * 
 * Unified field type definitions imported from shared JSON.
 * These values match the protobuf FieldType enum and the backend fieldtypes registry.
 */

import fieldTypesData from '@shared/constants/fieldTypes.json';

// Field type metadata from shared JSON
export interface FieldTypeMetadata {
    sqlType: string;
    icon: string;
    label: string;
    description: string;
    isSearchable: boolean;
    isGroupable: boolean;
    isSummable: boolean;
    operators: string[];
}

// Type for all field types
export type FieldTypeName = keyof typeof fieldTypesData;

// FieldType constant object matching protobuf enum values
export const FieldType = {
    TEXT: 'Text',
    NUMBER: 'Number',
    CURRENCY: 'Currency',
    DATE: 'Date',
    DATE_TIME: 'DateTime',
    PICKLIST: 'Picklist',
    EMAIL: 'Email',
    PHONE: 'Phone',
    TEXT_AREA: 'TextArea',
    LOOKUP: 'Lookup',
    URL: 'URL',
    BOOLEAN: 'Boolean',
    FORMULA: 'Formula',
    PERCENT: 'Percent',
    ROLLUP_SUMMARY: 'RollupSummary',
    JSON: 'JSON',
    LONG_TEXT_AREA: 'LongTextArea',
    RICH_TEXT: 'RichText',
    PASSWORD: 'Password',
    AUTO_NUMBER: 'AutoNumber',
} as const;

export type FieldTypeValue = typeof FieldType[keyof typeof FieldType];

// Get metadata for a field type
export function getFieldTypeMetadata(fieldType: FieldTypeName): FieldTypeMetadata | undefined {
    return fieldTypesData[fieldType] as FieldTypeMetadata;
}

// Get all field types with their metadata
export function getAllFieldTypes(): Array<{ type: FieldTypeName; metadata: FieldTypeMetadata }> {
    return Object.entries(fieldTypesData).map(([type, metadata]) => ({
        type: type as FieldTypeName,
        metadata: metadata as FieldTypeMetadata,
    }));
}



// Check if a field type is searchable
export function isFieldTypeSearchable(fieldType: FieldTypeName): boolean {
    const metadata = getFieldTypeMetadata(fieldType);
    return metadata?.isSearchable ?? false;
}

// Check if a field type is groupable
export function isFieldTypeGroupable(fieldType: FieldTypeName): boolean {
    const metadata = getFieldTypeMetadata(fieldType);
    return metadata?.isGroupable ?? false;
}

// Check if a field type is summable
export function isFieldTypeSummable(fieldType: FieldTypeName): boolean {
    const metadata = getFieldTypeMetadata(fieldType);
    return metadata?.isSummable ?? false;
}

// Check if a value is a valid field type
export function isValidFieldType(value: string): value is FieldTypeName {
    return value in fieldTypesData;
}

// Get icon for a field type (for UI rendering)
export function getFieldTypeIcon(fieldType: FieldTypeName): string {
    const metadata = getFieldTypeMetadata(fieldType);
    return metadata?.icon ?? 'Type';
}

// Get label for a field type (human-readable name)
export function getFieldTypeLabel(fieldType: FieldTypeName): string {
    const metadata = getFieldTypeMetadata(fieldType);
    return metadata?.label ?? fieldType;
}
