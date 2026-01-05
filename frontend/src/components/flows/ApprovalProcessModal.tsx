import React, { useState, useEffect } from 'react';
import { X, Save } from 'lucide-react';
import { ApprovalProcess, ObjectMetadata } from '../../types';

interface ApprovalProcessModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSave: (process: Partial<ApprovalProcess>) => Promise<void>;
    editingProcess: ApprovalProcess | null;
    schemas: ObjectMetadata[];
}

export const ApprovalProcessModal: React.FC<ApprovalProcessModalProps> = ({
    isOpen,
    onClose,
    onSave,
    editingProcess,
    schemas
}) => {
    const [formData, setFormData] = useState<Partial<ApprovalProcess>>({
        is_active: true,
        approver_type: 'User'
    });
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (isOpen) {
            if (editingProcess) {
                setFormData({ ...editingProcess });
            } else {
                setFormData({ is_active: true, approver_type: 'User' });
            }
            setError(null);
        }
    }, [isOpen, editingProcess]);

    const handleSave = async () => {
        if (!formData.name || !formData.object_api_name) {
            setError('Name and Target Object are required');
            return;
        }

        try {
            setSaving(true);
            setError(null);
            await onSave(formData);
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to save process');
        } finally {
            setSaving(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-2xl w-full max-w-lg p-6">
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                        {editingProcess ? 'Edit Approval Process' : 'New Approval Process'}
                    </h2>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {error && (
                    <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg border border-red-200">
                        {error}
                    </div>
                )}

                <div className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Process Name *</label>
                        <input
                            type="text"
                            value={formData.name || ''}
                            onChange={e => setFormData({ ...formData, name: e.target.value })}
                            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 focus:ring-2 focus:ring-purple-500 outline-none"
                            placeholder="e.g. Lead Approval"
                        />
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Target Object *</label>
                        <select
                            value={formData.object_api_name || ''}
                            onChange={e => setFormData({ ...formData, object_api_name: e.target.value })}
                            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 focus:ring-2 focus:ring-purple-500 outline-none"
                            disabled={!!editingProcess}
                        >
                            <option value="">Select Object...</option>
                            {schemas.map(s => (
                                <option key={s.api_name} value={s.api_name}>{s.label}</option>
                            ))}
                        </select>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Approver Type</label>
                            <select
                                value={formData.approver_type || 'User'}
                                onChange={e => setFormData({ ...formData, approver_type: e.target.value })}
                                className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 focus:ring-2 focus:ring-purple-500 outline-none"
                            >
                                <option value="User">Specific User</option>
                                <option value="Manager">Manager</option>
                                <option value="Self">Self (Auto-Approve)</option>
                            </select>
                        </div>
                        {formData.approver_type === 'User' && (
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">User ID</label>
                                <input
                                    type="text"
                                    value={formData.approver_id || ''}
                                    onChange={e => setFormData({ ...formData, approver_id: e.target.value })}
                                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 focus:ring-2 focus:ring-purple-500 outline-none"
                                    placeholder="User ID"
                                />
                            </div>
                        )}
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
                        <textarea
                            value={formData.description || ''}
                            onChange={e => setFormData({ ...formData, description: e.target.value })}
                            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 focus:ring-2 focus:ring-purple-500 outline-none"
                            rows={3}
                        />
                    </div>

                    <div className="flex items-center gap-2">
                        <label className="relative inline-flex items-center cursor-pointer">
                            <input
                                type="checkbox"
                                checked={formData.is_active || false}
                                onChange={e => setFormData({ ...formData, is_active: e.target.checked })}
                                className="sr-only peer"
                            />
                            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer 
                                    peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] 
                                    after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 
                                    after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-purple-600"></div>
                            <span className="ml-3 text-sm font-medium text-gray-900 dark:text-gray-300">Active</span>
                        </label>
                    </div>
                </div>

                <div className="flex justify-end gap-3 mt-6">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="flex items-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50 transition-colors"
                    >
                        <Save className="w-4 h-4" />
                        {saving ? 'Saving...' : 'Save'}
                    </button>
                </div>
            </div>
        </div>
    );
};
