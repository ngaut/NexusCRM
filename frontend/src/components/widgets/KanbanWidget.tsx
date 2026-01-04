import React from 'react';
import { WidgetRendererProps } from '../../types';
import { Loader2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useKanban } from '../../core/hooks';
import { getRecordDisplayName } from '../../core/utils/recordUtils';
import { SObject } from '../../types';
import { ROUTES } from '../../core/constants/Routes';

export const KanbanWidget: React.FC<WidgetRendererProps> = ({
    title,
    config,
    isEditing,
    isVisible,
    onToggle
}) => {
    const navigate = useNavigate();
    const objectApiName = config.query?.object_api_name;

    const { schema, columns, groupedRecords, loading, error } = useKanban({
        objectApiName,
        config: config as unknown as Record<string, unknown>
    });

    return (
        <div className={`bg-white rounded-lg border shadow-sm flex flex-col relative h-full overflow-hidden ${isEditing ? 'border-dashed border-2 border-slate-300' : 'border-slate-200'}`}>
            {isEditing && (
                <div className="absolute top-2 right-2 z-10">
                    <button onClick={onToggle} className={`p-1.5 rounded-full ${isVisible ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-500'}`}>
                        {isVisible ? "Hide" : "Show"}
                    </button>
                </div>
            )}

            <div className="p-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
                <h3 className="font-bold text-slate-800 truncate" title={title}>{title || (schema ? `${schema.plural_label} Board` : "Kanban")}</h3>
            </div>

            <div className="flex-1 overflow-x-auto overflow-y-hidden p-0 relative">
                {loading ? (
                    <div className="absolute inset-0 flex items-center justify-center text-slate-400">
                        <Loader2 className="animate-spin mr-2" /> Loading...
                    </div>
                ) : error ? (
                    <div className="absolute inset-0 flex items-center justify-center text-red-500 p-4 text-center text-sm">
                        {error}
                    </div>
                ) : schema ? (
                    <div className="h-full flex p-4 gap-4 min-w-[600px]">
                        {columns.map(col => (
                            <div key={col} className="w-64 flex-shrink-0 flex flex-col bg-slate-50 rounded-lg border border-slate-200 h-full max-h-full">
                                <div className="p-3 border-b border-slate-200 font-medium text-slate-700 flex justify-between items-center sticky top-0 bg-slate-50 rounded-t-lg z-10">
                                    <span className="truncate">{col}</span>
                                    <span className="bg-slate-200 text-slate-600 text-xs px-2 py-0.5 rounded-full">
                                        {groupedRecords[col]?.length || 0}
                                    </span>
                                </div>
                                <div className="p-2 flex-1 overflow-y-auto space-y-2">
                                    {groupedRecords[col]?.map((record: SObject) => (
                                        <div
                                            key={record.id}
                                            className="bg-white p-3 rounded border border-slate-200 shadow-sm hover:shadow-md cursor-pointer transition-shadow"
                                            onClick={() => navigate(ROUTES.OBJECT.DETAIL(schema.api_name, record.id as string))}
                                        >
                                            <div className="font-medium text-slate-800 mb-1 truncate">
                                                {getRecordDisplayName(record, schema)}
                                            </div>
                                            {/* Show extra fields? Maybe just one or two */}
                                            <div className="text-xs text-slate-500 flex justify-between items-center mt-2">
                                                <span>{(record.id as string).substring(0, 6)}</span>
                                            </div>
                                        </div>
                                    ))}
                                    {(!groupedRecords[col] || groupedRecords[col].length === 0) && (
                                        <div className="text-center py-8 text-slate-400 text-sm italic">
                                            Empty
                                        </div>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="absolute inset-0 flex items-center justify-center text-slate-400">
                        No object configured
                    </div>
                )}
            </div>
        </div>
    );
};
