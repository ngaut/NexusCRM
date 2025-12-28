import React, { useMemo, useState } from 'react';
import { useErrorToast } from './ui/Toast';
import { DndContext, DragOverlay, useDraggable, useDroppable, DragEndEvent, DragStartEvent } from '@dnd-kit/core';
import { ArrowLeft, Loader2 } from 'lucide-react';
import { ObjectMetadata, SObject } from '../types';

interface KanbanBoardProps {
    objectMetadata: ObjectMetadata;
    records: SObject[];
    groupByField?: string;
    onRecordClick?: (record: SObject) => void;
    onUpdateRecord?: (recordId: string, updates: Record<string, unknown>) => Promise<void>;
    loading?: boolean;
}

export const KanbanBoard: React.FC<KanbanBoardProps> = ({
    objectMetadata,
    records,
    groupByField: propGroupBy,
    onRecordClick,
    onUpdateRecord,
    loading
}) => {
    const errorToast = useErrorToast();
    const [activeId, setActiveId] = useState<string | null>(null);

    // Determine Group By Field
    const groupByField = propGroupBy || objectMetadata.kanban_group_by || 'status';

    // Get Columns
    const { columns, groupedRecords } = useMemo(() => {
        const fieldMeta = objectMetadata.fields?.find(f => f.api_name === groupByField);

        let cols: string[] = [];
        if (fieldMeta?.type === 'Picklist' && fieldMeta.options && fieldMeta.options.length > 0) {
            cols = fieldMeta.options;
        } else {
            // Fallback: derived from data
            const values = new Set(records.map(r => String(r[groupByField] || 'Unassigned')));
            cols = Array.from(values).sort();
        }

        const groups: Record<string, SObject[]> = {};
        cols.forEach(c => groups[c] = []);

        records.forEach(r => {
            const val = String(r[groupByField] || 'Unassigned');
            if (!groups[val]) {
                // Handle values outside defined columns (add dynamically if needed, or skip/bucket)
                if (!groups[val]) groups[val] = []; // Dynamically add if missing from options
                if (!cols.includes(val)) cols.push(val);
            }
            groups[val].push(r);
        });

        return { columns: cols, groupedRecords: groups };
    }, [objectMetadata, records, groupByField]);

    // Handlers
    const handleDragStart = (event: DragStartEvent) => {
        setActiveId(event.active.id as string);
    };

    const handleDragEnd = async (event: DragEndEvent) => {
        const { active, over } = event;
        setActiveId(null);

        if (!over) return;

        const recordId = active.id as string;
        const newStatus = over.id as string;

        // Find record
        const record = records.find(r => r.id === recordId);
        if (!record) return;

        const currentStatus = String(record[groupByField] || 'Unassigned');

        if (currentStatus !== newStatus) {
            // Optimistic update? Or loading state?
            // For now, call callback
            if (onUpdateRecord) {
                try {
                    await onUpdateRecord(recordId, { [groupByField]: newStatus });
                } catch {
                    errorToast("Failed to update status");
                }
            }
        }
    };

    if (loading) {
        return (
            <div className="flex h-96 items-center justify-center text-slate-400">
                <Loader2 className="animate-spin mr-2" /> Loading Board...
            </div>
        );
    }

    const activeRecord = activeId ? records.find(r => r.id === activeId) : null;

    return (
        <DndContext onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
            <div className="flex h-full gap-4 overflow-x-auto pb-4 p-1">
                {columns.map(col => (
                    <KanbanColumn
                        key={col}
                        id={col}
                        title={col}
                        records={groupedRecords[col] || []}
                        objectMetadata={objectMetadata}
                        onRecordClick={onRecordClick}
                    />
                ))}
            </div>
            <DragOverlay>
                {activeRecord ? (
                    <KanbanCard
                        record={activeRecord}
                        objectMetadata={objectMetadata}
                        isOverlay
                    />
                ) : null}
            </DragOverlay>
        </DndContext>
    );
};

interface KanbanColumnProps {
    id: string;
    title: string;
    records: SObject[];
    objectMetadata: ObjectMetadata;
    onRecordClick?: (record: SObject) => void;
}

const KanbanColumn: React.FC<KanbanColumnProps> = ({ id, title, records, objectMetadata, onRecordClick }) => {
    const { setNodeRef } = useDroppable({
        id: id,
    });

    return (
        <div ref={setNodeRef} className="w-72 flex-shrink-0 flex flex-col bg-slate-50/50 rounded-xl border border-slate-200/60 max-h-[calc(100vh-250px)]">
            {/* Header */}
            <div className="p-3 border-b border-slate-200/60 flex justify-between items-center sticky top-0 bg-slate-50/90 backdrop-blur rounded-t-xl z-10">
                <h3 className="font-semibold text-slate-700 truncate">{title}</h3>
                <span className="bg-slate-200 text-slate-600 text-xs px-2 py-0.5 rounded-full font-mono">
                    {records.length}
                </span>
            </div>

            {/* Cards Container */}
            <div className="p-2 flex-1 overflow-y-auto space-y-2 min-h-[100px]">
                {records.map(record => (
                    <DraggableKanbanCard
                        key={record.id as string}
                        record={record}
                        objectMetadata={objectMetadata}
                        onClick={() => onRecordClick && onRecordClick(record)}
                    />
                ))}
                {records.length === 0 && (
                    <div className="h-20 border-2 border-dashed border-slate-200 rounded-lg flex items-center justify-center text-slate-400 text-xs italic">
                        Drop items here
                    </div>
                )}
            </div>
        </div>
    );
};

interface KanbanCardProps {
    record: SObject;
    objectMetadata: ObjectMetadata;
    onClick?: () => void;
    isOverlay?: boolean;
}

const KanbanCard: React.FC<KanbanCardProps> = ({ record, objectMetadata, onClick, isOverlay }) => {
    const nameField = objectMetadata.fields?.find(f => f.is_name_field)?.api_name || 'name';

    // Find some secondary fields to show (e.g. priority, type, owner)
    // Priority 1: Priority
    // Priority 2: Type
    // Priority 3: CreatedDate
    const secondaryFields = (objectMetadata.fields || [])
        .filter(f => !f.is_system && f.api_name !== nameField && f.type !== 'LongTextArea')
        .slice(0, 3);

    return (
        <div
            onClick={onClick}
            className={`
                bg-white p-3 rounded-lg border shadow-sm 
                ${isOverlay ? 'shadow-xl rotate-2 scale-105 border-blue-400 cursor-grabbing' : 'border-slate-200 hover:shadow-md hover:border-blue-300 cursor-grab'}
                transition-all group relative select-none
            `}
        >
            <div className="font-medium text-slate-800 mb-2 truncate pr-6">
                {(record[nameField] as string) || 'Untitled'}
            </div>

            <div className="space-y-1">
                {secondaryFields.map(f => {
                    const val = record[f.api_name];
                    if (!val) return null;
                    return (
                        <div key={f.api_name} className="flex justify-between text-xs">
                            <span className="text-slate-400">{f.label}</span>
                            <span className="text-slate-600 font-medium truncate max-w-[120px]">{String(val)}</span>
                        </div>
                    )
                })}
            </div>

            <div className="mt-3 pt-2 border-t border-slate-50 flex justify-between items-center">
                <span className="text-[10px] text-slate-400 font-mono">{(record.id as string).substring(0, 6)}</span>
                {/* Avatar placeholder? */}
            </div>
        </div>
    );
};

const DraggableKanbanCard: React.FC<KanbanCardProps & { record: SObject }> = (props) => {
    const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
        id: props.record.id as string,
    });

    if (isDragging) {
        return (
            <div
                ref={setNodeRef}
                className="bg-slate-50 border-2 border-dashed border-slate-200 rounded-lg h-24 opacity-50"
            />
        );
    }

    return (
        <div ref={setNodeRef} {...listeners} {...attributes}>
            <KanbanCard {...props} />
        </div>
    );
}

