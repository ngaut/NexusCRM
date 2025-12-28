import React from 'react';

interface DropZoneProps {
    onDrop: (e: React.DragEvent) => void;
    onDragOver: (e: React.DragEvent) => void;
    onDragLeave: (e: React.DragEvent) => void;
    isOver: boolean;
    className?: string;
    children?: React.ReactNode;
}

export const DropZone: React.FC<DropZoneProps> = ({
    onDrop,
    onDragOver,
    onDragLeave,
    isOver,
    className = '',
    children
}) => {
    return (
        <div
            onDrop={onDrop}
            onDragOver={onDragOver}
            onDragLeave={onDragLeave}
            className={`
                transition-all duration-200
                ${className}
                ${isOver ? 'bg-blue-50 border-blue-400 ring-2 ring-blue-200 ring-opacity-50' : ''}
            `}
        >
            {children}
        </div>
    );
};
