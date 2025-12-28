import React, { useState, useEffect } from 'react';
import { X, Save, AlertCircle } from 'lucide-react';
import { usersAPI } from '../../infrastructure/api/users';
import { metadataAPI } from '../../infrastructure/api/metadata';
import type { ObjectMetadata, ObjectPermission, FieldPermission } from '../../types';

// Sub-components
import { ObjectPermissionsTable } from '../permissions/ObjectPermissionsTable';
import { FieldPermissionsEditor } from '../permissions/FieldPermissionsEditor';

export type PermissionEntityType = 'profile' | 'permission_set';

export interface PermissionEntity {
    id: string;
    name: string;
    type: PermissionEntityType;
}

interface PermissionEditorModalProps {
    entity: PermissionEntity;
    onClose: () => void;
    onSave: () => void;
}

export const PermissionEditorModal: React.FC<PermissionEditorModalProps> = ({ entity, onClose, onSave }) => {
    const [activeTab, setActiveTab] = useState<'objects' | 'fields'>('objects');
    const [permissions, setPermissions] = useState<ObjectPermission[]>([]); // Object Perms
    const [fieldPermissions, setFieldPermissions] = useState<FieldPermission[]>([]); // Field Perms
    const [schemas, setSchemas] = useState<ObjectMetadata[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Determine the entity ID field name based on type
    const entityIdField = entity.type === 'profile' ? 'profile_id' : 'permission_set_id';

    useEffect(() => {
        loadData();
    }, [entity.id]);

    const loadData = async () => {
        try {
            setLoading(true);

            // Use different API methods based on entity type
            const getObjPerms = entity.type === 'profile'
                ? usersAPI.getProfilePermissions(entity.id)
                : usersAPI.getPermissionSetPermissions(entity.id);

            const getFieldPerms = entity.type === 'profile'
                ? usersAPI.getProfileFieldPermissions(entity.id)
                : usersAPI.getPermissionSetFieldPermissions(entity.id);

            const [schemasResponse, permsResponse, fieldPermsResponse] = await Promise.all([
                metadataAPI.getSchemas(),
                getObjPerms,
                getFieldPerms
            ]);

            setSchemas(schemasResponse.schemas);

            // --- Object Permissions Merge ---
            const existingPerms = permsResponse.permissions || [];
            const mergedPerms: ObjectPermission[] = schemasResponse.schemas.map(schema => {
                const existing = existingPerms.find(p => p.object_api_name === schema.api_name);
                if (existing) return existing;

                return {
                    [entityIdField]: entity.id,
                    object_api_name: schema.api_name,
                    allow_read: false,
                    allow_create: false,
                    allow_edit: false,
                    allow_delete: false,
                    view_all: false,
                    modify_all: false
                };
            });
            setPermissions(mergedPerms);

            // --- Field Permissions Merge ---
            const existingFieldPerms = fieldPermsResponse.permissions || [];
            const allFieldPerms: FieldPermission[] = [];

            schemasResponse.schemas.forEach(schema => {
                const fields = schema.fields || [];
                fields.forEach(field => {
                    // Skip 'id' field
                    if (field.api_name === 'id') return;

                    const existing = existingFieldPerms.find(p =>
                        p.object_api_name === schema.api_name && p.field_api_name === field.api_name
                    );

                    allFieldPerms.push({
                        [entityIdField]: entity.id,
                        object_api_name: schema.api_name,
                        field_api_name: field.api_name,
                        readable: existing ? existing.readable : false, // Default to false? Or true? Usually default false for safety.
                        editable: existing ? existing.editable : false
                    });
                });
            });

            setFieldPermissions(allFieldPerms);

        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load permissions');
        } finally {
            setLoading(false);
        }
    };

    const handleToggleObject = (objectApiName: string, field: keyof ObjectPermission) => {
        setPermissions(prev => prev.map(p => {
            if (p.object_api_name !== objectApiName) return p;

            const updated = { ...p, [field]: !p[field] };

            // Logic cascading
            if (field === 'modify_all' && updated.modify_all) {
                updated.allow_read = true;
                updated.allow_create = true;
                updated.allow_edit = true;
                updated.allow_delete = true;
                updated.view_all = true;
            }
            if (field === 'view_all' && updated.view_all) {
                updated.allow_read = true;
            }
            if (!updated.allow_read && (updated.allow_create || updated.allow_edit || updated.allow_delete)) {
                updated.allow_read = true; // Auto-enable read if other rights exist
            }

            return updated;
        }));
    };

    const handleToggleField = (objectApiName: string, fieldApiName: string, type: 'readable' | 'editable') => {
        setFieldPermissions(prev => prev.map(p => {
            if (p.object_api_name !== objectApiName || p.field_api_name !== fieldApiName) return p;

            const updated = { ...p, [type]: !p[type] };

            // Logic: if editable is checked, readable must be checked
            if (type === 'editable' && updated.editable) {
                updated.readable = true;
            }
            // Logic: if readable is unchecked, editable must be unchecked
            if (type === 'readable' && !updated.readable) {
                updated.editable = false;
            }

            return updated;
        }));
    };

    const handleSave = async () => {
        try {
            setSaving(true);

            const fieldPermsPayload = fieldPermissions.map(p => ({
                object_api_name: p.object_api_name,
                field_api_name: p.field_api_name,
                readable: p.readable,
                editable: p.editable
            }));

            if (entity.type === 'profile') {
                await Promise.all([
                    usersAPI.updateProfilePermissions(entity.id, permissions),
                    usersAPI.updateProfileFieldPermissions(entity.id, fieldPermsPayload)
                ]);
            } else {
                await Promise.all([
                    usersAPI.updatePermissionSetPermissions(entity.id, permissions),
                    usersAPI.updatePermissionSetFieldPermissions(entity.id, fieldPermsPayload)
                ]);
            }
            onSave();
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to save permissions');
            setSaving(false);
        }
    };

    if (loading) return (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-50">
            <div className="bg-white rounded-xl shadow-2xl p-8">Loading permissions...</div>
        </div>
    );

    return (
        <div className="fixed inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-5xl max-h-[90vh] flex flex-col">
                <div className="flex items-center justify-between p-6 border-b border-slate-100">
                    <div>
                        <h2 className="text-xl font-bold text-slate-900">Edit Permissions: {entity.name}</h2>
                        <div className="flex gap-4 mt-4">
                            <button
                                onClick={() => setActiveTab('objects')}
                                className={`text-sm font-medium pb-1 border-b-2 transition-colors ${activeTab === 'objects' ? 'border-blue-600 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700'
                                    }`}
                            >
                                Object Permissions
                            </button>
                            <button
                                onClick={() => setActiveTab('fields')}
                                className={`text-sm font-medium pb-1 border-b-2 transition-colors ${activeTab === 'fields' ? 'border-blue-600 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700'
                                    }`}
                            >
                                Field Permissions
                            </button>
                        </div>
                    </div>
                    <button onClick={onClose} className="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-50">
                        <X size={20} />
                    </button>
                </div>

                <div className="flex-1 overflow-auto p-6 bg-slate-50">
                    {error && (
                        <div className="mb-6 p-4 bg-red-50 text-red-700 rounded-lg flex items-center gap-2 border border-red-100">
                            <AlertCircle size={20} />
                            {error}
                        </div>
                    )}

                    {activeTab === 'objects' ? (
                        <ObjectPermissionsTable
                            permissions={permissions}
                            schemas={schemas}
                            onToggle={handleToggleObject}
                        />
                    ) : (
                        <FieldPermissionsEditor
                            schemas={schemas}
                            fieldPermissions={fieldPermissions}
                            onToggle={handleToggleField}
                        />
                    )}
                </div>

                <div className="p-6 border-t border-slate-100 flex justify-end gap-3 bg-slate-50 rounded-b-xl">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-slate-700 hover:bg-white border border-transparent hover:border-slate-200 rounded-lg transition-all"
                        disabled={saving}
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="flex items-center gap-2 px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-all shadow-sm hover:shadow active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {saving ? (
                            <>
                                <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                                Saving...
                            </>
                        ) : (
                            <>
                                <Save size={18} />
                                Save Permissions
                            </>
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
};
