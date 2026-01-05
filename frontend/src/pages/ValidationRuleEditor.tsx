import React, { useState, useEffect } from 'react';
import { metadataAPI } from '../infrastructure/api/metadata';
import { ValidationRule } from '../types';
import { useNotification } from '../contexts/NotificationContext';

interface ValidationRuleEditorProps {
    objectApiName: string;
    rule: ValidationRule | null; // null means create mode
    onClose: () => void;
    onSave: () => void;
}

export const ValidationRuleEditor: React.FC<ValidationRuleEditorProps> = ({
    objectApiName,
    rule,
    onClose,
    onSave
}) => {
    const { success, error: showError } = useNotification();
    const [saving, setSaving] = useState(false);
    const [formData, setFormData] = useState({
        name: '',
        active: true,
        description: '',
        errorMessage: '',
        condition: ''
    });

    useEffect(() => {
        if (rule) {
            setFormData({
                name: rule.name,
                active: rule.active,
                description: '', // description not currently in type, omitted
                errorMessage: rule.error_message,
                condition: rule.condition
            });
        }
    }, [rule]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setSaving(true);

        try {
            const ruleData = {
                objectApiName,
                name: formData.name,
                active: formData.active,
                error_message: formData.errorMessage,
                condition: formData.condition
            };

            if (rule) {
                await metadataAPI.updateValidationRule(rule.id, ruleData);
                success('Rule Updated', 'Validation rule updated successfully');
            } else {
                await metadataAPI.createValidationRule(ruleData);
                success('Rule Created', 'Validation rule created successfully');
            }
            onSave();
        } catch (err) {
            showError('Save Failed', err instanceof Error ? err.message : 'Failed to save validation rule');
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100] p-4">
            <div className="bg-white/95 backdrop-blur-2xl rounded-3xl shadow-2xl w-full max-w-3xl max-h-[90vh] overflow-y-auto border border-white/40">
                <div className="px-6 py-4 border-b border-slate-200 flex justify-between items-center bg-slate-50 rounded-t-xl">
                    <h2 className="text-xl font-bold text-slate-800">
                        {rule ? 'Edit Validation Rule' : 'New Validation Rule'}
                    </h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-slate-600">
                        <span className="sr-only">Close</span>
                        <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">
                                Rule Name <span className="text-red-600">*</span>
                            </label>
                            <input
                                type="text"
                                required
                                value={formData.name}
                                onChange={e => setFormData({ ...formData, name: e.target.value })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                placeholder="e.g. High_Value_Opportunity"
                            />
                        </div>
                        <div className="flex items-center pt-6">
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="checkbox"
                                    checked={formData.active}
                                    onChange={e => setFormData({ ...formData, active: e.target.checked })}
                                    className="rounded border-slate-300 text-blue-600 focus:ring-blue-500 h-4 w-4"
                                />
                                <span className="text-sm font-medium text-slate-700">Active</span>
                            </label>
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">
                            Error Condition Formula <span className="text-red-600">*</span>
                        </label>
                        <p className="text-xs text-slate-500 mb-2">
                            Enter an expression that returns TRUE if the data is INVALID. Use field names like <code>record.amount</code>.
                        </p>
                        <div className="relative">
                            <textarea
                                required
                                value={formData.condition}
                                onChange={e => setFormData({ ...formData, condition: e.target.value })}
                                rows={4}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                                placeholder="record.amount < 0 || record.amount > 1000000"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">
                            Error Message <span className="text-red-600">*</span>
                        </label>
                        <p className="text-xs text-slate-500 mb-2">
                            The text to display to the user when the rule is violated.
                        </p>
                        <input
                            type="text"
                            required
                            value={formData.errorMessage}
                            onChange={e => setFormData({ ...formData, errorMessage: e.target.value })}
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            placeholder="Amount must be positive and less than 1,000,000"
                        />
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
                            disabled={saving}
                            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 flex items-center gap-2"
                        >
                            {saving ? (
                                <>
                                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                    Saving...
                                </>
                            ) : (
                                <>{rule ? 'Update Rule' : 'Save Rule'}</>
                            )}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};
