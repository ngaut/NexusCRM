import React from 'react';
import { GripVertical, Edit2, Eye, Trash2, Plus } from 'lucide-react';
import { DropZone } from '../DropZone';
import { UIRegistry } from '../../../registries/UIRegistry';
import type { PageLayout, ObjectMetadata, PageSection, FieldMetadata } from '../../../types';

interface LayoutCanvasProps {
    layout: PageLayout;
    metadata: ObjectMetadata;
    dragOverSection: string | null;
    onUpdateSection: (sectionId: string, updates: Partial<PageSection>) => void;
    onRemoveSection: (sectionId: string) => void;
    onAddSection: () => void;
    onEditSection: (sectionId: string, currentLabel: string) => void;
    onOpenVisibilityEditor: (e: React.MouseEvent, section: PageSection) => void;
    onDrop: (e: React.DragEvent, sectionId: string) => void;
    onDragOver: (sectionId: string) => void;
    onDragLeave: (e: React.DragEvent) => void;
    onDragStart: (e: React.DragEvent, field: FieldMetadata, sectionId: string) => void;
    onMoveField: (sectionId: string, fieldApiName: string, direction: 'up' | 'down') => void;
    onRemoveField: (sectionId: string, fieldApiName: string) => void;
}

export const LayoutCanvas: React.FC<LayoutCanvasProps> = ({
    layout,
    metadata,
    dragOverSection,
    onUpdateSection,
    onRemoveSection,
    onAddSection,
    onEditSection,
    onOpenVisibilityEditor,
    onDrop,
    onDragOver,
    onDragLeave,
    onDragStart,
    onMoveField,
    onRemoveField
}) => {
    return (
        <div className="flex-1 overflow-y-auto p-8">
            <div className="max-w-5xl mx-auto space-y-6">
                {layout.sections.map((section) => (
                    <div key={section.id} className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden group">
                        {/* Section Header */}
                        <div className="bg-slate-50 px-4 py-3 border-b border-slate-200 flex items-center justify-between">
                            <div className="flex items-center gap-3 flex-1">
                                <GripVertical className="text-slate-300 cursor-move" size={16} />
                                <input
                                    type="text"
                                    value={section.label}
                                    onChange={(e) => onUpdateSection(section.id, { label: e.target.value })}
                                    className="bg-transparent border-none focus:ring-0 font-semibold text-slate-700 p-0 w-full"
                                />
                            </div>
                            <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                <select
                                    value={section.columns}
                                    onChange={(e) => onUpdateSection(section.id, { columns: Number(e.target.value) as 1 | 2 })}
                                    className="text-xs border-slate-300 rounded bg-white py-1 pl-2 pr-6"
                                >
                                    <option value={1}>1 Column</option>
                                    <option value={2}>2 Columns</option>
                                </select>
                                <button
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        onEditSection(section.id, section.label);
                                    }}
                                    className="text-slate-400 hover:text-indigo-600 p-1"
                                    title="Edit Section Name"
                                >
                                    <Edit2 size={14} />
                                </button>
                                <button
                                    onClick={(e) => onOpenVisibilityEditor(e, section)}
                                    className={`p-1 ${section.visibility_condition ? 'text-indigo-600' : 'text-slate-400 hover:text-indigo-600'}`}
                                    title="Set Visibility Rules"
                                >
                                    <Eye size={14} />
                                </button>
                                <button
                                    onClick={() => onRemoveSection(section.id)}
                                    className="p-1 text-slate-400 hover:text-red-600 rounded hover:bg-red-50"
                                    title="Remove Section"
                                >
                                    <Trash2 size={16} />
                                </button>
                            </div>
                        </div>

                        {/* Section Content (Drop Zone) */}
                        <DropZone
                            onDrop={(e) => onDrop(e, section.id)}
                            onDragOver={(e) => {
                                e.preventDefault();
                                onDragOver(section.id);
                            }}
                            onDragLeave={onDragLeave}
                            isOver={dragOverSection === section.id}
                            className="p-4 min-h-[100px]"
                        >
                            <div className={`grid gap-4 ${section.columns === 2 ? 'grid-cols-2' : 'grid-cols-1'}`}>
                                {section.fields.map((fieldApiName, index) => {
                                    const field = metadata.fields.find(f => f.api_name === fieldApiName);
                                    if (!field) return null;
                                    return (
                                        <div
                                            key={fieldApiName}
                                            draggable
                                            onDragStart={(e) => onDragStart(e, field, section.id)}
                                            className="group/field relative p-3 bg-white border border-slate-200 rounded hover:border-blue-400 hover:shadow-sm transition-all cursor-move"
                                        >
                                            <div className="flex justify-between items-start">
                                                <div>
                                                    <div className="text-xs text-slate-500 mb-1">{field.label}</div>
                                                    <div className="pointer-events-none">
                                                        {(() => {
                                                            const Input = UIRegistry.getFieldInput(field.type);
                                                            return (
                                                                <Input
                                                                    field={field}
                                                                    value=""
                                                                    onChange={() => { }}
                                                                    disabled
                                                                    placeholder="Preview"
                                                                />
                                                            );
                                                        })()}
                                                    </div>
                                                </div>
                                                <div className="flex items-center gap-1 opacity-0 group-hover/field:opacity-100 transition-opacity">
                                                    <button
                                                        onClick={() => onMoveField(section.id, fieldApiName, 'up')}
                                                        disabled={index === 0}
                                                        className="p-1 text-slate-400 hover:text-blue-600 disabled:opacity-30"
                                                        title="Move Up"
                                                    >
                                                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" /></svg>
                                                    </button>
                                                    <button
                                                        onClick={() => onMoveField(section.id, fieldApiName, 'down')}
                                                        disabled={index === section.fields.length - 1}
                                                        className="p-1 text-slate-400 hover:text-blue-600 disabled:opacity-30"
                                                        title="Move Down"
                                                    >
                                                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" /></svg>
                                                    </button>
                                                    <button
                                                        onClick={() => onRemoveField(section.id, fieldApiName)}
                                                        className="p-1 text-slate-400 hover:text-red-500"
                                                        title="Remove Field"
                                                    >
                                                        <Trash2 size={14} />
                                                    </button>
                                                </div>
                                            </div>
                                        </div>
                                    );
                                })}
                                {section.fields.length === 0 && (
                                    <div className="col-span-full border-2 border-dashed border-slate-200 rounded-lg p-8 text-center text-slate-400 text-sm">
                                        Drop fields here
                                    </div>
                                )}
                            </div>
                        </DropZone>
                    </div>
                ))}

                <button
                    onClick={onAddSection}
                    className="w-full py-4 border-2 border-dashed border-slate-300 rounded-lg text-slate-500 hover:border-blue-400 hover:text-blue-600 hover:bg-blue-50 transition-all flex items-center justify-center gap-2 font-medium"
                >
                    <Plus size={20} />
                    Add New Section
                </button>
            </div>
        </div>
    );
};
