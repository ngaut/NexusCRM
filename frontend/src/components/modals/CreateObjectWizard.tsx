import React, { useState } from 'react';
import { ChevronDown, ChevronRight, Database } from 'lucide-react';
import { Button } from '../ui/Button';
import { BaseModal } from '../ui/BaseModal';
import { useSuccessToast, useErrorToast } from '../ui/Toast';
import { dataAPI } from '../../infrastructure/api/data';
import { formatApiError } from '../../core/utils/errorHandling';
import { TableObject } from '../../constants';
import { UI_CONFIG } from '../../core/constants/EnvironmentConfig';
import { ObjectMetadata } from '../../types';

interface CreateObjectWizardProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess?: (objectId: string) => void;
}

const SHARING_MODELS = [
    { value: 'PublicReadWrite', label: 'Public Read/Write', description: 'All users can view and edit' },
    { value: 'PublicRead', label: 'Public Read Only', description: 'All users can view, only owner can edit' },
    { value: 'Private', label: 'Private', description: 'Only owner and admins can access' },
];

export function CreateObjectWizard({ isOpen, onClose, onSuccess }: CreateObjectWizardProps) {
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    // Form state
    const [apiName, setApiName] = useState('');
    const [label, setLabel] = useState('');
    const [pluralLabel, setPluralLabel] = useState('');
    const [description, setDescription] = useState('');
    const [sharingModel, setSharingModel] = useState('PublicReadWrite');

    // UI state
    const [showAdvanced, setShowAdvanced] = useState(false);
    const [submitting, setSubmitting] = useState(false);

    // Auto-generate API name and plural from label (if enabled in config)
    const handleLabelChange = (value: string) => {
        setLabel(value);

        if (UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
            // Convert to snake_case API name
            const generated = value
                .toLowerCase()
                .replace(/[^a-z0-9]+/g, '_')
                .replace(/^_|_$/g, '')
                .substring(0, 40);
            setApiName(generated);
        }

        // Simple pluralization (always auto-fill plural)
        if (!value) {
            setPluralLabel('');
            return;
        }

        if (value.endsWith('s') || value.endsWith('x') || value.endsWith('ch') || value.endsWith('sh')) {
            setPluralLabel(value + 'es');
        } else if (value.endsWith('y') && !/[aeiou]y$/i.test(value)) {
            setPluralLabel(value.slice(0, -1) + 'ies');
        } else {
            setPluralLabel(value + 's');
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!apiName || !label) return;

        setSubmitting(true);
        try {
            const objectData: Partial<ObjectMetadata> = {
                api_name: apiName,
                label: label,
                plural_label: pluralLabel || label + 's',
                description: description,
                // sharing_model is likely custom for creation or needs to be added to type

                sharing_model: sharingModel as 'PublicReadWrite' | 'PublicRead' | 'Private',
                is_custom: true,
            };

            const result = await dataAPI.createRecord(TableObject, objectData as unknown as Record<string, unknown>);
            showSuccess(`Object "${label}" created successfully`);
            onSuccess?.(result.id);
            onClose();
            resetForm();
        } catch (err) {
            const apiError = formatApiError(err);
            showError(`Failed to create object: ${apiError.message}`);
        } finally {
            setSubmitting(false);
        }
    };

    const resetForm = () => {
        setApiName('');
        setLabel('');
        setPluralLabel('');
        setDescription('');
        setSharingModel('PublicReadWrite');
        setShowAdvanced(false);
    };

    if (!isOpen) return null;

    return (
        <BaseModal
            isOpen={isOpen}
            onClose={onClose}
            title="Create New Object"
            description="Define a new data model"
            icon={Database}
            iconBgClassName="bg-emerald-100"
            iconClassName="text-emerald-600"
            headerClassName="bg-gradient-to-r from-emerald-50 to-teal-50"
        >
            <form onSubmit={handleSubmit} className="p-6 space-y-5">
                {/* Label */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                        Object Name <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={label}
                        onChange={(e) => handleLabelChange(e.target.value)}
                        placeholder="e.g., Customer, Invoice, Project"
                        className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all text-lg"
                        required
                        autoFocus
                    />
                </div>

                {/* API Name (auto-generated, editable) */}
                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">
                            API Name <span className="text-red-500">*</span>
                        </label>
                        <input
                            type="text"
                            value={apiName}
                            onChange={(e) => setApiName(e.target.value)}
                            placeholder="Customer"
                            className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent font-mono text-sm"
                            required
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1.5">
                            Plural Name
                        </label>
                        <input
                            type="text"
                            value={pluralLabel}
                            onChange={(e) => setPluralLabel(e.target.value)}
                            placeholder="Customers"
                            className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent text-sm"
                        />
                    </div>
                </div>

                {/* Description */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                        Description
                    </label>
                    <textarea
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        placeholder="What is this object used for?"
                        rows={2}
                        className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-all text-sm"
                    />
                </div>

                {/* Advanced Options Toggle */}
                <button
                    type="button"
                    onClick={() => setShowAdvanced(!showAdvanced)}
                    className="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-700 transition-colors"
                >
                    {showAdvanced ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                    Advanced Options
                </button>

                {showAdvanced && (
                    <div className="space-y-4 p-4 bg-gray-50 rounded-lg border border-gray-100">
                        {/* Sharing Model */}
                        <div>
                            <label className="block text-sm font-medium text-gray-600 mb-2">
                                Sharing Model
                            </label>
                            <div className="space-y-2">
                                {SHARING_MODELS.map((model) => (
                                    <label
                                        key={model.value}
                                        className={`flex items-start gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all ${sharingModel === model.value
                                            ? 'border-emerald-500 bg-emerald-50'
                                            : 'border-gray-100 hover:border-gray-200'
                                            }`}
                                    >
                                        <input
                                            type="radio"
                                            name="sharingModel"
                                            value={model.value}
                                            checked={sharingModel === model.value}
                                            onChange={(e) => setSharingModel(e.target.value)}
                                            className="mt-0.5"
                                        />
                                        <div>
                                            <div className="text-sm font-medium text-gray-800">{model.label}</div>
                                            <div className="text-xs text-gray-500">{model.description}</div>
                                        </div>
                                    </label>
                                ))}
                            </div>
                        </div>
                    </div>
                )}

                {/* Actions */}
                <div className="flex justify-end gap-3 pt-4 border-t border-gray-100">
                    <Button variant="ghost" type="button" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button variant="primary" type="submit" loading={submitting}>
                        Create Object
                    </Button>
                </div>
            </form>
        </BaseModal>
    );
}
