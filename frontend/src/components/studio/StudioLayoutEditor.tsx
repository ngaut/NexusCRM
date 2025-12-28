import React, { useState, useEffect } from 'react';
import { GripVertical, Plus, Trash2, Save, ArrowLeft, Layout as LayoutIcon } from 'lucide-react';
import { UIRegistry } from '../../registries/UIRegistry';
import { metadataAPI } from '../../infrastructure/api/metadata';
import { useObjectMetadata, useLayout } from '../../core/hooks/useMetadata';
import { DraggableField } from '../admin/DraggableField';
import { DropZone } from '../admin/DropZone';
import { ConfirmationModal } from '../modals/ConfirmationModal';
import { useErrorToast } from '../ui/Toast';
import type { PageLayout, FieldMetadata, PageSection } from '../../types';

interface StudioLayoutEditorProps {
    objectApiName: string;
}

interface DragData {
    field: FieldMetadata;
    sourceSectionId?: string;
}

export const StudioLayoutEditor: React.FC<StudioLayoutEditorProps> = ({ objectApiName }) => {
    const errorToast = useErrorToast();
    const { metadata } = useObjectMetadata(objectApiName);
    const { layout: existingLayout } = useLayout(objectApiName);

    const [layout, setLayout] = useState<PageLayout | null>(null);
    const [draggedField, setDraggedField] = useState<FieldMetadata | null>(null);
    const [dragOverSection, setDragOverSection] = useState<string | null>(null);
    const [isSaving, setIsSaving] = useState(false);
    const [lastSaved, setLastSaved] = useState<Date | null>(null);

    // Confirmation Modal State
    const [sectionToRemove, setSectionToRemove] = useState<string | null>(null);
    const [showRemoveSectionModal, setShowRemoveSectionModal] = useState(false);

    // Initialize layout
    useEffect(() => {
        if (existingLayout) {
            setLayout(existingLayout);
        } else if (metadata) {
            // Default layout
            setLayout({
                id: crypto.randomUUID(),
                object_api_name: metadata.api_name,
                layout_name: 'Default Layout',
                type: 'Detail',
                compact_layout: [],
                sections: [
                    {
                        id: crypto.randomUUID(),
                        label: 'Information',
                        columns: 2,
                        fields: []
                    }
                ],
                related_lists: [],
                header_actions: [],
                quick_actions: []
            });
        }
    }, [existingLayout, metadata]);

    if (!metadata || !layout) return <div className="p-8 text-center animate-pulse">Loading layout...</div>;

    const usedFieldNames = new Set(layout.sections.flatMap(s => s.fields));
    const availableFields = metadata.fields.filter(f => !usedFieldNames.has(f.api_name));

    const handleDragStart = (e: React.DragEvent, field: FieldMetadata, sourceSectionId?: string) => {
        setDraggedField(field);
        e.dataTransfer.effectAllowed = 'copyMove';
        e.dataTransfer.setData('application/json', JSON.stringify({ field, sourceSectionId }));
    };

    const handleDrop = (e: React.DragEvent, targetSectionId: string) => {
        e.preventDefault();
        setDragOverSection(null);
        if (!draggedField) return;

        const data = JSON.parse(e.dataTransfer.getData('application/json')) as DragData;
        const sourceSectionId = data.sourceSectionId;

        setLayout(prev => {
            if (!prev) return null;
            if (sourceSectionId === targetSectionId) return prev;

            let newSections = [...prev.sections];

            // Remove from source
            if (sourceSectionId) {
                newSections = newSections.map(section => {
                    if (section.id === sourceSectionId) {
                        return {
                            ...section,
                            fields: section.fields.filter(f => f !== draggedField.api_name)
                        };
                    }
                    return section;
                });
            }

            // Add to target
            newSections = newSections.map(section => {
                if (section.id === targetSectionId) {
                    if (section.fields.includes(draggedField.api_name)) return section;
                    return {
                        ...section,
                        fields: [...section.fields, draggedField.api_name]
                    };
                }
                return section;
            });

            return { ...prev, sections: newSections };
        });
        setDraggedField(null);
    };

    const handleRemoveField = (sectionId: string, fieldApiName: string) => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.map(section => {
                    if (section.id === sectionId) {
                        return {
                            ...section,
                            fields: section.fields.filter(f => f !== fieldApiName)
                        };
                    }
                    return section;
                })
            };
        });
    };

    const handleMoveField = (sectionId: string, fieldApiName: string, direction: 'up' | 'down') => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.map(section => {
                    if (section.id === sectionId) {
                        const fields = [...section.fields];
                        const index = fields.indexOf(fieldApiName);
                        if (index === -1) return section;

                        if (direction === 'up' && index > 0) {
                            [fields[index], fields[index - 1]] = [fields[index - 1], fields[index]];
                        } else if (direction === 'down' && index < fields.length - 1) {
                            [fields[index], fields[index + 1]] = [fields[index + 1], fields[index]];
                        }
                        return { ...section, fields };
                    }
                    return section;
                })
            };
        });
    };

    const handleAddSection = () => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: [...prev.sections, {
                    id: crypto.randomUUID(),
                    label: 'New Section',
                    columns: 2,
                    fields: []
                }]
            };
        });
    };

    const handleUpdateSection = (sectionId: string, updates: Partial<PageSection>) => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.map(s => s.id === sectionId ? { ...s, ...updates } : s)
            };
        });
    };

    const handleRemoveSection = (sectionId: string) => {
        setSectionToRemove(sectionId);
        setShowRemoveSectionModal(true);
    };

    const confirmRemoveSection = () => {
        if (!sectionToRemove) return;
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.filter(s => s.id !== sectionToRemove)
            };
        });
        setShowRemoveSectionModal(false);
        setSectionToRemove(null);
    };

    const handleSave = async () => {
        if (!layout) return;
        setIsSaving(true);
        try {
            await metadataAPI.saveLayout(layout);
            setLastSaved(new Date());
        } catch {
            errorToast('Failed to save layout');
        } finally {
            setIsSaving(false);
        }
    };

    return (
        <div className="flex flex-col h-[calc(100vh-250px)]">
            {/* Toolbar */}
            <div className="flex justify-between items-center mb-4 px-1">
                <div className="text-sm text-slate-500">
                    {lastSaved ? `Last saved at ${lastSaved.toLocaleTimeString()}` : 'Unsaved changes'}
                </div>
                <button
                    onClick={handleSave}
                    disabled={isSaving}
                    className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm font-medium disabled:opacity-50"
                >
                    <Save size={14} />
                    {isSaving ? 'Saving...' : 'Save Layout'}
                </button>
            </div>

            <div className="flex-1 flex overflow-hidden border rounded-lg bg-slate-50">
                {/* Palette */}
                <div className="w-64 bg-white border-r border-slate-200 flex flex-col overflow-hidden">
                    <div className="p-3 border-b border-slate-200 bg-slate-50">
                        <h3 className="font-semibold text-slate-700 text-sm flex items-center gap-2">
                            <GripVertical size={14} /> Fields Palette
                        </h3>
                    </div>
                    <div className="flex-1 overflow-y-auto p-3 space-y-2">
                        {availableFields.map(field => (
                            <DraggableField
                                key={field.api_name}
                                field={field}
                                onDragStart={(e) => handleDragStart(e, field)}
                            />
                        ))}
                        {availableFields.length === 0 && (
                            <div className="text-center py-4 text-xs text-slate-400 italic">
                                All fields used
                            </div>
                        )}
                    </div>
                </div>

                {/* Canvas */}
                <div className="flex-1 overflow-y-auto p-6">
                    <div className="max-w-3xl mx-auto space-y-6">
                        {layout.sections.map((section) => (
                            <div key={section.id} className="bg-white rounded-lg border border-slate-200 shadow-sm overflow-hidden group">
                                <div className="bg-slate-50 px-3 py-2 border-b border-slate-200 flex items-center justify-between">
                                    <input
                                        type="text"
                                        value={section.label}
                                        onChange={(e) => handleUpdateSection(section.id, { label: e.target.value })}
                                        className="bg-transparent border-none focus:ring-0 font-semibold text-slate-700 text-sm p-0 w-full"
                                    />
                                    <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <select
                                            onChange={(e) => {
                                                const val = Number(e.target.value);
                                                if (val === 1 || val === 2) {
                                                    handleUpdateSection(section.id, { columns: val });
                                                }
                                            }}
                                            className="text-xs border-slate-300 rounded bg-white py-0.5 pl-1 pr-5"
                                        >
                                            <option value={1}>1 Col</option>
                                            <option value={2}>2 Col</option>
                                        </select>
                                        <button onClick={() => handleRemoveSection(section.id)} className="text-slate-400 hover:text-red-600">
                                            <Trash2 size={14} />
                                        </button>
                                    </div>
                                </div>

                                <DropZone
                                    onDrop={(e) => handleDrop(e, section.id)}
                                    onDragOver={(e) => { e.preventDefault(); setDragOverSection(section.id); }}
                                    onDragLeave={() => setDragOverSection(null)}
                                    isOver={dragOverSection === section.id}
                                    className="p-3 min-h-[50px]"
                                >
                                    <div className={`grid gap-3 ${section.columns === 2 ? 'grid-cols-2' : 'grid-cols-1'}`}>
                                        {section.fields.map((fieldApiName, index) => {
                                            const field = metadata.fields.find(f => f.api_name === fieldApiName);
                                            if (!field) return null;
                                            return (
                                                <div
                                                    key={fieldApiName}
                                                    draggable
                                                    onDragStart={(e) => handleDragStart(e, field, section.id)}
                                                    className="group/field relative p-2 bg-white border border-slate-200 rounded hover:border-blue-400 cursor-move"
                                                >
                                                    <div className="flex justify-between items-start">
                                                        <div>
                                                            <div className="text-xs text-slate-500 mb-1">{field.label}</div>
                                                            <div className="h-4 bg-slate-100 rounded w-full"></div>
                                                        </div>
                                                        <div className="absolute top-1 right-1 flex gap-1 opacity-0 group-hover/field:opacity-100 bg-white/80 rounded px-1">
                                                            <button onClick={() => handleMoveField(section.id, fieldApiName, 'up')} disabled={index === 0} className="hover:text-blue-600 disabled:opacity-30"><ArrowLeft className="rotate-90" size={12} /></button>
                                                            <button onClick={() => handleMoveField(section.id, fieldApiName, 'down')} disabled={index === section.fields.length - 1} className="hover:text-blue-600 disabled:opacity-30"><ArrowLeft className="-rotate-90" size={12} /></button>
                                                            <button onClick={() => handleRemoveField(section.id, fieldApiName)} className="hover:text-red-500"><Trash2 size={12} /></button>
                                                        </div>
                                                    </div>
                                                </div>
                                            );
                                        })}
                                        {section.fields.length === 0 && (
                                            <div className="col-span-full border border-dashed border-slate-200 rounded p-4 text-center text-xs text-slate-400">
                                                Drop fields here
                                            </div>
                                        )}
                                    </div>
                                </DropZone>
                            </div>
                        ))}
                        <button
                            onClick={handleAddSection}
                            className="w-full py-3 border-2 border-dashed border-slate-300 rounded-lg text-slate-500 hover:border-blue-400 hover:text-blue-600 hover:bg-blue-50 transition-all flex items-center justify-center gap-2 text-sm font-medium"
                        >
                            <Plus size={16} /> Add Section
                        </button>
                    </div>
                </div>
            </div>


            <ConfirmationModal
                isOpen={showRemoveSectionModal}
                onClose={() => {
                    setShowRemoveSectionModal(false);
                    setSectionToRemove(null);
                }}
                onConfirm={confirmRemoveSection}
                title="Remove Section"
                message="Are you sure you want to remove this section? All fields in it will be moved back to the palette."
                confirmLabel="Remove"
                cancelLabel="Cancel"
                variant="warning"
            />
        </div >
    );
};
