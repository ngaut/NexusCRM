import { FieldType } from '../types';
import * as Icons from 'lucide-react';
import { RegistryBase } from '@shared/utils';
import { metadataAPI } from '../infrastructure/api/metadata';
import { FieldMetadata } from '../types';
import {
    FIELD_TYPES,
    type FieldTypeDefinition as SharedFieldTypeDefinition,
    getOperatorsForType as getSharedOperators,
} from '../../../shared/generated/constants';
import { SYSTEM_FIELDS } from '../constants';

// Map operator names from shared constants to symbols for UI display
const operatorSymbolMap: Record<string, FilterOperator> = {
    equals: '=',
    not_equals: '!=',
    greater_than: '>',
    greater_or_equal: '>=',
    less_than: '<',
    less_or_equal: '<=',
    contains: 'contains',
    starts_with: 'startsWith',
    ends_with: 'endsWith',
};

export type FilterOperator = '=' | '!=' | '>' | '<' | '>=' | '<=' | 'contains' | 'startsWith' | 'endsWith';

export interface FieldTypeDefinition {
    type: FieldType;
    label: string;
    sqlType: string;
    icon: string;
    description: string;
    isSystemOnly?: boolean;
    isSearchable?: boolean;
    isGroupable?: boolean;
    isSummable?: boolean;
    isVirtual?: boolean;
    operators: FilterOperator[];
}

/**
 * Converts shared field type definition to the format expected by MetadataRegistry
 */
function convertSharedDefinition(type: string, shared: SharedFieldTypeDefinition): FieldTypeDefinition {
    const operators = shared.operators
        .map((op: string) => operatorSymbolMap[op])
        .filter((op): op is FilterOperator => !!op);

    return {
        type: type as FieldType,
        label: shared.label,
        sqlType: shared.sqlType || '',
        icon: shared.icon,
        description: shared.description,
        isSystemOnly: shared.isSystemOnly,
        isSearchable: shared.isSearchable,
        isGroupable: shared.isGroupable,
        isSummable: shared.isSummable,
        isVirtual: shared.isVirtual,
        operators,
    };
}

class MetadataRegistryClass extends RegistryBase<FieldTypeDefinition> {
    constructor() {
        super({
            validator: (key, value) => !!value.type && value.hasOwnProperty('sqlType') ? true : `Field definition "${key}" is invalid`,
            enableEvents: true
        });
        this.registerDefaults();
    }

    private registerDefaults() {
        // Load from shared constants (single source of truth)
        Object.entries(FIELD_TYPES).forEach(([type, shared]) => {
            const def = convertSharedDefinition(type, shared);
            this.register(def.type, def);
        });
    }

    // Instance method for easy access via the exported singleton
    getOperators(type: string): FilterOperator[] {
        const def = this.get(type);
        return def?.operators || ['=', '!='];
    }

    // Instance method for system field check - uses imported constants
    isSystemField(field: { is_system?: boolean; api_name: string }): boolean {
        return !!field.is_system || SYSTEM_FIELDS.includes(field.api_name);
    }

    // Helper to determine if a field is editable in forms
    isFieldEditable(field: { is_system?: boolean; api_name: string; type: FieldType; isReadOnly?: boolean }): boolean {
        // System fields are never editable
        if (this.isSystemField(field)) return false;

        // Calculated fields are not editable
        if (field.type === 'Formula' || field.type === 'RollupSummary') return false;

        // Virtual fields are not editable
        const def = this.get(field.type);
        if (def?.isVirtual) return false;

        // Read-only fields (if flagged) are not editable
        if (field.isReadOnly) return false;

        return true;
    }
    async loadDynamicTypes() {
        try {
            const response = await metadataAPI.getFieldTypes();
            const fieldTypes = response.fieldTypes;

            // Register plugin types (and overwrite built-ins if needed, though they should match)
            fieldTypes.forEach(ft => {
                if (ft.isPlugin) {
                    // Use a generic icon if the specific one isn't found in Lucide imports
                    // or just pass the string and let the UI handle fallback
                    this.register(ft.name as FieldType, {
                        type: ft.name as FieldType,
                        label: ft.label,
                        sqlType: ft.sqlType,
                        icon: ft.icon,
                        description: ft.description,
                        isSearchable: ft.isSearchable,
                        isGroupable: ft.isGroupable,
                        isSummable: ft.isSummable,
                        isVirtual: ft.isVirtual,
                        operators: ft.operators as FilterOperator[],
                    });
                }
            });

        } catch (error) {
            console.error("Failed to load dynamic field types:", error);
        }
    }
}

export const MetadataRegistry = new MetadataRegistryClass();

export const FIELD_DEFINITIONS = MetadataRegistry.getValues();

export const getSqlType = (type: FieldType): string => {
    const def = MetadataRegistry.get(type);
    return def ? def.sqlType : 'VARCHAR(255)';
};

export const getFieldIcon = (type: FieldType) => {
    const def = MetadataRegistry.get(type);
    const iconName = def?.icon || 'Box';
    return (Icons as unknown as Record<string, React.ComponentType>)[iconName];
};

export const isSearchableType = (type: string): boolean => {
    const def = MetadataRegistry.get(type);
    return def?.isSearchable || false;
};

export const isGroupableType = (type: string): boolean => {
    const def = MetadataRegistry.get(type);
    return def?.isGroupable || false;
};

export const isSummableType = (type: string): boolean => {
    const def = MetadataRegistry.get(type);
    return def?.isSummable || false;
};
