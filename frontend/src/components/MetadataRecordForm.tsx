import React, { useState, useEffect } from 'react';
import { useForm, Controller, useWatch } from 'react-hook-form';
import { Loader2, Save, X } from 'lucide-react';
import { Button } from './ui/Button';
import { useErrorToast, useSuccessToast } from './ui/Toast';
import { formatApiError, getOperationErrorMessage } from '../core/utils/errorHandling';
import { usePermissions } from '../contexts/PermissionContext';
import { SearchableLookup } from './SearchableLookup';
import { dataAPI } from '../infrastructure/api/data';
import type { ObjectMetadata, FieldMetadata, SObject } from '../types';
import { SYSTEM_FIELDS } from '../constants';
import { COMMON_FIELDS } from '../core/constants';
import { FIELD_TYPES, FieldType } from '@shared/generated/constants';

// Additional hidden fields beyond SYSTEM_FIELDS
const ADDITIONAL_HIDDEN = ['system_modstamp', 'stage_name'] as const;

interface MetadataRecordFormProps {
    objectMetadata: ObjectMetadata;
    recordId?: string; // If present, edit mode
    initialData?: SObject;
    onSuccess?: (record: SObject) => void;
    onCancel?: () => void;
}

export function MetadataRecordForm({
    objectMetadata,
    recordId,
    initialData,
    onSuccess,
    onCancel,
}: MetadataRecordFormProps) {
    const isEdit = !!recordId;
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    const { control, handleSubmit, setValue, formState: { errors, isSubmitting }, reset } = useForm({
        defaultValues: initialData || {}
    });

    // Watch all fields for changes to trigger calculations
    const values = useWatch({ control });

    // Apply default values from field metadata for new records
    useEffect(() => {
        if (!isEdit && !initialData) {
            const defaults: Record<string, unknown> = {};
            objectMetadata.fields.forEach(field => {
                if (field.default_value !== undefined && field.default_value !== null) {
                    defaults[field.api_name] = field.default_value;
                }
                // For picklists with options but no default, select first option
                if (field.type === 'Picklist' && field.options?.length > 0 && !field.default_value) {
                    defaults[field.api_name] = field.options[0];
                }
            });
            if (Object.keys(defaults).length > 0) {
                reset(defaults);
            }
        }
    }, [objectMetadata, isEdit, initialData, reset]);

    // Sync form with initialData whenever it changes (now handled by parent)
    useEffect(() => {
        if (initialData) {
            reset(initialData);
        }
    }, [initialData, reset]);

    // Calculation Effect
    useEffect(() => {
        // Debounce calculation
        const timer = setTimeout(async () => {
            // Basic optimization: only calculate if we have values
            if (Object.keys(values).length === 0) return;

            try {
                // We merge with initialData to ensure we have ID and other context if needed
                // If we fetched record, values should mostly be complete, but merging with initialData (if any) or existing values is good
                const payload = { ...initialData, ...values }; // values takes precedence
                const calculated = await dataAPI.calculate(objectMetadata.api_name, payload);

                // Update fields that changed
                Object.keys(calculated).forEach(key => {
                    // Only update if value is different to avoid unnecessary renders
                    // And typically matches formula fields (which are read-only in UI usually)
                    if (JSON.stringify(calculated[key]) !== JSON.stringify(values[key])) {
                        setValue(key, calculated[key]);
                    }
                });
            } catch (err) {
                // Silent fail for calculation - don't disrupt user
                // console.warn("Calculation failed", err);
            }
        }, 500); // 500ms debounce

        return () => clearTimeout(timer);
    }, [values, objectMetadata.api_name, initialData, setValue]);

    // if (isLoadingRecord) {
    //     return (
    //         <div className="flex justify-center items-center h-64">
    //             <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
    //         </div>
    //     );
    // }

    const onSubmit = async (data: Record<string, unknown>) => {
        try {
            // Filter out system fields and read-only fields from payload
            const cleanData: Record<string, unknown> = {};
            Object.keys(data).forEach(key => {
                const fieldMeta = objectMetadata.fields.find(f => f.api_name === key);
                const isSystem = (SYSTEM_FIELDS as readonly string[]).includes(key) ||
                    (ADDITIONAL_HIDDEN as readonly string[]).includes(key) ||
                    !!fieldMeta?.is_system;

                // Formula fields should be excluded.
                const isFormula = fieldMeta?.type === 'Formula' || !!fieldMeta?.formula;

                if (fieldMeta && !isSystem && !isFormula) {
                    cleanData[key] = data[key];
                }
            });

            let savedRecord;
            if (isEdit && recordId) {
                await dataAPI.updateRecord(objectMetadata.api_name, recordId, cleanData);
                savedRecord = { ...initialData, ...cleanData, [COMMON_FIELDS.ID]: recordId };
                showSuccess(`${objectMetadata.label} updated successfully`);
            } else {
                savedRecord = await dataAPI.createRecord(objectMetadata.api_name, cleanData);
                showSuccess(`${objectMetadata.label} created successfully`);
            }
            onSuccess?.(savedRecord);
        } catch (err: unknown) {
            const apiError = formatApiError(err);
            showError(getOperationErrorMessage(isEdit ? 'update' : 'create', objectMetadata.label, apiError));
        }
    };

    const { hasFieldPermission, objectPermissions } = usePermissions();

    const renderFieldInput = (field: FieldMetadata) => {
        // Check editability
        const isEditable = hasFieldPermission(objectMetadata.api_name, field.api_name, 'editable');
        const disabled = !isEditable || isSubmitting;

        // Simple rendering logic based on type (can be expanded to use complex inputs from UIRegistry)
        const commonClasses = "mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border disabled:bg-gray-100 disabled:text-gray-500";

        switch (field.type as FieldType) {
            case 'Boolean':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        render={({ field: { value, onChange } }) => (
                            <input
                                type="checkbox"
                                checked={!!value}
                                onChange={onChange}
                                disabled={disabled}
                                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 disabled:opacity-50"
                            />
                        )}
                    />
                );
            case 'TextArea':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <textarea
                                className={commonClasses}
                                rows={4}
                                value={value?.toString() || ''}
                                onChange={onChange}
                                disabled={disabled}
                            />
                        )}
                    />
                );
            case 'Picklist':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <select className={commonClasses} value={value?.toString() || ''} onChange={onChange} disabled={disabled}>
                                <option value="">-- Select --</option>
                                {field.options?.map(opt => (
                                    <option key={opt} value={opt}>{opt}</option>
                                ))}
                            </select>
                        )}
                    />
                );
            case 'Number':
            case 'Currency':
            case 'Percent':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <input
                                type="number"
                                className={commonClasses}
                                value={value?.toString() || ''}
                                onChange={onChange}
                                step={field.type === 'Currency' ? "0.01" : "1"}
                                disabled={disabled}
                            />
                        )}
                    />
                );
            case 'Password':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <input
                                type="password"
                                className={commonClasses}
                                value={value?.toString() || ''}
                                onChange={onChange}
                                disabled={disabled}
                            />
                        )}
                    />
                );
            case 'Date':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <input
                                type="date"
                                className={commonClasses}
                                value={value ? new Date(value as string).toISOString().split('T')[0] : ''}
                                onChange={onChange}
                                disabled={disabled}
                            />
                        )}
                    />
                );
            case 'Lookup':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <SearchableLookup
                                objectApiName={field.reference_to || ''}
                                objectType={(values as Record<string, unknown>)[field.api_name + '_type'] as string | undefined}
                                value={value as string}
                                onChange={(newValue, selectedRecord) => {
                                    onChange(newValue);
                                    // If polymorphic selection, update the type field
                                    if (selectedRecord && '_object_type' in selectedRecord) {
                                        setValue(field.api_name + '_type', (selectedRecord as SObject & { _object_type: string })._object_type);
                                    }
                                }}
                                disabled={disabled}
                            />
                        )}
                    />
                );
            case 'LongTextArea':
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <textarea
                                className={commonClasses}
                                rows={6}
                                value={value?.toString() || ''}
                                onChange={onChange}
                                disabled={disabled}
                            />
                        )}
                    />
                );
            case 'Text':
            case 'Email':
            case 'Phone':
            case 'Url':
            default:
                return (
                    <Controller
                        name={field.api_name}
                        control={control}
                        rules={{ required: field.required }}
                        render={({ field: { value, onChange } }) => (
                            <input
                                type="text"
                                className={commonClasses}
                                value={value?.toString() || ''}
                                onChange={onChange}
                                disabled={disabled}
                            />
                        )}
                    />
                );
        }
    };

    // Filter fields for the form:
    // - Exclude internal system fields (id, audit fields)
    // - Include standard system fields (Name, Amount, etc.)
    // - Exclude fields without read permission
    // ADDITIONAL_HIDDEN is now defined at module scope

    const fieldsToRender = objectMetadata.fields.filter(f => {
        const isSystem = (SYSTEM_FIELDS as readonly string[]).includes(f.api_name);
        const isHidden = (ADDITIONAL_HIDDEN as readonly string[]).includes(f.api_name);
        const hasPerm = hasFieldPermission(objectMetadata.api_name, f.api_name, 'readable');
        return !isSystem && !isHidden && hasPerm;
    });

    return (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6 relative">
            {/* Loading handled by parent */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {fieldsToRender.map(field => (
                    <div key={field.api_name} className={field.type === 'TextArea' || field.type === 'LongTextArea' ? 'col-span-2' : ''}>
                        <label className="block text-sm font-medium text-gray-700">
                            {field.label} {field.required && <span className="text-red-500">*</span>}
                        </label>
                        {renderFieldInput(field)}
                        {errors[field.api_name] && (
                            <p className="mt-1 text-sm text-red-600">This field is required</p>
                        )}
                    </div>
                ))}
            </div>

            <div className="flex justify-end gap-3 pt-6 border-t border-gray-200">
                <Button variant="ghost" type="button" onClick={onCancel}>
                    Cancel
                </Button>
                <Button variant="primary" type="submit" loading={isSubmitting}>
                    {isEdit ? 'Update' : 'Create'} {objectMetadata.label}
                </Button>
            </div>
        </form>
    );
}
