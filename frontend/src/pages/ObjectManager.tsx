import React, { useState, useMemo } from 'react';
import { useSchemas } from '../core/hooks/useMetadata';
import { Link, useNavigate } from 'react-router-dom';
import { Database, Plus, Edit, Trash2, Search, Filter, Box } from 'lucide-react';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { useErrorToast } from '../components/ui/Toast';
import { metadataAPI } from '../infrastructure/api/metadata';
import { UI_CONFIG } from '../core/constants/EnvironmentConfig';

export const ObjectManager: React.FC = () => {
    const errorToast = useErrorToast();
    const navigate = useNavigate();
    const { schemas, loading, error, refresh } = useSchemas();

    const [isModalOpen, setIsModalOpen] = useState(false);
    const [formData, setFormData] = useState({
        label: '',
        plural_label: '',
        api_name: '',
        description: '',
        searchable: true,
    });
    const [creating, setCreating] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [typeFilter, setTypeFilter] = useState<'all' | 'standard' | 'custom'>('all');

    // UX State
    const [apiNameTouched, setApiNameTouched] = useState(false);

    // Delete confirmation modal state
    const [deleteModalOpen, setDeleteModalOpen] = useState(false);
    const [schemaToDelete, setSchemaToDelete] = useState<{ api_name: string; label: string } | null>(null);
    const [deleting, setDeleting] = useState(false);

    const filteredSchemas = useMemo(() => {
        return schemas.filter(schema => {
            if (schema.api_name.startsWith('_')) return false;
            const matchesSearch = schema.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
                schema.api_name.toLowerCase().includes(searchQuery.toLowerCase());
            const matchesType = typeFilter === 'all' ||
                (typeFilter === 'standard' && schema.is_system) ||
                (typeFilter === 'custom' && !schema.is_system);
            return matchesSearch && matchesType;
        });
    }, [schemas, searchQuery, typeFilter]);

    if (loading) return <div className="p-8 text-center">Loading objects...</div>;
    if (error) return <div className="p-8 text-center text-red-500">Error: {error.message}</div>;

    const openCreateModal = () => {
        setIsModalOpen(true);
    };

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        setCreating(true);
        try {
            await metadataAPI.createSchema({
                label: formData.label,
                plural_label: formData.plural_label,
                api_name: formData.api_name,
                description: formData.description,
                searchable: formData.searchable,
                is_custom: true
            });
            setIsModalOpen(false);
            // reset form state not strictly necessary if redirecting, but good practice
            setFormData({
                label: '',
                plural_label: '',
                api_name: '',
                description: '',
                searchable: true
            });
            setApiNameTouched(false);

            // Redirect to the new object's detail page
            navigate(`/setup/objects/${formData.api_name}`);
        } catch (err) {
            // Extract error message from Error
            let errorMessage = err instanceof Error ? err.message : 'Failed to create object. Please check console for details.';

            // Check for nested error data (e.g. from backend validation or conflicts)
            const apiError = err as { data?: { error?: string } };
            if (apiError?.data?.error) {
                errorMessage = `${errorMessage} (${apiError.data.error})`;
            }

            errorToast(errorMessage);
        } finally {
            setCreating(false);
        }
    };

    const handleDeleteObject = async () => {
        if (!schemaToDelete) return;
        setDeleting(true);
        try {
            await metadataAPI.deleteSchema(schemaToDelete.api_name);
            await refresh();
            setDeleteModalOpen(false);
            setSchemaToDelete(null);
        } catch {
            errorToast('Failed to delete object. Check console for details.');
        } finally {
            setDeleting(false);
        }
    };

    return (
        <div className="max-w-7xl mx-auto p-6">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-blue-100 rounded-lg">
                        <Box className="text-blue-600" size={32} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">Object Manager</h1>
                        <p className="text-slate-500">Manage standard and custom objects for your organization.</p>
                    </div>
                </div>
                <button
                    onClick={openCreateModal}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm transition-colors"
                >
                    <Plus size={18} />
                    Create Object
                </button>
            </div>

            <div className="mb-6 flex gap-4">
                <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
                    <input
                        type="text"
                        placeholder="Search objects..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                </div>
                <div className="relative w-48">
                    <Filter className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
                    <select
                        value={typeFilter}
                        onChange={(e) => setTypeFilter(e.target.value as 'all' | 'standard' | 'custom')}
                        className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 appearance-none bg-white"
                    >
                        <option value="all">All Objects</option>
                        <option value="standard">Standard Objects</option>
                        <option value="custom">Custom Objects</option>
                    </select>
                </div>
            </div>

            <div className="bg-white/80 backdrop-blur-xl rounded-2xl border border-white/20 overflow-hidden shadow-xl">
                <table className="w-full text-left text-sm">
                    <thead className="bg-slate-50 border-b border-slate-200">
                        <tr>
                            <th className="px-6 py-3 font-semibold text-slate-700 uppercase tracking-wider text-xs">Label</th>
                            <th className="px-6 py-3 font-semibold text-slate-700 uppercase tracking-wider text-xs">API Name</th>
                            <th className="px-6 py-3 font-semibold text-slate-700 uppercase tracking-wider text-xs">Type</th>
                            <th className="px-6 py-3 font-semibold text-slate-700 uppercase tracking-wider text-xs">Description</th>
                            <th className="px-6 py-3 font-semibold text-slate-700 uppercase tracking-wider text-xs text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                        {filteredSchemas.map(schema => (
                            <tr key={schema.api_name} className="hover:bg-slate-50 transition-colors">
                                <td className="px-6 py-4 font-medium text-slate-900">
                                    <Link to={`/setup/objects/${schema.api_name}`} className="text-blue-600 hover:underline flex items-center gap-2">
                                        <Database size={16} className="text-slate-400" />
                                        {schema.label}
                                    </Link>
                                </td>
                                <td className="px-6 py-4 text-slate-600 font-mono text-xs">{schema.api_name}</td>
                                <td className="px-6 py-4">
                                    {schema.is_system ? (
                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-800 border border-slate-200">
                                            Standard
                                        </span>
                                    ) : (
                                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-700 border border-blue-100">
                                            Custom
                                        </span>
                                    )}
                                </td>
                                <td className="px-6 py-4 text-slate-500 truncate max-w-xs">
                                    {schema.description || '-'}
                                </td>
                                <td className="px-6 py-4 text-right flex justify-end gap-2">
                                    <Link
                                        to={`/setup/objects/${schema.api_name}`}
                                        className="p-1.5 text-slate-400 hover:text-blue-600 rounded hover:bg-blue-50 transition-colors"
                                        title="Edit"
                                    >
                                        <Edit size={16} />
                                    </Link>
                                    {!schema.is_system && (
                                        <button
                                            onClick={() => {
                                                setSchemaToDelete({ api_name: schema.api_name, label: schema.label });
                                                setDeleteModalOpen(true);
                                            }}
                                            className="p-1.5 text-slate-400 hover:text-red-600 rounded hover:bg-red-50 transition-colors"
                                            title="Delete"
                                        >
                                            <Trash2 size={16} />
                                        </button>
                                    )}
                                </td>
                            </tr>
                        ))}
                        {filteredSchemas.length === 0 && (
                            <tr>
                                <td colSpan={5} className="px-6 py-12 text-center text-slate-500">
                                    <div className="flex flex-col items-center gap-2">
                                        <Search size={32} className="text-slate-300" />
                                        <p>No objects found matching your search.</p>
                                    </div>
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>

            {/* Create Object Modal */}
            {isModalOpen && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                    <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
                        <div className="px-6 py-4 border-b border-slate-200 flex justify-between items-center bg-slate-50 rounded-t-xl">
                            <h2 className="text-xl font-bold text-slate-800">Create Custom Object</h2>
                            <button onClick={() => setIsModalOpen(false)} className="text-slate-400 hover:text-slate-600">
                                <span className="sr-only">Close</span>
                                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>

                        <form onSubmit={handleCreate} className="p-6 space-y-6">
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 mb-1">
                                        Label <span className="text-red-600">*</span>
                                    </label>
                                    <input
                                        type="text"
                                        required
                                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                        value={formData.label}
                                        onChange={e => {
                                            const label = e.target.value;
                                            const newFormData = { ...formData, label };
                                            // Only auto-fill if config allows and user hasn't manually edited
                                            if (!apiNameTouched && UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
                                                // Convert to snake_case API name (consistent with other components)
                                                newFormData.api_name = label.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '');
                                            }
                                            setFormData(newFormData);
                                        }}
                                        placeholder="e.g. Project"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 mb-1">
                                        Plural Label <span className="text-red-600">*</span>
                                    </label>
                                    <input
                                        type="text"
                                        required
                                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                        value={formData.plural_label}
                                        onChange={e => setFormData({ ...formData, plural_label: e.target.value })}
                                        placeholder="e.g. Projects"
                                    />
                                </div>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">
                                    API Name <span className="text-red-600">*</span>
                                </label>
                                <input
                                    type="text"
                                    required
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                                    value={formData.api_name}
                                    onChange={e => {
                                        setFormData({ ...formData, api_name: e.target.value });
                                        setApiNameTouched(true);
                                    }}
                                />
                                <p className="text-xs text-slate-500 mt-1">Unique identifier used in URLs and API</p>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">Description</label>
                                <textarea
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                    rows={3}
                                    value={formData.description}
                                    onChange={e => setFormData({ ...formData, description: e.target.value })}
                                    placeholder="Describe the purpose of this object..."
                                />
                            </div>

                            <div>
                                <label className="flex items-center gap-2 cursor-pointer">
                                    <input
                                        type="checkbox"
                                        className="w-4 h-4 text-blue-600 rounded border-slate-300 focus:ring-blue-500"
                                        checked={formData.searchable}
                                        onChange={e => setFormData({ ...formData, searchable: e.target.checked })}
                                    />
                                    <span className="text-sm font-medium text-slate-700">Allow Search</span>
                                </label>
                                <p className="text-xs text-slate-500 mt-1 ml-6">Allow records of this object to be found in Global Search.</p>
                            </div>

                            <div className="flex justify-end gap-3 pt-4 border-t border-slate-200">
                                <button
                                    type="button"
                                    onClick={() => setIsModalOpen(false)}
                                    className="px-4 py-2 text-slate-700 border border-slate-300 rounded-lg hover:bg-slate-50 font-medium"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={creating}
                                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 flex items-center gap-2"
                                >
                                    {creating ? (
                                        <>
                                            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                            Creating...
                                        </>
                                    ) : (
                                        'Create Object'
                                    )}
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

            {/* Delete Object Confirmation Modal */}
            <ConfirmationModal
                isOpen={deleteModalOpen}
                onClose={() => {
                    setDeleteModalOpen(false);
                    setSchemaToDelete(null);
                }}
                onConfirm={handleDeleteObject}
                title="Delete Object"
                message={`Are you sure you want to delete the object "${schemaToDelete?.label}"? This action cannot be undone.`}
                confirmLabel="Delete"
                cancelLabel="Cancel"
                variant="danger"
                loading={deleting}
            />
        </div>
    );
};
