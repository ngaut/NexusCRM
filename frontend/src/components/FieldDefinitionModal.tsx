import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { FieldMetadata, FieldType } from '../types';
import { metadataAPI } from '../infrastructure/api/metadata';
import { MetadataRegistry } from '../registries/MetadataRegistry';
import { useSchemas } from '../core/hooks/useMetadata';
import { VALIDATION_MESSAGES, VALIDATION_PATTERNS, EXCLUDED_FORMULA_RETURN_TYPES } from '../core/constants/ValidationConstants';
import { UI_CONFIG } from '../core/constants/EnvironmentConfig';

interface FieldDefinitionModalProps {
    isOpen: boolean;
    onClose: () => void;
    objectApiName: string;
    onSave: () => void;
    editingField?: FieldMetadata | null;
}

export const FieldDefinitionModal: React.FC<FieldDefinitionModalProps> = ({
    isOpen,
    onClose,
    objectApiName,
    onSave,
    editingField
}) => {
    const { schemas: allSchemas } = useSchemas();
    const [isSaving, setIsSaving] = useState(false);
    const [formData, setFormData] = useState({
        label: '',
        api_name: '',
        type: 'Text' as FieldType,
        required: false,
        unique: false,
        help_text: '',
        default_value: '',
        options: [] as string[],
        reference_to: '',
        formula: '',
        return_type: 'Text' as FieldType,
        max_length: 255,
        decimal_places: 0,
        display_format: '',
        starting_number: 1
    });

    const [error, setError] = useState<string | null>(null);

    const fieldTypes: FieldType[] = MetadataRegistry.getValues().map(def => def.type);

    useEffect(() => {
        if (isOpen) {
            setError(null); // Reset error on open
            if (editingField) {
                setFormData({
                    label: editingField.label,
                    api_name: editingField.api_name,
                    type: editingField.type,
                    required: editingField.required || false,
                    unique: editingField.unique || false,
                    help_text: editingField.help_text || '',
                    default_value: editingField.default_value || '',
                    options: editingField.options || [],
                    reference_to: (Array.isArray(editingField.reference_to) ? editingField.reference_to[0] : editingField.reference_to) || '',
                    formula: editingField.formula || '',
                    return_type: editingField.return_type || 'Text',
                    max_length: editingField.max_length || 255,
                    decimal_places: editingField.decimal_places || 0,
                    display_format: editingField.display_format || '',
                    starting_number: editingField.starting_number || 1
                });
            } else {
                // Reset form for new field
                setFormData({
                    label: '',
                    api_name: '',
                    type: 'Text',
                    required: false,
                    unique: false,
                    help_text: '',
                    default_value: '',
                    options: [],
                    reference_to: '',
                    formula: '',
                    return_type: 'Text' as FieldType,
                    max_length: 255,
                    decimal_places: 0,
                    display_format: '',
                    starting_number: 1
                });
            }
        }
    }, [isOpen, editingField]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSaving(true);
        setError(null);

        try {
            // Client-side validation
            if (formData.type === 'Picklist' && formData.options.length === 0) {
                setError(VALIDATION_MESSAGES.PICKLIST_NO_OPTIONS);
                setIsSaving(false);
                return;
            }
            if (formData.type === 'Lookup' && !formData.reference_to) {
                setError(VALIDATION_MESSAGES.LOOKUP_NO_REFERENCE);
                setIsSaving(false);
                return;
            }

            const fieldData: Partial<FieldMetadata> = {
                label: formData.label.trim(),
                api_name: formData.api_name.trim(),
                type: formData.type,
                required: formData.required,
                unique: formData.unique,
            };

            // Only include optional fields if they have values
            if (formData.help_text?.trim()) fieldData.help_text = formData.help_text.trim();
            if (formData.default_value?.trim()) fieldData.default_value = formData.default_value.trim();

            if (formData.type === 'Picklist') {
                fieldData.options = formData.options;
            }
            if (formData.type === 'Lookup') {
                fieldData.reference_to = [formData.reference_to];
            }
            if (formData.type === 'Formula' && formData.formula) {
                fieldData.formula = formData.formula;
                fieldData.return_type = formData.return_type;
            }
            if (formData.type === 'Text') {
                fieldData.max_length = Number(formData.max_length);
            }
            if (['Number', 'Currency', 'Percent'].includes(formData.type)) {
                fieldData.decimal_places = Number(formData.decimal_places);
                // Also support max_length or length for number precision if needed, but backend often handles decimal(18,scale)
                if (formData.max_length) fieldData.max_length = Number(formData.max_length);
            }
            if (formData.type === 'AutoNumber') {
                fieldData.display_format = formData.display_format;
                fieldData.starting_number = Number(formData.starting_number);
            }


            if (editingField) {
                await metadataAPI.updateField(objectApiName, editingField.api_name, fieldData);
            } else {
                await metadataAPI.createField(objectApiName, fieldData);
            }

            onSave();
            onClose();
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : VALIDATION_MESSAGES.SAVE_FAILED;
            setError(message);
        } finally {
            setIsSaving(false);
        }
    };

    if (!isOpen) return null;

    return createPortal(
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100] p-4">
            <div className="bg-white/95 backdrop-blur-2xl rounded-3xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto border border-white/40">

                <div className="px-6 py-4 border-b border-slate-200 flex justify-between items-center bg-slate-50 rounded-t-xl">
                    <h2 className="text-xl font-bold text-slate-800">
                        {editingField ? 'Edit Field' : 'New Field'}
                    </h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-slate-600">
                        <span className="sr-only">Close</span>
                        <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-6">
                    {error && (
                        <div className="p-4 bg-red-50 text-red-700 rounded-lg border border-red-200 text-sm">
                            {error}
                        </div>
                    )}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                Label <span className="text-red-600">*</span>
                            </label>
                            <input
                                type="text"
                                value={formData.label}
                                onChange={(e) => {
                                    const label = e.target.value;
                                    const newFormData = { ...formData, label };
                                    // Only auto-fill if config allows and not editing
                                    if (!editingField && UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
                                        // Convert to snake_case API name (consistent with other components)
                                        newFormData.api_name = label.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '');
                                    }
                                    setFormData(newFormData);
                                }}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                required
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                API Name <span className="text-red-600">*</span>
                            </label>
                            <input
                                type="text"
                                value={formData.api_name}
                                onChange={(e) => setFormData({ ...formData, api_name: e.target.value })}
                                // Pattern removed to prevent blocking
                                title={VALIDATION_PATTERNS.API_NAME_TITLE}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                                required
                                disabled={!!editingField}
                            />
                            <p className="text-xs text-slate-500 mt-1">Unique identifier (e.g., Status)</p>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">
                            Type <span className="text-red-600">*</span>
                        </label>
                        <select
                            value={formData.type}
                            onChange={(e) => setFormData({ ...formData, type: e.target.value as FieldType })}
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            required
                            disabled={!!editingField}
                        >
                            {fieldTypes.map(type => (
                                <option key={type} value={type}>{type}</option>
                            ))}
                        </select>
                    </div>

                    {formData.type === 'Text' && (
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                Length <span className="text-red-600">*</span>
                            </label>
                            <input
                                type="number"
                                value={formData.max_length}
                                onChange={(e) => setFormData({ ...formData, max_length: parseInt(e.target.value) || 255 })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                required
                                min={1}
                                max={255}
                            />
                            <p className="text-xs text-slate-500 mt-1">Maximum number of characters (1-255)</p>
                        </div>
                    )}

                    {['Number', 'Currency', 'Percent'].includes(formData.type) && (
                        <div className="grid grid-cols-2 gap-6">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Length <span className="text-red-600">*</span>
                                </label>
                                <input
                                    type="number"
                                    value={formData.max_length}
                                    onChange={(e) => setFormData({ ...formData, max_length: parseInt(e.target.value) || 18 })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    required
                                    min={1}
                                    max={18}
                                />
                                <p className="text-xs text-slate-500 mt-1">Total digits</p>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Decimal Places <span className="text-red-600">*</span>
                                </label>
                                <input
                                    type="number"
                                    value={formData.decimal_places}
                                    onChange={(e) => setFormData({ ...formData, decimal_places: parseInt(e.target.value) || 0 })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    required
                                    min={0}
                                    max={18}
                                />
                            </div>
                        </div>
                    )}

                    {formData.type === 'AutoNumber' && (
                        <div className="grid grid-cols-2 gap-6">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Display Format <span className="text-red-600">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={formData.display_format}
                                    onChange={(e) => setFormData({ ...formData, display_format: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    placeholder="A-{0000}"
                                    required
                                />
                                <p className="text-xs text-slate-500 mt-1">Example: INV-&#123;0000&#125;</p>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Starting Number <span className="text-red-600">*</span>
                                </label>
                                <input
                                    type="number"
                                    value={formData.starting_number}
                                    onChange={(e) => setFormData({ ...formData, starting_number: parseInt(e.target.value) || 1 })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    required
                                    min={1}
                                />
                            </div>
                        </div>
                    )}

                    {formData.type === 'Picklist' && (
                        <div className="bg-slate-50 p-4 rounded-lg border border-slate-200">
                            <label className="block text-sm font-medium text-slate-700 mb-2">
                                Picklist Options
                            </label>
                            <textarea
                                value={formData.options.join('\n')}
                                onChange={(e) => setFormData({
                                    ...formData,
                                    options: e.target.value.split('\n').filter(Boolean)
                                })}
                                placeholder="Enter each option on a new line"
                                rows={4}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                            <p className="text-xs text-slate-500 mt-1">One option per line</p>
                        </div>
                    )}

                    {formData.type === 'Lookup' && (
                        <div className="bg-slate-50 p-4 rounded-lg border border-slate-200">
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                Related Object <span className="text-red-600">*</span>
                            </label>
                            <select
                                value={formData.reference_to}
                                onChange={(e) => setFormData({ ...formData, reference_to: e.target.value })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                required
                            >
                                <option value="">Select an object...</option>
                                {allSchemas.map(schema => (
                                    <option key={schema.api_name} value={schema.api_name}>
                                        {schema.label} ({schema.api_name})
                                    </option>
                                ))}
                            </select>
                            <p className="text-xs text-slate-500 mt-1">Select the object to create a relationship with</p>
                        </div>
                    )}

                    {formData.type === 'Formula' && (
                        <div className="bg-slate-50 p-4 rounded-lg border border-slate-200 space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Return Type <span className="text-red-600">*</span>
                                </label>
                                <select
                                    value={formData.return_type}
                                    onChange={(e) => setFormData({ ...formData, return_type: e.target.value as FieldType })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    required
                                >
                                    {fieldTypes.filter(t => !(EXCLUDED_FORMULA_RETURN_TYPES as readonly string[]).includes(t)).map(type => (
                                        <option key={type} value={type}>{type}</option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Formula Expression <span className="text-red-600">*</span>
                                </label>
                                <textarea
                                    value={formData.formula}
                                    onChange={(e) => setFormData({ ...formData, formula: e.target.value })}
                                    placeholder="e.g. amount * 0.1 or first_name + ' ' + last_name"
                                    rows={3}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                                    required
                                />
                                <p className="text-xs text-slate-500 mt-1">
                                    Uses <a href="https://expr-lang.org/" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">expr-lang syntax</a>. Use field names directly (e.g., <code>amount</code>, <code>name</code>).
                                </p>
                            </div>
                        </div>
                    )}

                    <div className="flex gap-6">
                        <label className="flex items-center gap-2 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={formData.required}
                                onChange={(e) => setFormData({ ...formData, required: e.target.checked })}
                                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500 h-4 w-4"
                            />
                            <span className="text-sm text-slate-700">Required Field</span>
                        </label>

                        <label className="flex items-center gap-2 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={formData.unique}
                                onChange={(e) => setFormData({ ...formData, unique: e.target.checked })}
                                className="rounded border-slate-300 text-blue-600 focus:ring-blue-500 h-4 w-4"
                            />
                            <span className="text-sm text-slate-700">Unique Values</span>
                        </label>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                Default Value
                            </label>
                            <input
                                type="text"
                                value={formData.default_value}
                                onChange={(e) => setFormData({ ...formData, default_value: e.target.value })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                Help Text
                            </label>
                            <input
                                type="text"
                                value={formData.help_text}
                                onChange={(e) => setFormData({ ...formData, help_text: e.target.value })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                placeholder="Tooltip for users"
                            />
                        </div>
                    </div>

                    <div className="flex justify-end gap-3 pt-6 border-t border-slate-200">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-slate-700 border border-slate-300 rounded-lg hover:bg-slate-50 font-medium"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isSaving}
                            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 flex items-center gap-2"
                        >
                            {isSaving ? (
                                <>
                                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                    Saving...
                                </>
                            ) : (
                                <>{editingField ? 'Update Field' : 'Create Field'}</>
                            )}
                        </button>
                    </div>
                </form>
            </div>
        </div>,
        document.body
    );
};
