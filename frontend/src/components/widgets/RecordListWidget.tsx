
import React, { useEffect, useState } from 'react';
import { WidgetRendererProps } from '../../types';
import { MetadataRecordList } from '../MetadataRecordList';
import { metadataAPI } from '../../infrastructure/api/metadata';
import { ObjectMetadata } from '../../types';
import { Loader2 } from 'lucide-react';

export const RecordListWidget: React.FC<WidgetRendererProps> = ({
    title,
    config,
    isEditing,
    isVisible,
    onToggle
}) => {
    const [schema, setSchema] = useState<ObjectMetadata | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const objectApiName = config.query?.object_api_name;

    useEffect(() => {
        if (!objectApiName) {
            setLoading(false);
            return;
        }

        setLoading(true);
        metadataAPI.getSchema(objectApiName)
            .then(res => {
                setSchema(res.schema);
                setLoading(false);
            })
            .catch(() => {
                setError("Failed to load object metadata");
                setLoading(false);
            });
    }, [objectApiName]);

    return (
        <div className={`bg-white rounded-lg border shadow-sm flex flex-col relative h-full overflow-hidden ${isEditing ? 'border-dashed border-2 border-slate-300' : 'border-slate-200'}`}>
            {isEditing && (
                <div className="absolute top-2 right-2 z-10">
                    <button onClick={onToggle} className={`p-1.5 rounded-full ${isVisible ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-500'}`}>
                        {/* Eye icon would be passed or imported, simplifying for now */}
                        {isVisible ? "Hide" : "Show"}
                    </button>
                </div>
            )}

            <div className="p-4 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
                <h3 className="font-bold text-slate-800 truncate" title={title}>{title || schema?.plural_label || "Record List"}</h3>
            </div>

            <div className="flex-1 overflow-hidden p-0 relative">
                {loading ? (
                    <div className="absolute inset-0 flex items-center justify-center text-slate-400">
                        <Loader2 className="animate-spin mr-2" /> Loading...
                    </div>
                ) : error ? (
                    <div className="absolute inset-0 flex items-center justify-center text-red-500 p-4 text-center text-sm">
                        {error}
                    </div>
                ) : schema ? (
                    <div className="h-full overflow-auto p-2">
                        <MetadataRecordList
                            objectMetadata={schema}
                            filterExpr={config.query?.filter_expr}
                        />
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
