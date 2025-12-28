import React, { useState } from 'react';
import { ChevronDown, ChevronRight, Type, Hash, Calendar, ToggleLeft, Link2, List, FileText, Percent, DollarSign, Mail, Phone, Globe, Calculator, Zap } from 'lucide-react';
import { Button } from '../ui/Button';
import { BaseModal } from '../ui/BaseModal';
import { useSuccessToast, useErrorToast } from '../ui/Toast';
import { metadataAPI } from '../../infrastructure/api/metadata';
import { useSchemas } from '../../core/hooks/useMetadata';
import { formatApiError } from '../../core/utils/errorHandling';
import { UI_CONFIG } from '../../core/constants/EnvironmentConfig';
import { FieldMetadata } from '../../types';

import { PicklistConfig } from './wizard-steps/PicklistConfig';
import { LookupConfig } from './wizard-steps/LookupConfig';
import { RollupConfig } from './wizard-steps/RollupConfig';
import { FormulaConfig } from './wizard-steps/FormulaConfig';

interface CreateFieldWizardProps {
    isOpen: boolean;
    onClose: () => void;
    objectId: string;
    objectApiName: string;
    onSuccess?: () => void;
}

const FIELD_TYPES = [
    { value: 'Text', label: 'Text', icon: Type, description: 'Single line of text' },
    { value: 'TextArea', label: 'Text Area', icon: FileText, description: 'Multiple lines of text' },
    { value: 'Number', label: 'Number', icon: Hash, description: 'Numeric value' },
    { value: 'Currency', label: 'Currency', icon: DollarSign, description: 'Money amount' },
    { value: 'Percent', label: 'Percent', icon: Percent, description: 'Percentage value' },
    { value: 'Date', label: 'Date', icon: Calendar, description: 'Date picker' },
    { value: 'Boolean', label: 'Checkbox', icon: ToggleLeft, description: 'True/False toggle' },
    { value: 'Picklist', label: 'Picklist', icon: List, description: 'Dropdown selection' },
    { value: 'Lookup', label: 'Lookup', icon: Link2, description: 'Link to another object' },
    { value: 'RollupSummary', label: 'Roll-up', icon: Calculator, description: 'Aggregate child data' },
    { value: 'Formula', label: 'Formula', icon: Zap, description: 'Calculated field' },
    { value: 'Email', label: 'Email', icon: Mail, description: 'Email address' },
    { value: 'Phone', label: 'Phone', icon: Phone, description: 'Phone number' },
    { value: 'Url', label: 'URL', icon: Globe, description: 'Web address' },
];

export function CreateFieldWizard({ isOpen, onClose, objectId, objectApiName, onSuccess }: CreateFieldWizardProps) {
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    // Form state
    const [apiName, setApiName] = useState('');
    const [label, setLabel] = useState('');
    const [fieldType, setFieldType] = useState('Text');
    const [required, setRequired] = useState(false);
    const [unique, setUnique] = useState(false);
    const [defaultValue, setDefaultValue] = useState('');
    const [helpText, setHelpText] = useState('');

    // Type-specific state
    const [picklistValues, setPicklistValues] = useState('');
    const [lookupTarget, setLookupTarget] = useState('');
    const [isMasterDetail, setIsMasterDetail] = useState(false);

    // Rollup state
    const [rollupSummaryObject, setRollupSummaryObject] = useState('');
    const [rollupRelationshipField, setRollupRelationshipField] = useState('');
    const [rollupCalcType, setRollupCalcType] = useState<"COUNT" | "SUM" | "MIN" | "MAX" | "AVG">('COUNT');
    const [rollupSummaryField, setRollupSummaryField] = useState('');

    // Formula state
    const [formulaExpression, setFormulaExpression] = useState('');
    const [formulaReturnType, setFormulaReturnType] = useState('Number');

    // UI state
    const [showAdvanced, setShowAdvanced] = useState(false);
    const [submitting, setSubmitting] = useState(false);
    const { schemas } = useSchemas();

    // Auto-generate API name from label (if enabled in config)
    const handleLabelChange = (value: string) => {
        setLabel(value);
        if (UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
            // Convert to snake_case API name
            const generated = value
                .toLowerCase()
                .replace(/[^a-z0-9\s]/g, '')
                .replace(/\s+/g, '_')
                .substring(0, 40);
            setApiName(generated);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!apiName || !label || !fieldType) return;

        setSubmitting(true);
        try {
            // Extended type to support non-standard properties during creation
            type FieldCreationPayload = Partial<FieldMetadata> & {
                is_master_detail?: boolean;
                delete_rule?: string;
                rollup_config?: Record<string, unknown>;
                return_type?: string;
                formula?: string;
                object_id?: string;
            };

            const fieldData: FieldCreationPayload = {
                api_name: apiName,
                label: label,
                type: fieldType as FieldMetadata['type'],
                object_id: objectId,
                required: required,
                unique: unique,
                is_system: false,
                is_name_field: false,
            };

            if (defaultValue) fieldData.default_value = defaultValue;
            if (helpText) fieldData.help_text = helpText;

            // Type-specific options
            if (fieldType === 'Picklist' && picklistValues) {
                fieldData.options = picklistValues.split('\n').map(v => v.trim()).filter(Boolean);
            }
            if (fieldType === 'Lookup' && lookupTarget) {
                fieldData.reference_to = [lookupTarget]; // Expects array
                // is_master_detail not standard prop
                if (isMasterDetail) {
                    fieldData.is_master_detail = true;
                    fieldData.delete_rule = 'Cascade';
                }
            }
            if (fieldType === 'RollupSummary') {
                fieldData.rollup_config = {
                    summary_object: rollupSummaryObject,
                    relationship_field: rollupRelationshipField,
                    calc_type: rollupCalcType,
                    summary_field: rollupSummaryField || undefined,
                };
            }
            if (fieldType === 'Formula' && formulaExpression) {
                fieldData.formula = formulaExpression;
                fieldData.return_type = formulaReturnType;
            }

            await metadataAPI.createField(objectApiName, fieldData as unknown as Partial<FieldMetadata>);
            showSuccess(`Field "${label}" created successfully`);
            onSuccess?.();
            onClose();
            resetForm();
        } catch (err) {
            const apiError = formatApiError(err);
            showError(`Failed to create field: ${apiError.message}`);
        } finally {
            setSubmitting(false);
        }
    };

    const resetForm = () => {
        setApiName('');
        setLabel('');
        setFieldType('Text');
        setRequired(false);
        setUnique(false);
        setDefaultValue('');
        setHelpText('');
        setPicklistValues('');
        setLookupTarget('');
        setIsMasterDetail(false);
        setRollupSummaryObject('');
        setRollupRelationshipField('');
        setRollupCalcType('COUNT');
        setRollupSummaryField('');
        setFormulaExpression('');
        setFormulaReturnType('Number');
        setShowAdvanced(false);
    };

    if (!isOpen) return null;

    return (
        <BaseModal
            isOpen={isOpen}
            onClose={onClose}
            title="Create New Field"
            description={`Add a field to ${objectApiName}`}
            headerClassName="bg-gradient-to-r from-blue-50 to-indigo-50"
        >
            <form onSubmit={handleSubmit} className="p-6 space-y-5">
                {/* Label */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                        Label <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={label}
                        onChange={(e) => handleLabelChange(e.target.value)}
                        placeholder="e.g., Customer Name"
                        className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                        required
                    />
                </div>

                {/* API Name (auto-generated, editable) */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1.5">
                        API Name <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={apiName}
                        onChange={(e) => setApiName(e.target.value)}
                        placeholder="customer_name"
                        className="w-full px-4 py-2.5 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all font-mono text-sm"
                        required
                    />
                    <p className="mt-1 text-xs text-gray-400">Auto-generated from label, editable</p>
                </div>

                {/* Field Type - Visual Selector */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Field Type <span className="text-red-500">*</span>
                    </label>
                    <div className="grid grid-cols-4 gap-2">
                        {FIELD_TYPES.slice(0, 8).map((type) => {
                            const Icon = type.icon;
                            const isSelected = fieldType === type.value;
                            return (
                                <button
                                    key={type.value}
                                    type="button"
                                    onClick={() => setFieldType(type.value)}
                                    className={`flex flex-col items-center p-3 rounded-lg border-2 transition-all ${isSelected
                                        ? 'border-blue-500 bg-blue-50 text-blue-700'
                                        : 'border-gray-100 hover:border-gray-200 hover:bg-gray-50'
                                        }`}
                                >
                                    <Icon className={`w-5 h-5 mb-1 ${isSelected ? 'text-blue-600' : 'text-gray-400'}`} />
                                    <span className="text-xs font-medium">{type.label}</span>
                                </button>
                            );
                        })}
                    </div>
                    {/* More types dropdown */}
                    <details className="mt-2">
                        <summary className="text-xs text-gray-500 cursor-pointer hover:text-gray-700">
                            More field types...
                        </summary>
                        <div className="grid grid-cols-4 gap-2 mt-2">
                            {FIELD_TYPES.slice(8).map((type) => {
                                const Icon = type.icon;
                                const isSelected = fieldType === type.value;
                                return (
                                    <button
                                        key={type.value}
                                        type="button"
                                        onClick={() => setFieldType(type.value)}
                                        className={`flex flex-col items-center p-3 rounded-lg border-2 transition-all ${isSelected
                                            ? 'border-blue-500 bg-blue-50 text-blue-700'
                                            : 'border-gray-100 hover:border-gray-200 hover:bg-gray-50'
                                            }`}
                                    >
                                        <Icon className={`w-5 h-5 mb-1 ${isSelected ? 'text-blue-600' : 'text-gray-400'}`} />
                                        <span className="text-xs font-medium">{type.label}</span>
                                    </button>
                                );
                            })}
                        </div>
                    </details>
                </div>

                {/* Type-specific options */}
                {fieldType === 'Picklist' && (
                    <PicklistConfig
                        picklistValues={picklistValues}
                        onChange={setPicklistValues}
                    />
                )}

                {fieldType === 'Lookup' && (
                    <LookupConfig
                        lookupTarget={lookupTarget}
                        setLookupTarget={setLookupTarget}
                        isMasterDetail={isMasterDetail}
                        setIsMasterDetail={setIsMasterDetail}
                        setRequired={setRequired}
                        schemas={schemas}
                    />
                )}

                {fieldType === 'RollupSummary' && (
                    <RollupConfig
                        rollupSummaryObject={rollupSummaryObject}
                        setRollupSummaryObject={setRollupSummaryObject}
                        rollupRelationshipField={rollupRelationshipField}
                        setRollupRelationshipField={setRollupRelationshipField}
                        rollupCalcType={rollupCalcType}
                        setRollupCalcType={setRollupCalcType}
                        rollupSummaryField={rollupSummaryField}
                        setRollupSummaryField={setRollupSummaryField}
                        schemas={schemas}
                        objectApiName={objectApiName}
                    />
                )}

                {fieldType === 'Formula' && (
                    <FormulaConfig
                        formulaExpression={formulaExpression}
                        setFormulaExpression={setFormulaExpression}
                        formulaReturnType={formulaReturnType}
                        setFormulaReturnType={setFormulaReturnType}
                    />
                )}

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
                        {/* Required & Unique */}
                        <div className="flex gap-6">
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="checkbox"
                                    checked={required}
                                    onChange={(e) => setRequired(e.target.checked)}
                                    disabled={isMasterDetail} // Enforced true by Master-Detail
                                    className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500 disabled:opacity-50"
                                />
                                <span className="text-sm text-gray-700">Required</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="checkbox"
                                    checked={unique}
                                    onChange={(e) => setUnique(e.target.checked)}
                                    className="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                                />
                                <span className="text-sm text-gray-700">Unique</span>
                            </label>
                        </div>

                        {/* Default Value */}
                        <div>
                            <label className="block text-sm font-medium text-gray-600 mb-1">
                                Default Value
                            </label>
                            <input
                                type="text"
                                value={defaultValue}
                                onChange={(e) => setDefaultValue(e.target.value)}
                                placeholder="Optional default value"
                                className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                            />
                        </div>

                        {/* Help Text */}
                        <div>
                            <label className="block text-sm font-medium text-gray-600 mb-1">
                                Help Text
                            </label>
                            <input
                                type="text"
                                value={helpText}
                                onChange={(e) => setHelpText(e.target.value)}
                                placeholder="Instructions for users"
                                className="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                            />
                        </div>
                    </div>
                )}

                {/* Actions */}
                <div className="flex justify-end gap-3 pt-4 border-t border-gray-100">
                    <Button variant="ghost" type="button" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button variant="primary" type="submit" loading={submitting}>
                        Create Field
                    </Button>
                </div>
            </form>
        </BaseModal>
    );
}
