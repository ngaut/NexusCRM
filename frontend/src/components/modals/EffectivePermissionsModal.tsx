import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { Shield, Loader2, X, Check, Search, Eye, Edit2, Trash2, Plus, Info } from 'lucide-react';
import { usersAPI } from '../../infrastructure/api/users';
import { metadataAPI } from '../../infrastructure/api/metadata';
import type { ObjectMetadata, FieldMetadata, ObjectPermission, FieldPermission } from '../../types';

interface EffectivePermissionsModalProps {
    user: { id?: string; name: string };
    onClose: () => void;
}

export const EffectivePermissionsModal: React.FC<EffectivePermissionsModalProps> = ({ user, onClose }) => {
    const [activeTab, setActiveTab] = useState<'objects' | 'fields'>('objects');
    const [permissions, setPermissions] = useState<ObjectPermission[]>([]);
    const [fieldPermissions, setFieldPermissions] = useState<FieldPermission[]>([]);
    const [schemas, setSchemas] = useState<ObjectMetadata[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Filter state for fields tab
    const [selectedObject, setSelectedObject] = useState<string | null>(null);
    const [searchTerm, setSearchTerm] = useState('');

    useEffect(() => {
        if (user.id) {
            loadData();
        }
    }, [user.id]);

    const loadData = async () => {
        try {
            setLoading(true);
            const [schemasResponse, permsResponse, fieldPermsResponse] = await Promise.all([
                metadataAPI.getSchemas(),
                usersAPI.getUserEffectivePermissions(user.id),
                usersAPI.getUserEffectiveFieldPermissions(user.id)
            ]);

            setSchemas(schemasResponse.schemas);
            if (schemasResponse.schemas.length > 0) {
                setSelectedObject(schemasResponse.schemas[0].api_name);
            }

            // Map effective permissions
            // Backend returns aggregated rows. Use them directly.
            setPermissions(permsResponse.data || []);
            setFieldPermissions(fieldPermsResponse.data || []);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load effective permissions');
        } finally {
            setLoading(false);
        }
    };

    const getObjectPermission = (objectName: string) => {
        return permissions.find(p => p.object_api_name === objectName) || {
            object_api_name: objectName,
            allow_read: false,
            allow_create: false,
            allow_edit: false,
            allow_delete: false,
            view_all: false,
            modify_all: false
        } as ObjectPermission;
    };

    const getFieldPermission = (objectName: string, fieldName: string) => {
        // If specific field perm exists, use it
        const fieldPerm = fieldPermissions.find(p => p.object_api_name === objectName && p.field_api_name === fieldName);
        if (fieldPerm) return fieldPerm;

        // Fallback to object-level read permission for readability? 
        // Backend CheckFieldVisibilityWithUser logic:
        // If field perm exists, use it. Else check object read.
        // But for "Effective View", we should probably show exactly what the backend calculated.
        // Wait, backend `GetEffectiveFieldPermissions` only returns rows where MAX(readable) or MAX(editable) is found.
        // If no row, it means no explicit perm.
        // But we want to show "True" if object is readable (implicit).
        // Let's implement frontend logic matching backend fallbacks for display.

        const objPerm = getObjectPermission(objectName);
        return {
            object_api_name: objectName,
            field_api_name: fieldName,
            readable: objPerm.allow_read,
            editable: objPerm.allow_edit
        } as FieldPermission;
    };

    const StatusIcon = ({ check }: { check: boolean }) => (
        check ? <Check size={18} className="text-green-600 inline" /> : <X size={18} className="text-slate-300 inline" />
    );

    return createPortal(
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-5xl max-h-[90vh] flex flex-col">
                <div className="flex items-center justify-between p-6 border-b border-slate-100">
                    <div>
                        <div className="flex items-center gap-2">
                            <Shield className="text-purple-600" size={24} />
                            <h2 className="text-xl font-bold text-slate-900">Effective Permissions: {user.name}</h2>
                        </div>
                        <p className="text-sm text-slate-500 mt-1">
                            Combined permissions from Profile and all assigned Permission Sets.
                        </p>
                    </div>

                    <div className="flex gap-4">
                        <div className="flex bg-slate-100 p-1 rounded-lg">
                            <button
                                onClick={() => setActiveTab('objects')}
                                className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${activeTab === 'objects'
                                    ? 'bg-white text-blue-600 shadow-sm'
                                    : 'text-slate-600 hover:text-slate-900'
                                    }`}
                            >
                                Objects
                            </button>
                            <button
                                onClick={() => setActiveTab('fields')}
                                className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${activeTab === 'fields'
                                    ? 'bg-white text-blue-600 shadow-sm'
                                    : 'text-slate-600 hover:text-slate-900'
                                    }`}
                            >
                                Fields
                            </button>
                        </div>
                        <button
                            onClick={onClose}
                            className="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-50"
                        >
                            <X size={24} />
                        </button>
                    </div>
                </div>

                <div className="flex-1 overflow-y-auto p-6 bg-slate-50">
                    {loading ? (
                        <div className="flex items-center justify-center h-64">
                            <Loader2 className="animate-spin text-blue-600" size={32} />
                        </div>
                    ) : error ? (
                        <div className="text-center text-red-500 py-8 bg-red-50 rounded-lg border border-red-100">
                            {error}
                        </div>
                    ) : activeTab === 'objects' ? (
                        <div className="bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden">
                            <table className="w-full text-sm text-left">
                                <thead className="bg-slate-50 border-b border-slate-200">
                                    <tr>
                                        <th className="px-6 py-3 font-semibold text-slate-700">Object</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-24">Read</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-24">Create</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-24">Edit</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-24">Delete</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-24">View All</th>
                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-24">Modify All</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-100">
                                    {schemas.map(schema => {
                                        const perm = getObjectPermission(schema.api_name);
                                        return (
                                            <tr key={schema.api_name} className="hover:bg-slate-50">
                                                <td className="px-6 py-3 font-medium text-slate-900 border-r border-slate-100">
                                                    {schema.label}
                                                    <span className="ml-2 text-xs text-slate-400 font-mono font-normal">{schema.api_name}</span>
                                                </td>
                                                <td className="px-6 py-3 text-center bg-blue-50/30">
                                                    <StatusIcon check={perm.allow_read} />
                                                </td>
                                                <td className="px-6 py-3 text-center">
                                                    <StatusIcon check={perm.allow_create} />
                                                </td>
                                                <td className="px-6 py-3 text-center">
                                                    <StatusIcon check={perm.allow_edit} />
                                                </td>
                                                <td className="px-6 py-3 text-center">
                                                    <StatusIcon check={perm.allow_delete} />
                                                </td>
                                                <td className="px-6 py-3 text-center bg-yellow-50/30 border-l border-slate-100">
                                                    <StatusIcon check={perm.view_all} />
                                                </td>
                                                <td className="px-6 py-3 text-center bg-yellow-50/30">
                                                    <StatusIcon check={perm.modify_all} />
                                                </td>
                                            </tr>
                                        );
                                    })}
                                </tbody>
                            </table>
                        </div>
                    ) : (
                        <div className="flex gap-6 h-full">
                            {/* Object List */}
                            <div className="w-1/4 min-w-[250px] bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden flex flex-col">
                                <div className="p-3 border-b border-slate-200">
                                    <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2">Objects</h3>
                                    <div className="relative">
                                        <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400" />
                                        <input
                                            type="text"
                                            placeholder="Search objects..."
                                            value={searchTerm}
                                            onChange={e => setSearchTerm(e.target.value)}
                                            className="w-full pl-8 pr-3 py-1.5 text-sm border border-slate-200 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                                        />
                                    </div>
                                </div>
                                <div className="flex-1 overflow-y-auto">
                                    {schemas
                                        .filter(s => s.label.toLowerCase().includes(searchTerm.toLowerCase()) || s.api_name.toLowerCase().includes(searchTerm.toLowerCase()))
                                        .map(schema => (
                                            <button
                                                key={schema.api_name}
                                                onClick={() => setSelectedObject(schema.api_name)}
                                                className={`w-full text-left px-4 py-3 text-sm flex items-center justify-between border-b border-slate-50 hover:bg-slate-50 transition-colors ${selectedObject === schema.api_name ? 'bg-blue-50 text-blue-700 font-medium' : 'text-slate-700'
                                                    }`}
                                            >
                                                <span>{schema.label}</span>
                                            </button>
                                        ))}
                                </div>
                            </div>

                            {/* Field Table */}
                            <div className="flex-1 bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden flex flex-col">
                                {selectedObject ? (
                                    <>
                                        <div className="p-4 border-b border-slate-200 bg-slate-50/50 flex items-center justify-between">
                                            <h3 className="font-semibold text-slate-800">
                                                {schemas.find(s => s.api_name === selectedObject)?.label} Fields
                                            </h3>
                                        </div>
                                        <div className="flex-1 overflow-y-auto">
                                            <table className="w-full text-sm text-left">
                                                <thead className="bg-slate-50 border-b border-slate-200 sticky top-0">
                                                    <tr>
                                                        <th className="px-6 py-3 font-semibold text-slate-700">Field Label</th>
                                                        <th className="px-6 py-3 font-semibold text-slate-700">API Name</th>
                                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-32">Read Access</th>
                                                        <th className="px-6 py-3 font-semibold text-slate-700 text-center w-32">Edit Access</th>
                                                    </tr>
                                                </thead>
                                                <tbody className="divide-y divide-slate-100">
                                                    {(schemas.find(s => s.api_name === selectedObject)?.fields || []).map(field => {
                                                        const perm = getFieldPermission(selectedObject, field.api_name);
                                                        return (
                                                            <tr key={field.api_name} className="hover:bg-slate-50">
                                                                <td className="px-6 py-3 text-slate-900 font-medium">{field.label}</td>
                                                                <td className="px-6 py-3 text-slate-500 font-mono text-xs">{field.api_name}</td>
                                                                <td className="px-6 py-3 text-center">
                                                                    <StatusIcon check={perm.readable} />
                                                                </td>
                                                                <td className="px-6 py-3 text-center">
                                                                    <StatusIcon check={perm.editable} />
                                                                </td>
                                                            </tr>
                                                        );
                                                    })}
                                                </tbody>
                                            </table>
                                        </div>
                                    </>
                                ) : (
                                    <div className="flex flex-col items-center justify-center h-full text-slate-400">
                                        <p>Select an object to view field permissions</p>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                <div className="p-4 border-t border-slate-200 bg-slate-50/50 flex justify-end">
                    <button
                        onClick={onClose}
                        className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm"
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
};
