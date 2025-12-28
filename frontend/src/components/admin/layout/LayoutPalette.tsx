import React from 'react';
import { GripVertical } from 'lucide-react';
import { DraggableField } from '../DraggableField';
import type { FieldMetadata, ObjectMetadata } from '../../../types';

interface LayoutPaletteProps {
    availableFields: FieldMetadata[];
    metadata: ObjectMetadata;
    usedFieldNames: Set<string>;
    onDragStart: (e: React.DragEvent, field: FieldMetadata) => void;
}

export const LayoutPalette: React.FC<LayoutPaletteProps> = ({
    availableFields,
    metadata,
    usedFieldNames,
    onDragStart
}) => {
    return (
        <div className="w-80 bg-white border-r border-slate-200 flex flex-col shadow-inner z-0">
            <div className="p-4 border-b border-slate-200 bg-slate-50">
                <h2 className="font-semibold text-slate-700 flex items-center gap-2">
                    <GripVertical size={16} />
                    Fields Palette
                </h2>
                <p className="text-xs text-slate-500 mt-1">Drag fields onto the canvas.</p>
            </div>
            <div className="flex-1 overflow-y-auto p-4 space-y-2">
                <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">Available Fields</div>
                {availableFields.map(field => (
                    <DraggableField
                        key={field.api_name}
                        field={field}
                        onDragStart={(e) => onDragStart(e, field)}
                    />
                ))}
                {availableFields.length === 0 && (
                    <div className="text-center py-8 text-slate-400 text-sm italic">
                        All fields are used in the layout.
                    </div>
                )}

                <div className="mt-6 pt-6 border-t border-slate-100">
                    <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">Used Fields</div>
                    {metadata.fields.filter(f => usedFieldNames.has(f.api_name)).map(field => (
                        <div key={field.api_name} className="flex items-center gap-2 p-2 rounded border border-slate-200 bg-slate-50 text-slate-400 cursor-not-allowed">
                            <GripVertical size={14} className="text-slate-300" />
                            <div className="flex-1 truncate">
                                <span className="font-medium">{field.label}</span>
                                <span className="ml-2 text-xs text-slate-400 font-mono">{field.type}</span>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};
