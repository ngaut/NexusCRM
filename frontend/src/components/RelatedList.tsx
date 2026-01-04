import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { dataAPI } from '../infrastructure/api/data';
import { COMMON_FIELDS } from '../core/constants';
import { useObjectMetadata } from '../core/hooks/useMetadata';
import { UIRegistry } from '../registries/UIRegistry';
import type { SObject, RelatedListConfig } from '../types';
import { Plus, Loader2, ChevronRight } from 'lucide-react';
import { ROUTES, buildRoute } from '../core/constants/Routes';

interface RelatedListProps {
    config: RelatedListConfig;
    parentRecordId: string;
    parentObjectApiName: string;
    refreshKey?: number;  // External trigger to force reload
}

export const RelatedList: React.FC<RelatedListProps> = ({
    config,
    parentRecordId,
    parentObjectApiName,
    refreshKey = 0
}) => {
    const [records, setRecords] = useState<SObject[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // Get metadata for the related object
    const { metadata } = useObjectMetadata(config.object_api_name);

    useEffect(() => {
        loadRelatedRecords();
    }, [parentRecordId, config.object_api_name, config.lookup_field, refreshKey]);

    const loadRelatedRecords = async () => {
        if (!parentRecordId) return;

        setLoading(true);
        setError(null);

        try {
            // Query records where the lookup field matches the parent record ID
            const relatedRecords = await dataAPI.query({
                objectApiName: config.object_api_name,
                filterExpr: `${config.lookup_field} == '${parentRecordId}'`,
                sortField: COMMON_FIELDS.CREATED_DATE,
                sortDirection: 'DESC',
                limit: 50
            });

            setRecords(relatedRecords);
        } catch (err: unknown) {
            setError(err instanceof Error ? err.message : 'Failed to load related records');
        } finally {
            setLoading(false);
        }
    };

    // Get fields to display (use config.fields or default to first few)
    const displayFields = config.fields && config.fields.length > 0
        ? config.fields
        : metadata?.fields
            .filter(f => !f.is_system && f.api_name !== COMMON_FIELDS.ID)
            .slice(0, 4)
            .map(f => f.api_name) || [];

    const getFieldMetadata = (fieldApiName: string) => {
        return metadata?.fields.find(f => f.api_name === fieldApiName);
    };

    if (loading) {
        return (
            <div className="bg-white rounded-lg border border-slate-200 p-6 shadow-sm">
                <h3 className="text-lg font-bold text-slate-800 mb-4">{config.label}</h3>
                <div className="flex items-center justify-center py-8 text-slate-500">
                    <Loader2 className="animate-spin mr-2" size={20} />
                    Loading...
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-white rounded-lg border border-slate-200 p-6 shadow-sm">
                <h3 className="text-lg font-bold text-slate-800 mb-4">{config.label}</h3>
                <div className="text-center py-8 text-red-500">{error}</div>
            </div>
        );
    }

    return (
        <div className="bg-white rounded-lg border border-slate-200 shadow-sm">
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200">
                <h3 className="text-lg font-bold text-slate-800 flex items-center gap-2">
                    {config.label}
                    <span className="text-sm font-normal text-slate-500">
                        ({records.length})
                    </span>
                </h3>
                <Link
                    to={buildRoute(ROUTES.OBJECT.NEW(config.object_api_name), { [config.lookup_field]: parentRecordId })}
                    className="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                >
                    <Plus size={16} />
                    New
                </Link>
            </div>

            {/* Records */}
            {records.length === 0 ? (
                <div className="px-6 py-8 text-center text-slate-500 text-sm">
                    No {config.label.toLowerCase()} to display
                </div>
            ) : (
                <div className="overflow-x-auto">
                    <table className="w-full">
                        <thead className="bg-slate-50 border-b border-slate-200">
                            <tr>
                                {displayFields.map(fieldApiName => {
                                    const field = getFieldMetadata(fieldApiName);
                                    return (
                                        <th
                                            key={fieldApiName}
                                            className="px-6 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider"
                                        >
                                            {field?.label || fieldApiName}
                                        </th>
                                    );
                                })}
                                <th className="px-6 py-3 w-12"></th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                            {records.map((record) => (
                                <tr
                                    key={record[COMMON_FIELDS.ID] as string}
                                    className="hover:bg-slate-50 transition-colors cursor-pointer"
                                    onClick={() => {
                                        window.location.href = ROUTES.OBJECT.DETAIL(config.object_api_name, record[COMMON_FIELDS.ID] as string);
                                    }}
                                >
                                    {displayFields.map(fieldApiName => {
                                        const field = getFieldMetadata(fieldApiName);
                                        const value = record[fieldApiName];

                                        if (!field) {
                                            return (
                                                <td key={fieldApiName} className="px-6 py-4 text-sm text-slate-900">
                                                    {String(value ?? '')}
                                                </td>
                                            );
                                        }

                                        const Renderer = UIRegistry.getFieldRenderer(field.type);

                                        return (
                                            <td key={fieldApiName} className="px-6 py-4 text-sm">
                                                <Renderer
                                                    field={field}
                                                    value={value}
                                                    record={record}
                                                    variant="table"
                                                    onNavigate={(obj, id) => {
                                                        window.location.href = ROUTES.OBJECT.DETAIL(obj, id);
                                                    }}
                                                />
                                            </td>
                                        );
                                    })}
                                    <td className="px-6 py-4 text-right">
                                        <Link
                                            to={ROUTES.OBJECT.DETAIL(config.object_api_name, record[COMMON_FIELDS.ID] as string)}
                                            className="text-blue-600 hover:text-blue-800"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            <ChevronRight size={16} />
                                        </Link>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Footer - View All Link */}
            {records.length > 0 && (
                <div className="px-6 py-3 border-t border-slate-200 bg-slate-50">
                    <Link
                        to={buildRoute(ROUTES.OBJECT.LIST(config.object_api_name), { [config.lookup_field]: parentRecordId })}
                        className="text-sm text-blue-600 hover:text-blue-800 font-medium"
                    >
                        View All ({records.length})
                    </Link>
                </div>
            )}
        </div>
    );
};
