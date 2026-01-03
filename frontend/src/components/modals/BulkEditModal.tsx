import React, { useState, useMemo } from 'react';
import { Edit3, Loader2, AlertCircle } from 'lucide-react';
import { ObjectMetadata, FieldMetadata } from '../../types';
import { dataAPI } from '../../infrastructure/api/data';
import { COMMON_FIELDS } from '../../core/constants';
import { Button } from '../ui/Button';
import { BaseModal } from '../ui/BaseModal';
import { useSuccessToast, useErrorToast } from '../ui/Toast';

// ============================================================================
// Types
// ============================================================================

interface BulkEditModalProps {
    /** Whether the modal is open */
    isOpen: boolean;
    /** Handler to close the modal */
    onClose: () => void;
    /** Metadata for the object being edited */
    objectMetadata: ObjectMetadata;
    /** IDs of records to bulk edit */
    selectedRecordIds: string[];
    /** Callback after successful update */
    onSuccess: () => void;
}

export const BulkEditModal: React.FC<BulkEditModalProps> = ({
    isOpen,
    onClose,
    objectMetadata,
    selectedRecordIds,
    onSuccess
}) => {
    const [updates, setUpdates] = useState<Record<string, string | number | boolean | null>>({});
    const [enabledFields, setEnabledFields] = useState<Set<string>>(new Set());
    const [saving, setSaving] = useState(false);
    const [progress, setProgress] = useState({ current: 0, total: 0 });
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    // Reset state when modal opens
    React.useEffect(() => {
        if (isOpen) {
            setUpdates({});
            setEnabledFields(new Set());
            setSaving(false);
            setProgress({ current: 0, total: 0 });
        }
    }, [isOpen]);

    // Get editable fields (non-system, non-formula, non-auto-number)
    const editableFields = useMemo(() => {
        return (objectMetadata.fields || []).filter(f =>
            !f.is_system &&
            !f.formula &&
            f.type !== 'AutoNumber' &&
            f.api_name !== COMMON_FIELDS.ID
        );
    }, [objectMetadata.fields]);

    const handleFieldToggle = (apiName: string) => {
        const newEnabled = new Set(enabledFields);
        if (newEnabled.has(apiName)) {
            newEnabled.delete(apiName);
            const newUpdates = { ...updates };
            delete newUpdates[apiName];
            setUpdates(newUpdates);
        } else {
            newEnabled.add(apiName);
        }
        setEnabledFields(newEnabled);
    };

    const handleFieldChange = (apiName: string, value: string | number | boolean | null) => {
        setUpdates(prev => ({ ...prev, [apiName]: value }));
    };

    const handleSave = async () => {
        // Only update fields that are enabled and have values
        const fieldsToUpdate = Object.entries(updates).filter(
            ([key]) => enabledFields.has(key)
        );

        if (fieldsToUpdate.length === 0) {
            showError('Please select and set at least one field to update');
            return;
        }

        setSaving(true);
        setProgress({ current: 0, total: selectedRecordIds.length });

        let successCount = 0;
        let errorCount = 0;

        for (let i = 0; i < selectedRecordIds.length; i++) {
            const recordId = selectedRecordIds[i];
            try {
                const updatePayload: Record<string, any> = {};
                fieldsToUpdate.forEach(([key, value]) => {
                    updatePayload[key] = value;
                });

                await dataAPI.updateRecord(objectMetadata.api_name, recordId, updatePayload as Record<string, unknown>);
                successCount++;
            } catch {
                errorCount++;
            }
            setProgress({ current: i + 1, total: selectedRecordIds.length });
        }

        setSaving(false);

        if (errorCount === 0) {
            showSuccess(`Successfully updated ${successCount} record${successCount !== 1 ? 's' : ''}`);
            onSuccess();
            onClose();
        } else {
            showError(`Updated ${successCount} records, failed ${errorCount}`);
            if (successCount > 0) {
                onSuccess();
            }
        }
    };

    const renderFieldInput = (field: FieldMetadata) => {
        const value = updates[field.api_name] ?? '';
        const isEnabled = enabledFields.has(field.api_name);

        const inputClasses = `w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 ${!isEnabled ? 'bg-gray-100 text-gray-400' : ''
            }`;

        switch (field.type) {
            case 'Boolean':
            case 'Checkbox':
                return (
                    <input
                        type="checkbox"
                        checked={value === true}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.checked)}
                        disabled={!isEnabled}
                        className="w-5 h-5 rounded border-gray-300 pointer-events-auto cursor-pointer"
                    />
                );

            case 'Picklist':
                return (
                    <select
                        value={String(value)}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value)}
                        disabled={!isEnabled}
                        className={inputClasses}
                    >
                        <option value="">-- Select --</option>
                        {(field.options || []).map(opt => (
                            <option key={opt} value={opt}>{opt}</option>
                        ))}
                    </select>
                );

            case 'Date':
                return (
                    <input
                        type="date"
                        value={String(value)}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value)}
                        disabled={!isEnabled}
                        className={inputClasses}
                    />
                );

            case 'DateTime':
                return (
                    <input
                        type="datetime-local"
                        value={String(value)}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value)}
                        disabled={!isEnabled}
                        className={inputClasses}
                    />
                );

            case 'Number':
            case 'Currency':
            case 'Percent':
                return (
                    <input
                        type="number"
                        value={value as number}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value ? Number(e.target.value) : '')}
                        disabled={!isEnabled}
                        className={inputClasses}
                        step={field.type === 'Currency' ? '0.01' : 'any'}
                        onFocus={(e) => e.target.select()}
                    />
                );

            case 'JSON':
                return (
                    <textarea
                        value={String(value)}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value)}
                        disabled={!isEnabled}
                        className={`${inputClasses} font-mono text-sm`}
                        rows={6}
                        placeholder='{ "key": "value" }'
                        onFocus={(e) => e.target.select()}
                    />
                );

            case 'TextArea':
            case 'LongTextArea':
                return (
                    <textarea
                        value={String(value)}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value)}
                        disabled={!isEnabled}
                        className={inputClasses}
                        rows={3}
                        onFocus={(e) => e.target.select()}
                    />
                );

            case 'Url':
                return (
                    <input
                        type="url"
                        value={String(value)}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value)}
                        disabled={!isEnabled}
                        className={inputClasses}
                        placeholder="https://example.com"
                        onFocus={(e) => e.target.select()}
                    />
                );

            default:
                return (
                    <input
                        type="text"
                        value={String(value)}
                        onChange={(e) => handleFieldChange(field.api_name, e.target.value)}
                        disabled={!isEnabled}
                        className={inputClasses}
                        placeholder={`Enter ${field.label}...`}
                        onFocus={(e) => e.target.select()}
                    />
                );
        }
    };

    if (!isOpen) return null;

    return (
        <BaseModal
            isOpen={isOpen}
            onClose={onClose}
            title="Bulk Edit"
            description={`Editing ${selectedRecordIds.length} ${objectMetadata.label} record${selectedRecordIds.length !== 1 ? 's' : ''}`}
            icon={Edit3}
            maxWidth="2xl"
            closeButtonLight={true}
            iconClassName="text-white"
            iconBgClassName="bg-white/20"
            headerClassName="bg-gradient-to-r from-amber-500 to-orange-500 text-white"
            footer={
                <div className="flex justify-between items-center w-full">
                    <p className="text-sm text-gray-500">
                        {enabledFields.size} field{enabledFields.size !== 1 ? 's' : ''} selected
                    </p>
                    <div className="flex gap-3">
                        <Button variant="ghost" onClick={onClose} disabled={saving}>
                            Cancel
                        </Button>
                        <Button
                            variant="primary"
                            onClick={handleSave}
                            disabled={saving || enabledFields.size === 0}
                            icon={saving ? <Loader2 className="w-4 h-4 animate-spin" /> : undefined}
                        >
                            {saving ? 'Saving...' : `Update ${selectedRecordIds.length} Records`}
                        </Button>
                    </div>
                </div>
            }
        >
            <div className="p-6 space-y-4">
                {/* Instructions */}
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6 flex items-start gap-3">
                    <AlertCircle className="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" />
                    <div className="text-sm text-blue-700">
                        <p className="font-medium">Select fields to update</p>
                        <p className="text-blue-600">Check the box next to each field you want to modify. Only checked fields will be updated.</p>
                    </div>
                </div>

                {/* Fields */}
                <div className="space-y-4">
                    {editableFields.map(field => (
                        <div key={field.api_name} className="flex items-start gap-4">
                            <div className="pt-2">
                                <input
                                    type="checkbox"
                                    checked={enabledFields.has(field.api_name)}
                                    onChange={() => handleFieldToggle(field.api_name)}
                                    className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                />
                            </div>
                            <div className="flex-1">
                                <label className={`block text-sm font-medium mb-1 ${enabledFields.has(field.api_name) ? 'text-gray-900' : 'text-gray-400'
                                    }`}>
                                    {field.label}
                                    {field.required && <span className="text-red-500 ml-1">*</span>}
                                </label>
                                {renderFieldInput(field)}
                            </div>
                        </div>
                    ))}

                    {editableFields.length === 0 && (
                        <p className="text-gray-500 text-center py-8">
                            No editable fields available for this object.
                        </p>
                    )}
                </div>

                {/* Progress bar */}
                {saving && (
                    <div className="pt-4">
                        <div className="bg-gray-200 rounded-full h-2 overflow-hidden">
                            <div
                                className="bg-blue-500 h-full transition-all duration-300"
                                style={{ width: `${(progress.current / progress.total) * 100}%` }}
                            />
                        </div>
                        <p className="text-xs text-gray-500 text-center mt-1">
                            Updating {progress.current} of {progress.total}...
                        </p>
                    </div>
                )}
            </div>
        </BaseModal>
    );
};

export default BulkEditModal;
