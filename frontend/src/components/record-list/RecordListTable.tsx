import React from 'react';
import { Plus } from 'lucide-react';
import { SObject, FieldMetadata, ObjectMetadata } from '../../types';
import { UIRegistry } from '../../registries/UIRegistry';
import { COMMON_FIELDS } from '../../core/constants';

interface RecordListTableProps {
    records: SObject[];
    displayFields: FieldMetadata[];
    selectedRecords: Set<string>;
    toggleSelectAll: () => void;
    toggleSelectRecord: (id: string) => void;
    sortField: string;
    sortDirection: 'asc' | 'desc';
    onSort: (field: string) => void;
    handleRecordClick: (record: SObject) => void;
    objectMetadata: ObjectMetadata;
    onNavigate: (obj: string | undefined, id: string) => void;
    onCreateNew?: () => void;
}

export const RecordListTable: React.FC<RecordListTableProps> = ({
    records,
    displayFields,
    selectedRecords,
    toggleSelectAll,
    toggleSelectRecord,
    sortField,
    sortDirection,
    onSort,
    handleRecordClick,
    objectMetadata,
    onNavigate,
    onCreateNew
}) => {
    return (
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
            <div className="overflow-x-auto">
                <table className="w-full">
                    <thead className="bg-gray-50 border-b border-gray-200">
                        <tr>
                            <th className="w-12 px-4 py-3">
                                <input
                                    type="checkbox"
                                    checked={selectedRecords.size === records.length && records.length > 0}
                                    onChange={toggleSelectAll}
                                    className="rounded border-gray-300"
                                />
                            </th>
                            {displayFields.map(field => (
                                <th
                                    key={field.api_name}
                                    className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                                    onClick={() => onSort(field.api_name)}
                                >
                                    <div className="flex items-center gap-1">
                                        {field.label}
                                        {sortField === field.api_name && (
                                            <span className="text-blue-600">
                                                {sortDirection === 'asc' ? '↑' : '↓'}
                                            </span>
                                        )}
                                    </div>
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                        {records.length === 0 ? (
                            <tr>
                                <td colSpan={displayFields.length + 1} className="px-4 py-12 text-center">
                                    <div className="flex flex-col items-center gap-3">
                                        <p className="text-slate-500">No records found</p>
                                        {onCreateNew && (
                                            <button
                                                onClick={onCreateNew}
                                                className="inline-flex items-center gap-2 px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                                            >
                                                <Plus size={16} />
                                                Create {objectMetadata.label}
                                            </button>
                                        )}
                                    </div>
                                </td>
                            </tr>
                        ) : (
                            records.map(record => (
                                <tr
                                    key={record[COMMON_FIELDS.ID] as string}
                                    className="hover:bg-gray-50 cursor-pointer transition-colors"
                                    onClick={() => handleRecordClick(record)}
                                >
                                    <td
                                        className="px-4 py-3"
                                        onClick={(e) => e.stopPropagation()}
                                    >
                                        <input
                                            type="checkbox"
                                            checked={selectedRecords.has(record[COMMON_FIELDS.ID] as string)}
                                            onChange={() => toggleSelectRecord(record[COMMON_FIELDS.ID] as string)}
                                            className="rounded border-gray-300"
                                        />
                                    </td>
                                    {displayFields.map(field => {
                                        const Renderer = UIRegistry.getFieldRenderer(field.type);
                                        return (
                                            <td key={field.api_name} className="px-4 py-3 text-sm text-gray-900">
                                                <Renderer
                                                    field={field}
                                                    value={record[field.api_name]}
                                                    record={record}
                                                    variant="table"
                                                    onNavigate={onNavigate}
                                                />
                                            </td>
                                        );
                                    })}
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
