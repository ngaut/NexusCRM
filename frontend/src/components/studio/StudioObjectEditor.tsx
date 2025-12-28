import React, { useState, useEffect } from 'react';
import * as Icons from 'lucide-react';
import { metadataAPI } from '../../infrastructure/api/metadata';
import type { ObjectMetadata, FieldMetadata } from '../../types';
import { StudioFieldEditor } from './StudioFieldEditor';
import { StudioLayoutEditor } from './StudioLayoutEditor';
import { ConfirmationModal } from '../modals/ConfirmationModal';
import { useErrorToast } from '../ui/Toast';

// Sub-components
import { StudioFieldList } from './object-editor/StudioFieldList';
import { StudioObjectDetails } from './object-editor/StudioObjectDetails';

interface StudioObjectEditorProps {
    objectApiName: string;
    onObjectUpdated?: () => void;
}

export const StudioObjectEditor: React.FC<StudioObjectEditorProps> = ({
    objectApiName,
    onObjectUpdated,
}) => {
    const errorToast = useErrorToast();
    const [object, setObject] = useState<ObjectMetadata | null>(null);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState<'details' | 'fields' | 'layout'>('fields');
    const [showFieldEditor, setShowFieldEditor] = useState(false);
    const [editingField, setEditingField] = useState<FieldMetadata | null>(null);
    const [fieldToDelete, setFieldToDelete] = useState<string | null>(null);

    useEffect(() => {
        loadObject();
    }, [objectApiName]);

    const loadObject = async () => {
        setLoading(true);
        try {
            const response = await metadataAPI.getSchema(objectApiName);
            setObject(response.schema);
        } catch {
            // Object loading handled via UI empty state
        } finally {
            setLoading(false);
        }
    };

    const handleAddField = () => {
        setEditingField(null);
        setShowFieldEditor(true);
    };

    const handleEditField = (field: FieldMetadata) => {
        setEditingField(field);
        setShowFieldEditor(true);
    };

    const handleFieldSaved = () => {
        setShowFieldEditor(false);
        setEditingField(null);
        loadObject();
        onObjectUpdated?.();
    };

    const handleDeleteField = (fieldApiName: string) => {
        setFieldToDelete(fieldApiName);
    };

    const executeDeleteField = async () => {
        if (!fieldToDelete) return;
        try {
            await metadataAPI.deleteField(objectApiName, fieldToDelete);
            loadObject();
            onObjectUpdated?.();
        } catch (error) {
            errorToast(error instanceof Error ? error.message : 'Failed to delete field');
        } finally {
            setFieldToDelete(null);
        }
    };

    if (loading) {
        return (
            <div className="bg-white rounded-xl border border-slate-200 p-8">
                <div className="animate-pulse space-y-4">
                    <div className="h-8 bg-slate-200 rounded w-1/3"></div>
                    <div className="h-4 bg-slate-200 rounded w-1/2"></div>
                    <div className="h-64 bg-slate-200 rounded"></div>
                </div>
            </div>
        );
    }

    if (!object) {
        return (
            <div className="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-500">
                Object not found
            </div>
        );
    }

    const IconComponent = Icons[object.icon as keyof typeof Icons] as React.ComponentType<{ size?: number | string; className?: string }> || Icons.Database;
    const customFields = object.fields?.filter(f => !f.is_system) || [];
    const systemFields = object.fields?.filter(f => f.is_system) || [];

    return (
        <div className="space-y-4">
            {/* Object Header */}
            <div className="bg-white rounded-xl border border-slate-200 p-5">
                <div className="flex items-center gap-4">
                    <div className="w-14 h-14 rounded-xl bg-blue-100 flex items-center justify-center">
                        <IconComponent size={28} className="text-blue-600" />
                    </div>
                    <div className="flex-1">
                        <h1 className="text-xl font-bold text-slate-800">{object.label}</h1>
                        <p className="text-sm text-slate-500 font-mono">{object.api_name}</p>
                    </div>
                    {object.is_custom && (
                        <span className="px-3 py-1 bg-purple-100 text-purple-700 text-xs font-medium rounded-full">
                            Custom Object
                        </span>
                    )}
                </div>
            </div>

            {/* Tabs */}
            <div className="bg-white rounded-xl border border-slate-200 overflow-hidden">
                <div className="flex border-b">
                    {(['fields', 'details', 'layout'] as const).map(tab => (
                        <button
                            key={tab}
                            onClick={() => setActiveTab(tab)}
                            className={`px-5 py-3 text-sm font-medium capitalize transition-colors ${activeTab === tab
                                ? 'text-blue-600 border-b-2 border-blue-600 bg-blue-50/50'
                                : 'text-slate-500 hover:text-slate-700 hover:bg-slate-50'
                                }`}
                        >
                            {tab}
                        </button>
                    ))}
                </div>

                {/* Fields Tab */}
                {activeTab === 'fields' && (
                    <StudioFieldList
                        customFields={customFields}
                        systemFields={systemFields}
                        onAddField={handleAddField}
                        onEditField={handleEditField}
                        onDeleteField={handleDeleteField}
                    />
                )}

                {/* Details Tab */}
                {activeTab === 'details' && (
                    <StudioObjectDetails
                        objectApiName={objectApiName}
                        object={object}
                        onUpdate={() => {
                            loadObject();
                            onObjectUpdated?.();
                        }}
                    />
                )}

                {/* Layout Tab */}
                {activeTab === 'layout' && (
                    <div className="p-5">
                        <StudioLayoutEditor objectApiName={objectApiName} />
                    </div>
                )}
            </div>

            {/* Field Editor Slide-out */}
            {showFieldEditor && (
                <StudioFieldEditor
                    objectApiName={objectApiName}
                    field={editingField}
                    onSave={handleFieldSaved}
                    onClose={() => {
                        setShowFieldEditor(false);
                        setEditingField(null);
                    }}
                />
            )}

            <ConfirmationModal
                isOpen={!!fieldToDelete}
                onClose={() => setFieldToDelete(null)}
                onConfirm={executeDeleteField}
                title="Delete Field"
                message={`Are you sure you want to delete field "${fieldToDelete}"? This cannot be undone.`}
                confirmLabel="Delete"
                variant="danger"
            />
        </div>
    );
};
