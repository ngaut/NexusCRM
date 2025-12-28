import React from 'react';
import { UIRegistry } from '../registries/UIRegistry';
import { getHighlightFields } from '../core/utils/recordUtils';
import { SObject, FieldMetadata, PageLayout, ObjectMetadata } from '../types';

interface HighlightsPanelProps {
    record: SObject;
    layout: PageLayout | null;
    fields: FieldMetadata[];
    loading?: boolean;
}

export const HighlightsPanel: React.FC<HighlightsPanelProps> = ({
    record,
    layout,
    fields,
    loading
}) => {
    if (loading) {
        return (
            <div className="bg-white border-b border-slate-200 p-4 animate-pulse">
                <div className="flex gap-8">
                    {[1, 2, 3, 4].map(i => (
                        <div key={i} className="space-y-2">
                            <div className="h-3 w-20 bg-slate-200 rounded"></div>
                            <div className="h-4 w-32 bg-slate-200 rounded"></div>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    // Use compact layout fields if available, otherwise default to first 5 fields
    const displayFields = layout?.compact_layout && layout.compact_layout.length > 0
        ? layout.compact_layout
        : getHighlightFields({ fields } as unknown as ObjectMetadata, 5).map(f => f.api_name);

    return (
        <div className="bg-white border-b border-slate-200 px-6 py-4">
            <div className="flex flex-wrap gap-x-12 gap-y-4">
                {displayFields.map(fieldApiName => {
                    const field = fields.find(f => f.api_name === fieldApiName);
                    if (!field) return null;

                    const Renderer = UIRegistry.getFieldRenderer(field.type);

                    return (
                        <div key={fieldApiName} className="flex flex-col min-w-[100px]">
                            <span className="text-xs text-slate-500 font-medium uppercase tracking-wide mb-1">
                                {field.label}
                            </span>
                            <span className="text-sm font-semibold text-slate-900 truncate">
                                <Renderer
                                    field={field}
                                    value={record[fieldApiName]}
                                    record={record}
                                    variant="detail"
                                />
                            </span>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};
