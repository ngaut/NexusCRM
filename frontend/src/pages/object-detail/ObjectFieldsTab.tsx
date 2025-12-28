import React, { useMemo, useState } from 'react';
import { Search, Filter, Plus, Edit, Trash2 } from 'lucide-react';
import { FieldMetadata, FieldType, ObjectMetadata } from '../../types';
import { MetadataRegistry } from '../../registries/MetadataRegistry';

interface ObjectFieldsTabProps {
    metadata: ObjectMetadata;
    onOpenCreate: () => void;
    onOpenEdit: (field: FieldMetadata) => void;
    onDelete: (fieldApiName: string) => void;
}

export const ObjectFieldsTab: React.FC<ObjectFieldsTabProps> = ({
    metadata,
    onOpenCreate,
    onOpenEdit,
    onDelete
}) => {
    const [searchQuery, setSearchQuery] = useState('');
    const [typeFilter, setTypeFilter] = useState<string>('all');
    const fieldTypes: FieldType[] = MetadataRegistry.getValues().map(def => def.type);

    const filteredFields = useMemo(() => {
        if (!metadata?.fields) return [];
        return metadata.fields.filter(field => {
            const matchesSearch = field.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
                field.api_name.toLowerCase().includes(searchQuery.toLowerCase());
            const matchesType = typeFilter === 'all' || field.type === typeFilter;
            return matchesSearch && matchesType;
        });
    }, [metadata?.fields, searchQuery, typeFilter]);

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <div className="flex gap-4 flex-1 max-w-2xl">
                    <div className="relative flex-1">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
                        <input
                            type="text"
                            placeholder="Search fields..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                    <div className="relative w-48">
                        <Filter className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
                        <select
                            value={typeFilter}
                            onChange={(e) => setTypeFilter(e.target.value)}
                            className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 appearance-none bg-white"
                        >
                            <option value="all">All Types</option>
                            {fieldTypes.map(type => (
                                <option key={type} value={type}>{type}</option>
                            ))}
                        </select>
                    </div>
                </div>
                <button
                    onClick={onOpenCreate}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                    <Plus size={18} />
                    New Field
                </button>
            </div>

            <div className="bg-white/80 backdrop-blur-xl border border-white/20 shadow-xl overflow-hidden rounded-2xl">
                <table className="w-full">
                    <thead className="bg-slate-50 border-b border-slate-200">
                        <tr>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">Field Label</th>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">API Name</th>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">Type</th>
                            <th className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">Properties</th>
                            <th className="px-6 py-3 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-200">
                        {filteredFields.map((field) => (
                            <tr key={field.api_name} className="hover:bg-slate-50">
                                <td className="px-6 py-4 whitespace-nowrap">
                                    <div className="flex items-center">
                                        <span className="font-medium text-slate-900">{field.label}</span>
                                        {field.is_name_field && (
                                            <span className="ml-2 px-2 py-0.5 text-xs bg-blue-100 text-blue-800 rounded border border-blue-200">Name Field</span>
                                        )}
                                    </div>
                                    {field.description && (
                                        <p className="text-xs text-slate-500 mt-0.5">{field.description}</p>
                                    )}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap font-mono text-sm text-slate-600">
                                    {field.api_name}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-600">
                                    {field.type}
                                    {field.type === 'Lookup' && field.reference_to && (
                                        <span className="ml-1 text-slate-400">({field.reference_to})</span>
                                    )}
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-sm">
                                    <div className="flex gap-2">
                                        {field.required && (
                                            <span className="px-2 py-0.5 bg-red-50 text-red-700 rounded text-xs border border-red-100">Required</span>
                                        )}
                                        {field.unique && (
                                            <span className="px-2 py-0.5 bg-purple-50 text-purple-700 rounded text-xs border border-purple-100">Unique</span>
                                        )}
                                        {field.is_system && (
                                            <span className="px-2 py-0.5 bg-slate-100 text-slate-600 rounded text-xs border border-slate-200">System</span>
                                        )}
                                    </div>
                                </td>
                                <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                    <div className="flex justify-end gap-2">
                                        {!field.is_system && (
                                            <>
                                                <button
                                                    onClick={() => onOpenEdit(field)}
                                                    className="p-1.5 text-blue-600 hover:bg-blue-50 rounded transition-colors"
                                                    title="Edit"
                                                >
                                                    <Edit size={16} />
                                                </button>
                                                {!field.is_name_field && (
                                                    <button
                                                        onClick={() => onDelete(field.api_name)}
                                                        className="p-1.5 text-red-600 hover:bg-red-50 rounded transition-colors"
                                                        title="Delete"
                                                    >
                                                        <Trash2 size={16} />
                                                    </button>
                                                )}
                                            </>
                                        )}
                                    </div>
                                </td>
                            </tr>
                        ))}
                        {filteredFields.length === 0 && (
                            <tr>
                                <td colSpan={5} className="px-6 py-8 text-center text-slate-500">
                                    No fields found matching your search.
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
