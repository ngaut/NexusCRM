import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useObjectMetadata } from '../core/hooks/useMetadata';
import { metadataAPI } from '../infrastructure/api/metadata';
import { useNotification } from '../contexts/NotificationContext';
import { Settings, Layout, Shield, Database, Edit, Trash2 } from 'lucide-react';
import type { FieldMetadata } from '../types';
import { ValidationRulesList } from './ValidationRulesList';
import { FieldDefinitionModal } from '../components/FieldDefinitionModal';
import { Skeleton } from '../components/ui/Skeleton';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';

// Sub-components
import { ObjectHeader } from './object-detail/ObjectHeader';
import { ObjectFieldsTab } from './object-detail/ObjectFieldsTab';
import { ObjectDetailsTab } from './object-detail/ObjectDetailsTab';

type Tab = 'fields' | 'details' | 'layouts' | 'validation';

export const ObjectDetail: React.FC = () => {
    const { objectApiName } = useParams<{ objectApiName: string }>();
    const { metadata, loading, error, refresh } = useObjectMetadata(objectApiName || '');
    const { success, error: showError } = useNotification();

    if (!objectApiName) return <div className="p-8 text-center text-red-500">Error: Object API Name is missing</div>;

    const [activeTab, setActiveTab] = useState<Tab>('fields');

    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingField, setEditingField] = useState<FieldMetadata | null>(null);

    // Edit Object State
    const [isEditObjectModalOpen, setIsEditObjectModalOpen] = useState(false);
    const [editObjectData, setEditObjectData] = useState({
        label: '',
        plural_label: '',
        description: ''
    });
    const [savingObject, setSavingObject] = useState(false);

    // Field delete confirmation state
    const [deleteFieldModalOpen, setDeleteFieldModalOpen] = useState(false);
    const [fieldToDelete, setFieldToDelete] = useState<string | null>(null);
    const [deletingField, setDeletingField] = useState(false);

    if (loading) {
        return (
            <div className="max-w-7xl mx-auto p-6">
                <div className="mb-6">
                    <div className="flex items-center gap-2 mb-4">
                        <Skeleton variant="text" width={100} height={20} />
                    </div>
                    <div className="flex items-center justify-between mb-6">
                        <div className="flex items-center gap-4">
                            <Skeleton variant="rectangular" width={56} height={56} className="rounded-lg" />
                            <div>
                                <Skeleton variant="text" width={200} height={32} className="mb-2" />
                                <Skeleton variant="text" width={150} height={20} />
                            </div>
                        </div>
                    </div>
                    <div className="border-b border-slate-200 mb-6">
                        <div className="flex space-x-8">
                            <Skeleton variant="text" width={100} height={40} />
                            <Skeleton variant="text" width={100} height={40} />
                            <Skeleton variant="text" width={100} height={40} />
                            <Skeleton variant="text" width={100} height={40} />
                        </div>
                    </div>
                    <div className="space-y-4">
                        <Skeleton variant="rectangular" height={400} />
                    </div>
                </div>
            </div>
        );
    }
    if (error) return <div className="p-8 text-center text-red-500">Error: {error.message}</div>;
    if (!metadata) return <div className="p-8 text-center">Object not found</div>;

    const handleOpenCreate = () => {
        setEditingField(null);
        setIsModalOpen(true);
    };

    const handleOpenEdit = (field: FieldMetadata) => {
        setEditingField(field);
        setIsModalOpen(true);
    };

    const handleDelete = (fieldApiName: string) => {
        setFieldToDelete(fieldApiName);
        setDeleteFieldModalOpen(true);
    };

    const confirmDeleteField = async () => {
        if (!fieldToDelete) return;
        setDeletingField(true);
        try {
            await metadataAPI.deleteField(objectApiName, fieldToDelete);
            refresh();
            setDeleteFieldModalOpen(false);
            setFieldToDelete(null);
        } catch {
            showError('Delete Failed', 'Failed to delete field. Please check console for details.');
        } finally {
            setDeletingField(false);
        }
    };

    const handleOpenEditObject = () => {
        setEditObjectData({
            label: metadata.label,
            plural_label: metadata.plural_label,
            description: metadata.description || ''
        });
        setIsEditObjectModalOpen(true);
    };

    const handleSaveObject = async (e: React.FormEvent) => {
        e.preventDefault();
        setSavingObject(true);
        try {
            await metadataAPI.updateSchema(objectApiName, {
                label: editObjectData.label,
                plural_label: editObjectData.plural_label,
                description: editObjectData.description
            });
            await refresh();
            setIsEditObjectModalOpen(false);
            success('Object Updated', 'Object properties verified successfully.');
        } catch {
            showError('Update Failed', 'Failed to update object properties.');
        } finally {
            setSavingObject(false);
        }
    };

    return (
        <div className="max-w-7xl mx-auto p-6">
            <ObjectHeader
                metadata={metadata}
                onEditObject={handleOpenEditObject}
            />

            {/* Tabs */}
            <div className="border-b border-slate-200 mb-6">
                <nav className="-mb-px flex space-x-8">
                    <button
                        onClick={() => setActiveTab('fields')}
                        className={`
                                pb-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2
                                ${activeTab === 'fields'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
                            `}
                    >
                        <Settings size={16} />
                        Fields & Relationships
                    </button>
                    <button
                        onClick={() => setActiveTab('layouts')}
                        className={`
                                pb-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2
                                ${activeTab === 'layouts'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
                            `}
                    >
                        <Layout size={16} />
                        Page Layouts
                    </button>
                    <button
                        onClick={() => setActiveTab('validation')}
                        className={`
                                pb-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2
                                ${activeTab === 'validation'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
                            `}
                    >
                        <Shield size={16} />
                        Validation Rules
                    </button>
                    <button
                        onClick={() => setActiveTab('details')}
                        className={`
                                pb-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2
                                ${activeTab === 'details'
                                ? 'border-blue-500 text-blue-600'
                                : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}
                            `}
                    >
                        <Database size={16} />
                        Details
                    </button>
                </nav>
            </div>

            {/* Tab Content */}
            {activeTab === 'fields' && (
                <ObjectFieldsTab
                    metadata={metadata}
                    onOpenCreate={handleOpenCreate}
                    onOpenEdit={handleOpenEdit}
                    onDelete={handleDelete}
                />
            )}

            {activeTab === 'validation' && (
                <ValidationRulesList objectApiName={objectApiName} />
            )}

            {activeTab === 'details' && (
                <ObjectDetailsTab
                    metadata={metadata}
                    refresh={refresh}
                />
            )}

            {activeTab === 'layouts' && (
                <div className="bg-white rounded-lg shadow border border-slate-200 p-8 text-center">
                    <Layout className="mx-auto h-12 w-12 text-slate-400 mb-4" />
                    <h3 className="text-lg font-medium text-slate-900 mb-2">Page Layouts</h3>
                    <p className="text-slate-500 mb-6">Manage how this object appears to users.</p>
                    <Link
                        to={`/setup/objects/${objectApiName}/layout`}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium inline-flex items-center gap-2"
                    >
                        <Edit size={16} />
                        Edit Layout
                    </Link>
                </div>
            )}

            {/* Field Modal */}
            <FieldDefinitionModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                objectApiName={objectApiName}
                onSave={refresh}
                editingField={editingField}
            />

            {/* Edit Object Modal */}
            {
                isEditObjectModalOpen && (
                    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
                        <div className="bg-white rounded-xl shadow-2xl w-full max-w-lg">
                            <div className="px-6 py-4 border-b border-slate-200 flex justify-between items-center bg-slate-50 rounded-t-xl">
                                <h2 className="text-xl font-bold text-slate-800">Edit Object</h2>
                                <button onClick={() => setIsEditObjectModalOpen(false)} className="text-slate-400 hover:text-slate-600">
                                    <div className="sr-only">Close</div>
                                    <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                    </svg>
                                </button>
                            </div>

                            <form onSubmit={handleSaveObject} className="p-6 space-y-6">
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium text-slate-700 mb-1">
                                            Label <span className="text-red-600">*</span>
                                        </label>
                                        <input
                                            type="text"
                                            required
                                            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                            value={editObjectData.label}
                                            onChange={e => setEditObjectData({ ...editObjectData, label: e.target.value })}
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
                                            value={editObjectData.plural_label}
                                            onChange={e => setEditObjectData({ ...editObjectData, plural_label: e.target.value })}
                                        />
                                    </div>
                                </div>

                                <div>
                                    <label className="block text-sm font-medium text-slate-700 mb-1">Description</label>
                                    <textarea
                                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                        rows={3}
                                        value={editObjectData.description}
                                        onChange={e => setEditObjectData({ ...editObjectData, description: e.target.value })}
                                    />
                                </div>

                                <div className="flex justify-end gap-3 pt-4 border-t border-slate-200">
                                    <button
                                        type="button"
                                        onClick={() => setIsEditObjectModalOpen(false)}
                                        className="px-4 py-2 text-slate-700 border border-slate-300 rounded-lg hover:bg-slate-50 font-medium"
                                    >
                                        Cancel
                                    </button>
                                    <button
                                        type="submit"
                                        disabled={savingObject}
                                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 flex items-center gap-2"
                                    >
                                        {savingObject ? 'Saving...' : 'Save Changes'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )
            }

            {/* Delete Field Confirmation Modal */}
            <ConfirmationModal
                isOpen={deleteFieldModalOpen}
                onClose={() => {
                    setDeleteFieldModalOpen(false);
                    setFieldToDelete(null);
                }}
                onConfirm={confirmDeleteField}
                title="Delete Field"
                message={`Are you sure you want to delete the field "${fieldToDelete}"? This action cannot be undone.`}
                confirmLabel="Delete"
                cancelLabel="Cancel"
                variant="danger"
                loading={deletingField}
            />
        </div>
    );
};
