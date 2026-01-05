import React, { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { metadataAPI } from '../../infrastructure/api/metadata';
import type { FieldMetadata, FieldType, ObjectMetadata } from '../../types';
import { UI_CONFIG } from '../../core/constants/EnvironmentConfig';

// Sub-components
import { FieldTypeSelector } from './field-editor/FieldTypeSelector';
import { FIELD_TYPE_OPTIONS } from '../../core/constants/ui/FieldUIConstants';
import { FieldForm } from './field-editor/FieldForm';

interface StudioFieldEditorProps {
    objectApiName: string;
    field: FieldMetadata | null; // null = creating new
    onSave: () => void;
    onClose: () => void;
}

export const StudioFieldEditor: React.FC<StudioFieldEditorProps> = ({
    objectApiName,
    field,
    onSave,
    onClose,
}) => {
    const isEditing = !!field;

    const [formData, setFormData] = useState({
        label: '',
        api_name: '',
        type: 'Text' as FieldType,
        required: false,
        unique: false,
        options: '', // For picklist, newline-separated
        reference_to: [] as string[],
        description: '',
        help_text: '',
        default_value: '',
        is_master_detail: false,
        relationship_name: '',
    });

    const [availableObjects, setAvailableObjects] = useState<ObjectMetadata[]>([]);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [step, setStep] = useState<'type' | 'details'>(isEditing ? 'details' : 'type');

    useEffect(() => {
        loadObjects();
        if (field) {
            setFormData({
                label: field.label,
                api_name: field.api_name,
                type: field.type,
                required: field.required || false,
                unique: field.unique || false,
                options: field.options?.join('\n') || '',
                reference_to: field.reference_to || [],
                description: field.description || '',
                help_text: field.help_text || '',
                default_value: field.default_value || '',
                is_master_detail: field.is_master_detail || false,
                relationship_name: field.relationship_name || '',
            });
        }
    }, [field]);

    const loadObjects = async () => {
        try {
            const response = await metadataAPI.getSchemas();
            setAvailableObjects(response.schemas || []);
        } catch {
            // Objects loaded for dropdown; failure silently falls back to empty list
        }
    };

    const handleFormChange = (updates: Partial<typeof formData>) => {
        setFormData(prev => {
            const next = { ...prev, ...updates };

            // Auto-generate API name if label changed and logic applies
            if (updates.label !== undefined && (isEditing || !UI_CONFIG.ENABLE_AUTO_FILL_API_NAME)) {
                // If editing or auto-fill disabled, keep existing or user input unless explicitly cleared/changed
                // Wait, if updates.label is passed, we check if we should auto-fill
                if (!isEditing && UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
                    next.api_name = updates.label.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '');
                }
            }
            // Explicit API name update handled directly by spread
            return next;
        });
    };

    // Custom wrapper for label change to handle the specific auto-fill logic which depends on previous state
    const handleFieldFormChange = (updates: Partial<typeof formData>) => {
        setFormData(prev => {
            const next = { ...prev, ...updates };
            // Replicate the logic: if label changed, try to auto-fill API name
            if (updates.label !== undefined && !isEditing && !UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
                // Do nothing, let user type
            } else if (updates.label !== undefined && !isEditing && UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
                next.api_name = updates.label.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '');
            }
            return next;
        });
    };

    const handleTypeSelect = (type: FieldType | 'MasterDetail') => {
        if (type === 'MasterDetail') {
            setFormData(prev => ({
                ...prev,
                type: 'Lookup' as FieldType,
                is_master_detail: true,
                required: true, // MD is always required
                delete_rule: 'Cascade', // MD always cascades
            }));
        } else {
            setFormData(prev => ({
                ...prev,
                type: type as FieldType,
                is_master_detail: false,
                required: false,
            }));
        }
        setStep('details');
    };

    const handleSave = async () => {
        if (!formData.label.trim()) {
            setError('Label is required');
            return;
        }
        if (!formData.api_name.trim()) {
            setError('API Name is required');
            return;
        }
        if (formData.type === 'Picklist' && !formData.options.trim()) {
            setError('Picklist options are required');
            return;
        }
        if (formData.type === 'Lookup' && (!formData.reference_to || formData.reference_to.length === 0)) {
            setError('Target object is required for Lookup fields');
            return;
        }

        setSaving(true);
        setError(null);

        try {
            const fieldData: Partial<FieldMetadata> = {
                label: formData.label.trim(),
                api_name: formData.api_name.trim(),
                type: formData.type,
                required: formData.required,
                unique: formData.unique,
                description: formData.description || undefined,
                help_text: formData.help_text || undefined,
                default_value: formData.default_value || undefined,
            };

            if (formData.type === 'Picklist') {
                fieldData.options = formData.options.split('\n').map(s => s.trim()).filter(Boolean);
            }

            if (formData.type === 'Lookup') {
                // Ensure it's an array
                const refs = Array.isArray(formData.reference_to) ? formData.reference_to : [formData.reference_to as string];
                fieldData.reference_to = refs.filter(Boolean);
                fieldData.is_polymorphic = refs.length > 1;
            }

            if (formData.is_master_detail) {
                fieldData.is_master_detail = true;
                fieldData.delete_rule = 'Cascade';
                // relationship_name would go here if backend supported it fully, passing for consistency
                fieldData.relationship_name = formData.relationship_name;
            }

            if (isEditing) {
                await metadataAPI.updateField(objectApiName, formData.api_name, fieldData);
            } else {
                await metadataAPI.createField(objectApiName, fieldData);
            }

            onSave();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to save field');
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg max-h-[85vh] overflow-hidden flex flex-col">
                {/* Header */}
                <div className="flex items-center justify-between px-5 py-4 border-b flex-shrink-0">
                    <h2 className="text-lg font-semibold text-slate-800">
                        {isEditing ? `Edit Field: ${field?.label}` : 'New Field'}
                    </h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-slate-600">
                        <X size={20} />
                    </button>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-5">
                    {error && (
                        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
                            {error}
                        </div>
                    )}

                    {/* Step 1: Type Selection */}
                    {step === 'type' && (
                        <FieldTypeSelector onSelect={handleTypeSelect} />
                    )}

                    {/* Step 2: Field Details */}
                    {step === 'details' && (
                        <FieldForm
                            formData={formData}
                            isEditing={isEditing}
                            availableObjects={availableObjects}
                            onChange={handleFieldFormChange}
                            onChangeType={() => setStep('type')}

                        />
                    )}
                </div>

                {/* Footer */}
                <div className="flex justify-between items-center px-5 py-4 border-t bg-slate-50 flex-shrink-0">
                    {step === 'details' && !isEditing ? (
                        <button
                            onClick={() => setStep('type')}
                            className="px-4 py-2 text-slate-600 hover:bg-slate-200 rounded-lg text-sm"
                        >
                            Back
                        </button>
                    ) : (
                        <div />
                    )}
                    <div className="flex gap-2">
                        <button
                            onClick={onClose}
                            className="px-4 py-2 text-slate-600 hover:bg-slate-200 rounded-lg text-sm"
                        >
                            Cancel
                        </button>
                        {step === 'details' && (
                            <button
                                onClick={handleSave}
                                disabled={saving}
                                className="px-5 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
                            >
                                {saving ? 'Saving...' : isEditing ? 'Update Field' : 'Create Field'}
                            </button>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};
