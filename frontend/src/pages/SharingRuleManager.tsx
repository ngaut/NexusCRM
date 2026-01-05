import React, { useState, useEffect } from 'react';
import { Shield, Plus, Search, Trash2, Edit2 } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { metadataAPI } from '../infrastructure/api/metadata';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { useErrorToast, useSuccessToast } from '../components/ui/Toast';
import { FormulaEditor } from '../components/formula/FormulaEditor';
import { FieldMetadata, SharingRule, ObjectMetadata, Role } from '../types';

export const SharingRuleManager: React.FC = () => {
    const errorToast = useErrorToast();
    const successToast = useSuccessToast();
    const [rules, setRules] = useState<SharingRule[]>([]);
    const [roles, setRoles] = useState<Role[]>([]);
    const [objects, setObjects] = useState<ObjectMetadata[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [editingRule, setEditingRule] = useState<SharingRule | null>(null);
    const [deletingRule, setDeletingRule] = useState<SharingRule | null>(null);

    // Current object fields for editor
    const [currentFields, setCurrentFields] = useState<FieldMetadata[]>([]);

    const [formData, setFormData] = useState({
        name: '',
        object_api_name: '',
        criteria: '',
        access_level: 'Read' as 'Read' | 'Edit',
        share_with_role_id: ''
    });

    useEffect(() => {
        loadData();
    }, []);

    // Load fields when object is selected
    useEffect(() => {
        if (formData.object_api_name) {
            loadObjectFields(formData.object_api_name);
        } else {
            setCurrentFields([]);
        }
    }, [formData.object_api_name]);

    const loadObjectFields = async (objectName: string) => {
        try {
            const response = await metadataAPI.getSchema(objectName);
            if (response?.schema?.fields) {
                setCurrentFields(response.schema.fields);
            }
        } catch {
            errorToast('Failed to load object fields');
        }
    };

    const loadData = async () => {
        setLoading(true);
        try {
            // Load sharing rules
            const rulesRecords = await dataAPI.query({ objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_SHARINGRULE });
            setRules(rulesRecords as unknown as SharingRule[]);

            // Load roles for dropdown
            const rolesRecords = await dataAPI.query({ objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_ROLE });
            setRoles(rolesRecords as unknown as Role[]);

            // Load objects for dropdown
            const objectsRes = await metadataAPI.getSchemas();
            setObjects(objectsRes.schemas || []);
        } catch {
            errorToast('Failed to load sharing rules');
        } finally {
            setLoading(false);
        }
    };

    const handleCreate = async () => {
        try {
            await dataAPI.createRecord(SYSTEM_TABLE_NAMES.SYSTEM_SHARINGRULE, formData);
            successToast('Sharing rule created successfully');
            setShowCreateModal(false);
            resetForm();
            loadData();
        } catch {
            errorToast('Failed to create sharing rule');
        }
    };

    const handleUpdate = async () => {
        if (!editingRule) return;
        try {
            await dataAPI.updateRecord(SYSTEM_TABLE_NAMES.SYSTEM_SHARINGRULE, editingRule.id, formData);
            successToast('Sharing rule updated successfully');
            setEditingRule(null);
            resetForm();
            loadData();
        } catch {
            errorToast('Failed to update sharing rule');
        }
    };

    const handleDelete = async () => {
        if (!deletingRule) return;
        try {
            await dataAPI.deleteRecord(SYSTEM_TABLE_NAMES.SYSTEM_SHARINGRULE, deletingRule.id);
            successToast('Sharing rule deleted successfully');
            setDeletingRule(null);
            loadData();
        } catch {
            errorToast('Failed to delete sharing rule');
        }
    };

    const resetForm = () => {
        setFormData({
            name: '',
            object_api_name: '',
            criteria: '',
            access_level: 'Read',
            share_with_role_id: ''
        });
    };

    const openEditModal = (rule: SharingRule) => {
        setFormData({
            name: rule.name,
            object_api_name: rule.object_api_name,
            criteria: rule.criteria || '',
            access_level: rule.access_level,
            share_with_role_id: rule.share_with_role_id
        });
        setEditingRule(rule);
    };

    const getRoleName = (roleId: string) => {
        const role = roles.find(r => r.id === roleId);
        return role?.name || roleId;
    };

    const getObjectLabel = (apiName: string) => {
        const obj = objects.find(o => o.api_name === apiName);
        return obj?.label || apiName;
    };

    const filteredRules = rules.filter(r =>
        r.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        r.object_api_name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="max-w-7xl mx-auto p-6">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-purple-100 rounded-lg">
                        <Shield className="text-purple-600" size={32} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">Sharing Rules</h1>
                        <p className="text-slate-500">Configure criteria-based record sharing with roles.</p>
                    </div>
                </div>
                <button
                    onClick={() => setShowCreateModal(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm transition-colors"
                >
                    <Plus size={18} />
                    New Sharing Rule
                </button>
            </div>

            {/* Search Bar */}
            <div className="mb-6">
                <div className="relative">
                    <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input
                        type="text"
                        placeholder="Search sharing rules..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>
            </div>

            {/* Rules Table */}
            <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
                {loading ? (
                    <div className="text-center py-8 text-slate-500">Loading...</div>
                ) : filteredRules.length === 0 ? (
                    <div className="text-center py-8 text-slate-500">
                        {searchQuery ? 'No rules match your search.' : 'No sharing rules created yet.'}
                    </div>
                ) : (
                    <table className="w-full text-left text-sm">
                        <thead className="bg-slate-50 border-b border-slate-200">
                            <tr>
                                <th className="px-6 py-3 font-semibold text-slate-700">Name</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Object</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Access</th>
                                <th className="px-6 py-3 font-semibold text-slate-700">Share With</th>
                                <th className="px-6 py-3 font-semibold text-slate-700 text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                            {filteredRules.map(rule => (
                                <tr key={rule.id} className="hover:bg-slate-50">
                                    <td className="px-6 py-4 font-medium text-slate-900">{rule.name}</td>
                                    <td className="px-6 py-4 text-slate-600">{getObjectLabel(rule.object_api_name)}</td>
                                    <td className="px-6 py-4">
                                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${rule.access_level === 'Edit'
                                            ? 'bg-amber-50 text-amber-700 border border-amber-100'
                                            : 'bg-green-50 text-green-700 border border-green-100'
                                            }`}>
                                            {rule.access_level === 'Edit' ? 'Read/Write' : 'Read Only'}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 text-slate-600">{getRoleName(rule.share_with_role_id)}</td>
                                    <td className="px-6 py-4 text-right whitespace-nowrap">
                                        <button
                                            onClick={() => openEditModal(rule)}
                                            className="text-sm font-medium text-blue-600 hover:text-blue-800 mr-3"
                                        >
                                            <Edit2 size={14} className="inline mr-1" />
                                            Edit
                                        </button>
                                        <button
                                            onClick={() => setDeletingRule(rule)}
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
            {(showCreateModal || editingRule) && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]">
                    <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4">
                        <div className="p-6 border-b border-slate-200 flex items-center justify-between">
                            <h2 className="text-xl font-semibold text-slate-800">
                                {editingRule ? 'Edit Sharing Rule' : 'Create Sharing Rule'}
                            </h2>
                            <button
                                onClick={() => {
                                    setShowCreateModal(false);
                                    setEditingRule(null);
                                    resetForm();
                                }}
                                className="text-slate-400 hover:text-slate-600 transition-colors"
                            >
                                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                        <div className="p-6 space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Rule Name <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={formData.name}
                                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                    className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 ${!formData.name && showCreateModal ? 'border-red-300' : 'border-slate-300'
                                        }`}
                                    placeholder="e.g., Share Leads with VP"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Object *
                                </label>
                                <select
                                    value={formData.object_api_name}
                                    onChange={(e) => setFormData({ ...formData, object_api_name: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="">Select Object</option>
                                    {objects.map(obj => (
                                        <option key={obj.api_name} value={obj.api_name}>{obj.label}</option>
                                    ))}
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Access Level *
                                </label>
                                <select
                                    value={formData.access_level}
                                    onChange={(e) => setFormData({ ...formData, access_level: e.target.value as 'Read' | 'Edit' })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="Read">Read Only</option>
                                    <option value="Edit">Read/Write</option>
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    Share With Role *
                                </label>
                                <select
                                    value={formData.share_with_role_id}
                                    onChange={(e) => setFormData({ ...formData, share_with_role_id: e.target.value })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                >
                                    <option value="">Select Role</option>
                                    {roles.map(role => (
                                        <option key={role.id} value={role.id}>{role.name}</option>
                                    ))}
                                </select>
                            </div>
                            <FormulaEditor
                                value={formData.criteria}
                                onChange={(val) => setFormData({ ...formData, criteria: val })}
                                fields={currentFields}
                                label="Criteria"
                                helpText="Define conditions for sharing. Use Builder for simple rules, or Code for complex formulas."
                            />
                        </div>
                        <div className="p-6 border-t border-slate-200 flex justify-end gap-3">
                            <button
                                onClick={() => {
                                    setShowCreateModal(false);
                                    setEditingRule(null);
                                    resetForm();
                                }}
                                className="px-4 py-2 text-slate-600 hover:text-slate-800 font-medium"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={editingRule ? handleUpdate : handleCreate}
                                disabled={!formData.name || !formData.object_api_name || !formData.share_with_role_id}
                                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {editingRule ? 'Save Changes' : 'Create'}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Delete Confirmation */}
            <ConfirmationModal
                isOpen={!!deletingRule}
                onClose={() => setDeletingRule(null)}
                onConfirm={handleDelete}
                title="Delete Sharing Rule"
                message={`Are you sure you want to delete "${deletingRule?.name}"? Users who rely on this rule will lose access.`}
                confirmLabel="Delete"
                variant="danger"
            />
        </div>
    );
};
