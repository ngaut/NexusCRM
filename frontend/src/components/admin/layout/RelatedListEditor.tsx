import React, { useState, useEffect } from 'react';
import { Trash2 } from 'lucide-react';
import { PageLayout, ObjectMetadata } from '../../../types';
import { metadataAPI } from '../../../infrastructure/api/metadata';

interface DiscoveredRelationship {
    objectApiName: string;
    label: string;
    fieldApiName: string;
    fieldLabel: string;
}

interface RelatedListEditorProps {
    layout: PageLayout;
    metadata: ObjectMetadata;
    onUpdateLayout: (updater: (prev: PageLayout | null) => PageLayout | null) => void;
    onRemoveRelatedList: (id: string) => void;
}

export const RelatedListEditor: React.FC<RelatedListEditorProps> = ({
    layout,
    metadata,
    onUpdateLayout,
    onRemoveRelatedList
}) => {
    const [allSchemas, setAllSchemas] = useState<ObjectMetadata[]>([]);
    const [availableRelatedLists, setAvailableRelatedLists] = useState<DiscoveredRelationship[]>([]);

    // Fetch all schemas to discover relationships
    useEffect(() => {
        if (allSchemas.length === 0) {
            metadataAPI.getSchemas().then(response => {
                setAllSchemas(response.schemas);
            }).catch(() => { });
        }
    }, [allSchemas.length]);

    // Derive available related lists from schemas
    useEffect(() => {
        if (!metadata || allSchemas.length === 0) return;

        const currentObjectApiName = metadata.api_name;
        const relationships = allSchemas.flatMap(schema => {
            return schema.fields
                .filter(field => field.type === 'Lookup' && field.reference_to?.includes(currentObjectApiName))
                .map(field => ({
                    objectApiName: schema.api_name,
                    label: schema.plural_label,
                    fieldApiName: field.api_name,
                    fieldLabel: field.label
                }));
        });
        setAvailableRelatedLists(relationships);
    }, [metadata, allSchemas]);

    const handleAddRelatedList = (rel: DiscoveredRelationship) => {
        onUpdateLayout(prev => {
            if (!prev) return null;
            if (prev.related_lists?.some(r => r.object_api_name === rel.objectApiName && r.lookup_field === rel.fieldApiName)) {
                return prev;
            }
            const newList = {
                id: crypto.randomUUID(),
                label: rel.label,
                object_api_name: rel.objectApiName,
                lookup_field: rel.fieldApiName,
                fields: ['name', 'created_date']
            };
            return {
                ...prev,
                related_lists: [...(prev.related_lists || []), newList]
            };
        });
    };

    const handleMoveRelatedList = (index: number, direction: 'up' | 'down') => {
        onUpdateLayout(prev => {
            if (!prev || !prev.related_lists) return prev;
            const newLists = [...prev.related_lists];
            if (direction === 'up' && index > 0) {
                [newLists[index], newLists[index - 1]] = [newLists[index - 1], newLists[index]];
            } else if (direction === 'down' && index < newLists.length - 1) {
                [newLists[index], newLists[index + 1]] = [newLists[index + 1], newLists[index]];
            }
            return { ...prev, related_lists: newLists };
        });
    };

    return (
        <div className="bg-white rounded-lg border border-slate-200 p-6">
            <h3 className="font-semibold text-slate-700 mb-4">Related Lists</h3>
            <p className="text-slate-500 mb-6">Manage the related lists that appear on the record detail page.</p>

            <div className="space-y-4">
                {(layout.related_lists || []).map((list, index) => (
                    <div key={list.id} className="flex justify-between items-center p-3 border rounded-lg bg-slate-50">
                        <span className="font-medium">{list.label}</span>
                        <div className="flex gap-2">
                            <button onClick={() => handleMoveRelatedList(index, 'up')} disabled={index === 0} className="text-slate-400 hover:text-blue-600">Up</button>
                            <button onClick={() => handleMoveRelatedList(index, 'down')} disabled={index === (layout.related_lists?.length || 0) - 1} className="text-slate-400 hover:text-blue-600">Down</button>
                            <button onClick={() => onRemoveRelatedList(list.id)} className="text-red-500 hover:text-red-700">Remove</button>
                        </div>
                    </div>
                ))}
            </div>

            <div className="mt-8">
                <h4 className="font-semibold text-slate-700 mb-2">Available Relationships</h4>
                <div className="grid grid-cols-2 gap-4">
                    {availableRelatedLists.map((rel, idx) => (
                        <button
                            key={idx}
                            onClick={() => handleAddRelatedList(rel)}
                            className="text-left p-3 border border-dashed rounded hover:bg-blue-50 hover:border-blue-300 transition-colors"
                        >
                            <div className="font-medium text-slate-700">{rel.label}</div>
                            <div className="text-xs text-slate-500">via {rel.fieldLabel}</div>
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
};
