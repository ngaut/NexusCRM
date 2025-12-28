import React, { useState } from 'react';
import { Check } from 'lucide-react';
import type { ObjectMetadata, FieldPermission } from '../../types';

interface FieldPermissionsEditorProps {
    schemas: ObjectMetadata[];
    fieldPermissions: FieldPermission[];
    onToggle: (objectApiName: string, fieldApiName: string, type: 'readable' | 'editable') => void;
}

export const FieldPermissionsEditor: React.FC<FieldPermissionsEditorProps> = ({
    schemas,
    fieldPermissions,
    onToggle,
}) => {
    const [selectedObject, setSelectedObject] = useState<string | null>(schemas.length > 0 ? schemas[0].api_name : null);
    const [searchTerm, setSearchTerm] = useState('');

    return (
        <div className="flex gap-6 h-full">
            {/* Object Selector */}
            <div className="w-1/4 bg-white rounded-lg border border-slate-200 flex flex-col overflow-hidden">
                <div className="p-3 bg-slate-50 border-b border-slate-200 flex flex-col gap-2">
                    <div className="font-medium text-xs text-slate-500 uppercase tracking-wider">
                        Select Object
                    </div>
                    <div className="relative">
                        <input
                            type="text"
                            placeholder="Search objects..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="w-full text-xs px-2 py-1.5 border border-slate-300 rounded focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                </div>
                <div className="flex-1 overflow-y-auto">
                    {schemas
                        .filter(s => s.label.toLowerCase().includes(searchTerm.toLowerCase()) || s.api_name.toLowerCase().includes(searchTerm.toLowerCase()))
                        .sort((a, b) => a.label.localeCompare(b.label))
                        .map(schema => (
                            <button
                                key={schema.api_name}
                                onClick={() => setSelectedObject(schema.api_name)}
                                className={`w-full text-left px-4 py-3 text-sm flex items-center justify-between hover:bg-slate-50 transition-colors ${selectedObject === schema.api_name ? 'bg-blue-50 text-blue-700 font-medium' : 'text-slate-700'
                                    }`}
                            >
                                {schema.label}
                                {selectedObject === schema.api_name && <Check size={14} />}
                            </button>
                        ))}
                </div>
            </div>

            {/* Fields Table */}
            <div className="flex-1 bg-white rounded-lg border border-slate-200 flex flex-col overflow-hidden">
                {selectedObject ? (
                    <>
                        <div className="p-3 bg-slate-50 border-b border-slate-200 font-medium text-xs text-slate-500 uppercase tracking-wider flex justify-between">
                            <span>Fields for {schemas.find(s => s.api_name === selectedObject)?.label}</span>
                            <span className="text-xs normal-case text-slate-400">
                                {fieldPermissions.filter(p => p.object_api_name === selectedObject).length} fields
                            </span>
                        </div>
                        <div className="flex-1 overflow-y-auto">
                            <table className="w-full text-sm text-left">
                                <thead className="bg-slate-50 text-slate-600 font-medium border-b border-slate-200 sticky top-0">
                                    <tr>
                                        <th className="px-4 py-3">Field Name</th>
                                        <th className="px-4 py-3 text-center w-24">Read Access</th>
                                        <th className="px-4 py-3 text-center w-24">Edit Access</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-100">
                                    {fieldPermissions
                                        .filter(p => p.object_api_name === selectedObject)
                                        .map(perm => {
                                            const schema = schemas.find(s => s.api_name === selectedObject);
                                            const field = schema?.fields.find(f => f.api_name === perm.field_api_name);

                                            return (
                                                <tr key={perm.field_api_name} className="hover:bg-slate-50">
                                                    <td className="px-4 py-3">
                                                        <div className="font-medium text-slate-900">{field?.label || perm.field_api_name}</div>
                                                        <div className="text-xs text-slate-400 font-mono">{perm.field_api_name}</div>
                                                    </td>
                                                    <td className="px-4 py-3 text-center">
                                                        <input
                                                            type="checkbox"
                                                            checked={perm.readable}
                                                            onChange={() => onToggle(perm.object_api_name, perm.field_api_name, 'readable')}
                                                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                                                        />
                                                    </td>
                                                    <td className="px-4 py-3 text-center">
                                                        <input
                                                            type="checkbox"
                                                            checked={perm.editable}
                                                            onChange={() => onToggle(perm.object_api_name, perm.field_api_name, 'editable')}
                                                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                                                        />
                                                    </td>
                                                </tr>
                                            );
                                        })}
                                </tbody>
                            </table>
                        </div>
                    </>
                ) : (
                    <div className="flex-1 flex items-center justify-center text-slate-400">
                        Select an object to view fields
                    </div>
                )}
            </div>
        </div>
    );
};
