import React from 'react';
import { Plus, ArrowLeft, Trash2 } from 'lucide-react';
import type { ObjectMetadata } from '../../../types';

interface ListViewEditorProps {
    metadata: ObjectMetadata;
    listFields: string[];
    onToggleField: (fieldApiName: string, targetList: 'list') => void;
    onMoveField: (fieldApiName: string, direction: 'up' | 'down', targetList: 'list') => void;
}

export const ListViewEditor: React.FC<ListViewEditorProps> = ({
    metadata,
    listFields,
    onToggleField,
    onMoveField
}) => {
    return (
        <div className="flex-1 overflow-y-auto p-8 flex gap-8">
            {/* Available Fields */}
            <div className="w-1/3 bg-white rounded-lg border border-slate-200 shadow-sm flex flex-col">
                <div className="p-4 border-b border-slate-200 bg-slate-50 font-semibold text-slate-700">Available Fields</div>
                <div className="flex-1 overflow-y-auto p-2 space-y-1">
                    {metadata.fields
                        .filter(f => !listFields.includes(f.api_name))
                        .map(field => (
                            <div
                                key={field.api_name}
                                className="p-2 bg-white border border-slate-200 rounded hover:bg-slate-50 cursor-pointer flex justify-between items-center group"
                                onClick={() => onToggleField(field.api_name, 'list')}
                            >
                                <span className="text-sm font-medium text-slate-700">{field.label}</span>
                                <Plus size={16} className="text-blue-500 opacity-0 group-hover:opacity-100" />
                            </div>
                        ))}
                </div>
            </div>

            {/* Selected Fields */}
            <div className="w-1/3 bg-white rounded-lg border border-slate-200 shadow-sm flex flex-col">
                <div className="p-4 border-b border-slate-200 bg-slate-50 font-semibold text-slate-700 flex justify-between">
                    <span>Selected Columns</span>
                    <span className="text-xs font-normal text-slate-500 self-center">{listFields.length} selected</span>
                </div>
                <div className="flex-1 overflow-y-auto p-2 space-y-1">
                    {listFields.map((apiName, index) => {
                        const field = metadata.fields.find(f => f.api_name === apiName);
                        if (!field) return null;
                        return (
                            <div key={apiName} className="p-2 bg-blue-50 border border-blue-200 rounded flex justify-between items-center group">
                                <span className="text-sm font-medium text-slate-800">{field.label}</span>
                                <div className="flex items-center gap-1">
                                    <button onClick={() => onMoveField(apiName, 'up', 'list')} disabled={index === 0} className="p-1 text-slate-400 hover:text-blue-600 disabled:opacity-30"><ArrowLeft className="rotate-90" size={14} /></button>
                                    <button onClick={() => onMoveField(apiName, 'down', 'list')} disabled={index === listFields.length - 1} className="p-1 text-slate-400 hover:text-blue-600 disabled:opacity-30"><ArrowLeft className="-rotate-90" size={14} /></button>
                                    <button onClick={() => onToggleField(apiName, 'list')} className="p-1 text-slate-400 hover:text-red-600"><Trash2 size={14} /></button>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </div>

            {/* Preview / Info */}
            <div className="w-1/3 bg-slate-50 rounded-lg p-6 border border-slate-200">
                <h3 className="font-semibold text-slate-700 mb-2">List View Configuration</h3>
                <p className="text-sm text-slate-500">
                    Select the fields that should appear as columns in the default list view for this object.
                    The order of fields here determines the column order.
                </p>
            </div>
        </div>
    );
};
