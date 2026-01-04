import React, { useMemo } from 'react';
import { Plus, Trash2, ArrowRight } from 'lucide-react';
import { ObjectMetadata, FieldMetadata } from '../../types';
import { SYSTEM_FIELDS } from '../../core/constants/CommonFields';

interface FieldMappingBuilderProps {
    targetObjectApiName: string;
    schemas: ObjectMetadata[];
    mapping: Record<string, string>;
    onChange: (mapping: Record<string, string>) => void;
}

// Fields that should not be mapped (system + additional read-only)
const READ_ONLY_FIELDS = [...SYSTEM_FIELDS, 'system_modstamp'];

export const FieldMappingBuilder: React.FC<FieldMappingBuilderProps> = ({
    targetObjectApiName,
    schemas,
    mapping,
    onChange
}) => {
    const targetSchema = useMemo(() =>
        schemas.find(s => s.api_name === targetObjectApiName),
        [schemas, targetObjectApiName]
    );

    const availableFields = useMemo(() => {
        if (!targetSchema) return [];
        return targetSchema.fields
            .filter(f => !READ_ONLY_FIELDS.includes(f.api_name))
            .sort((a, b) => a.label.localeCompare(b.label));
    }, [targetSchema]);

    const handleAddRow = () => {
        onChange({ ...mapping, '': '' });
    };

    const handleUpdateRow = (oldKey: string, newKey: string, newValue: string) => {
        const newMapping: Record<string, string> = {};
        Object.entries(mapping).forEach(([k, v]) => {
            if (k === oldKey) {
                // If taking a new key (dropdown change), use it.
                // If just updating value (input change), keep old key.
                if (newKey !== oldKey) {
                    // Verify new key isn't already taken ?? logic complexity here.
                    // Simple approach: just rebuild
                    newMapping[newKey] = v; // This passes value to new key
                } else {
                    newMapping[k] = newValue;
                }
            } else {
                newMapping[k] = v;
            }
        });

        // This logic is slightly flawed for renaming keys in an object order-preserving way or uniqueness.
        // Better: Convert to array, modify, convert back.
        // But for simplicity, let's just use the entry array approach.
    };

    // Better state approach for editability: Convert config object to array of entries
    const entries = Object.entries(mapping);

    const updateEntry = (index: number, field: 'key' | 'value', val: string) => {
        const newEntries = [...entries];
        if (field === 'key') {
            newEntries[index][0] = val;
        } else {
            newEntries[index][1] = val;
        }

        // Reconstruct object
        const newMap = newEntries.reduce((acc, [k, v]) => {
            if (k) acc[k] = v; // Only add if key is not empty? Or allow empty key transiently?
            // If we filter empty keys, the UI input for key might verify focus.
            // Let's allow empty string keys in the object temporarily, though backend might reject.
            // Actually, object keys must be unique. 
            // If user selects same field twice, one will overwrite. 
            // We should filter functionality or alert. 
            return acc;
        }, {} as Record<string, string>);

        // Simple overwrite protection: 
        // If we change key to an existing key, we merge? No, duplication in UI is bad.

        // Let's copy simple key assignment
        const constructed: Record<string, string> = {};
        newEntries.forEach(([k, v], i) => {
            if (index === i && field === 'key') k = val;
            if (index === i && field === 'value') v = val;
            if (k) constructed[k] = v;
        });
        onChange(constructed);
    };

    const removeEntry = (keyToRemove: string) => {
        const newMap = { ...mapping };
        delete newMap[keyToRemove];
        onChange(newMap);
    };

    if (!targetObjectApiName) {
        return (
            <div className="p-4 text-center text-gray-500 bg-gray-50 dark:bg-gray-800 rounded-lg border border-dashed border-gray-300 dark:border-gray-700">
                Please select a Target Object first
            </div>
        );
    }

    if (!targetSchema) {
        return (
            <div className="p-4 text-center text-orange-500 bg-orange-50 dark:bg-orange-900/20 rounded-lg">
                Schema not found for {targetObjectApiName}
            </div>
        );
    }

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                    Field Mappings
                </label>
                <button
                    type="button"
                    onClick={handleAddRow}
                    className="text-xs flex items-center gap-1 px-2 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded hover:bg-purple-200"
                >
                    <Plus className="w-3 h-3" /> Add Field
                </button>
            </div>

            <div className="space-y-2">
                {entries.length === 0 && (
                    <div className="text-sm text-gray-500 italic p-2">
                        No fields mapped. Click "Add Field" to start.
                    </div>
                )}

                {entries.map(([key, value], index) => (
                    <div key={index} className="flex items-center gap-2 group">
                        <div className="flex-1 min-w-[30%]">
                            <select
                                value={key}
                                onChange={(e) => updateEntry(index, 'key', e.target.value)}
                                className="w-full px-2 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 dark:text-gray-200 focus:ring-1 focus:ring-purple-500"
                            >
                                <option value="">Select Field...</option>
                                {availableFields.map(f => (
                                    <option key={f.api_name} value={f.api_name} disabled={key !== f.api_name && mapping.hasOwnProperty(f.api_name)}>
                                        {f.label} ({f.api_name})
                                    </option>
                                ))}
                                {/* If key is not in availableFields (e.g. unknown manual entry), show it */}
                                {!availableFields.some(f => f.api_name === key) && key && (
                                    <option value={key}>{key}</option>
                                )}
                            </select>
                        </div>

                        <ArrowRight className="w-3 h-3 text-gray-400 flex-shrink-0" />

                        <div className="flex-1">
                            <input
                                type="text"
                                value={value}
                                onChange={(e) => updateEntry(index, 'value', e.target.value)}
                                placeholder="Value / Formula (e.g. {!Name})"
                                className="w-full px-2 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 dark:text-gray-200 focus:ring-1 focus:ring-purple-500 font-mono"
                            />
                        </div>

                        <button
                            type="button"
                            onClick={() => removeEntry(key)}
                            className="p-1.5 text-gray-400 hover:text-red-500 transition-colors opacity-0 group-hover:opacity-100"
                            title="Remove Mapping"
                        >
                            <Trash2 className="w-4 h-4" />
                        </button>
                    </div>
                ))}
            </div>

            {/* Helper Hint */}
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                Tip: Use <code>{'{!FieldName}'}</code> to reference fields from the trigger object.
            </p>
        </div>
    );
};
