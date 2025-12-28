import React, { useEffect, useState } from 'react';
import { WidgetConfig, ObjectMetadata } from '../../../types';
import { Settings, Type, Database, Palette as PaletteIcon } from 'lucide-react';

interface DashboardInspectorProps {
    selectedWidget: WidgetConfig | null;
    schemas: ObjectMetadata[];
    onUpdate: (updates: Partial<WidgetConfig>) => void;
}

export const DashboardInspector: React.FC<DashboardInspectorProps> = ({
    selectedWidget,
    schemas,
    onUpdate
}) => {
    if (!selectedWidget) {
        return (
            <div className="w-80 bg-white border-l border-slate-200 p-8 text-center flex flex-col items-center justify-center h-full text-slate-500">
                <Settings className="w-12 h-12 mb-4 text-slate-300" />
                <p className="font-medium">No Widget Selected</p>
                <p className="text-sm mt-2">Select a widget on the canvas to edit its properties.</p>
            </div>
        );
    }

    const { type, query } = selectedWidget;
    const isChart = type.startsWith('chart') || type === 'metric';
    const needsDataSource = isChart || type === 'record-list' || type === 'kanban';

    // Helper to get fields for selected object
    const getObjectFields = () => {
        if (!query?.object_api_name) {

            return [];
        }
        const schema = schemas.find(s => s.api_name === query.object_api_name);
        if (!schema) {
            console.warn("DashboardInspector: Object schema not found for", query.object_api_name);
            return [];
        }
        if (!Array.isArray(schema.fields)) {
            console.warn("DashboardInspector: Schema fields is not an array", schema);
            return [];
        }
        return schema.fields;
    };

    return (
        <div className="w-80 bg-white border-l border-slate-200 flex flex-col h-full shadow-lg z-10">
            <div className="p-4 border-b border-slate-200 bg-slate-50">
                <h3 className="font-semibold text-slate-900">Properties</h3>
                <p className="text-xs text-slate-500 font-mono mt-1">{selectedWidget.id}</p>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-6">

                {/* General Settings */}
                <div className="space-y-4">
                    <h4 className="text-xs font-bold text-slate-400 uppercase tracking-wider flex items-center gap-2">
                        <Type size={12} />
                        General
                    </h4>

                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">Title</label>
                        <input
                            type="text"
                            value={selectedWidget.title}
                            onChange={(e) => onUpdate({ title: e.target.value })}
                            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>

                    {type === 'text' && (
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">Content (Markdown)</label>
                            <textarea
                                value={selectedWidget.content || ''}
                                onChange={(e) => onUpdate({ content: e.target.value })}
                                rows={6}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm font-mono focus:ring-2 focus:ring-blue-500"
                                placeholder="# Heading..."
                            />
                        </div>
                    )}

                    {type === 'image' && (
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">Image URL</label>
                            <input
                                type="text"
                                value={selectedWidget.imageUrl || ''}
                                onChange={(e) => onUpdate({ imageUrl: e.target.value })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500"
                                placeholder="https://..."
                            />
                        </div>
                    )}
                </div>

                {/* Data Configuration */}
                {needsDataSource && (
                    <div className="space-y-4 pt-4 border-t border-slate-100">
                        <h4 className="text-xs font-bold text-slate-400 uppercase tracking-wider flex items-center gap-2">
                            <Database size={12} />
                            Data Source
                        </h4>

                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">Object</label>
                            <select
                                value={query?.object_api_name || ''}
                                onChange={(e) => {
                                    const schema = schemas.find(s => s.api_name === e.target.value);
                                    const updates: Partial<WidgetConfig> = { query: { ...query, object_api_name: e.target.value, field: undefined } };

                                    // Auto-title if default
                                    if (selectedWidget.title === 'New Widget' && schema) {
                                        updates.title = `Total ${schema.plural_label}`;
                                    }
                                    onUpdate(updates);
                                }}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white"
                            >
                                <option value="">-- Select Object --</option>
                                {schemas.map(s => (
                                    <option key={s.api_name} value={s.api_name}>{s.plural_label}</option>
                                ))}
                            </select>
                        </div>

                        {/* Chart/Metric Operations */}
                        {(isChart) && (
                            <>
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 mb-1">Operation</label>
                                    <select
                                        value={query?.operation || 'count'}
                                        onChange={(e) => onUpdate({
                                            query: { ...query, operation: e.target.value as 'count' | 'sum' | 'avg' | 'group_by' }
                                        })}
                                        className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white"
                                    >
                                        <option value="count">Count Records</option>
                                        <option value="sum">Sum</option>
                                        <option value="avg">Average</option>
                                        {type.startsWith('chart') && <option value="group_by">Group By</option>}
                                    </select>
                                </div>

                                {(query?.operation === 'sum' || query?.operation === 'avg' || query?.operation === 'group_by') && (
                                    <div>
                                        <label className="block text-sm font-medium text-slate-700 mb-1">
                                            {query.operation === 'group_by' ? 'Aggregation Field (Value)' : 'Field'}
                                        </label>
                                        <select
                                            value={query.field || ''}
                                            onChange={(e) => {
                                                const fieldL = getObjectFields().find(f => f.api_name === e.target.value)?.label;
                                                const updates: Partial<WidgetConfig> = { query: { ...query, field: e.target.value } };

                                                if (selectedWidget.title === 'New Widget' && fieldL) {
                                                    const opMap: Record<string, string> = { sum: 'Total', avg: 'Avg' };
                                                    const prefix = opMap[query.operation] || '';
                                                    updates.title = prefix ? `${prefix} ${fieldL}` : fieldL;
                                                }
                                                onUpdate(updates);
                                            }}
                                            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white"
                                        >
                                            <option value="">-- Select Field --</option>
                                            {(getObjectFields() || [])
                                                .filter(f => ['Currency', 'Number', 'Percent', 'Decimal', 'Int'].includes(f.type))
                                                .map(f => (
                                                    <option key={f.api_name} value={f.api_name}>{f.label}</option>
                                                ))}
                                        </select>
                                    </div>
                                )}

                                {query?.operation === 'group_by' && (
                                    <div>
                                        <label className="block text-sm font-medium text-slate-700 mb-1">Group By (Category)</label>
                                        <select
                                            value={query.group_by || ''}
                                            onChange={(e) => onUpdate({
                                                query: { ...query, group_by: e.target.value }
                                            })}
                                            className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white"
                                        >
                                            <option value="">-- Select Field --</option>
                                            {(getObjectFields() || [])
                                                .filter(f => ['Picklist', 'Lookup', 'Text', 'Boolean'].includes(f.type))
                                                .map(f => (
                                                    <option key={f.api_name} value={f.api_name}>{f.label}</option>
                                                ))}
                                        </select>
                                    </div>
                                )}
                            </>
                        )}

                        {/* Kanban Settings */}
                        {type === 'kanban' && (
                            <div>
                                <label className="block text-sm font-medium text-slate-700 mb-1">Group By (Column)</label>
                                <select
                                    value={query?.group_by || ''}
                                    onChange={(e) => onUpdate({
                                        query: { ...query, group_by: e.target.value }
                                    })}
                                    className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white"
                                >
                                    <option value="">-- Default (Status) --</option>
                                    {(getObjectFields() || [])
                                        .filter(f => ['Picklist', 'Status'].includes(f.type) || f.api_name === 'status' || f.api_name === 'stage')
                                        .map(f => (
                                            <option key={f.api_name} value={f.api_name}>{f.label}</option>
                                        ))}
                                </select>
                            </div>
                        )}
                    </div>
                )}

                {/* Appearance */}
                <div className="space-y-4 pt-4 border-t border-slate-100">
                    <h4 className="text-xs font-bold text-slate-400 uppercase tracking-wider flex items-center gap-2">
                        <PaletteIcon size={12} />
                        Appearance
                    </h4>

                    {type === 'metric' && (
                        <div>
                            <label className="block text-sm font-medium text-slate-700 mb-1">Color</label>
                            <select
                                value={selectedWidget.color || 'blue'}
                                onChange={(e) => onUpdate({ color: e.target.value })}
                                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm bg-white"
                            >
                                <option value="blue">Blue</option>
                                <option value="emerald">Emerald</option>
                                <option value="purple">Purple</option>
                                <option value="orange">Orange</option>
                                <option value="rose">Rose</option>
                            </select>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};
