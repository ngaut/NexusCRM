import React, { useState, useEffect } from 'react';
import { metadataAPI } from '../infrastructure/api/metadata';
import { ValidationRule } from '../types';
import { useNotification } from '../contexts/NotificationContext';
import { Plus, Edit, Trash2, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';
import { ValidationRuleEditor } from './ValidationRuleEditor';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';

interface ValidationRulesListProps {
    objectApiName: string;
}

export const ValidationRulesList: React.FC<ValidationRulesListProps> = ({ objectApiName }) => {
    const [rules, setRules] = useState<ValidationRule[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isEditorOpen, setIsEditorOpen] = useState(false);
    const [editingRule, setEditingRule] = useState<ValidationRule | null>(null);
    const { success, error: showError } = useNotification();

    // Delete confirmation state
    const [deleteModalOpen, setDeleteModalOpen] = useState(false);
    const [ruleToDelete, setRuleToDelete] = useState<ValidationRule | null>(null);
    const [deleting, setDeleting] = useState(false);

    const fetchRules = async () => {
        setLoading(true);
        try {
            const response = await metadataAPI.getValidationRules(objectApiName);
            setRules(response.rules || []);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load validation rules');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchRules();
    }, [objectApiName]);

    const handleDelete = (rule: ValidationRule) => {
        setRuleToDelete(rule);
        setDeleteModalOpen(true);
    };

    const confirmDelete = async () => {
        if (!ruleToDelete) return;
        setDeleting(true);
        try {
            await metadataAPI.deleteValidationRule(ruleToDelete.id);
            success('Rule Deleted', 'Validation rule deleted successfully');
            fetchRules();
            setDeleteModalOpen(false);
            setRuleToDelete(null);
        } catch (err) {
            showError('Delete Failed', err instanceof Error ? err.message : 'Failed to delete validation rule');
        } finally {
            setDeleting(false);
        }
    };

    const handleEdit = (rule: ValidationRule) => {
        setEditingRule(rule);
        setIsEditorOpen(true);
    };

    const handleCreate = () => {
        setEditingRule(null);
        setIsEditorOpen(true);
    };

    const handleSave = () => {
        setIsEditorOpen(false);
        fetchRules();
    };

    if (loading) return <div className="p-8 text-center text-slate-500">Loading rules...</div>;
    if (error) return <div className="p-8 text-center text-red-500">Error: {error}</div>;

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center mb-4">
                <div>
                    <h3 className="text-lg font-medium text-slate-900">Validation Rules</h3>
                    <p className="text-sm text-slate-500">Defines standards that your data must verify before being saved.</p>
                </div>
                <button
                    onClick={handleCreate}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                    <Plus size={18} />
                    New Rule
                </button>
            </div>

            <div className="bg-white/80 backdrop-blur-xl shadow-xl overflow-hidden border border-white/20 rounded-2xl">
                <table className="w-full">
                    <thead className="bg-slate-50 border-b border-slate-200">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">Rule Name</th>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">Active</th>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">Error Message</th>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">Condition</th>
                            <th className="px-6 py-3 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-200">
                        {rules.map((rule) => (
                            <tr key={rule.id} className="hover:bg-slate-50">
                                <td className="px-6 py-4 whitespace-nowrap">
                                    <span className="font-medium text-slate-900">{rule.name}</span>
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap">
                                    {rule.active ? (
                                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                                            <CheckCircle size={12} className="mr-1" /> Active
                                        </span>
                                    ) : (
                                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-800">
                                            <XCircle size={12} className="mr-1" /> Inactive
                                        </span>
                                    )}
                                </td>
                                <td className="px-6 py-4 text-sm text-slate-600 truncate max-w-xs" title={rule.error_message}>
                                    {rule.error_message}
                                </td>
                                <td className="px-6 py-4 text-sm text-slate-600 font-mono text-xs truncate max-w-xs" title={rule.condition}>
                                    {rule.condition}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <div className="flex justify-end gap-2">
                                        <button
                                            onClick={() => handleEdit(rule)}
                                            className="p-1.5 text-blue-600 hover:bg-blue-50 rounded transition-colors"
                                            title="Edit"
                                        >
                                            <Edit size={16} />
                                        </button>
                                        <button
                                            onClick={() => handleDelete(rule)}
                                            className="p-1.5 text-red-600 hover:bg-red-50 rounded transition-colors"
                                            title="Delete"
                                        >
                                            <Trash2 size={16} />
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        ))}
                        {rules.length === 0 && (
                            <tr>
                                <td colSpan={5} className="px-6 py-12 text-center text-slate-500">
                                    <div className="flex flex-col items-center gap-2">
                                        <AlertTriangle size={32} className="text-slate-300" />
                                        <p>No validation rules defined for this object.</p>
                                    </div>
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>

            {isEditorOpen && (
                <ValidationRuleEditor
                    objectApiName={objectApiName}
                    rule={editingRule}
                    onClose={() => setIsEditorOpen(false)}
                    onSave={handleSave}
                />
            )}

            {/* Delete Rule Confirmation Modal */}
            <ConfirmationModal
                isOpen={deleteModalOpen}
                onClose={() => {
                    setDeleteModalOpen(false);
                    setRuleToDelete(null);
                }}
                onConfirm={confirmDelete}
                title="Delete Validation Rule"
                message={`Are you sure you want to delete the rule "${ruleToDelete?.name}"?`}
                confirmLabel="Delete"
                cancelLabel="Cancel"
                variant="danger"
                loading={deleting}
            />
        </div>
    );
};
