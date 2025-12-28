import React from 'react';
import { Database } from 'lucide-react';
import type { ObjectMetadata, ObjectPermission } from '../../types';

interface ObjectPermissionsTableProps {
    permissions: ObjectPermission[];
    schemas: ObjectMetadata[];
    onToggle: (objectApiName: string, field: keyof ObjectPermission) => void;
}

export const ObjectPermissionsTable: React.FC<ObjectPermissionsTableProps> = ({
    permissions,
    schemas,
    onToggle,
}) => {
    return (
        <div className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden">
            <table className="w-full text-sm text-left">
                <thead className="bg-slate-50 text-slate-600 font-medium border-b border-slate-200">
                    <tr>
                        <th className="px-4 py-3 sticky top-0 bg-slate-50">Object</th>
                        <th className="px-4 py-3 sticky top-0 bg-slate-50 text-center w-24">Read</th>
                        <th className="px-4 py-3 sticky top-0 bg-slate-50 text-center w-24">Create</th>
                        <th className="px-4 py-3 sticky top-0 bg-slate-50 text-center w-24">Edit</th>
                        <th className="px-4 py-3 sticky top-0 bg-slate-50 text-center w-24">Delete</th>
                        <th className="px-4 py-3 sticky top-0 bg-slate-50 text-center w-24 border-l border-slate-200">View All</th>
                        <th className="px-4 py-3 sticky top-0 bg-slate-50 text-center w-24">Modify All</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                    {permissions.map(perm => {
                        const schema = schemas.find(s => s.api_name === perm.object_api_name);
                        const label = schema?.label || perm.object_api_name;

                        return (
                            <tr key={perm.object_api_name} className="hover:bg-slate-50">
                                <td className="px-4 py-3 font-medium text-slate-900 flex items-center gap-2">
                                    <Database size={16} className="text-slate-400" />
                                    {label}
                                    <span className="text-xs text-slate-400 font-normal font-mono ml-1">({perm.object_api_name})</span>
                                </td>
                                <td className="px-4 py-3 text-center">
                                    <input type="checkbox" checked={perm.allow_read} onChange={() => onToggle(perm.object_api_name, 'allow_read')}
                                        className="rounded border-slate-300 text-blue-600 focus:ring-blue-500" />
                                </td>
                                <td className="px-4 py-3 text-center">
                                    <input type="checkbox" checked={perm.allow_create} onChange={() => onToggle(perm.object_api_name, 'allow_create')}
                                        className="rounded border-slate-300 text-blue-600 focus:ring-blue-500" />
                                </td>
                                <td className="px-4 py-3 text-center">
                                    <input type="checkbox" checked={perm.allow_edit} onChange={() => onToggle(perm.object_api_name, 'allow_edit')}
                                        className="rounded border-slate-300 text-blue-600 focus:ring-blue-500" />
                                </td>
                                <td className="px-4 py-3 text-center">
                                    <input type="checkbox" checked={perm.allow_delete} onChange={() => onToggle(perm.object_api_name, 'allow_delete')}
                                        className="rounded border-slate-300 text-blue-600 focus:ring-blue-500" />
                                </td>
                                <td className="px-4 py-3 text-center border-l border-slate-200">
                                    <input type="checkbox" checked={!!perm.view_all} onChange={() => onToggle(perm.object_api_name, 'view_all')}
                                        className="rounded border-slate-300 text-purple-600 focus:ring-purple-500" />
                                </td>
                                <td className="px-4 py-3 text-center">
                                    <input type="checkbox" checked={!!perm.modify_all} onChange={() => onToggle(perm.object_api_name, 'modify_all')}
                                        className="rounded border-slate-300 text-purple-600 focus:ring-purple-500" />
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
};
