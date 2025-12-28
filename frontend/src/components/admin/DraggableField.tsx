import React from 'react';
import { GripVertical } from 'lucide-react';
import type { FieldMetadata } from '../../types';

interface DraggableFieldProps {
    field: FieldMetadata;
    onDragStart: (e: React.DragEvent, field: FieldMetadata) => void;
    isUsed?: boolean;
}

export const DraggableField: React.FC<DraggableFieldProps> = ({ field, onDragStart, isUsed }) => {
    return (
        <div
            draggable={!isUsed}
            onDragStart={(e) => onDragStart(e, field)}
            className={`
                flex items-center gap-2 p-2 rounded border text-sm cursor-move transition-colors
                ${isUsed
                    ? 'bg-slate-50 border-slate-200 text-slate-400 cursor-not-allowed'
                    : 'bg-white border-slate-200 text-slate-700 hover:border-blue-300 hover:shadow-sm'}
            `}
        >
            <GripVertical size={14} className={isUsed ? 'text-slate-300' : 'text-slate-400'} />
            <div className="flex-1 truncate">
                <span className="font-medium">{field.label}</span>
                <span className="ml-2 text-xs text-slate-400 font-mono">{field.type}</span>
            </div>
        </div>
    );
};
