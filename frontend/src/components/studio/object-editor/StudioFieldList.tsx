import React from 'react';
import { Plus, Edit2, Trash2 } from 'lucide-react';
import * as Icons from 'lucide-react';
import type { FieldMetadata, FieldType } from '../../../types';

interface StudioFieldListProps {
    customFields: FieldMetadata[];
    systemFields: FieldMetadata[];
    onAddField: () => void;
    onEditField: (field: FieldMetadata) => void;
    onDeleteField: (fieldApiName: string) => void;
}

export const StudioFieldList: React.FC<StudioFieldListProps> = ({
    customFields,
    systemFields,
    onAddField,
    onEditField,
    onDeleteField,
}) => {
    const getFieldTypeIcon = (type: FieldType): React.ComponentType<{ size?: number | string; className?: string }> => {
        const iconMap: Record<string, keyof typeof Icons> = {
            Text: 'Type',
            TextArea: 'AlignLeft',
            Number: 'Hash',
            Currency: 'DollarSign',
            Percent: 'Percent',
            Date: 'Calendar',
            DateTime: 'Clock',
            Boolean: 'ToggleLeft',
            Email: 'Mail',
            Phone: 'Phone',
            Url: 'Link',
            Picklist: 'List',
            Lookup: 'Link2',
            Formula: 'Sigma',
            RollupSummary: 'Calculator',
            JSON: 'Braces',
        };
        const iconName = iconMap[type] || 'Circle';
        return Icons[iconName as keyof typeof Icons] as React.ComponentType<{ size?: number | string; className?: string }>;
    };

    const getFieldTypeBadgeColor = (type: FieldType): string => {
        const colorMap: Record<string, string> = {
            Text: 'bg-slate-100 text-slate-600',
            TextArea: 'bg-slate-100 text-slate-600',
            Number: 'bg-blue-100 text-blue-700',
            Currency: 'bg-emerald-100 text-emerald-700',
            Percent: 'bg-blue-100 text-blue-700',
            Date: 'bg-amber-100 text-amber-700',
            DateTime: 'bg-amber-100 text-amber-700',
            Boolean: 'bg-purple-100 text-purple-700',
            Picklist: 'bg-indigo-100 text-indigo-700',
            Lookup: 'bg-cyan-100 text-cyan-700',
            Formula: 'bg-pink-100 text-pink-700',
        };
        return colorMap[type] || 'bg-slate-100 text-slate-600';
    };

    return (
        <div className="p-5">
            {/* Add Field Button */}
            <div className="flex justify-between items-center mb-4">
                <div>
                    <h3 className="font-semibold text-slate-800">Fields ({customFields.length})</h3>
                    <p className="text-sm text-slate-500">Define the data structure for this object</p>
                </div>
                <button
                    onClick={onAddField}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium"
                >
                    <Plus size={16} />
                    New Field
                </button>
            </div>

            {/* Fields List */}
            <div className="border rounded-lg overflow-hidden">
                {customFields.length === 0 ? (
                    <div className="px-6 py-12 text-center">
                        <Icons.Columns size={40} className="mx-auto text-slate-300 mb-3" />
                        <p className="text-slate-500 mb-4">No custom fields yet</p>
                        <button
                            onClick={onAddField}
                            className="text-blue-600 hover:text-blue-700 font-medium text-sm"
                        >
                            + Add your first field
                        </button>
                    </div>
                ) : (
                    <table className="w-full">
                        <thead className="bg-slate-50 border-b">
                            <tr>
                                <th className="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                                    Field
                                </th>
                                <th className="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                                    Type
                                </th>
                                <th className="text-left px-4 py-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">
                                    Properties
                                </th>
                                <th className="w-20"></th>
                            </tr>
                        </thead>
                        <tbody className="divide-y">
                            {customFields.map(field => {
                                const FieldIcon = getFieldTypeIcon(field.type);
                                return (
                                    <tr key={field.api_name} className="hover:bg-slate-50 group">
                                        <td className="px-4 py-3">
                                            <div className="font-medium text-slate-800">{field.label}</div>
                                            <div className="text-xs text-slate-500 font-mono">{field.api_name}</div>
                                        </td>
                                        <td className="px-4 py-3">
                                            <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${getFieldTypeBadgeColor(field.type)}`}>
                                                <FieldIcon size={12} />
                                                {field.type}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3">
                                            <div className="flex gap-2">
                                                {field.required && (
                                                    <span className="px-2 py-0.5 bg-red-100 text-red-700 rounded text-xs">Required</span>
                                                )}
                                                {field.unique && (
                                                    <span className="px-2 py-0.5 bg-amber-100 text-amber-700 rounded text-xs">Unique</span>
                                                )}
                                                {field.type === 'Lookup' && field.reference_to && (
                                                    <span className="px-2 py-0.5 bg-cyan-100 text-cyan-700 rounded text-xs">
                                                        → {field.reference_to}
                                                    </span>
                                                )}
                                            </div>
                                        </td>
                                        <td className="px-4 py-3">
                                            <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                                <button
                                                    onClick={() => onEditField(field)}
                                                    className="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 rounded"
                                                >
                                                    <Edit2 size={14} />
                                                </button>
                                                <button
                                                    onClick={() => onDeleteField(field.api_name)}
                                                    className="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded"
                                                >
                                                    <Trash2 size={14} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                )}
            </div>

            {/* System Fields (Collapsed) */}
            {systemFields.length > 0 && (
                <details className="mt-6">
                    <summary className="text-sm text-slate-500 cursor-pointer hover:text-slate-700">
                        System Fields ({systemFields.length})
                    </summary>
                    <div className="mt-2 border rounded-lg overflow-hidden bg-slate-50">
                        <table className="w-full text-sm">
                            <tbody className="divide-y divide-slate-200">
                                {systemFields.map(field => (
                                    <tr key={field.api_name}>
                                        <td className="px-4 py-2 text-slate-600">{field.label}</td>
                                        <td className="px-4 py-2 text-slate-500 font-mono text-xs">{field.api_name}</td>
                                        <td className="px-4 py-2 text-slate-500">{field.type}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </details>
            )}
        </div>
    );
};
