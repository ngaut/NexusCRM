import React, { useState, useEffect } from 'react';
import { Key, Plus, Search, Trash2, Edit2, Check, X, Shield } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { PermissionEditorModal, PermissionEntity } from '../components/modals/PermissionEditorModal';
import { useErrorToast, useSuccessToast } from '../components/ui/Toast';
import type { PermissionSet } from '../types';

export const PermissionSetManager: React.FC = () => {
    const errorToast = useErrorToast();
    const successToast = useSuccessToast();
    const [permissionSets, setPermissionSets] = useState<PermissionSet[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [editingPermSet, setEditingPermSet] = useState<PermissionSet | null>(null);
    const [deletingPermSet, setDeletingPermSet] = useState<PermissionSet | null>(null);
    const [editingPermissions, setEditingPermissions] = useState<PermissionEntity | null>(null);

    // Form state
    const [formData, setFormData] = useState({
        name: '',
        label: '',
        description: '',
        is_active: true
    });

    useEffect(() => {
        loadPermissionSets();
    }, []);

    const loadPermissionSets = async () => {
        setLoading(true);
        try {
            const records = await dataAPI.query({ objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_PERMISSIONSET });
            setPermissionSets(records as unknown as PermissionSet[]);
        } catch {
            errorToast('Failed to load permission sets');
        } finally {
            setLoading(false);
        }
    };

    const handleCreate = async () => {
        try {
            await dataAPI.createRecord(SYSTEM_TABLE_NAMES.SYSTEM_PERMISSIONSET, formData);
            successToast('Permission Set created successfully');
            setShowCreateModal(false);
            setFormData({ name: '', label: '', description: '', is_active: true });
            loadPermissionSets();
        } catch {
            errorToast('Failed to create permission set');
        }
    };

    const handleUpdate = async () => {
        if (!editingPermSet) return;
        try {
            await dataAPI.updateRecord(SYSTEM_TABLE_NAMES.SYSTEM_PERMISSIONSET, editingPermSet.id, formData);
            successToast('Permission Set updated successfully');
            setEditingPermSet(null);
            setFormData({ name: '', label: '', description: '', is_active: true });
            loadPermissionSets();
        } catch {
            errorToast('Failed to update permission set');
        }
    };

    const handleDelete = async () => {
        if (!deletingPermSet) return;
        try {
            await dataAPI.deleteRecord(SYSTEM_TABLE_NAMES.SYSTEM_PERMISSIONSET, deletingPermSet.id);
            successToast('Permission Set deleted successfully');
            setDeletingPermSet(null);
            loadPermissionSets();
        } catch {
            errorToast('Failed to delete permission set');
        }
    };

    const openEditModal = (permSet: PermissionSet) => {
        setFormData({
            name: permSet.name,
            label: permSet.label,
            description: permSet.description || '',
            is_active: permSet.is_active
        });
        setEditingPermSet(permSet);
    };

    const filteredPermSets = permissionSets.filter(ps =>
        ps.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        ps.label.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="max-w-7xl mx-auto p-6">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-amber-100 rounded-lg">
                        <Key className="text-amber-600" size={32} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">Permission Sets</h1>
                        <p className="text-slate-500">Grant additive permissions to users without changing their profile.</p>
                    </div>
                </div>
                <button
                    onClick={() => setShowCreateModal(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm transition-colors"
                >
                    <Plus size={18} />
                    New Permission Set
                </button>
            </div>

            {/* Search Bar */}
            <div className="mb-6">
                <div className="relative">
                    <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input
                        type="text"
                        placeholder="Search permission sets..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>
            </div>

            {/* Permission Sets Table */}
            <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
                {loading ? (
                    <div className="text-center py-8 text-slate-500">Loading...</div>
                ) : filteredPermSets.length === 0 ? (
                    <div className="text-center py-8 text-slate-500">
                        {searchQuery ? 'No permission sets match your search.' : 'No permission sets created yet.'}
                    </div>
                ) : (
                    <table className="w-full text-left text-sm">
                        <thead className="bg-slate-50 border-b border-slate-200">
                            <tr>
                                <th className="px-6 py-3 font-semibold text-slate-700">Name</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Label</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Description</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Status</th>
                                <th className="px-6 py-3 font-semibold text-slate-700 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                            {filteredPermSets.map(permSet => (
                                <tr key={permSet.id} className="hover:bg-slate-50">
                                    <td className="px-6 py-4 font-medium text-slate-900">{permSet.name}</td>
                                    <td className="px-6 py-4 text-slate-600">{permSet.label}</td>
                                    <td className="px-6 py-4 text-slate-500 max-w-xs truncate">
                                        {permSet.description || '-'}
                                    </td>
                                    <td className="px-6 py-4">
                                        {permSet.is_active ? (
                                            <span className="inline-flex items-center gap-1 text-green-600">
                                                <Check size={14} /> Active
                                            </span>
                                        ) : (
                                            <span className="inline-flex items-center gap-1 text-slate-500">
                                                <X size={14} /> Inactive
                                            </span>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 text-right whitespace-nowrap">
                                        <button
                                            onClick={() => setEditingPermissions({ id: permSet.id, name: permSet.label, type: 'permission_set' })}
                                            className="text-sm font-medium text-purple-600 hover:text-purple-800 mr-3"
                                        >
                                            <Shield size={14} className="inline mr-1" />
                                            Permissions
                                        </button>
                                        <button
                                            onClick={() => openEditModal(permSet)}
                                            className="text-sm font-medium text-blue-600 hover:text-blue-800 mr-3"
                                        >
                                            <Edit2 size={14} className="inline mr-1" />
                                            Edit
                                        </button>
                                        <button
                                            onClick={() => setDeletingPermSet(permSet)}
                                            className="text-sm font-medium text-red-600 hover:text-red-800"
                                        >
                                            <Trash2 size={14} className="inline mr-1" />
                                            Delete
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            {/* Create/Edit Modal */}
            {(showCreateModal || editingPermSet) && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-white rounded-xl shadow-2xl w-full max-w-md mx-4">
                        <div className="p-6 border-b border-slate-200">
                            <h2 className="text-xl font-semibold text-slate-800">
                                {editingPermSet ? 'Edit Permission Set' : 'Create Permission Set'}
                            </h2>
                        </div>
                        <div className="p-6 space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    API Name *
                                </label>
                                <input
                                    type="text"
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    placeholder="e.g., sales_user_enhanced"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Label *
                                </label>
                                <input
                                    type="text"
                                    value={formData.label}
                                    onChange={(e) => setFormData({ ...formData, label: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    placeholder="e.g., Sales User Enhanced"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Description
                                </label>
                                <textarea
                                    value={formData.description}
                                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    rows={3}
                                    placeholder="Describe what this permission set grants..."
                                />
                            </div>
                            <div className="flex items-center gap-2">
                                <input
                                    type="checkbox"
                                    id="is_active"
                                    checked={formData.is_active}
                                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                                    className="w-4 h-4 text-blue-600 rounded"
                                />
                                <label htmlFor="is_active" className="text-sm text-slate-700">
                                    Active
                                </label>
                            </div>
                        </div>
                        <div className="p-6 border-t border-slate-200 flex justify-end gap-3">
                            <button
                                onClick={() => {
                                    setShowCreateModal(false);
                                    setEditingPermSet(null);
                                    setFormData({ name: '', label: '', description: '', is_active: true });
                                }}
                                className="px-4 py-2 text-slate-600 hover:text-slate-800 font-medium"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={editingPermSet ? handleUpdate : handleCreate}
                                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
                            >
                                {editingPermSet ? 'Save Changes' : 'Create'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Delete Confirmation */}
            <ConfirmationModal
                isOpen={!!deletingPermSet}
                onClose={() => setDeletingPermSet(null)}
                onConfirm={handleDelete}
                title="Delete Permission Set"
                message={`Are you sure you want to delete "${deletingPermSet?.label}"? This will also remove all user assignments.`}
                confirmLabel="Delete"
                variant="danger"
            />

            {/* Permission Editor Modal */}
            {editingPermissions && (
                <PermissionEditorModal
                    entity={editingPermissions}
                    onClose={() => setEditingPermissions(null)}
                    onSave={() => {
                        successToast('Permissions saved successfully');
                    }}
                />
            )}
        </div>
    );
};
