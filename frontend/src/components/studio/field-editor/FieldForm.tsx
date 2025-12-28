import React from 'react';
import type { FieldType, ObjectMetadata } from '../../../types';
import { FIELD_TYPES } from './FieldTypeSelector';

interface FieldFormData {
    label: string;
    api_name: string;
    type: FieldType;
    required: boolean;
    unique: boolean;
    options: string;
    reference_to: string[];
    description: string;
    help_text: string;
    default_value: string;
    is_master_detail: boolean;
    relationship_name: string;
}

interface FieldFormProps {
    formData: FieldFormData;
    isEditing: boolean;
    availableObjects: ObjectMetadata[];
    onChange: (updates: Partial<FieldFormData>) => void;
    onChangeType: () => void;
}

export const FieldForm: React.FC<FieldFormProps> = ({
    formData,
    isEditing,
    availableObjects,
    onChange,
    onChangeType,
}) => {
    const selectedType = FIELD_TYPES.find(t =>
        formData.is_master_detail ? t.type === 'MasterDetail' : t.type === formData.type
    );

    const handleLabelChange = (value: string) => {
        onChange({
            label: value,
        });
    };

    return (
        <div className="space-y-4">
            {/* Type Badge */}
            {selectedType && !isEditing && (
                <button
                    onClick={onChangeType}
                    className="inline-flex items-center gap-2 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 rounded-lg text-sm transition-colors"
                >
                    <selectedType.icon size={16} className="text-slate-600" />
                    <span className="font-medium text-slate-700">{selectedType.label}</span>
                    <span className="text-slate-400">Change</span>
                </button>
            )}

            {/* Label + API Name */}
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        Label <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={formData.label}
                        onChange={(e) => handleLabelChange(e.target.value)}
                        placeholder="e.g. Status, Amount"
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                        autoFocus
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        API Name <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={formData.api_name}
                        onChange={(e) => onChange({ api_name: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '') })}
                        disabled={isEditing}
                        className="w-full px-3 py-2 border rounded-lg font-mono text-sm focus:ring-2 focus:ring-blue-500 disabled:bg-slate-50 disabled:text-slate-500"
                    />
                </div>
            </div>

            {/* Picklist Options */}
            {(formData.type === 'Picklist') && (
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        Options <span className="text-red-500">*</span>
                    </label>
                    <textarea
                        value={formData.options}
                        onChange={(e) => onChange({ options: e.target.value })}
                        placeholder="Enter one option per line&#10;To Do&#10;In Progress&#10;Done"
                        rows={4}
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                    />
                    <p className="text-xs text-slate-500 mt-1">One option per line</p>
                </div>
            )}

            {/* Master-Detail Relationship Name */}
            {formData.is_master_detail && (
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        Child Relationship Name <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={formData.relationship_name}
                        onChange={(e) => onChange({ relationship_name: e.target.value })}
                        placeholder="e.g. Project_Tasks"
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                    />
                    <p className="text-xs text-slate-500 mt-1">Used for sub-queries and related lists on the parent.</p>
                </div>
            )}

            {/* Lookup Target */}
            {formData.type === 'Lookup' && (
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        Related To <span className="text-red-500">*</span>
                    </label>
                    <select
                        multiple
                        value={formData.reference_to}
                        onChange={(e) => {
                            const selected = Array.from(e.target.selectedOptions, option => option.value);
                            onChange({ reference_to: selected });
                        }}
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 min-h-[120px]"
                    >
                        <optgroup label="Custom Objects">
                            {availableObjects.filter(o => o.is_custom).map(obj => (
                                <option key={obj.api_name} value={obj.api_name}>
                                    {obj.label} ({obj.api_name})
                                </option>
                            ))}
                        </optgroup>
                        <optgroup label="System Objects">
                            {availableObjects.filter(o => !o.is_custom).map(obj => (
                                <option key={obj.api_name} value={obj.api_name}>
                                    {obj.label} ({obj.api_name})
                                </option>
                            ))}
                        </optgroup>
                    </select>
                    <p className="text-xs text-slate-500 mt-1">Hold Cmd/Ctrl to select multiple objects (Polymorphic Lookup)</p>
                </div>
            )}

            {/* Checkboxes */}
            <div className="flex gap-6">
                <label className="flex items-center gap-2 cursor-pointer">
                    <input
                        type="checkbox"
                        checked={formData.required}
                        onChange={(e) => onChange({ required: e.target.checked })}
                        disabled={formData.is_master_detail} // Locked for Master-Detail
                        className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                    />
                    <span className="text-sm text-slate-700">Required</span>
                </label>
                {['Text', 'Email', 'Phone', 'Number'].includes(formData.type) && (
                    <label className="flex items-center gap-2 cursor-pointer">
                        <input
                            type="checkbox"
                            checked={formData.unique}
                            onChange={(e) => onChange({ unique: e.target.checked })}
                            className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                        />
                        <span className="text-sm text-slate-700">Unique</span>
                    </label>
                )}
            </div>

            {/* Help Text */}
            <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Help Text</label>
                <input
                    type="text"
                    value={formData.help_text}
                    onChange={(e) => onChange({ help_text: e.target.value })}
                    placeholder="Optional tooltip for users"
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                />
            </div>
        </div>
    );
};
