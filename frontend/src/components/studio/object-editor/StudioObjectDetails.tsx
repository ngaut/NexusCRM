import React, { useState, useEffect } from 'react';
import { Edit2 } from 'lucide-react';
import { metadataAPI } from '../../../infrastructure/api/metadata';
import type { ObjectMetadata } from '../../../types';
import { useErrorToast } from '../../ui/Toast';

interface StudioObjectDetailsProps {
    objectApiName: string;
    object: ObjectMetadata;
    onUpdate: () => void;
}

export const StudioObjectDetails: React.FC<StudioObjectDetailsProps> = ({
    objectApiName,
    object,
    onUpdate,
}) => {
    const errorToast = useErrorToast();
    const [isEditing, setIsEditing] = useState(false);
    const [saving, setSaving] = useState(false);
    const [formData, setFormData] = useState({
        label: '',
        plural_label: '',
        description: '',
    });

    useEffect(() => {
        if (object) {
            setFormData({
                label: object.label,
                plural_label: object.plural_label,
                description: object.description || '',
            });
        }
    }, [object]);

    const handleSave = async () => {
        setSaving(true);
        try {
            await metadataAPI.updateSchema(objectApiName, {
                ...object,
                label: formData.label,
                plural_label: formData.plural_label,
                description: formData.description,
            });
            onUpdate();
            setIsEditing(false);
        } catch (error) {
            errorToast(error instanceof Error ? error.message : 'Failed to update object details');
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="p-5 space-y-4">
            <div className="flex justify-end mb-2">
                {!isEditing ? (
                    <button
                        onClick={() => setIsEditing(true)}
                        className="flex items-center gap-2 px-3 py-1.5 text-blue-600 hover:bg-blue-50 rounded-lg text-sm font-medium transition-colors"
                    >
                        <Edit2 size={16} />
                        Edit Details
                    </button>
                ) : (
                    <div className="flex gap-2">
                        <button
                            onClick={() => {
                                setIsEditing(false);
                                setFormData({
                                    label: object.label,
                                    plural_label: object.plural_label,
                                    description: object.description || '',
                                });
                            }}
                            className="px-3 py-1.5 text-slate-600 hover:bg-slate-100 rounded-lg text-sm font-medium"
                        >
                            Cancel
                        </button>
                        <button
                            onClick={handleSave}
                            disabled={saving}
                            className="px-3 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
                        >
                            {saving ? 'Saving...' : 'Save Changes'}
                        </button>
                    </div>
                )}
            </div>

            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Label</label>
                    <input
                        type="text"
                        value={isEditing ? formData.label : object.label}
                        onChange={(e) => setFormData(prev => ({ ...prev, label: e.target.value }))}
                        disabled={!isEditing}
                        className={`w-full px-3 py-2 border rounded-lg transition-colors ${isEditing
                            ? 'bg-white border-slate-300 focus:ring-2 focus:ring-blue-500'
                            : 'bg-slate-50 border-transparent text-slate-600'}`}
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Plural Label</label>
                    <input
                        type="text"
                        value={isEditing ? formData.plural_label : object.plural_label}
                        onChange={(e) => setFormData(prev => ({ ...prev, plural_label: e.target.value }))}
                        disabled={!isEditing}
                        className={`w-full px-3 py-2 border rounded-lg transition-colors ${isEditing
                            ? 'bg-white border-slate-300 focus:ring-2 focus:ring-blue-500'
                            : 'bg-slate-50 border-transparent text-slate-600'}`}
                    />
                </div>
            </div>
            <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Description</label>
                <textarea
                    value={isEditing ? formData.description : (object.description || '')}
                    onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                    disabled={!isEditing}
                    rows={3}
                    className={`w-full px-3 py-2 border rounded-lg transition-colors ${isEditing
                        ? 'bg-white border-slate-300 focus:ring-2 focus:ring-blue-500'
                        : 'bg-slate-50 border-transparent text-slate-600'}`}
                />
            </div>
        </div>
    );
};
